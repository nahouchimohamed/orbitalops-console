package usecase

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/nahouchimohamed/orbital-ops-console/services/orbit-service/internal/domain"
)

type OrbitService struct {
	prop   domain.Propagator
	logger *slog.Logger
}

func NewOrbitService(prop domain.Propagator, logger *slog.Logger) *OrbitService {
	return &OrbitService{prop: prop, logger: logger}
}

func (s *OrbitService) Propagate(ctx context.Context, tle domain.TLE, start, end time.Time, stepSec int) ([]domain.GeodeticSample, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if stepSec <= 0 {
		stepSec = 60
	}
	step := time.Duration(stepSec) * time.Second
	samples, err := s.prop.Propagate(tle, start, end, step)
	if err != nil {
		return nil, fmt.Errorf("propagate: %w", err)
	}
	return samples, nil
}

func (s *OrbitService) FindPasses(ctx context.Context, tle domain.TLE, station domain.GroundStation, start, end time.Time, stepSec int) ([]domain.PassWindow, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if stepSec <= 0 {
		stepSec = 30
	}
	return s.prop.FindPasses(tle, station, start, end, time.Duration(stepSec)*time.Second)
}

// PropagateMany propagates multiple TLEs concurrently (constellation scale).
func (s *OrbitService) PropagateMany(ctx context.Context, tles []domain.TLE, start, end time.Time, stepSec int) (map[string][]domain.GeodeticSample, error) {
	out := make(map[string][]domain.GeodeticSample, len(tles))
	var mu sync.Mutex
	var wg sync.WaitGroup
	errCh := make(chan error, len(tles))

	for _, tle := range tles {
		wg.Add(1)
		go func(tle domain.TLE) {
			defer wg.Done()
			select {
			case <-ctx.Done():
				errCh <- ctx.Err()
				return
			default:
			}
			samples, err := s.Propagate(ctx, tle, start, end, stepSec)
			if err != nil {
				errCh <- err
				return
			}
			key := tle.Name
			if key == "" {
				key = tle.Line1
			}
			mu.Lock()
			out[key] = samples
			mu.Unlock()
		}(tle)
	}
	wg.Wait()
	close(errCh)
	for err := range errCh {
		if err != nil {
			return nil, err
		}
	}
	return out, nil
}
