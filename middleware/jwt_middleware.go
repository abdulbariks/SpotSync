package middleware

import (
	"net/http"
	"strings"

	"spotsync/service"

	"github.com/labstack/echo/v4"
)

type JWTMiddleware struct {
	service *service.Service
}

func NewJWTMiddleware(service *service.Service) *JWTMiddleware {
	return &JWTMiddleware{service: service}
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