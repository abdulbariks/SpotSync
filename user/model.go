package user

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
	ID        uint     `gorm:"primaryKey"`
	Name      string   `gorm:"not null"`
	Email     string   `gorm:"unique;not null"`
	Password  string   `gorm:"not null"`
	Role      UserRole `gorm:"default:driver;not null"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

type JWTClaims struct {
	UserID uint    `json:"user_id"`
	Role   UserRole `json:"role"`
	jwt.RegisteredClaims
}