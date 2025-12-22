package postgres

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nahouchimohamed/orbital-ops-console/services/planning-service/internal/domain"
)

type Repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

func (r *Repository) ListSatellites(ctx context.Context) ([]domain.Satellite, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, name, norad_id, tle_line1, tle_line2, created_at
		FROM satellites ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.Satellite
	for rows.Next() {
		var s domain.Satellite
		if err := rows.Scan(&s.ID, &s.Name, &s.NoradID, &s.TLELine1, &s.TLELine2, &s.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, s)
	}
	return out, rows.Err()
}

func (r *Repository) GetSatellite(ctx context.Context, id uuid.UUID) (*domain.Satellite, error) {
	var s domain.Satellite
	err := r.pool.QueryRow(ctx, `
		SELECT id, name, norad_id, tle_line1, tle_line2, created_at
		FROM satellites WHERE id=$1`, id).
		Scan(&s.ID, &s.Name, &s.NoradID, &s.TLELine1, &s.TLELine2, &s.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *Repository) ListGroundStations(ctx context.Context) ([]domain.GroundStation, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, name, latitude, longitude, altitude_m, min_elevation_deg, created_at
		FROM ground_stations ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.GroundStation
	for rows.Next() {
		var g domain.GroundStation
		if err := rows.Scan(&g.ID, &g.Name, &g.Latitude, &g.Longitude, &g.AltitudeM, &g.MinElevationDeg, &g.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, g)
	}
	return out, rows.Err()
}

func (r *Repository) GetGroundStation(ctx context.Context, id uuid.UUID) (*domain.GroundStation, error) {
	var g domain.GroundStation
	err := r.pool.QueryRow(ctx, `
		SELECT id, name, latitude, longitude, altitude_m, min_elevation_deg, created_at
		FROM ground_stations WHERE id=$1`, id).
		Scan(&g.ID, &g.Name, &g.Latitude, &g.Longitude, &g.AltitudeM, &g.MinElevationDeg, &g.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &g, nil
}

func (r *Repository) CreateObservationRequest(ctx context.Context, req domain.ObservationRequest) (*domain.ObservationRequest, error) {
	if req.ID == uuid.Nil {
		req.ID = uuid.New()
	}
	if req.Status == "" {
		req.Status = "pending"
	}
	if req.Priority == 0 {
		req.Priority = 1
	}
	err := r.pool.QueryRow(ctx, `
		INSERT INTO observation_requests (id, name, aoi_geojson, priority, status)
		VALUES ($1,$2,$3,$4,$5)
		RETURNING created_at`,
		req.ID, req.Name, req.AOI, req.Priority, req.Status,
	).Scan(&req.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &req, nil
}

func (r *Repository) ListObservationRequests(ctx context.Context) ([]domain.ObservationRequest, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, name, aoi_geojson, priority, status, created_at
		FROM observation_requests ORDER BY created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.ObservationRequest
	for rows.Next() {
		var o domain.ObservationRequest
		if err := rows.Scan(&o.ID, &o.Name, &o.AOI, &o.Priority, &o.Status, &o.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, o)
	}
	return out, rows.Err()
}

func (r *Repository) GetObservationRequest(ctx context.Context, id uuid.UUID) (*domain.ObservationRequest, error) {
	var o domain.ObservationRequest
	err := r.pool.QueryRow(ctx, `
		SELECT id, name, aoi_geojson, priority, status, created_at
		FROM observation_requests WHERE id=$1`, id).
		Scan(&o.ID, &o.Name, &o.AOI, &o.Priority, &o.Status, &o.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &o, nil
}

func (r *Repository) CreateImagingPlan(ctx context.Context, plan domain.ImagingPlan) (*domain.ImagingPlan, error) {
	if plan.ID == uuid.Nil {
		plan.ID = uuid.New()
	}
	if plan.Status == "" {
		plan.Status = "draft"
	}
	if plan.LookSide == "" {
		plan.LookSide = "right"
	}
	cmdJSON, err := json.Marshal(plan.CommandSequence)
	if err != nil {
		return nil, err
	}
	err = r.pool.QueryRow(ctx, `
		INSERT INTO imaging_plans (id, request_id, satellite_id, start_time, end_time, look_side, status, command_sequence)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8)
		RETURNING created_at`,
		plan.ID, plan.RequestID, plan.SatelliteID, plan.StartTime, plan.EndTime, plan.LookSide, plan.Status, cmdJSON,
	).Scan(&plan.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &plan, nil
}

func (r *Repository) ListImagingPlans(ctx context.Context) ([]domain.ImagingPlan, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT p.id, p.request_id, p.satellite_id, s.name, p.start_time, p.end_time, p.look_side, p.status, p.command_sequence, p.created_at
		FROM imaging_plans p
		JOIN satellites s ON s.id = p.satellite_id
		ORDER BY p.start_time`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.ImagingPlan
	for rows.Next() {
		var p domain.ImagingPlan
		var cmdJSON []byte
		if err := rows.Scan(&p.ID, &p.RequestID, &p.SatelliteID, &p.SatelliteName, &p.StartTime, &p.EndTime, &p.LookSide, &p.Status, &cmdJSON, &p.CreatedAt); err != nil {
			return nil, err
		}
		_ = json.Unmarshal(cmdJSON, &p.CommandSequence)
		out = append(out, p)
	}
	return out, rows.Err()
}

func (r *Repository) UpdateRequestStatus(ctx context.Context, id uuid.UUID, status string) error {
	_, err := r.pool.Exec(ctx, `UPDATE observation_requests SET status=$2 WHERE id=$1`, id, status)
	return err
}
