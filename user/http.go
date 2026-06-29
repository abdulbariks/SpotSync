package user

import (
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
)

type HTTPHandler struct {
	service *Service
}

func NewHTTPHandler(svc *Service) *HTTPHandler {
	return &HTTPHandler{service: svc}
}

func (h *HTTPHandler) Register(c echo.Context) error {
	var req RegisterRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"success": false, "message": "Invalid request body", "errors": err.Error()})
	}

	if err := c.Validate(&req); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"success": false, "message": "Validation failed", "errors": err.Error()})
	}

	resp, err := h.service.Register(&req)
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"success": false, "message": err.Error()})
	}

	return c.JSON(http.StatusCreated, echo.Map{
		"success": true,
		"message": "User registered successfully",
		"data":    resp,
	})
}

func (h *HTTPHandler) Login(c echo.Context) error {
	var req LoginRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"success": false, "message": "Invalid request body", "errors": err.Error()})
	}

	if err := c.Validate(&req); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"success": false, "message": "Validation failed", "errors": err.Error()})
	}

	resp, err := h.service.Login(&req)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, echo.Map{"success": false, "message": err.Error()})
	}

	return c.JSON(http.StatusOK, echo.Map{
		"success": true,
		"message": "Login successful",
		"data":    resp,
	})
}

type JWTMiddleware struct {
	service *Service
}

func NewJWTMiddleware(svc *Service) *JWTMiddleware {
	return &JWTMiddleware{service: svc}
}

func (m *JWTMiddleware) VerifyJWT(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		authHeader := c.Request().Header.Get("Authorization")
		if authHeader == "" {
			return c.JSON(http.StatusUnauthorized, echo.Map{"success": false, "message": "Missing authorization token"})
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			return c.JSON(http.StatusUnauthorized, echo.Map{"success": false, "message": "Invalid authorization header format"})
		}

		claims, err := m.service.ValidateJWT(parts[1])
		if err != nil {
			return c.JSON(http.StatusUnauthorized, echo.Map{"success": false, "message": "Invalid or expired token"})
		}

		c.Set("user", claims)
		return next(c)
	}
}

func AdminOnly(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		u, ok := c.Get("user").(*JWTClaims)
		if !ok || u.Role != RoleAdmin {
			return c.JSON(http.StatusForbidden, echo.Map{"success": false, "message": "Admin access required"})
		}
		return next(c)
	}
}