package shipments

import (
	"fmt"
	"gorm.io/gorm"
	"log"
	"personal-homepage-service/workers/shipments/repository"
)

type ShipmentWorker struct {
	repo repository.ShipmentRepository
}

func NewShipmentWorker(db *gorm.DB) *ShipmentWorker {
	repo := repository.NewShipmentRepository(db)
	return &ShipmentWorker{*repo}
}

func (w *ShipmentWorker) Init() {

	// Get all shipments
	shipments, shipmentsErr := w.repo.GetAllShipments()

	if shipmentsErr != nil {
		log.Fatal(shipmentsErr)
	}

	for _, s := range shipments {
		fmt.Printf("Shipment #%d\n", s.ID)
		fmt.Printf("  Label: %s\n", s.Label)
		fmt.Printf("  Tracking #: %s\n", s.TrackingNumber)
		if s.TrackingURL != "" {
			fmt.Printf("  Tracking URL: %s\n", s.TrackingURL)
		}
		if s.LastLocation != "" {
			fmt.Printf("  Last Location: %s\n", s.LastLocation)
		}
		if s.ExpectedAt != nil {
			fmt.Printf("  Expected: %s\n", s.ExpectedAt.Format("2006-01-02 15:04:05"))
		}
		if s.LastCheckedAt != nil {
			fmt.Printf("  Last Checked: %s\n", s.LastCheckedAt.Format("2006-01-02 15:04:05"))
		}
		if s.ThumbnailURL != "" {
			fmt.Printf("  Thumbnail: %s\n", s.ThumbnailURL)
		}
		if s.Status != nil {
			fmt.Printf("  Status: %s (%s)\n", s.Status.Label, s.Status.Key)
		}
		if s.Carrier != nil {
			fmt.Printf("  Carrier: %s\n", s.Carrier.Label)
		}
		fmt.Println()
	}
}
