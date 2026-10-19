package domain

import "testing"

func TestValidateAOIGeoJSONOK(t *testing.T) {
	raw := []byte(`{"type":"Polygon","coordinates":[[[106.7,-6.35],[106.95,-6.35],[106.95,-6.1],[106.7,-6.1],[106.7,-6.35]]]}`)
	if err := ValidateAOIGeoJSON(raw); err != nil {
		t.Fatal(err)
	}
}

func TestValidateAOIGeoJSONTooLarge(t *testing.T) {
	raw := []byte(`{"type":"Polygon","coordinates":[[[0,0],[50,0],[50,10],[0,10],[0,0]]]}`)
	if err := ValidateAOIGeoJSON(raw); err == nil {
		t.Fatal("expected error")
	}
}
