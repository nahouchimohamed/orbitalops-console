package grpcserver

import (
	"context"
	"time"

	"github.com/nahouchimohamed/orbital-ops-console/services/orbit-service/internal/domain"
	orbitv1 "github.com/nahouchimohamed/orbital-ops-console/services/orbit-service/api/orbit/v1"
	"github.com/nahouchimohamed/orbital-ops-console/services/orbit-service/internal/usecase"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Server struct {
	orbitv1.UnimplementedOrbitServiceServer
	svc *usecase.OrbitService
}

func New(svc *usecase.OrbitService) *Server {
	return &Server{svc: svc}
}

func (s *Server) Propagate(ctx context.Context, req *orbitv1.PropagateRequest) (*orbitv1.PropagateResponse, error) {
	if req.GetTle() == nil {
		return nil, status.Error(codes.InvalidArgument, "tle required")
	}
	start, end, err := parseWindow(req.GetStartTimeRfc3339(), req.GetEndTimeRfc3339())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "time window: %v", err)
	}
	tle := domain.TLE{
		Name:  req.Tle.GetName(),
		Line1: req.Tle.GetLine1(),
		Line2: req.Tle.GetLine2(),
	}
	samples, err := s.svc.Propagate(ctx, tle, start, end, int(req.GetStepSeconds()))
	if err != nil {
		return nil, status.Errorf(codes.Internal, "%v", err)
	}
	out := &orbitv1.PropagateResponse{}
	for _, sample := range samples {
		out.Samples = append(out.Samples, &orbitv1.GeodeticSample{
			TimeRfc3339:  sample.Time.UTC().Format(time.RFC3339),
			LatitudeDeg:  sample.Latitude,
			LongitudeDeg: sample.Longitude,
			AltitudeKm:   sample.AltitudeKM,
		})
	}
	return out, nil
}

func (s *Server) FindPasses(ctx context.Context, req *orbitv1.FindPassesRequest) (*orbitv1.FindPassesResponse, error) {
	if req.GetTle() == nil || req.GetStation() == nil {
		return nil, status.Error(codes.InvalidArgument, "tle and station required")
	}
	start, end, err := parseWindow(req.GetStartTimeRfc3339(), req.GetEndTimeRfc3339())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "time window: %v", err)
	}
	tle := domain.TLE{Name: req.Tle.Name, Line1: req.Tle.Line1, Line2: req.Tle.Line2}
	st := domain.GroundStation{
		Name:            req.Station.Name,
		LatitudeDeg:     req.Station.LatitudeDeg,
		LongitudeDeg:    req.Station.LongitudeDeg,
		AltitudeM:       req.Station.AltitudeM,
		MinElevationDeg: req.Station.MinElevationDeg,
	}
	passes, err := s.svc.FindPasses(ctx, tle, st, start, end, int(req.GetStepSeconds()))
	if err != nil {
		return nil, status.Errorf(codes.Internal, "%v", err)
	}
	out := &orbitv1.FindPassesResponse{}
	for _, p := range passes {
		out.Passes = append(out.Passes, &orbitv1.PassWindow{
			AosRfc3339:               p.AOS.UTC().Format(time.RFC3339),
			LosRfc3339:               p.LOS.UTC().Format(time.RFC3339),
			MaxElevationDeg:          p.MaxElevationDeg,
			MaxElevationTimeRfc3339:  p.MaxElevationTime.UTC().Format(time.RFC3339),
		})
	}
	return out, nil
}

func parseWindow(startS, endS string) (time.Time, time.Time, error) {
	start, err := time.Parse(time.RFC3339, startS)
	if err != nil {
		return time.Time{}, time.Time{}, err
	}
	end, err := time.Parse(time.RFC3339, endS)
	if err != nil {
		return time.Time{}, time.Time{}, err
	}
	return start, end, nil
}
