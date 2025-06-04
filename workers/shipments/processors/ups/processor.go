package ups

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"io"
	"net/http"
	"net/url"
	"personal-homepage-service/config"
	"personal-homepage-service/workers/shipments/processors"
	"strings"
	"time"
)

type TrackingProcessor struct {
	config *config.UpsApiConfig
}

func NewTrackingProcessor() *TrackingProcessor {
	cfg := config.LoadConfig()
	return &TrackingProcessor{cfg.UPSApi}
}

func (p *TrackingProcessor) Process(trackingNumber string) (*processors.CarrierTrackingResults, error) {
	details, err := p.getTrackingDetails(trackingNumber)
	if err != nil {
		return nil, err
	}

	shipment := details.Response.Shipments[0]
	pkg := shipment.Packages[0]
	return &processors.CarrierTrackingResults{
		TrackingNumber: trackingNumber,
		TrackingURL:    "https://www.ups.com/track?&loc=en_US&requester=ST/trackdetails&tracknum=" + trackingNumber,
		ExpectedAt:     time.Now(),
		LastLocation:   pkg.Activity[0].Location.Address.City,
		LastCheckedAt:  time.Now(),
		Status:         pkg.CurrentStatus.StatusCode,
	}, nil
}

func basicAuth(username, password string) string {
	auth := username + ":" + password
	return base64.StdEncoding.EncodeToString([]byte(auth))
}

func (p *TrackingProcessor) getAccessToken() (string, error) {
	u, err := url.Parse(p.config.BaseUri + "/security/v1/oauth/token")
	if err != nil {
		fmt.Println("URL parse error:", err)
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
	resp, err2 := client.Do(req)
	if err2 != nil {
		panic(err2)
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
