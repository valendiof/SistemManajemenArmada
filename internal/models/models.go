package models

type VehicleLocation struct {
	VehicleID string  `json:"vehicle_id"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Timestamp int64   `json:"timestamp"`
	DateTime  string  `json:"date_time"`
}

type GeofenceEvent struct {
	VehicleID string `json:"vehicle_id"`
	Event     string `json:"event"`
	Location  LatLng `json:"location"`
	Timestamp int64  `json:"timestamp"`
	DateTime  string `json:"date_time"`
}

type LatLng struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}
