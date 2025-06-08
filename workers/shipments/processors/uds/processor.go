package uds

import (
	"github.com/gocolly/colly/v2"
	"go.uber.org/zap"
	"personal-homepage-service/workers/shipments/models"
	"personal-homepage-service/workers/shipments/processors"
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

	c := colly.NewCollector()

	wg := &sync.WaitGroup{}
	wg.Add(1)

	c.OnHTML(".multi-step.numbered li.current", func(e *colly.HTMLElement) {
		defer wg.Done()
		title = e.ChildText(".wrap > p.title")
	})

	if err := c.Visit(url); err != nil {
		return nil, err
	}

	wg.Wait()

	return &processors.CarrierTrackingResults{
		TrackingNumber: trackingNumber,
		LastCheckedAt:  &now,
		Status:         getStatusKey(title),
	}, nil
}

func getStatusKey(title string) string {
	status, ok := statusMap[title]
	if !ok {
		return "unknown"
	}
	return status
}
