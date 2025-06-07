package unsupported

import (
	"go.uber.org/zap"
	"personal-homepage-service/workers/shipments/processors"
	"time"
)

type TrackingProcessor struct {
	logger *zap.Logger
}

func NewTrackingProcessor(logger *zap.Logger) *TrackingProcessor {
	return &TrackingProcessor{logger}
}

func (p *TrackingProcessor) Process(trackingNumber string) (*processors.CarrierTrackingResults, error) {
	now := time.Now()

	return &processors.CarrierTrackingResults{
		TrackingNumber: trackingNumber,
		LastCheckedAt:  &now,
		Status:         "unsupported",
	}, nil
}
