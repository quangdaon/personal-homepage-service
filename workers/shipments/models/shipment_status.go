package models

// ShipmentStatus represents shipment_statuses table
type ShipmentStatus struct {
	ID      uint   `gorm:"primaryKey;autoIncrement"`
	Key     string `gorm:"size:50;not null;unique"`
	Label   string `gorm:"size:50;not null;unique"`
	IsFinal bool   `gorm:"not null"`
}
