package domain

import "time"

type TLE struct {
	Name  string
	Line1 string
	Line2 string
}

type GeodeticSample struct {
	Time      time.Time
	Latitude  float64
	Longitude float64
	AltitudeKM float64
}

type GroundStation struct {
	Name             string
	LatitudeDeg      float64
	LongitudeDeg     float64
	AltitudeM        float64
	MinElevationDeg  float64
}

type PassWindow struct {
	AOS               time.Time
	LOS               time.Time
	MaxElevationDeg   float64
	MaxElevationTime  time.Time
}

type Propagator interface {
	Propagate(tle TLE, start, end time.Time, step time.Duration) ([]GeodeticSample, error)
	FindPasses(tle TLE, station GroundStation, start, end time.Time, step time.Duration) ([]PassWindow, error)
}
