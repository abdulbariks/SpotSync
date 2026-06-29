package zone

import "time"

type ZoneType string

const (
	TypeGeneral    ZoneType = "general"
	TypeEVCharging ZoneType = "ev_charging"
	TypeCovered    ZoneType = "covered"
)

type ParkingZone struct {
	ID           uint     `gorm:"primaryKey"`
	Name         string   `gorm:"not null"`
	Type         ZoneType `gorm:"not null"`
	TotalCapacity int     `gorm:"not null"`
	PricePerHour  float64  `gorm:"not null"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
}