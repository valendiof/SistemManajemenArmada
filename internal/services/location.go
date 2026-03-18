package services

import (
	"encoding/json"
	"fmt"
	"time"

	"fleet-management/internal/models"
	"fleet-management/internal/repository"
)

type LocationService struct {
	repo      *repository.VehicleLocationRepository
	rabbitPub *RabbitMQPublisher
}

func NewLocationService(repo *repository.VehicleLocationRepository, rabbitPub *RabbitMQPublisher) *LocationService {
	return &LocationService{repo: repo, rabbitPub: rabbitPub}
}

func (s *LocationService) ProcessMQTTMessage(payload []byte) error {
	var loc models.VehicleLocation
	if err := json.Unmarshal(payload, &loc); err != nil {
		return fmt.Errorf("invalid json: %w", err)
	}

	if loc.VehicleID == "" {
		return fmt.Errorf("vehicle_id is required")
	}
	if loc.Latitude < -90 || loc.Latitude > 90 {
		return fmt.Errorf("latitude must be between -90 and 90")
	}
	if loc.Longitude < -180 || loc.Longitude > 180 {
		return fmt.Errorf("longitude must be between -180 and 180")
	}

	loc.DateTime = time.Unix(loc.Timestamp, 0).In(time.FixedZone("WIB", 7*3600)).Format("02-01-2006 15:04:05")

	if err := s.repo.Insert(&loc); err != nil {
		return err
	}

	s.rabbitPub.PublishIfInGeofence(&loc)
	return nil
}

func (s *LocationService) GetLatest(vehicleID string) (*models.VehicleLocation, error) {
	return s.repo.GetLatest(vehicleID)
}

func (s *LocationService) GetHistory(vehicleID string, start, end int64) ([]models.VehicleLocation, error) {
	return s.repo.GetHistory(vehicleID, start, end)
}

func (s *LocationService) GetHistoryToday(vehicleID string) ([]models.VehicleLocation, error) {
	loc := time.FixedZone("WIB", 7*3600)
	now := time.Now().In(loc)

	// Start of today: 00:00:00
	start := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, loc).Unix()
	// End of today: 23:59:59
	end := time.Date(now.Year(), now.Month(), now.Day(), 23, 59, 59, 0, loc).Unix()

	return s.repo.GetHistory(vehicleID, start, end)
}
