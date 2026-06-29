package service

import (
	"spotsync/dto"
	"spotsync/models"
	"spotsync/repository"
	"time"
)

type ZoneService struct {
	repo *repository.ZoneRepository
}

func NewZoneService(repo *repository.ZoneRepository) *ZoneService {
	return &ZoneService{repo: repo}
}

func (s *ZoneService) Create(req *dto.CreateZoneRequest) (*dto.ParkingZoneResponse, error) {
	zone := &models.ParkingZone{
		Name:         req.Name,
		Type:         models.ParkingZoneType(req.Type),
		TotalCapacity: req.TotalCapacity,
		PricePerHour:  req.PricePerHour,
	}

	if err := s.repo.Create(zone); err != nil {
		return nil, err
	}

	return &dto.ParkingZoneResponse{
		ID:            zone.ID,
		Name:          zone.Name,
		Type:          string(zone.Type),
		TotalCapacity: zone.TotalCapacity,
		AvailableSpots: zone.TotalCapacity,
		PricePerHour:  zone.PricePerHour,
		CreatedAt:     zone.CreatedAt.Format(time.RFC3339),
	}, nil
}

func (s *ZoneService) GetAll() ([]dto.ParkingZoneResponse, error) {
	zones, err := s.repo.FindAll()
	if err != nil {
		return nil, err
	}

	var responses []dto.ParkingZoneResponse
	for _, zone := range zones {
		activeCount, _ := s.repo.CountActiveReservations(zone.ID)
		responses = append(responses, dto.ParkingZoneResponse{
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

func (s *ZoneService) GetByID(id uint) (*dto.ParkingZoneResponse, error) {
	zone, err := s.repo.FindByID(id)
	if err != nil {
		return nil, err
	}

	activeCount, _ := s.repo.CountActiveReservations(zone.ID)

	return &dto.ParkingZoneResponse{
		ID:             zone.ID,
		Name:           zone.Name,
		Type:           string(zone.Type),
		TotalCapacity:  zone.TotalCapacity,
		AvailableSpots: int(zone.TotalCapacity) - int(activeCount),
		PricePerHour:   zone.PricePerHour,
		CreatedAt:      zone.CreatedAt.Format(time.RFC3339),
	}, nil
}

func (s *ZoneService) Update(id uint, req *dto.CreateZoneRequest) (*dto.ParkingZoneResponse, error) {
	zone, err := s.repo.FindByID(id)
	if err != nil {
		return nil, err
	}

	zone.Name = req.Name
	zone.Type = models.ParkingZoneType(req.Type)
	zone.TotalCapacity = req.TotalCapacity
	zone.PricePerHour = req.PricePerHour

	if err := s.repo.Update(zone); err != nil {
		return nil, err
	}

	activeCount, _ := s.repo.CountActiveReservations(zone.ID)

	return &dto.ParkingZoneResponse{
		ID:             zone.ID,
		Name:           zone.Name,
		Type:           string(zone.Type),
		TotalCapacity:  zone.TotalCapacity,
		AvailableSpots: int(zone.TotalCapacity) - int(activeCount),
		PricePerHour:   zone.PricePerHour,
		CreatedAt:      zone.CreatedAt.Format(time.RFC3339),
	}, nil
}

func (s *ZoneService) Delete(id uint) error {
	return s.repo.Delete(id)
}