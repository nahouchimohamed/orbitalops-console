package domain

import (
	"testing"
	"time"
)

func TestPropagateProducesSamples(t *testing.T) {
	p := NewKeplerPropagator()
	tle := TLE{
		Name:  "demo",
		Line1: "1 25544U 98067A   24200.50000000  .00016717  00000-0  10270-3 0  9024",
		Line2: "2 25544  51.6400 200.0000 0004000  50.0000  80.0000 15.50000000100000",
	}
	start := time.Date(2024, 7, 18, 12, 0, 0, 0, time.UTC)
	end := start.Add(30 * time.Minute)
	samples, err := p.Propagate(tle, start, end, 60*time.Second)
	if err != nil {
		t.Fatalf("propagate: %v", err)
	}
	if len(samples) < 20 {
		t.Fatalf("expected >=20 samples, got %d", len(samples))
	}
	s0 := samples[0]
	if s0.Latitude < -90 || s0.Latitude > 90 {
		t.Fatalf("bad lat %v", s0.Latitude)
	}
	if s0.Longitude < -180 || s0.Longitude > 180 {
		t.Fatalf("bad lon %v", s0.Longitude)
	}
	if s0.AltitudeKM < 100 || s0.AltitudeKM > 2000 {
		t.Fatalf("unexpected altitude %v", s0.AltitudeKM)
	}
}

func TestFindPassesTokyo(t *testing.T) {
	p := NewKeplerPropagator()
	tle := TLE{
		Line1: "1 25544U 98067A   24200.50000000  .00016717  00000-0  10270-3 0  9024",
		Line2: "2 25544  51.6400 200.0000 0004000  50.0000  80.0000 15.50000000100000",
	}
	station := GroundStation{
		Name:            "Tokyo-GS",
		LatitudeDeg:     35.6895,
		LongitudeDeg:    139.6917,
		AltitudeM:       40,
		MinElevationDeg: 5,
	}
	start := time.Date(2024, 7, 18, 0, 0, 0, 0, time.UTC)
	end := start.Add(24 * time.Hour)
	passes, err := p.FindPasses(tle, station, start, end, 60*time.Second)
	if err != nil {
		t.Fatalf("find passes: %v", err)
	}
	// May be zero depending on geometry; ensure call succeeds and windows are ordered.
	for i, pass := range passes {
		if !pass.LOS.After(pass.AOS) {
			t.Fatalf("pass %d: LOS not after AOS", i)
		}
	}
}
