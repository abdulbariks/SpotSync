package repository

import (
	"errors"

	"spotsync/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var (
	ErrZoneFull       = errors.New("parking zone is at full capacity")
	ErrZoneNotFound   = errors.New("parking zone not found")
	ErrUserNotFound   = errors.New("user not found")
	ErrReservationNotFound = errors.New("reservation not found")
	ErrLicensePlateExists  = errors.New("license plate already has active reservation")
)

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) CreateUser(user *models.User) error {
	return r.db.Create(user).Error
}

func (r *Repository) FindUserByEmail(email string) (*models.User, error) {
	var user models.User
	if err := r.db.Where("email = ?", email).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	return &user, nil
}

func (r *Repository) FindUserByID(id uint) (*models.User, error) {
	var user models.User
	if err := r.db.First(&user, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	return &user, nil
}

func (r *Repository) CreateParkingZone(zone *models.ParkingZone) error {
	return r.db.Create(zone).Error
}

func (r *Repository) FindAllParkingZones() ([]models.ParkingZone, error) {
	var zones []models.ParkingZone
	err := r.db.Order("created_at DESC").Find(&zones).Error
	return zones, err
}

func (r *Repository) FindParkingZoneByID(id uint) (*models.ParkingZone, error) {
	var zone models.ParkingZone
	if err := r.db.First(&zone, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrZoneNotFound
		}
		return nil, err
	}
	return &zone, nil
}

func (r *Repository) UpdateParkingZone(zone *models.ParkingZone) error {
	return r.db.Save(zone).Error
}

func (r *Repository) DeleteParkingZone(id uint) error {
	return r.db.Delete(&models.ParkingZone{}, id).Error
}

func (r *Repository) CreateReservation(reservation *models.Reservation) error {
	return r.db.Create(reservation).Error
}

func (r *Repository) FindReservationByID(id uint) (*models.Reservation, error) {
	var reservation models.Reservation
	if err := r.db.Preload("User").Preload("Zone").First(&reservation, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrReservationNotFound
		}
		return nil, err
	}
	return &reservation, nil
}

func (r *Repository) FindReservationsByUserID(userID uint) ([]models.Reservation, error) {
	var reservations []models.Reservation
	err := r.db.Where("user_id = ?", userID).Preload("Zone").Order("created_at DESC").Find(&reservations).Error
	return reservations, err
}

func (r *Repository) FindAllReservations() ([]models.Reservation, error) {
	var reservations []models.Reservation
	err := r.db.Preload("User").Preload("Zone").Order("created_at DESC").Find(&reservations).Error
	return reservations, err
}

func (r *Repository) UpdateReservation(reservation *models.Reservation) error {
	return r.db.Save(reservation).Error
}

func (r *Repository) CountActiveReservationsByZone(zoneID uint) (int64, error) {
	var count int64
	return count, r.db.Model(&models.Reservation{}).Where("zone_id = ? AND status = ?", zoneID, models.StatusActive).Count(&count).Error
}

func (r *Repository) FindActiveReservationByLicensePlate(licensePlate string) (*models.Reservation, error) {
	var reservation models.Reservation
	if err := r.db.Where("license_plate = ? AND status = ?", licensePlate, models.StatusActive).First(&reservation).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &reservation, nil
}

func (r *Repository) ReserveSpot(zoneID uint, reservation *models.Reservation) error {
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