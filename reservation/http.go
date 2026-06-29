package reservation

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
)

type HTTPHandler struct {
	service *Service
}

func NewHTTPHandler(svc *Service) *HTTPHandler {
	return &HTTPHandler{service: svc}
}

func (h *HTTPHandler) CreateReservation(c echo.Context) error {
	u, ok := c.Get("user").(*JWTClaims)
	if !ok {
		return c.JSON(http.StatusUnauthorized, echo.Map{"success": false, "message": "Unauthorized"})
	}

	var req CreateRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"success": false, "message": "Invalid request body", "errors": err.Error()})
	}

	if err := c.Validate(&req); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"success": false, "message": "Validation failed", "errors": err.Error()})
	}

	resp, err := h.service.Create(&req, u.UserID)
	if err != nil {
		if err == ErrZoneFull {
			return c.JSON(http.StatusConflict, echo.Map{"success": false, "message": "Parking zone is at full capacity"})
		}
		if err == ErrZoneNotFound {
			return c.JSON(http.StatusNotFound, echo.Map{"success": false, "message": "Parking zone not found"})
		}
		if err == ErrLicensePlateExists {
			return c.JSON(http.StatusConflict, echo.Map{"success": false, "message": "License plate already has active reservation"})
		}
		return c.JSON(http.StatusInternalServerError, echo.Map{"success": false, "message": "Failed to create reservation"})
	}

	return c.JSON(http.StatusCreated, echo.Map{
		"success": true,
		"message": "Reservation confirmed successfully",
		"data":    resp,
	})
}

func (h *HTTPHandler) GetMyReservations(c echo.Context) error {
	u, ok := c.Get("user").(*JWTClaims)
	if !ok {
		return c.JSON(http.StatusUnauthorized, echo.Map{"success": false, "message": "Unauthorized"})
	}

	reservations, err := h.service.GetMyReservations(u.UserID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"success": false, "message": "Failed to fetch reservations"})
	}

	return c.JSON(http.StatusOK, echo.Map{
		"success": true,
		"message": "My reservations retrieved successfully",
		"data":    reservations,
	})
}

func (h *HTTPHandler) CancelReservation(c echo.Context) error {
	u, ok := c.Get("user").(*JWTClaims)
	if !ok {
		return c.JSON(http.StatusUnauthorized, echo.Map{"success": false, "message": "Unauthorized"})
	}

	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)

	if err := h.service.Cancel(uint(id), u.UserID); err != nil {
		if err.Error() == "unauthorized: cannot cancel another user's reservation" {
			return c.JSON(http.StatusForbidden, echo.Map{"success": false, "message": "Cannot cancel another user's reservation"})
		}
		if errors.Is(err, ErrReservationNotFound) {
			return c.JSON(http.StatusNotFound, echo.Map{"success": false, "message": "Reservation not found"})
		}
		return c.JSON(http.StatusInternalServerError, echo.Map{"success": false, "message": "Failed to cancel reservation"})
	}

	return c.JSON(http.StatusOK, echo.Map{
		"success": true,
		"message": "Reservation cancelled successfully",
	})
}

func (h *HTTPHandler) GetAllReservations(c echo.Context) error {
	reservations, err := h.service.GetAll()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"success": false, "message": "Failed to fetch reservations"})
	}

	return c.JSON(http.StatusOK, echo.Map{
		"success": true,
		"message": "All reservations retrieved successfully",
		"data":    reservations,
	})
}