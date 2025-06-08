package processors

import "personal-homepage-service/workers/shipments/models"

type CarrierTrackingProcessor interface {
	Process(shipment models.Shipment) (*CarrierTrackingResults, error)
}
