package models

type ShipmentCarrier struct {
	ID    uint   `gorm:"primaryKey;autoIncrement"`
	Label string `gorm:"size:50;not null;unique"`
	Icon  string `gorm:"size:256"`
}
