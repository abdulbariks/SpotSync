package service

import (
	"errors"
	"time"

	"spotsync/dto"
	"spotsync/models"
	"spotsync/repository"
	"golang.org/x/crypto/bcrypt"
	"github.com/golang-jwt/jwt/v5"
)

type Service struct {
	repo        *repository.Repository
	jwtSecret   string
	jwtExpiry   time.Duration
}

func NewService(repo *repository.Repository, jwtSecret string, jwtExpiryHours int) *Service {
	return &Service{
		repo:      repo,
		jwtSecret: jwtSecret,
		jwtExpiry: time.Duration(jwtExpiryHours) * time.Hour,
	}
}

func (s *Service) Register(req *dto.RegisterRequest) (*dto.RegisterResponse, error) {
	existingUser, _ := s.repo.FindUserByEmail(req.Email)
	if existingUser != nil {
		return nil, errors.New("email already registered")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), 10)
	if err != nil {
		return nil, err
	}

	role := models.RoleDriver
	if req.Role == "admin" {
		role = models.RoleAdmin
	}

	user := &models.User{
		Name:     req.Name,
		Email:    req.Email,
		Password: string(hashedPassword),
		Role:     role,
	}

	if err := s.repo.CreateUser(user); err != nil {
		return nil, err
	}

	return &dto.RegisterResponse{
		ID:        user.ID,
		Name:      user.Name,
		Email:     user.Email,
		Role:      string(user.Role),
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}, nil
}

func (s *Service) Login(req *dto.LoginRequest) (*dto.LoginResponse, error) {
	user, err := s.repo.FindUserByEmail(req.Email)
	if err != nil {
		return nil, errors.New("invalid credentials")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return nil, errors.New("invalid credentials")
	}

	claims := &models.JWTClaims{
		UserID: user.ID,
		Role:   user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(s.jwtExpiry)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(s.jwtSecret))
	if err != nil {
		return nil, err
	}

	resp := &dto.LoginResponse{}
	resp.Token = tokenString
	resp.User.ID = user.ID
	resp.User.Name = user.Name
	resp.User.Email = user.Email
	resp.User.Role = string(user.Role)

	return resp, nil
}

func (s *Service) ValidateJWT(tokenString string) (*models.JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &models.JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(s.jwtSecret), nil
	})
	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*models.JWTClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token")
}

func (s *Service) CreateParkingZone(req *dto.CreateZoneRequest) (*dto.ParkingZoneResponse, error) {
	zoneType := models.ParkingZoneType(req.Type)
	zone := &models.ParkingZone{
		Name:         req.Name,
		Type:         zoneType,
		TotalCapacity: req.TotalCapacity,
		PricePerHour:  req.PricePerHour,
	}

	if err := s.repo.CreateParkingZone(zone); err != nil {
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

func (s *Service) GetAllParkingZones() ([]dto.ParkingZoneResponse, error) {
	zones, err := s.repo.FindAllParkingZones()
	if err != nil {
		return nil, err
	}

	var responses []dto.ParkingZoneResponse
	for _, zone := range zones {
		activeCount, _ := s.repo.CountActiveReservationsByZone(zone.ID)
		responses = append(responses, dto.ParkingZoneResponse{
			ID:            zone.ID,
			Name:          zone.Name,
			Type:          string(zone.Type),
			TotalCapacity: zone.TotalCapacity,
			AvailableSpots: int(zone.TotalCapacity) - int(activeCount),
			PricePerHour:  zone.PricePerHour,
			CreatedAt:     zone.CreatedAt.Format(time.RFC3339),
		})
	}

	return responses, nil
}

func (s *Service) GetParkingZoneByID(id uint) (*dto.ParkingZoneResponse, error) {
	zone, err := s.repo.FindParkingZoneByID(id)
	if err != nil {
		return nil, err
	}

	activeCount, _ := s.repo.CountActiveReservationsByZone(zone.ID)

	return &dto.ParkingZoneResponse{
		ID:            zone.ID,
		Name:          zone.Name,
		Type:          string(zone.Type),
		TotalCapacity: zone.TotalCapacity,
		AvailableSpots: int(zone.TotalCapacity) - int(activeCount),
		PricePerHour:  zone.PricePerHour,
		CreatedAt:     zone.CreatedAt.Format(time.RFC3339),
	}, nil
}

func (s *Service) UpdateParkingZone(id uint, req *dto.CreateZoneRequest) (*dto.ParkingZoneResponse, error) {
	zone, err := s.repo.FindParkingZoneByID(id)
	if err != nil {
		return nil, err
	}

	zone.Name = req.Name
	zone.Type = models.ParkingZoneType(req.Type)
	zone.TotalCapacity = req.TotalCapacity
	zone.PricePerHour = req.PricePerHour

	if err := s.repo.UpdateParkingZone(zone); err != nil {
		return nil, err
	}

	activeCount, _ := s.repo.CountActiveReservationsByZone(zone.ID)

	return &dto.ParkingZoneResponse{
		ID:            zone.ID,
		Name:          zone.Name,
		Type:          string(zone.Type),
		TotalCapacity: zone.TotalCapacity,
		AvailableSpots: int(zone.TotalCapacity) - int(activeCount),
		PricePerHour:  zone.PricePerHour,
		CreatedAt:     zone.CreatedAt.Format(time.RFC3339),
	}, nil
}

func (s *Service) DeleteParkingZone(id uint) error {
	return s.repo.DeleteParkingZone(id)
}

func (s *Service) ReserveSpot(req *dto.ReserveRequest, userID uint) (*dto.ReservationResponse, error) {
	existingReservation, _ := s.repo.FindActiveReservationByLicensePlate(req.LicensePlate)
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
		if errors.Is(err, repository.ErrZoneFull) {
			return nil, err
		}
		if errors.Is(err, repository.ErrZoneNotFound) {
			return nil, err
		}
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

func (s *Service) GetMyReservations(userID uint) ([]dto.MyReservationResponse, error) {
	reservations, err := s.repo.FindReservationsByUserID(userID)
	if err != nil {
		return nil, err
	}

	var responses []dto.MyReservationResponse
	for _, r := range reservations {
		resp := dto.MyReservationResponse{
			ID:           r.ID,
			LicensePlate:   r.LicensePlate,
			Status:         string(r.Status),
			CreatedAt:      r.CreatedAt.Format(time.RFC3339),
		}
		resp.Zone.ID = r.Zone.ID
		resp.Zone.Name = r.Zone.Name
		resp.Zone.Type = string(r.Zone.Type)
		responses = append(responses, resp)
	}

	return responses, nil
}

func (s *Service) CancelReservation(id uint, userID uint) error {
	reservation, err := s.repo.FindReservationByID(id)
	if err != nil {
		return err
	}

	if reservation.UserID != userID {
		return errors.New("unauthorized: cannot cancel another user's reservation")
	}

	return s.repo.UpdateReservation(&models.Reservation{
		ID:     reservation.ID,
		Status: models.StatusCancelled,
	})
}

func (s *Service) GetAllReservations() ([]dto.ReservationResponse, error) {
	reservations, err := s.repo.FindAllReservations()
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