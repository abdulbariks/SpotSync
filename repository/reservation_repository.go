package repository

import (
	"errors"

	"spotsync/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var (
	ErrReservationNotFound = errors.New("reservation not found")
	ErrZoneFull          = errors.New("parking zone is at full capacity")
	ErrLicensePlateExists = errors.New("license plate already has active reservation")
	ErrZoneNotFound       = errors.New("parking zone not found")
)

type ReservationRepository struct {
	db *gorm.DB
}

func NewReservationRepository(db *gorm.DB) *ReservationRepository {
	return &ReservationRepository{db: db}
}

func (r *ReservationRepository) Create(reservation *models.Reservation) error {
	return r.db.Create(reservation).Error
}

func (r *ReservationRepository) FindByID(id uint) (*models.Reservation, error) {
	var reservation models.Reservation
	if err := r.db.Preload("User").Preload("Zone").First(&reservation, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrReservationNotFound
		}
		return nil, err
	}
	return &reservation, nil
}

func (r *ReservationRepository) FindByUserID(userID uint) ([]models.Reservation, error) {
	var reservations []models.Reservation
	err := r.db.Where("user_id = ?", userID).Preload("Zone").Order("created_at DESC").Find(&reservations).Error
	return reservations, err
}

func (r *ReservationRepository) FindAll() ([]models.Reservation, error) {
	var reservations []models.Reservation
	err := r.db.Preload("User").Preload("Zone").Order("created_at DESC").Find(&reservations).Error
	return reservations, err
}

func (r *ReservationRepository) Update(reservation *models.Reservation) error {
	return r.db.Save(reservation).Error
}

func (r *ReservationRepository) FindActiveByLicensePlate(licensePlate string) (*models.Reservation, error) {
	var reservation models.Reservation
	if err := r.db.Where("license_plate = ? AND status = ?", licensePlate, models.StatusActive).First(&reservation).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &reservation, nil
}

func (r *ReservationRepository) ReserveSpot(zoneID uint, reservation *models.Reservation) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		var zone models.ParkingZone
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&zone, zoneID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrZoneNotFound
			}
			return err
		}

		var activeCount int64
		if err := tx.Model(&models.Reservation{}).Where("zone_id = ? AND status = ?", zoneID, models.StatusActive).Count(&activeCount).Error; err != nil {
			return err
		}

		if activeCount >= int64(zone.TotalCapacity) {
			return ErrZoneFull
		}

		return tx.Create(reservation).Error
	})
}