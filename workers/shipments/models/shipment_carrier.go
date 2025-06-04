package models

type ShipmentCarrier struct {
	ID    uint   `gorm:"primaryKey;autoIncrement"`
	Key   string `gorm:"size:50;not null;unique"`
	Label string `gorm:"size:50;not null;unique"`
	Icon  string `gorm:"size:256"`
}
