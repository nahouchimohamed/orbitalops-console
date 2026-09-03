package usecase

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"math"
	"sort"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/nahouchimohamed/orbital-ops-console/services/planning-service/internal/domain"
)

type PlanningService struct {
	repo   domain.Repository
	orbit  domain.OrbitPort
	logger *slog.Logger
}

func NewPlanningService(repo domain.Repository, orbit domain.OrbitPort, logger *slog.Logger) *PlanningService {
	return &PlanningService{repo: repo, orbit: orbit, logger: logger}
}

func (s *PlanningService) ListSatellites(ctx context.Context) ([]domain.Satellite, error) {
	return s.repo.ListSatellites(ctx)
}

func (s *PlanningService) ListGroundStations(ctx context.Context) ([]domain.GroundStation, error) {
	return s.repo.ListGroundStations(ctx)
}

func (s *PlanningService) ListRequests(ctx context.Context) ([]domain.ObservationRequest, error) {
	return s.repo.ListObservationRequests(ctx)
}

func (s *PlanningService) ListPlans(ctx context.Context) ([]domain.ImagingPlan, error) {
	return s.repo.ListImagingPlans(ctx)
}

func (s *PlanningService) GroundTrack(ctx context.Context, satID uuid.UUID, hours float64, stepSec int) (*domain.GroundTrack, error) {
	sat, err := s.repo.GetSatellite(ctx, satID)
	if err != nil {
		return nil, err
	}
	if hours <= 0 {
		hours = 1.5
	}
	if hours > 24 {
		hours = 24
	}
	if stepSec <= 0 {
		stepSec = 60
	}
	if stepSec < 15 {
		stepSec = 15
	}
	if stepSec > 300 {
		stepSec = 300
	}
	start := time.Now().UTC()
	end := start.Add(time.Duration(hours * float64(time.Hour)))
	samples, err := s.orbit.Propagate(ctx, sat.TLELine1, sat.TLELine2, sat.Name, start, end, stepSec)
	if err != nil {
		return nil, err
	}
	return &domain.GroundTrack{
		SatelliteID:   sat.ID,
		SatelliteName: sat.Name,
		Samples:       samples,
	}, nil
}

func (s *PlanningService) FindPasses(ctx context.Context, stationID uuid.UUID, hours float64) ([]domain.PassWindow, error) {
	station, err := s.repo.GetGroundStation(ctx, stationID)
	if err != nil {
		return nil, err
	}
	sats, err := s.repo.ListSatellites(ctx)
	if err != nil {
		return nil, err
	}
	if hours <= 0 {
		hours = 24
	}
	if hours > 72 {
		hours = 72
	}
	start := time.Now().UTC()
	end := start.Add(time.Duration(hours * float64(time.Hour)))

	var (
		mu   sync.Mutex
		wg   sync.WaitGroup
		out  []domain.PassWindow
		errCh = make(chan error, len(sats))
	)

	for _, sat := range sats {
		wg.Add(1)
		go func(sat domain.Satellite) {
			defer wg.Done()
			passes, err := s.orbit.FindPasses(ctx, sat.TLELine1, sat.TLELine2, sat.Name, *station, start, end, 30)
			if err != nil {
				errCh <- err
				return
			}
			mu.Lock()
			for _, p := range passes {
				p.SatelliteID = sat.ID
				p.SatelliteName = sat.Name
				p.StationID = station.ID
				p.StationName = station.Name
				p.Quality = passQualityTier(p.MaxElevationDeg)
				out = append(out, p)
			}
			mu.Unlock()
		}(sat)
	}
	wg.Wait()
	close(errCh)
	for err := range errCh {
		if err != nil {
			return nil, err
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].AOS.Before(out[j].AOS) })
	return out, nil
}

type CreateRequestInput struct {
	Name     string          `json:"name"`
	AOI      json.RawMessage `json:"aoi_geojson"`
	Priority int             `json:"priority"`
}

func (s *PlanningService) CreateRequestAndPlan(ctx context.Context, in CreateRequestInput) (*domain.ObservationRequest, []domain.ImagingPlan, error) {
	in.Name = domain.SanitizeRequestName(in.Name)
	if in.Name == "" {
		return nil, nil, fmt.Errorf("name required")
	}
	if len(in.AOI) == 0 {
		return nil, nil, fmt.Errorf("aoi_geojson required")
	}
	in.Priority = domain.ClampPriority(in.Priority)
	if err := domain.ValidateAOIGeoJSON(in.AOI); err != nil {
		return nil, nil, err
	}
	req, err := s.repo.CreateObservationRequest(ctx, domain.ObservationRequest{
		Name:     in.Name,
		AOI:      in.AOI,
		Priority: in.Priority,
		Status:   "pending",
	})
	if err != nil {
		return nil, nil, err
	}

	sats, err := s.repo.ListSatellites(ctx)
	if err != nil {
		return nil, nil, err
	}

	centroid := aoiCentroid(in.AOI)
	start := time.Now().UTC()
	end := start.Add(12 * time.Hour)

	var plans []domain.ImagingPlan
	for _, sat := range sats {
		samples, err := s.orbit.Propagate(ctx, sat.TLELine1, sat.TLELine2, sat.Name, start, end, 60)
		if err != nil {
			return nil, nil, err
		}
		window := nearestOverflight(samples, centroid.lat, centroid.lon)
		if window == nil {
			continue
		}
		cmds := []domain.CommandStep{
			{Time: window.start.Add(-2 * time.Minute).Format(time.RFC3339), Command: "ATTITUDE_SLEW", Args: map[string]any{"look_side": "right"}},
			{Time: window.start.Format(time.RFC3339), Command: "SAR_IMAGING_START", Args: map[string]any{"mode": "stripmap", "request": req.Name}},
			{Time: window.end.Format(time.RFC3339), Command: "SAR_IMAGING_STOP", Args: map[string]any{}},
			{Time: window.end.Add(1 * time.Minute).Format(time.RFC3339), Command: "PAYLOAD_SAFE", Args: map[string]any{}},
		}
		conflict := false
		for _, existing := range plans {
			if existing.SatelliteID == sat.ID &&
				domain.WindowsOverlap(existing.StartTime, existing.EndTime, window.start, window.end) {
				conflict = true
				break
			}
		}
		if conflict {
			continue
		}
		plan, err := s.repo.CreateImagingPlan(ctx, domain.ImagingPlan{
			RequestID:       req.ID,
			SatelliteID:     sat.ID,
			SatelliteName:   sat.Name,
			StartTime:       window.start,
			EndTime:         window.end,
			LookSide:        "right",
			Status:          "draft",
			CommandSequence: cmds,
		})
		if err != nil {
			return nil, nil, err
		}
		plans = append(plans, *plan)
		if len(plans) >= 2 {
			break
		}
	}

	_ = s.repo.UpdateRequestStatus(ctx, req.ID, "planned")
	req.Status = "planned"
	s.logger.Info("observation planned",
		"request", req.Name,
		"priority", req.Priority,
		"plans", len(plans),
		"band", domain.PriorityBand(req.Priority),
	)
	return req, plans, nil
}

type latLon struct{ lat, lon float64 }

type overflight struct{ start, end time.Time }

func aoiCentroid(raw json.RawMessage) latLon {
	var gj map[string]any
	if err := json.Unmarshal(raw, &gj); err != nil {
		return latLon{lat: -6.2, lon: 106.85}
	}
	// Accept Polygon or Feature/FeatureCollection loosely.
	coords := extractCoords(gj)
	if len(coords) == 0 {
		return latLon{lat: -6.2, lon: 106.85}
	}
	var sumLat, sumLon float64
	for _, c := range coords {
		sumLon += c[0]
		sumLat += c[1]
	}
	n := float64(len(coords))
	return latLon{lat: sumLat / n, lon: sumLon / n}
}

func extractCoords(gj map[string]any) [][]float64 {
	switch gj["type"] {
	case "Feature":
		if geom, ok := gj["geometry"].(map[string]any); ok {
			return extractCoords(geom)
		}
	case "FeatureCollection":
		if feats, ok := gj["features"].([]any); ok && len(feats) > 0 {
			if f, ok := feats[0].(map[string]any); ok {
				return extractCoords(f)
			}
		}
	case "Polygon":
		if coords, ok := gj["coordinates"].([]any); ok && len(coords) > 0 {
			if ring, ok := coords[0].([]any); ok {
				return toLonLat(ring)
			}
		}
	case "Point":
		if c, ok := gj["coordinates"].([]any); ok && len(c) >= 2 {
			return [][]float64{{asFloat(c[0]), asFloat(c[1])}}
		}
	}
	return nil
}

func toLonLat(ring []any) [][]float64 {
	out := make([][]float64, 0, len(ring))
	for _, p := range ring {
		arr, ok := p.([]any)
		if !ok || len(arr) < 2 {
			continue
		}
		out = append(out, []float64{asFloat(arr[0]), asFloat(arr[1])})
	}
	return out
}

func asFloat(v any) float64 {
	switch t := v.(type) {
	case float64:
		return t
	case int:
		return float64(t)
	default:
		return 0
	}
}

func nearestOverflight(samples []domain.TrackSample, lat, lon float64) *overflight {
	bestIdx := -1
	bestDist := math.MaxFloat64
	for i, s := range samples {
		d := haversineKM(lat, lon, s.Latitude, s.Longitude)
		if d < bestDist {
			bestDist = d
			bestIdx = i
		}
	}
	if bestIdx < 0 || bestDist > 1500 {
		return nil
	}
	start := samples[bestIdx].Time.Add(-45 * time.Second)
	end := samples[bestIdx].Time.Add(90 * time.Second)
	return &overflight{start: start, end: end}
}

func haversineKM(lat1, lon1, lat2, lon2 float64) float64 {
	const R = 6371.0
	dLat := (lat2 - lat1) * math.Pi / 180
	dLon := (lon2 - lon1) * math.Pi / 180
	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(lat1*math.Pi/180)*math.Cos(lat2*math.Pi/180)*math.Sin(dLon/2)*math.Sin(dLon/2)
	return 2 * R * math.Asin(math.Sqrt(a))
}

func passQualityTier(maxElDeg float64) string {
	switch {
	case maxElDeg >= 40:
		return "excellent"
	case maxElDeg >= 20:
		return "good"
	case maxElDeg >= 10:
		return "fair"
	default:
		return "marginal"
	}
}
