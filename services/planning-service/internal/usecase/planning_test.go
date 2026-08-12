package usecase

import (
	"encoding/json"
	"testing"
)

func TestAOICentroidPolygon(t *testing.T) {
	raw := json.RawMessage(`{
		"type":"Polygon",
		"coordinates":[[[106.7,-6.35],[106.95,-6.35],[106.95,-6.1],[106.7,-6.1],[106.7,-6.35]]]
	}`)
	c := aoiCentroid(raw)
	if c.lat < -6.35 || c.lat > -6.1 {
		t.Fatalf("unexpected lat %v", c.lat)
	}
	if c.lon < 106.7 || c.lon > 106.95 {
		t.Fatalf("unexpected lon %v", c.lon)
	}
}

func TestHaversineZero(t *testing.T) {
	if d := haversineKM(-6.2, 106.8, -6.2, 106.8); d != 0 {
		t.Fatalf("expected 0 got %v", d)
	}
}
