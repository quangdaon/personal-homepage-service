package repositories

import (
	"gorm.io/gorm"
	"personal-homepage-service/workers/shipments/models"
)

// ShipmentRepository is the repo for accessing shipments and related data
type ShipmentRepository struct {
	db *gorm.DB
}

// NewShipmentRepository creates a new repositories with DB dependency
func NewShipmentRepository(db *gorm.DB) *ShipmentRepository {
	return &ShipmentRepository{db: db}
}

// GetAllShipments returns all shipments with related status and carrier
func (r *ShipmentRepository) GetAllShipments() ([]models.Shipment, error) {
	var shipments []models.Shipment
	err := r.db.Preload("Status").Preload("Carrier").Find(&shipments).Error
	return shipments, err
}

// GetAllCarriers returns all shipment carriers
func (r *ShipmentRepository) GetAllCarriers() ([]models.ShipmentCarrier, error) {
	var carriers []models.ShipmentCarrier
	err := r.db.Find(&carriers).Error
	return carriers, err
}

// SaveShipment creates or updates a shipment
func (r *ShipmentRepository) SaveShipment(shipment *models.Shipment) error {
	return r.db.Save(shipment).Error
}
