package repository

import (
	"database/sql"
	"time"

	"fleet-management/internal/models"

	_ "github.com/lib/pq"
)

type VehicleLocationRepository struct {
	db *sql.DB
}

func NewVehicleLocationRepository(db *sql.DB) *VehicleLocationRepository {
	return &VehicleLocationRepository{db: db}
}

func (r *VehicleLocationRepository) Insert(loc *models.VehicleLocation) error {
	_, err := r.db.Exec(
		`INSERT INTO vehicle_locations (vehicle_id, latitude, longitude, timestamp) VALUES ($1, $2, $3, $4)`,
		loc.VehicleID, loc.Latitude, loc.Longitude, loc.Timestamp,
	)
	return err
}

func (r *VehicleLocationRepository) GetLatest(vehicleID string) (*models.VehicleLocation, error) {
	var loc models.VehicleLocation
	err := r.db.QueryRow(
		`SELECT vehicle_id, latitude, longitude, timestamp FROM vehicle_locations WHERE vehicle_id = $1 ORDER BY timestamp DESC LIMIT 1`,
		vehicleID,
	).Scan(&loc.VehicleID, &loc.Latitude, &loc.Longitude, &loc.Timestamp)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	loc.DateTime = time.Unix(loc.Timestamp, 0).In(time.FixedZone("WIB", 7*3600)).Format("02-01-2006 15:04:05")
	return &loc, nil
}

func (r *VehicleLocationRepository) GetHistory(vehicleID string, start, end int64) ([]models.VehicleLocation, error) {
	rows, err := r.db.Query(
		`SELECT vehicle_id, latitude, longitude, timestamp FROM vehicle_locations WHERE vehicle_id = $1 AND timestamp >= $2 AND timestamp <= $3 ORDER BY timestamp ASC`,
		vehicleID, start, end,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var locations []models.VehicleLocation
	for rows.Next() {
		var loc models.VehicleLocation
		if err := rows.Scan(&loc.VehicleID, &loc.Latitude, &loc.Longitude, &loc.Timestamp); err != nil {
			return nil, err
		}
		loc.DateTime = time.Unix(loc.Timestamp, 0).In(time.FixedZone("WIB", 7*3600)).Format("02-01-2006 15:04:05")
		locations = append(locations, loc)
	}

	return locations, rows.Err()
}
