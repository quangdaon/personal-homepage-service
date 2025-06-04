package models

import "time"

type Shipment struct {
	ID             uint   `gorm:"primaryKey;autoIncrement"`
	Label          string `gorm:"not null"`
	TrackingNumber string `gorm:"size:100;not null;unique"`
	TrackingURL    string `gorm:"size:256"`
	ExpectedAt     *time.Time
	LastLocation   string `gorm:"size:100"`
	LastCheckedAt  *time.Time
	ThumbnailURL   string `gorm:"size:256"`

	// Foreign keys
	StatusID  *uint
	Status    *ShipmentStatus `gorm:"foreignKey:StatusID;references:ID"`
	CarrierID *uint
	Carrier   *ShipmentCarrier `gorm:"foreignKey:CarrierID;references:ID"`
}
