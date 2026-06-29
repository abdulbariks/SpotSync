package reservation

import (
	"errors"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
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
	ID           uint    `gorm:"primaryKey"`
	Name         string  `gorm:"not null"`
	Type         string  `gorm:"not null"`
	TotalCapacity int    `gorm:"not null"`
}

type JWTClaims struct {
	UserID uint   `json:"user_id"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

type CreateRequest struct {
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

var (
	ErrReservationNotFound = errors.New("reservation not found")
	ErrZoneFull           = errors.New("parking zone is at full capacity")
	ErrLicensePlateExists = errors.New("license plate already has active reservation")
	ErrZoneNotFound        = errors.New("parking zone not found")
)

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Create(reservation *Reservation) error {
	return r.db.Create(reservation).Error
}

func (r *Repository) FindByID(id uint) (*Reservation, error) {
	var reservation Reservation
	if err := r.db.Preload("User").Preload("Zone").First(&reservation, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrReservationNotFound
		}
		return nil, err
	}
	return &reservation, nil
}

func (r *Repository) FindByUserID(userID uint) ([]Reservation, error) {
	var reservations []Reservation
	return reservations, r.db.Where("user_id = ?", userID).Preload("Zone").Order("created_at DESC").Find(&reservations).Error
}

func (r *Repository) FindAll() ([]Reservation, error) {
	var reservations []Reservation
	return reservations, r.db.Preload("User").Preload("Zone").Order("created_at DESC").Find(&reservations).Error
}

func (r *Repository) Update(reservation *Reservation) error {
	return r.db.Save(reservation).Error
}

func (r *Repository) FindActiveByLicensePlate(licensePlate string) (*Reservation, error) {
	var reservation Reservation
	if err := r.db.Where("license_plate = ? AND status = ?", licensePlate, StatusActive).First(&reservation).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &reservation, nil
}

func (r *Repository) ReserveSpot(zoneID uint, reservation *Reservation) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		var zone ParkingZone
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&zone, zoneID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrZoneNotFound
			}
			return err
		}

		var activeCount int64
		if err := tx.Model(&Reservation{}).Where("zone_id = ? AND status = ?", zoneID, StatusActive).Count(&activeCount).Error; err != nil {
			return err
		}

		if activeCount >= int64(zone.TotalCapacity) {
			return ErrZoneFull
		}

		return tx.Create(reservation).Error
	})
}

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Create(req *CreateRequest, userID uint) (*ReservationResponse, error) {
	existingReservation, _ := s.repo.FindActiveByLicensePlate(req.LicensePlate)
	if existingReservation != nil {
		return nil, ErrLicensePlateExists
	}

	reservation := &Reservation{
		UserID:       userID,
		ZoneID:       req.ZoneID,
		LicensePlate: req.LicensePlate,
		Status:       StatusActive,
	}

	if err := s.repo.ReserveSpot(req.ZoneID, reservation); err != nil {
		return nil, err
	}

	return &ReservationResponse{
		ID:           reservation.ID,
		UserID:       reservation.UserID,
		ZoneID:       reservation.ZoneID,
		LicensePlate: reservation.LicensePlate,
		Status:       string(reservation.Status),
		CreatedAt:    reservation.CreatedAt.Format(time.RFC3339),
		UpdatedAt:    reservation.UpdatedAt.Format(time.RFC3339),
	}, nil
}

func (s *Service) GetMyReservations(userID uint) ([]MyReservationResponse, error) {
	reservations, err := s.repo.FindByUserID(userID)
	if err != nil {
		return nil, err
	}

	var responses []MyReservationResponse
	for _, r := range reservations {
		resp := MyReservationResponse{
			ID:           r.ID,
			LicensePlate: r.LicensePlate,
			Status:       string(r.Status),
			CreatedAt:    r.CreatedAt.Format(time.RFC3339),
		}
		resp.Zone.ID = r.Zone.ID
		resp.Zone.Name = r.Zone.Name
		resp.Zone.Type = r.Zone.Type
		responses = append(responses, resp)
	}

	return responses, nil
}

func (s *Service) Cancel(id uint, userID uint) error {
	reservation, err := s.repo.FindByID(id)
	if err != nil {
		return err
	}

	if reservation.UserID != userID {
		return errors.New("unauthorized: cannot cancel another user's reservation")
	}

	return s.repo.Update(&Reservation{
		ID:     reservation.ID,
		Status: StatusCancelled,
	})
}

func (s *Service) GetAll() ([]ReservationResponse, error) {
	reservations, err := s.repo.FindAll()
	if err != nil {
		return nil, err
	}

	var responses []ReservationResponse
	for _, r := range reservations {
		responses = append(responses, ReservationResponse{
			ID:           r.ID,
			UserID:       r.UserID,
			ZoneID:       r.ZoneID,
			LicensePlate: r.LicensePlate,
			Status:       string(r.Status),
			CreatedAt:    r.CreatedAt.Format(time.RFC3339),
			UpdatedAt:    r.UpdatedAt.Format(time.RFC3339),
		})
	}

	return responses, nil
}

type HTTPHandler struct {
	service *Service
}

func NewHTTPHandler(svc *Service) *HTTPHandler {
	return &HTTPHandler{service: svc}
}

func (h *HTTPHandler) CreateReservation(c echo.Context) error {
	u, ok := c.Get("user").(*JWTClaims)
	if !ok {
		return c.JSON(http.StatusUnauthorized, echo.Map{"success": false, "message": "Unauthorized"})
	}

	var req CreateRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"success": false, "message": "Invalid request body", "errors": err.Error()})
	}

	if err := c.Validate(&req); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"success": false, "message": "Validation failed", "errors": err.Error()})
	}

	resp, err := h.service.Create(&req, u.UserID)
	if err != nil {
		if err == ErrZoneFull {
			return c.JSON(http.StatusConflict, echo.Map{"success": false, "message": "Parking zone is at full capacity"})
		}
		if err == ErrZoneNotFound {
			return c.JSON(http.StatusNotFound, echo.Map{"success": false, "message": "Parking zone not found"})
		}
		if err == ErrLicensePlateExists {
			return c.JSON(http.StatusConflict, echo.Map{"success": false, "message": "License plate already has active reservation"})
		}
		return c.JSON(http.StatusInternalServerError, echo.Map{"success": false, "message": "Failed to create reservation"})
	}

	return c.JSON(http.StatusCreated, echo.Map{
		"success": true,
		"message": "Reservation confirmed successfully",
		"data":    resp,
	})
}

func (h *HTTPHandler) GetMyReservations(c echo.Context) error {
	u, ok := c.Get("user").(*JWTClaims)
	if !ok {
		return c.JSON(http.StatusUnauthorized, echo.Map{"success": false, "message": "Unauthorized"})
	}

	reservations, err := h.service.GetMyReservations(u.UserID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"success": false, "message": "Failed to fetch reservations"})
	}

	return c.JSON(http.StatusOK, echo.Map{
		"success": true,
		"message": "My reservations retrieved successfully",
		"data":    reservations,
	})
}

func (h *HTTPHandler) CancelReservation(c echo.Context) error {
	u, ok := c.Get("user").(*JWTClaims)
	if !ok {
		return c.JSON(http.StatusUnauthorized, echo.Map{"success": false, "message": "Unauthorized"})
	}

	id := parseUint(c.Param("id"))

	if err := h.service.Cancel(id, u.UserID); err != nil {
		if err.Error() == "unauthorized: cannot cancel another user's reservation" {
			return c.JSON(http.StatusForbidden, echo.Map{"success": false, "message": "Cannot cancel another user's reservation"})
		}
		if errors.Is(err, ErrReservationNotFound) {
			return c.JSON(http.StatusNotFound, echo.Map{"success": false, "message": "Reservation not found"})
		}
		return c.JSON(http.StatusInternalServerError, echo.Map{"success": false, "message": "Failed to cancel reservation"})
	}

	return c.JSON(http.StatusOK, echo.Map{
		"success": true,
		"message": "Reservation cancelled successfully",
	})
}

func (h *HTTPHandler) GetAllReservations(c echo.Context) error {
	reservations, err := h.service.GetAll()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"success": false, "message": "Failed to fetch reservations"})
	}

	return c.JSON(http.StatusOK, echo.Map{
		"success": true,
		"message": "All reservations retrieved successfully",
		"data":    reservations,
	})
}

func parseUint(s string) uint {
	var result uint
	for _, c := range s {
		result = result*10 + uint(c-'0')
	}
	return result
}