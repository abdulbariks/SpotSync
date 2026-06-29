package user

import (
	"net/http"
	"time"

	"golang.org/x/crypto/bcrypt"
	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
)

type Service struct {
	repo      *Repository
	jwtSecret string
	jwtExpiry time.Duration
}

func NewService(repo *Repository, jwtSecret string, jwtExpiryHours int) *Service {
	return &Service{
		repo:      repo,
		jwtSecret: jwtSecret,
		jwtExpiry: time.Duration(jwtExpiryHours) * time.Hour,
	}
}

func (s *Service) Register(req *RegisterRequest) (*RegisterResponse, error) {
	existingUser, _ := s.repo.FindByEmail(req.Email)
	if existingUser != nil {
		return nil, echo.NewHTTPError(http.StatusBadRequest, "email already registered")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), 10)
	if err != nil {
		return nil, err
	}

	role := RoleDriver
	if req.Role == "admin" {
		role = RoleAdmin
	}

	u := &User{
		Name:     req.Name,
		Email:    req.Email,
		Password: string(hashedPassword),
		Role:     role,
	}

	if err := s.repo.Create(u); err != nil {
		return nil, err
	}

	return &RegisterResponse{
		ID:        u.ID,
		Name:      u.Name,
		Email:     u.Email,
		Role:      string(u.Role),
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
	}, nil
}

func (s *Service) Login(req *LoginRequest) (*LoginResponse, error) {
	u, err := s.repo.FindByEmail(req.Email)
	if err != nil {
		return nil, echo.NewHTTPError(http.StatusUnauthorized, "invalid credentials")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(req.Password)); err != nil {
		return nil, echo.NewHTTPError(http.StatusUnauthorized, "invalid credentials")
	}

	claims := &JWTClaims{
		UserID: u.ID,
		Role:   u.Role,
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

	resp := &LoginResponse{}
	resp.Token = tokenString
	resp.User.ID = u.ID
	resp.User.Name = u.Name
	resp.User.Email = u.Email
	resp.User.Role = string(u.Role)

	return resp, nil
}

func (s *Service) ValidateJWT(tokenString string) (*JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(s.jwtSecret), nil
	})
	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*JWTClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, echo.NewHTTPError(http.StatusUnauthorized, "invalid token")
}