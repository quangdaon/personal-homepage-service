package ups

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"io"
	"net/http"
	"net/url"
	"personal-homepage-service/config"
	"personal-homepage-service/workers/shipments/processors"
	"strings"
	"time"
)

var upsCodeMap = map[string]string{
	"003": "pending",            // Shipment Ready for UPS
	"005": "in_transit",         // In Transit
	"006": "out_for_delivery",   // Out for Delivery Today
	"007": "cancelled",          // Shipment Canceled
	"011": "delivered",          // Delivered
	"012": "in_transit",         // Clearance in Progress
	"013": "in_transit",         // Update
	"014": "in_transit",         // Cleared Customs
	"016": "exception",          // Held in Warehouse
	"017": "delivered",          // Held for Customer Pickup
	"018": "exception",          // Hold for Pickup Requested
	"019": "delayed",            // Delivery Rescheduled
	"021": "out_for_delivery",   // Out for Delivery Today
	"022": "attempted_delivery", // Delivery Attempted
	"023": "attempted_delivery", // Delivery Attempted
	"024": "attempted_delivery", // IsFinal Delivery Attempt Made
	"025": "in_transit",         // In Transit
	"026": "delivered",          // Delivered by Local Post Office
	"027": "in_transit",         // Address Change Requested
	"028": "in_transit",         // Delivery Address Changed
	"029": "exception",          // Address Information Required
	"030": "delayed",            // Local Post Office Delay
	"032": "delayed",            // Weather May Cause Delay
	"033": "returned",           // Return Requested
	"035": "returned",           // Returning to Sender
	"038": "accepted",           // Picked Up
	"040": "delivered",          // Delivered to UPS Access Point™
	"042": "in_transit",         // Service Upgraded
	"044": "in_transit",         // On Its Way to UPS
	"045": "in_transit",         // Order Processed: On its Way to UPS
	"046": "delayed",            // Delay
	"047": "in_transit",         // In Transit
	"048": "delayed",            // Delay
	"049": "exception",          // Delay: Attention Needed
	"050": "exception",          // Address Information Required
	"051": "delayed",            // Delay: Emergency Situation or Severe Weather
	"052": "delayed",            // Severe Weather Delay
	"053": "delayed",            // Severe Weather Delay
	"054": "delayed",            // Delivery Change Requested
	"055": "delayed",            // Rescheduled Delivery
	"057": "in_transit",         // On Its Way to a Local UPS Access Point™
	"058": "exception",          // Clearance Information Required
	"065": "attempted_delivery", // Pickup Attempted
	"070": "in_transit",         // On Its Way to a Local UPS Access Point™
	"071": "out_for_delivery",   // Preparing for Delivery Today
	"072": "out_for_delivery",   // Loaded on Delivery Vehicle
	"077": "delivered",          // Scheduled for Pickup Today
}

type TrackingProcessor struct {
	config *config.UpsApiConfig
	logger *zap.Logger
}

func NewTrackingProcessor(logger *zap.Logger) *TrackingProcessor {
	cfg := config.LoadConfig()
	return &TrackingProcessor{cfg.UPSApi, logger}
}

func (p *TrackingProcessor) Process(trackingNumber string) (*processors.CarrierTrackingResults, error) {
	details, err := p.getTrackingDetails(trackingNumber)
	if err != nil {
		return nil, err
	}

	shipment := details.Response.Shipments[0]
	pkg := shipment.Packages[0]
	now := time.Now()
	delStart, delEnd, delErr := getExpectedDeliveryWindow(pkg)
	if delErr != nil {
		p.logger.Error("Error parsing datetime:" + delErr.Error())
	}

	return &processors.CarrierTrackingResults{
		TrackingNumber:      trackingNumber,
		DeliveryWindowStart: delStart,
		DeliveryWindowEnd:   delEnd,
		LastLocation:        getLastLocation(pkg.Activity),
		LastCheckedAt:       &now,
		Status:              getStatusKey(pkg.CurrentStatus.Code),
	}, nil
}

func getExpectedDeliveryWindow(p Package) (*time.Time, *time.Time, error) {
	date := p.DeliveryDate[0].Date

	end, endErr := parseDatetime(date, p.DeliveryTime.EndTime)

	if endErr != nil || p.DeliveryTime.StartTime == "" {
		return nil, end, endErr
	}

	start, startErr := parseDatetime(date, p.DeliveryTime.StartTime)

	if startErr != nil {
		return nil, end, startErr
	}

	return start, end, nil
}

func parseDatetime(date string, timeStr string) (*time.Time, error) {
	datetimeStr := date + timeStr
	loc := time.Now().Location()

	const layout = "20060102150405"

	parsedTime, err := time.ParseInLocation(layout, datetimeStr, loc)
	if err != nil {
		return nil, err
	}

	return &parsedTime, nil
}

func getLastLocation(activity []Activity) string {
	lastLocation := activity[0].Location
	region := lastLocation.Address.CountryCode

	if lastLocation.Address.CountryCode == "US" {
		region = lastLocation.Address.State
	}

	return lastLocation.Address.City + ", " + region
}

func getStatusKey(code string) string {
	status, ok := upsCodeMap[code]
	if !ok {
		return "unknown"
	}
	return status
}

func basicAuth(username, password string) string {
	auth := username + ":" + password
	return base64.StdEncoding.EncodeToString([]byte(auth))
}

func (p *TrackingProcessor) getAccessToken() (string, error) {
	u, err := url.Parse(p.config.BaseUri + "/security/v1/oauth/token")
	if err != nil {
		return "", err
	}

	data := url.Values{}
	data.Set("grant_type", "client_credentials")

	req, err := http.NewRequest("POST", u.String(), strings.NewReader(data.Encode()))

	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Authorization", "Basic "+basicAuth(p.config.ClientId, p.config.ClientSecret))

	client := &http.Client{Timeout: 20 * time.Second}
	resp, err2 := client.Do(req)
	if err2 != nil {
		panic(err2)
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var authResponse OAuthResponse
	if err := json.NewDecoder(resp.Body).Decode(&authResponse); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	return authResponse.AccessToken, nil
}

func (p *TrackingProcessor) getTrackingDetails(trackingNumber string) (*ApiResponse, error) {
	endpoint := "/api/track/v1/details/" + trackingNumber

	u, err := url.Parse(p.config.BaseUri + endpoint)
	if err != nil {
		fmt.Println("URL parse error:", err)
		return nil, err
	}

	q := u.Query()
	q.Set("locale", "en_US")
	q.Set("returnSignature", "false")
	q.Set("returnMilestones", "false")
	q.Set("returnPOD", "false")
	u.RawQuery = q.Encode()

	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return nil, err
	}

	accessToken, tokenErr := p.getAccessToken()

	if tokenErr != nil {
		return nil, tokenErr
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("transId", uuid.New().String())
	req.Header.Set("transactionSrc", "personal_homepage")

	client := &http.Client{Timeout: 20 * time.Second}
	resp, clientErr := client.Do(req)
	if clientErr != nil {
		panic(clientErr)
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var apiResponse ApiResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResponse); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &apiResponse, nil

}
