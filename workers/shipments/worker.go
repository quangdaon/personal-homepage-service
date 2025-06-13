package shipments

import (
	"go.uber.org/zap"
	"gorm.io/gorm"
	"log"
	"personal-homepage-service/workers/shipments/models"
	"personal-homepage-service/workers/shipments/processors"
	"personal-homepage-service/workers/shipments/processors/uds"
	"personal-homepage-service/workers/shipments/processors/unsupported"
	"personal-homepage-service/workers/shipments/processors/ups"
	"personal-homepage-service/workers/shipments/repositories"
	"sync"
	"time"
)

type Worker struct {
	logger     *zap.Logger
	repo       *repositories.Repository
	processors map[string]processors.CarrierTrackingProcessor
	mu         sync.Mutex
	busy       bool
}

func NewWorker(logger *zap.Logger, db *gorm.DB) *Worker {
	repo := repositories.NewRepository(db)
	return &Worker{
		logger:     logger,
		repo:       repo,
		processors: make(map[string]processors.CarrierTrackingProcessor),
	}
}

func (w *Worker) Schedule() string {
	return "*/5 * * * *"
}

func (w *Worker) Ready(time.Time) bool {
	return !w.busy
}

func (w *Worker) Execute() {
	w.busy = true
	defer func() {
		w.busy = false
	}()

	shipments, err := w.repo.GetOpenShipments()
	if err != nil {
		log.Fatal(err)
		return
	}

	if len(shipments) == 0 {
		w.logger.Info("No active shipments found. Shipment work skipped 😴")
		return
	}

	shipmentsToProcess := w.getShipmentsToProcess(shipments)

	if len(shipmentsToProcess) == 0 {
		w.logger.Info("No shipments are ready to be processed. Shipment work skipped 😴")
		return
	}

	w.logger.Info("Starting shipment processing.")

	var wg sync.WaitGroup
	for _, shipment := range shipmentsToProcess {
		wg.Add(1)
		go func(sh models.Shipment) {
			defer wg.Done()
			w.processShipment(sh)
		}(shipment)
	}

	wg.Wait()
	w.logger.Info("Shipment work completed 😴")
}

func (w *Worker) getShipmentsToProcess(ss []models.Shipment) (ret []models.Shipment) {
	for _, s := range ss {
		if w.shouldCheck(s) {
			ret = append(ret, s)
		}
	}
	return
}

func (w *Worker) shouldCheck(shipment models.Shipment) bool {
	const (
		frequencyThreshold = 6 * time.Hour
		soonThreshold      = 2 * time.Hour
		recheckDelay       = 15 * time.Minute
	)

	if shipment.Status.IsFinal {
		return false
	}

	if shipment.Status.Key == "unchecked" || shipment.LastCheckedAt == nil {
		return true
	}

	now := time.Now()
	timeSinceLastCheck := now.Sub(*shipment.LastCheckedAt)

	if timeSinceLastCheck > frequencyThreshold {
		return true
	}

	if shipment.DeliveryWindowEnd == nil {
		return false
	}

	timeUntilExpected := shipment.DeliveryWindowEnd.Sub(now)

	return timeUntilExpected < soonThreshold && timeSinceLastCheck > recheckDelay
}
func (w *Worker) processShipment(sh models.Shipment) {
	processor := w.getProcessor(sh.Carrier.Key)

	result, err := processor.Process(sh)
	if err != nil {
		w.logger.Error("Failed to process shipment",
			zap.String("tracking_number", sh.TrackingNumber),
			zap.Error(err),
		)
		return
	}

	status, err := w.repo.GetStatus(result.Status)
	if err != nil {
		w.logger.Error("Failed to get shipment status",
			zap.String("tracking_number", result.TrackingNumber),
			zap.String("status_key", result.Status),
			zap.Error(err),
		)
		return
	}

	w.updateShipmentFromResult(&sh, result, &status)

	if err := w.repo.SaveShipment(&sh); err != nil {
		w.logger.Error("Failed to save shipment",
			zap.String("tracking_number", sh.TrackingNumber),
			zap.Error(err),
		)
		return
	}

	w.logger.Info("Shipment successfully processed",
		zap.String("tracking_number", sh.TrackingNumber),
	)
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

func (w *Worker) getProcessor(carrier string) processors.CarrierTrackingProcessor {
	w.mu.Lock()
	defer w.mu.Unlock()

	if processor, exists := w.processors[carrier]; exists {
		return processor
	}

	var processor processors.CarrierTrackingProcessor

	switch carrier {
	case "ups":
		processor = ups.NewTrackingProcessor(w.logger)
	case "uds":
		processor = uds.NewTrackingProcessor(w.logger)
	default:
		processor = unsupported.NewTrackingProcessor(w.logger)
	}

	w.processors[carrier] = processor
	return processor
}
