package repositories

import (
	"gorm.io/gorm"
	"personal-homepage-service/workers/shipments/models"
)

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) GetAllShipments() ([]models.Shipment, error) {
	var shipments []models.Shipment
	err := r.db.Preload("Status").Preload("Carrier").Find(&shipments).Error
	return shipments, err
}

func (r *Repository) GetOpenShipments() ([]models.Shipment, error) {
	var shipments []models.Shipment
	err := r.db.Joins("Status").
		Preload("Status").
		Preload("Carrier").
		Where("\"Status\".is_final = ?", false).
		Find(&shipments).Error
	return shipments, err
}

func (r *Repository) GetStatus(key string) (models.ShipmentStatus, error) {
	var status models.ShipmentStatus
	err := r.db.Where("key = ?", key).First(&status).Error
	return status, err
}

func (r *Repository) SaveShipment(shipment *models.Shipment) error {
	return r.db.Save(shipment).Error
}
