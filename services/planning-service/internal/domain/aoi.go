package domain

import (
	"encoding/json"
	"fmt"
)

func ValidateAOIGeoJSON(raw json.RawMessage) error {
	var gj map[string]any
	if err := json.Unmarshal(raw, &gj); err != nil {
		return fmt.Errorf("invalid aoi geojson")
	}
	coords := extractLonLat(gj)
	if len(coords) < 3 {
		return fmt.Errorf("AOI needs a polygon ring")
	}
	minLon, maxLon := coords[0][0], coords[0][0]
	minLat, maxLat := coords[0][1], coords[0][1]
	for _, c := range coords {
		if c[0] < minLon {
			minLon = c[0]
		}
		if c[0] > maxLon {
			maxLon = c[0]
		}
		if c[1] < minLat {
			minLat = c[1]
		}
		if c[1] > maxLat {
			maxLat = c[1]
		}
	}
	if maxLon-minLon > 40 || maxLat-minLat > 40 {
		return fmt.Errorf("AOI too large for single strip planning")
	}
	return nil
}

func extractLonLat(gj map[string]any) [][]float64 {
	switch gj["type"] {
	case "Feature":
		if geom, ok := gj["geometry"].(map[string]any); ok {
			return extractLonLat(geom)
		}
	case "Polygon":
		if coords, ok := gj["coordinates"].([]any); ok && len(coords) > 0 {
			if ring, ok := coords[0].([]any); ok {
				out := make([][]float64, 0, len(ring))
				for _, p := range ring {
					arr, ok := p.([]any)
					if !ok || len(arr) < 2 {
						continue
					}
					lon, _ := arr[0].(float64)
					lat, _ := arr[1].(float64)
					out = append(out, []float64{lon, lat})
				}
				return out
			}
		}
	}
	return nil
}
