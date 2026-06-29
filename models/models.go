package models

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type UserRole string

const (
	RoleDriver UserRole = "driver"
	RoleAdmin  UserRole = "admin"
)

type User struct {
	ID        uint      `gorm:"primaryKey"`
	Name      string    `gorm:"not null"`
	Email     string    `gorm:"unique;not null"`
	Password  string    `gorm:"not null"`
	Role      UserRole  `gorm:"default:driver;not null"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

type ParkingZoneType string

const (
	TypeGeneral    ParkingZoneType = "general"
	TypeEVCharging ParkingZoneType = "ev_charging"
	TypeCovered    ParkingZoneType = "covered"
)

type ParkingZone struct {
	ID           uint            `gorm:"primaryKey"`
	Name         string          `gorm:"not null"`
	Type         ParkingZoneType `gorm:"not null"`
	TotalCapacity int            `gorm:"not null"`
	PricePerHour  float64        `gorm:"not null"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

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
	User        User       `gorm:"foreignKey:UserID"`
	Zone        ParkingZone `gorm:"foreignKey:ZoneID"`
}

type JWTClaims struct {
	UserID uint    `json:"user_id"`
	Role   UserRole `json:"role"`
	jwt.RegisteredClaims
}