package zone

import "time"

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