package reservation

import (
	"errors"
	"time"
)

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