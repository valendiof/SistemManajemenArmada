package handlers

import (
	"net/http"
	"strconv"

	"fleet-management/internal/services"

	"github.com/gin-gonic/gin"
)

type VehicleHandler struct {
	locationService *services.LocationService
}

func NewVehicleHandler(locationService *services.LocationService) *VehicleHandler {
	return &VehicleHandler{locationService: locationService}
}

func (h *VehicleHandler) GetLocation(c *gin.Context) {
	vehicleID := c.Param("vehicle_id")
	if vehicleID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "vehicle_id is required"})
		return
	}

	loc, err := h.locationService.GetLatest(vehicleID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if loc == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "no location found for vehicle"})
		return
	}

	c.JSON(http.StatusOK, loc)
}

func (h *VehicleHandler) GetHistory(c *gin.Context) {
	vehicleID := c.Param("vehicle_id")
	if vehicleID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "vehicle_id is required"})
		return
	}

	startStr := c.Query("start")
	endStr := c.Query("end")
	if startStr == "" || endStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "start and end query params are required"})
		return
	}

	start, err := strconv.ParseInt(startStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid start timestamp"})
		return
	}
	end, err := strconv.ParseInt(endStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid end timestamp"})
		return
	}
	if start > end {
		c.JSON(http.StatusBadRequest, gin.H{"error": "start must be less than or equal to end"})
		return
	}

	locations, err := h.locationService.GetHistory(vehicleID, start, end)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, locations)
}

func (h *VehicleHandler) GetHistoryToday(c *gin.Context) {
	vehicleID := c.Param("vehicle_id")
	if vehicleID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "vehicle_id is required"})
		return
	}

	locations, err := h.locationService.GetHistoryToday(vehicleID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, locations)
}
