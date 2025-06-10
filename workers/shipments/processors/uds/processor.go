package uds

import (
	"fmt"
	"github.com/gocolly/colly/v2"
	"go.uber.org/zap"
	"personal-homepage-service/workers/shipments/models"
	"personal-homepage-service/workers/shipments/processors"
	"regexp"
	"strings"
	"sync"
	"time"
)

var statusMap = map[string]string{
	"Shipment Notification": "pending",
	"Received":              "in_transit",
	"Out for Delivery":      "out_for_delivery",
	"Delivered":             "delivered",
}

type TrackingProcessor struct {
	logger *zap.Logger
}

func NewTrackingProcessor(logger *zap.Logger) *TrackingProcessor {
	return &TrackingProcessor{logger}
}

func (p *TrackingProcessor) Process(shipment models.Shipment) (*processors.CarrierTrackingResults, error) {
	trackingNumber := shipment.TrackingNumber
	url := shipment.TrackingURL
	now := time.Now()
	title := ""
	lastLoc := shipment.LastLocation
	expected := shipment.DeliveryWindowEnd

	c := colly.NewCollector()

	wg := &sync.WaitGroup{}
	wg.Add(1)

	c.OnHTML(".multi-step.numbered li.current", func(e *colly.HTMLElement) {
		title = e.ChildText(".wrap > p.title")
	})

	c.OnHTML(".multi-step.numbered + table", func(e *colly.HTMLElement) {
		// Check if the table contains the header cell with the expected text
		header := e.DOM.Find("td.dkBlue").First()
		if !strings.Contains(header.Text(), "Expected Delivery Day:") {
			return
		}

		e.ForEach("tr", func(_ int, row *colly.HTMLElement) {
			cells := row.DOM.Find("td")
			if cells.Length() == 2 {
				dateStr := strings.TrimSpace(cells.Eq(0).Text()) // e.g. "Mon Jun 9"
				timeStr := strings.TrimSpace(cells.Eq(1).Text()) // e.g. "by\n8:00 PM"
				timeStr = strings.ReplaceAll(timeStr, "by", " ")

				// Combine with current year
				year := time.Now().Year()
				loc := time.Now().Location()
				combined := fmt.Sprintf("%s %d %s", dateStr, year, timeStr) // e.g. "Mon Jun 9 2025 8:00 PM"

				// Parse the combined string
				parsedTime, err := time.ParseInLocation("Mon Jan 2 2006 3:04 PM", combined, loc)
				if err != nil {
					fmt.Println("Failed to parse date:", err)
					return
				}

				expected = &parsedTime
			}
		})
	})

	c.OnHTML("td", func(e *colly.HTMLElement) {
		text := strings.ReplaceAll(e.Text, "\u00a0", " ") // normalize &nbsp;
		text = strings.TrimSpace(text)

		// Check for the known delivery phrase
		if strings.Contains(text, "The package has departed") &&
			strings.Contains(text, "sort facility and is out for delivery") {

			// Example: "The package has departed CITY, ST sort facility and is out for delivery."
			// Extract substring between "departed " and " sort facility"
			prefix := "departed "
			suffix := " sort facility"

			start := strings.Index(text, prefix)
			end := strings.Index(text, suffix)

			if start != -1 && end != -1 && end > start+len(prefix) {
				lastLoc = strings.TrimSpace(text[start+len(prefix) : end])
			}
		}

		if strings.Contains(text, "The package is delivered.") {
			// Split by lines
			re := regexp.MustCompile(`(\d{4}-\d{2}-\d{2})\s*-\s*(\d{1,2}:\d{2}:\d{2}\s*[AP]M)`)
			matches := re.FindStringSubmatch(text)
			if len(matches) == 3 {
				datePart := matches[1]
				timePart := matches[2]
				combined := fmt.Sprintf("%s %s", datePart, timePart) // "2025-06-09 12:13:03 PM"

				// Parse into time.Time
				loc := time.Now().Location()
				parsedTime, err := time.ParseInLocation("2006-01-02 3:04:05 PM", combined, loc)
				if err != nil {
					p.logger.Error("Failed to parse delivery time:", zap.String("datetime", combined), zap.Error(err))
					return
				}

				expected = &parsedTime
			}
		}
	})

	c.OnScraped(func(r *colly.Response) {
		wg.Done()
	})

	if err := c.Visit(url); err != nil {
		return nil, err
	}

	wg.Wait()

	return &processors.CarrierTrackingResults{
		TrackingNumber:    trackingNumber,
		DeliveryWindowEnd: expected,
		LastLocation:      lastLoc,
		LastCheckedAt:     &now,
		Status:            getStatusKey(title),
	}, nil
}

func getStatusKey(title string) string {
	status, ok := statusMap[title]
	if !ok {
		return "unknown"
	}
	return status
}
