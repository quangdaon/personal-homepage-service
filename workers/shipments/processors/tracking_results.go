package processors

import "time"

type CarrierTrackingResults struct {
	TrackingNumber string
	TrackingURL    string
	ExpectedAt     time.Time
	LastLocation   string
	LastCheckedAt  time.Time
	Status         string
}
