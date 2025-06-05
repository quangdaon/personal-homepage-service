package shipments

import (
	"fmt"
	"gorm.io/gorm"
	"log"
	"personal-homepage-service/workers/shipments/models"
	"personal-homepage-service/workers/shipments/processors"
	"personal-homepage-service/workers/shipments/processors/ups"
	"personal-homepage-service/workers/shipments/repositories"
	"sync"
	"time"
)

type Worker struct {
	repo       *repositories.Repository
	processors map[string]processors.CarrierTrackingProcessor
	mu         sync.Mutex
}

func NewWorker(db *gorm.DB) *Worker {
	repo := repositories.NewRepository(db)
	return &Worker{
		repo:       repo,
		processors: make(map[string]processors.CarrierTrackingProcessor),
	}
}

func (w *Worker) Execute() {
	shipments, err := w.repo.GetAllShipments()
	if err != nil {
		log.Fatal(err)
	}

	if len(shipments) == 0 {
		return
	}

	var wg sync.WaitGroup
	for _, shipment := range shipments {
		if !w.shouldCheck(shipment) {
			continue
		}

		wg.Add(1)
		go func(sh models.Shipment) {
			defer wg.Done()
			w.processShipment(sh)
		}(shipment)
	}

	wg.Wait()
}

func (w *Worker) shouldCheck(shipment models.Shipment) bool {
	const (
		day           = 24 * time.Hour
		soonThreshold = 2 * time.Hour
		recheckDelay  = 15 * time.Minute
	)

	if shipment.Status.IsFinal {
		return false
	}

	if shipment.Status.Key == "unchecked" || shipment.LastCheckedAt == nil {
		return true
	}

	now := time.Now()
	timeSinceLastCheck := now.Sub(*shipment.LastCheckedAt)

	if timeSinceLastCheck > day {
		return true
	}

	if shipment.DeliveryWindowEnd == nil {
		return false
	}

	timeUntilExpected := shipment.DeliveryWindowEnd.Sub(now)

	return timeUntilExpected < soonThreshold && timeSinceLastCheck > recheckDelay
}

func (w *Worker) processShipment(sh models.Shipment) {
	processor, err := w.getProcessor(sh.Carrier.Key)
	if err != nil {
		log.Printf("Failed to get processor for %s: %v", sh.TrackingNumber, err)
		return
	}

	result, err := processor.Process(sh.TrackingNumber)
	if err != nil {
		log.Printf("Failed to process %s: %v", sh.TrackingNumber, err)
		return
	}

	status, err := w.repo.GetStatus(result.Status)
	if err != nil {
		log.Printf("Failed to get status for %s: %v", result.TrackingNumber, err)
		return
	}

	w.updateShipmentFromResult(&sh, result, &status)

	if err := w.repo.SaveShipment(&sh); err != nil {
		log.Printf("Failed to save shipment %s: %v", sh.TrackingNumber, err)
		return
	}

	log.Printf("Shipment %s was successfully processed", sh.TrackingNumber)
}

func (w *Worker) updateShipmentFromResult(sh *models.Shipment, result *processors.CarrierTrackingResults, status *models.ShipmentStatus) {
	sh.Status = status
	sh.LastLocation = result.LastLocation

	if result.LastCheckedAt != nil {
		utc := result.LastCheckedAt.UTC()
		sh.LastCheckedAt = &utc
	}

	if result.DeliveryWindowStart != nil {
		utc := result.DeliveryWindowStart.UTC()
		sh.DeliveryWindowStart = &utc
	}

	if result.DeliveryWindowEnd != nil {
		utc := result.DeliveryWindowEnd.UTC()
		sh.DeliveryWindowEnd = &utc
	}
}

func (w *Worker) getProcessor(carrier string) (processors.CarrierTrackingProcessor, error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	if processor, exists := w.processors[carrier]; exists {
		return processor, nil
	}

	var processor processors.CarrierTrackingProcessor

	switch carrier {
	case "ups":
		processor = ups.NewTrackingProcessor()
	default:
		return nil, fmt.Errorf("unsupported carrier: %s", carrier)
	}

	w.processors[carrier] = processor
	return processor, nil
}
