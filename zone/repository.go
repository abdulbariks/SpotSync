package zone

import (
	"errors"

	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

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
		if errors.Is(err, gorm.ErrRecordNotFound) {
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