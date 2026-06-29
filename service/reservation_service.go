package service

import (
	"errors"
	"time"

	"spotsync/dto"
	"spotsync/models"
	"spotsync/repository"
)

type ReservationService struct {
	repo *repository.ReservationRepository
}

func NewReservationService(repo *repository.ReservationRepository) *ReservationService {
	return &ReservationService{repo: repo}
}

func (s *ReservationService) Create(req *dto.ReserveRequest, userID uint) (*dto.ReservationResponse, error) {
	existingReservation, _ := s.repo.FindActiveByLicensePlate(req.LicensePlate)
	if existingReservation != nil {
		return nil, repository.ErrLicensePlateExists
	}

	reservation := &models.Reservation{
		UserID:       userID,
		ZoneID:       req.ZoneID,
		LicensePlate: req.LicensePlate,
		Status:       models.StatusActive,
	}

	if err := s.repo.ReserveSpot(req.ZoneID, reservation); err != nil {
		return nil, err
	}

	return &dto.ReservationResponse{
		ID:           reservation.ID,
		UserID:       reservation.UserID,
		ZoneID:       reservation.ZoneID,
		LicensePlate: reservation.LicensePlate,
		Status:       string(reservation.Status),
		CreatedAt:    reservation.CreatedAt.Format(time.RFC3339),
		UpdatedAt:    reservation.UpdatedAt.Format(time.RFC3339),
	}, nil
}

func (s *ReservationService) GetMyReservations(userID uint) ([]dto.MyReservationResponse, error) {
	reservations, err := s.repo.FindByUserID(userID)
	if err != nil {
		return nil, err
	}

	var responses []dto.MyReservationResponse
	for _, r := range reservations {
		resp := dto.MyReservationResponse{
			ID:           r.ID,
			LicensePlate: r.LicensePlate,
			Status:       string(r.Status),
			CreatedAt:    r.CreatedAt.Format(time.RFC3339),
		}
		resp.Zone.ID = r.Zone.ID
		resp.Zone.Name = r.Zone.Name
		resp.Zone.Type = string(r.Zone.Type)
		responses = append(responses, resp)
	}

	return responses, nil
}

func (s *ReservationService) Cancel(id uint, userID uint) error {
	reservation, err := s.repo.FindByID(id)
	if err != nil {
		return err
	}

	if reservation.UserID != userID {
		return errors.New("unauthorized: cannot cancel another user's reservation")
	}

	return s.repo.Update(&models.Reservation{
		ID:     reservation.ID,
		Status: models.StatusCancelled,
	})
}

func (s *ReservationService) GetAll() ([]dto.ReservationResponse, error) {
	reservations, err := s.repo.FindAll()
	if err != nil {
		return nil, err
	}

	var responses []dto.ReservationResponse
	for _, r := range reservations {
		responses = append(responses, dto.ReservationResponse{
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