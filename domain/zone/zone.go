package zone

import (
	"time"

	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

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

type CreateRequest struct {
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

var ErrZoneNotFound = echo.NewHTTPError(404, "parking zone not found")

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Create(zone *ParkingZone) error {
	return r.db.Create(zone).Error
}

func (r *Repository) FindAll() ([]ParkingZone, error) {
	var zones []ParkingZone
	return zones, r.db.Order("created_at DESC").Find(&zones).Error
}

func (r *Repository) FindByID(id uint) (*ParkingZone, error) {
	var zone ParkingZone
	if err := r.db.First(&zone, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrZoneNotFound
		}
		return nil, err
	}
	return &zone, nil
}

func (r *Repository) Update(zone *ParkingZone) error {
	return r.db.Save(zone).Error
}

func (r *Repository) Delete(id uint) error {
	return r.db.Delete(&ParkingZone{}, id).Error
}

func (r *Repository) CountActiveReservations(zoneID uint) (int64, error) {
	var count int64
	return count, r.db.Table("reservations").Where("zone_id = ? AND status = ?", zoneID, "active").Count(&count).Error
}

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Create(req *CreateRequest) (*ParkingZoneResponse, error) {
	zone := &ParkingZone{
		Name:        req.Name,
		Type:        ZoneType(req.Type),
		TotalCapacity: req.TotalCapacity,
		PricePerHour: req.PricePerHour,
	}

	if err := s.repo.Create(zone); err != nil {
		return nil, err
	}

	return &ParkingZoneResponse{
		ID:            zone.ID,
		Name:          zone.Name,
		Type:          string(zone.Type),
		TotalCapacity: zone.TotalCapacity,
		AvailableSpots: zone.TotalCapacity,
		PricePerHour:  zone.PricePerHour,
		CreatedAt:     zone.CreatedAt.Format(time.RFC3339),
	}, nil
}

func (s *Service) GetAll() ([]ParkingZoneResponse, error) {
	zones, err := s.repo.FindAll()
	if err != nil {
		return nil, err
	}

	var responses []ParkingZoneResponse
	for _, zone := range zones {
		activeCount, _ := s.repo.CountActiveReservations(zone.ID)
		responses = append(responses, ParkingZoneResponse{
			ID:             zone.ID,
			Name:           zone.Name,
			Type:           string(zone.Type),
			TotalCapacity:  zone.TotalCapacity,
			AvailableSpots: int(zone.TotalCapacity) - int(activeCount),
			PricePerHour:   zone.PricePerHour,
			CreatedAt:      zone.CreatedAt.Format(time.RFC3339),
		})
	}

	return responses, nil
}

func (s *Service) GetByID(id uint) (*ParkingZoneResponse, error) {
	zone, err := s.repo.FindByID(id)
	if err != nil {
		return nil, err
	}

	activeCount, _ := s.repo.CountActiveReservations(zone.ID)

	return &ParkingZoneResponse{
		ID:             zone.ID,
		Name:           zone.Name,
		Type:           string(zone.Type),
		TotalCapacity:  zone.TotalCapacity,
		AvailableSpots: int(zone.TotalCapacity) - int(activeCount),
		PricePerHour:   zone.PricePerHour,
		CreatedAt:      zone.CreatedAt.Format(time.RFC3339),
	}, nil
}

func (s *Service) Update(id uint, req *CreateRequest) (*ParkingZoneResponse, error) {
	zone, err := s.repo.FindByID(id)
	if err != nil {
		return nil, err
	}

	zone.Name = req.Name
	zone.Type = ZoneType(req.Type)
	zone.TotalCapacity = req.TotalCapacity
	zone.PricePerHour = req.PricePerHour

	if err := s.repo.Update(zone); err != nil {
		return nil, err
	}

	activeCount, _ := s.repo.CountActiveReservations(zone.ID)

	return &ParkingZoneResponse{
		ID:             zone.ID,
		Name:           zone.Name,
		Type:           string(zone.Type),
		TotalCapacity:  zone.TotalCapacity,
		AvailableSpots: int(zone.TotalCapacity) - int(activeCount),
		PricePerHour:   zone.PricePerHour,
		CreatedAt:      zone.CreatedAt.Format(time.RFC3339),
	}, nil
}

func (s *Service) Delete(id uint) error {
	return s.repo.Delete(id)
}

type HTTPHandler struct {
	service *Service
}

func NewHTTPHandler(svc *Service) *HTTPHandler {
	return &HTTPHandler{service: svc}
}

func (h *HTTPHandler) CreateZone(c echo.Context) error {
	var req CreateRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(400, echo.Map{"success": false, "message": "Invalid request body", "errors": err.Error()})
	}

	if err := c.Validate(&req); err != nil {
		return c.JSON(400, echo.Map{"success": false, "message": "Validation failed", "errors": err.Error()})
	}

	resp, err := h.service.Create(&req)
	if err != nil {
		return c.JSON(500, echo.Map{"success": false, "message": "Failed to create zone"})
	}

	return c.JSON(201, echo.Map{
		"success": true,
		"message": "Parking zone created successfully",
		"data":    resp,
	})
}

func (h *HTTPHandler) GetAllZones(c echo.Context) error {
	zones, err := h.service.GetAll()
	if err != nil {
		return c.JSON(500, echo.Map{"success": false, "message": "Failed to fetch zones"})
	}

	return c.JSON(200, echo.Map{
		"success": true,
		"message": "Parking zones retrieved successfully",
		"data":    zones,
	})
}

func (h *HTTPHandler) GetZone(c echo.Context) error {
	var id uint
	for i := 0; i < len(c.Param("id")); i++ {
		if c.Param("id")[i] < '0' || c.Param("id")[i] > '9' {
			return c.JSON(400, echo.Map{"success": false, "message": "Invalid zone ID"})
		}
	}
	id = parseUint(c.Param("id"))

	zone, err := h.service.GetByID(id)
	if err != nil {
		if err == ErrZoneNotFound {
			return c.JSON(404, echo.Map{"success": false, "message": "Parking zone not found"})
		}
		return c.JSON(500, echo.Map{"success": false, "message": "Failed to fetch zone"})
	}

	return c.JSON(200, echo.Map{
		"success": true,
		"message": "Parking zone retrieved successfully",
		"data":    zone,
	})
}

func (h *HTTPHandler) UpdateZone(c echo.Context) error {
	id := parseUint(c.Param("id"))

	var req CreateRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(400, echo.Map{"success": false, "message": "Invalid request body", "errors": err.Error()})
	}

	if err := c.Validate(&req); err != nil {
		return c.JSON(400, echo.Map{"success": false, "message": "Validation failed", "errors": err.Error()})
	}

	resp, err := h.service.Update(id, &req)
	if err != nil {
		if err == ErrZoneNotFound {
			return c.JSON(404, echo.Map{"success": false, "message": "Parking zone not found"})
		}
		return c.JSON(500, echo.Map{"success": false, "message": "Failed to update zone"})
	}

	return c.JSON(200, echo.Map{
		"success": true,
		"message": "Parking zone updated successfully",
		"data":    resp,
	})
}

func (h *HTTPHandler) DeleteZone(c echo.Context) error {
	id := parseUint(c.Param("id"))
	if err := h.service.Delete(id); err != nil {
		return c.JSON(500, echo.Map{"success": false, "message": "Failed to delete zone"})
	}

	return c.JSON(200, echo.Map{
		"success": true,
		"message": "Parking zone deleted successfully",
	})
}

func parseUint(s string) uint {
	var result uint
	for _, c := range s {
		result = result*10 + uint(c-'0')
	}
	return result
}