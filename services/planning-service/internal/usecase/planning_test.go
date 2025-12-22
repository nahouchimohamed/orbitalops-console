package usecase

import (
	"encoding/json"
	"testing"
)

func TestAOICentroidPolygon(t *testing.T) {
	raw := json.RawMessage(`{
		"type":"Polygon",
		"coordinates":[[[139.6,35.6],[139.8,35.6],[139.8,35.8],[139.6,35.8],[139.6,35.6]]]
	}`)
	c := aoiCentroid(raw)
	if c.lat < 35.6 || c.lat > 35.8 {
		t.Fatalf("unexpected lat %v", c.lat)
	}
	if c.lon < 139.6 || c.lon > 139.8 {
		t.Fatalf("unexpected lon %v", c.lon)
	}
}

func TestHaversineZero(t *testing.T) {
	if d := haversineKM(35, 139, 35, 139); d != 0 {
		t.Fatalf("expected 0 got %v", d)
	}
}
