package reservation

import (
	"errors"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

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
	if err := r.db.Preload("Zone").First(&reservation, id).Error; err != nil {
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
	return reservations, r.db.Preload("Zone").Order("created_at DESC").Find(&reservations).Error
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