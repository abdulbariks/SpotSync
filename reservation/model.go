package reservation

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type ReservationStatus string

const (
	StatusActive    ReservationStatus = "active"
	StatusCompleted ReservationStatus = "completed"
	StatusCancelled ReservationStatus = "cancelled"
)

type Reservation struct {
	ID          uint              `gorm:"primaryKey"`
	UserID      uint              `gorm:"not null"`
	ZoneID      uint              `gorm:"not null"`
	LicensePlate string           `gorm:"not null;size:15"`
	Status      ReservationStatus `gorm:"default:active;not null"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
	Zone        ParkingZone `gorm:"foreignKey:ZoneID"`
}

type ParkingZone struct {
	ID           uint   `gorm:"primaryKey"`
	Name         string `gorm:"not null"`
	Type         string `gorm:"not null"`
	TotalCapacity int   `gorm:"not null"`
}

type JWTClaims struct {
	UserID uint   `json:"user_id"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}