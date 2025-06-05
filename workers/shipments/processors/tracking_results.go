package processors

import "time"

type CarrierTrackingResults struct {
	TrackingNumber      string
	DeliveryWindowStart *time.Time
	DeliveryWindowEnd   *time.Time
	LastLocation        string
	LastCheckedAt       *time.Time
	Status              string
}
