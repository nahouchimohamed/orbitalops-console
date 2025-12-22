package domain

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type Satellite struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	NoradID   *int      `json:"norad_id,omitempty"`
	TLELine1  string    `json:"tle_line1"`
	TLELine2  string    `json:"tle_line2"`
	CreatedAt time.Time `json:"created_at"`
}

type GroundStation struct {
	ID              uuid.UUID `json:"id"`
	Name            string    `json:"name"`
	Latitude        float64   `json:"latitude"`
	Longitude       float64   `json:"longitude"`
	AltitudeM       float64   `json:"altitude_m"`
	MinElevationDeg float64   `json:"min_elevation_deg"`
	CreatedAt       time.Time `json:"created_at"`
}

type ObservationRequest struct {
	ID        uuid.UUID       `json:"id"`
	Name      string          `json:"name"`
	AOI       json.RawMessage `json:"aoi_geojson"`
	Priority  int             `json:"priority"`
	Status    string          `json:"status"`
	CreatedAt time.Time       `json:"created_at"`
}

type CommandStep struct {
	Time    string `json:"time"`
	Command string `json:"command"`
	Args    map[string]any `json:"args,omitempty"`
}

type ImagingPlan struct {
	ID              uuid.UUID     `json:"id"`
	RequestID       uuid.UUID     `json:"request_id"`
	SatelliteID     uuid.UUID     `json:"satellite_id"`
	SatelliteName   string        `json:"satellite_name,omitempty"`
	StartTime       time.Time     `json:"start_time"`
	EndTime         time.Time     `json:"end_time"`
	LookSide        string        `json:"look_side"`
	Status          string        `json:"status"`
	CommandSequence []CommandStep `json:"command_sequence"`
	CreatedAt       time.Time     `json:"created_at"`
}

type PassWindow struct {
	SatelliteID     uuid.UUID `json:"satellite_id"`
	SatelliteName   string    `json:"satellite_name"`
	StationID       uuid.UUID `json:"station_id"`
	StationName     string    `json:"station_name"`
	AOS             time.Time `json:"aos"`
	LOS             time.Time `json:"los"`
	MaxElevationDeg float64   `json:"max_elevation_deg"`
	MaxElevationAt  time.Time `json:"max_elevation_at"`
}

type TrackSample struct {
	Time      time.Time `json:"time"`
	Latitude  float64   `json:"latitude"`
	Longitude float64   `json:"longitude"`
	AltitudeKM float64  `json:"altitude_km"`
}

type GroundTrack struct {
	SatelliteID   uuid.UUID     `json:"satellite_id"`
	SatelliteName string        `json:"satellite_name"`
	Samples       []TrackSample `json:"samples"`
}
