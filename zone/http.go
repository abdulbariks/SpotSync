package zone

import (
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

func (h *HTTPHandler) CreateZone(c echo.Context) error {
	var req CreateRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"success": false, "message": "Invalid request body", "errors": err.Error()})
	}

	if err := c.Validate(&req); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"success": false, "message": "Validation failed", "errors": err.Error()})
	}

	resp, err := h.service.Create(&req)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"success": false, "message": "Failed to create zone"})
	}

	return c.JSON(http.StatusCreated, echo.Map{
		"success": true,
		"message": "Parking zone created successfully",
		"data":    resp,
	})
}

func (h *HTTPHandler) GetAllZones(c echo.Context) error {
	zones, err := h.service.GetAll()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"success": false, "message": "Failed to fetch zones"})
	}

	return c.JSON(http.StatusOK, echo.Map{
		"success": true,
		"message": "Parking zones retrieved successfully",
		"data":    zones,
	})
}

func (h *HTTPHandler) GetZone(c echo.Context) error {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)
	zone, err := h.service.GetByID(uint(id))
	if err != nil {
		if err == ErrZoneNotFound {
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

func (h *HTTPHandler) UpdateZone(c echo.Context) error {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)

	var req CreateRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"success": false, "message": "Invalid request body", "errors": err.Error()})
	}

	if err := c.Validate(&req); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"success": false, "message": "Validation failed", "errors": err.Error()})
	}

	resp, err := h.service.Update(uint(id), &req)
	if err != nil {
		if err == ErrZoneNotFound {
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

func (h *HTTPHandler) DeleteZone(c echo.Context) error {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)
	if err := h.service.Delete(uint(id)); err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"success": false, "message": "Failed to delete zone"})
	}

	return c.JSON(http.StatusOK, echo.Map{
		"success": true,
		"message": "Parking zone deleted successfully",
	})
}