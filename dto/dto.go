package dto

import "time"

type RegisterRequest struct {
	Name     string `json:"name" validate:"required,min=2,max=100"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=6"`
	Role     string `json:"role" validate:"omitempty,oneof=driver admin"`
}

type RegisterResponse struct {
	ID        uint      `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Role      string    `json:"role"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type LoginResponse struct {
	Token string `json:"token"`
	User  struct {
		ID    uint   `json:"id"`
		Name  string `json:"name"`
		Email string `json:"email"`
		Role  string `json:"role"`
	} `json:"user"`
}

type CreateZoneRequest struct {
	Name         string  `json:"name" validate:"required,min=2,max=100"`
	Type         string  `json:"type" validate:"required,oneof=general ev_charging covered"`
	TotalCapacity int    `json:"total_capacity" validate:"required,gt=0"`
	PricePerHour float64 `json:"price_per_hour" validate:"required,gt=0"`
}

type ParkingZoneResponse struct {
	ID            uint    `json:"id"`
	Name          string  `json:"name"`
	Type          string  `json:"type"`
	TotalCapacity int     `json:"total_capacity"`
	AvailableSpots int   `json:"available_spots"`
	PricePerHour  float64 `json:"price_per_hour"`
	CreatedAt     string  `json:"created_at"`
}

type ReserveRequest struct {
	ZoneID       uint   `json:"zone_id" validate:"required"`
	LicensePlate string `json:"license_plate" validate:"required,min=2,max=15"`
}

type ReservationResponse struct {
	ID           uint   `json:"id"`
	UserID       uint   `json:"user_id"`
	ZoneID       uint   `json:"zone_id"`
	LicensePlate string `json:"license_plate"`
	Status       string `json:"status"`
	CreatedAt    string `json:"created_at"`
	UpdatedAt    string `json:"updated_at"`
}

type MyReservationResponse struct {
	ID          uint `json:"id"`
	LicensePlate string `json:"license_plate"`
	Status      string `json:"status"`
	Zone struct {
		ID   uint   `json:"id"`
		Name string `json:"name"`
		Type string `json:"type"`
	} `json:"zone"`
	CreatedAt string `json:"created_at"`
}