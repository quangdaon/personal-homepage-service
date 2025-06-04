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
)

type ShipmentWorker struct {
	repo       *repositories.ShipmentRepository
	processors map[string]processors.CarrierTrackingProcessor
	mu         sync.Mutex
}

func NewShipmentWorker(db *gorm.DB) *ShipmentWorker {
	repo := repositories.NewShipmentRepository(db)
	return &ShipmentWorker{
		repo:       repo,
		processors: make(map[string]processors.CarrierTrackingProcessor),
	}
}

func (w *ShipmentWorker) Execute() {
	shipments, shipmentsErr := w.repo.GetAllShipments()

	if shipmentsErr != nil {
		log.Fatal(shipmentsErr)
	}

	if len(shipments) == 0 {
		return
	}

	var wg sync.WaitGroup
	for _, shipment := range shipments {
		wg.Add(1)

		go func(sh models.Shipment) {
			defer wg.Done()

			processor, getProcessorErr := w.getProcessor(sh.Carrier.Key)
			if getProcessorErr != nil {
				log.Println(getProcessorErr.Error())
				return
			}

			result, processorErr := processor.Process(sh.TrackingNumber)
			if processorErr != nil {
				log.Printf("Failed to process %s: %v", sh.TrackingNumber, processorErr)
				return
			}

			fmt.Println(result.LastLocation)
		}(shipment)
	}

	wg.Wait()
}

func (w *ShipmentWorker) getProcessor(carrier string) (processors.CarrierTrackingProcessor, error) {
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
