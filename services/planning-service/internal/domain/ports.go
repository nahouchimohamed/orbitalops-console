package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type Repository interface {
	ListSatellites(ctx context.Context) ([]Satellite, error)
	GetSatellite(ctx context.Context, id uuid.UUID) (*Satellite, error)
	ListGroundStations(ctx context.Context) ([]GroundStation, error)
	GetGroundStation(ctx context.Context, id uuid.UUID) (*GroundStation, error)
	CreateObservationRequest(ctx context.Context, req ObservationRequest) (*ObservationRequest, error)
	ListObservationRequests(ctx context.Context) ([]ObservationRequest, error)
	GetObservationRequest(ctx context.Context, id uuid.UUID) (*ObservationRequest, error)
	CreateImagingPlan(ctx context.Context, plan ImagingPlan) (*ImagingPlan, error)
	ListImagingPlans(ctx context.Context) ([]ImagingPlan, error)
	UpdateRequestStatus(ctx context.Context, id uuid.UUID, status string) error
}

type OrbitPort interface {
	Propagate(ctx context.Context, line1, line2, name string, start, end time.Time, stepSec int) ([]TrackSample, error)
	FindPasses(ctx context.Context, line1, line2, name string, station GroundStation, start, end time.Time, stepSec int) ([]PassWindow, error)
}
