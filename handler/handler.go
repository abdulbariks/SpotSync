package handler

import (
	"net/http"
	"strconv"

	"spotsync/dto"
	"spotsync/models"
	"spotsync/repository"
	"spotsync/service"

	"github.com/labstack/echo/v4"
)

type Handler struct {
	service *service.Service
}

func NewHandler(svc *service.Service) *Handler {
	return &Handler{service: svc}
}

func (h *Handler) Register(c echo.Context) error {
	var req dto.RegisterRequest
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

func (h *Handler) Login(c echo.Context) error {
	var req dto.LoginRequest
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

func (h *Handler) CreateZone(c echo.Context) error {
	var req dto.CreateZoneRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"success": false, "message": "Invalid request body", "errors": err.Error()})
	}

	if err := c.Validate(&req); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"success": false, "message": "Validation failed", "errors": err.Error()})
	}

	resp, err := h.service.CreateParkingZone(&req)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"success": false, "message": "Failed to create zone"})
	}

	return c.JSON(http.StatusCreated, echo.Map{
		"success": true,
		"message": "Parking zone created successfully",
		"data":    resp,
	})
}

func (h *Handler) GetAllZones(c echo.Context) error {
	zones, err := h.service.GetAllParkingZones()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"success": false, "message": "Failed to fetch zones"})
	}

	return c.JSON(http.StatusOK, echo.Map{
		"success": true,
		"message": "Parking zones retrieved successfully",
		"data":    zones,
	})
}

func (h *Handler) GetZone(c echo.Context) error {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)
	zone, err := h.service.GetParkingZoneByID(uint(id))
	if err != nil {
		if err == repository.ErrZoneNotFound {
			return c.JSON(http.StatusNotFound, echo.Map{"success": false, "message": "Parking zone not found"})
		}
		return c.JSON(http.StatusInternalServerError, echo.Map{"success": false, "message": "Failed to fetch zone"})
	}

	return c.JSON(http.StatusOK, echo.Map{
		"success": true,
		"message": "Parking zone retrieved successfully",
		"data":    zone,
	})
}

func (h *Handler) UpdateZone(c echo.Context) error {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)

	var req dto.CreateZoneRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"success": false, "message": "Invalid request body", "errors": err.Error()})
	}

	if err := c.Validate(&req); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"success": false, "message": "Validation failed", "errors": err.Error()})
	}

	resp, err := h.service.UpdateParkingZone(uint(id), &req)
	if err != nil {
		if err == repository.ErrZoneNotFound {
			return c.JSON(http.StatusNotFound, echo.Map{"success": false, "message": "Parking zone not found"})
		}
		return c.JSON(http.StatusInternalServerError, echo.Map{"success": false, "message": "Failed to update zone"})
	}

	return c.JSON(http.StatusOK, echo.Map{
		"success": true,
		"message": "Parking zone updated successfully",
		"data":    resp,
	})
}

func (h *Handler) DeleteZone(c echo.Context) error {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)
	if err := h.service.DeleteParkingZone(uint(id)); err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"success": false, "message": "Failed to delete zone"})
	}

	return c.JSON(http.StatusOK, echo.Map{
		"success": true,
		"message": "Parking zone deleted successfully",
	})
}

func (h *Handler) CreateReservation(c echo.Context) error {
	user, ok := c.Get("user").(*models.JWTClaims)
	if !ok {
		return c.JSON(http.StatusUnauthorized, echo.Map{"success": false, "message": "Unauthorized"})
	}

	var req dto.ReserveRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"success": false, "message": "Invalid request body", "errors": err.Error()})
	}

	if err := c.Validate(&req); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"success": false, "message": "Validation failed", "errors": err.Error()})
	}

	resp, err := h.service.ReserveSpot(&req, user.UserID)
	if err != nil {
		if err == repository.ErrZoneFull {
			return c.JSON(http.StatusConflict, echo.Map{"success": false, "message": "Parking zone is at full capacity"})
		}
		if err == repository.ErrZoneNotFound {
			return c.JSON(http.StatusNotFound, echo.Map{"success": false, "message": "Parking zone not found"})
		}
		if err == repository.ErrLicensePlateExists {
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

func (h *Handler) GetMyReservations(c echo.Context) error {
	user, ok := c.Get("user").(*models.JWTClaims)
	if !ok {
		return c.JSON(http.StatusUnauthorized, echo.Map{"success": false, "message": "Unauthorized"})
	}

	reservations, err := h.service.GetMyReservations(user.UserID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"success": false, "message": "Failed to fetch reservations"})
	}

	return c.JSON(http.StatusOK, echo.Map{
		"success": true,
		"message": "My reservations retrieved successfully",
		"data":    reservations,
	})
}

func (h *Handler) CancelReservation(c echo.Context) error {
	user, ok := c.Get("user").(*models.JWTClaims)
	if !ok {
		return c.JSON(http.StatusUnauthorized, echo.Map{"success": false, "message": "Unauthorized"})
	}

	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)

	if err := h.service.CancelReservation(uint(id), user.UserID); err != nil {
		if err.Error() == "unauthorized: cannot cancel another user's reservation" {
			return c.JSON(http.StatusForbidden, echo.Map{"success": false, "message": "Cannot cancel another user's reservation"})
		}
		if err == repository.ErrReservationNotFound {
			return c.JSON(http.StatusNotFound, echo.Map{"success": false, "message": "Reservation not found"})
		}
		return c.JSON(http.StatusInternalServerError, echo.Map{"success": false, "message": "Failed to cancel reservation"})
	}

	return c.JSON(http.StatusOK, echo.Map{
		"success": true,
		"message": "Reservation cancelled successfully",
	})
}

func (h *Handler) GetAllReservations(c echo.Context) error {
	reservations, err := h.service.GetAllReservations()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"success": false, "message": "Failed to fetch reservations"})
	}

	return c.JSON(http.StatusOK, echo.Map{
		"success": true,
		"message": "All reservations retrieved successfully",
		"data":    reservations,
	})
}