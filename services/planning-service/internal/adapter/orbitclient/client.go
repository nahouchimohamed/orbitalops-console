package orbitclient

import (
	"context"
	"time"

	"github.com/nahouchimohamed/orbital-ops-console/services/planning-service/internal/domain"
	orbitv1 "github.com/nahouchimohamed/orbital-ops-console/services/orbit-service/api/orbit/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Client struct {
	conn   *grpc.ClientConn
	client orbitv1.OrbitServiceClient
}

func New(addr string) (*Client, error) {
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}
	return &Client{conn: conn, client: orbitv1.NewOrbitServiceClient(conn)}, nil
}

func (c *Client) Close() error {
	return c.conn.Close()
}

func (c *Client) Propagate(ctx context.Context, line1, line2, name string, start, end time.Time, stepSec int) ([]domain.TrackSample, error) {
	ctx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()
	resp, err := c.client.Propagate(ctx, &orbitv1.PropagateRequest{
		Tle:               &orbitv1.TLE{Line1: line1, Line2: line2, Name: name},
		StartTimeRfc3339:  start.UTC().Format(time.RFC3339),
		EndTimeRfc3339:    end.UTC().Format(time.RFC3339),
		StepSeconds:       int32(stepSec),
	})
	if err != nil {
		return nil, err
	}
	out := make([]domain.TrackSample, 0, len(resp.Samples))
	for _, s := range resp.Samples {
		ts, _ := time.Parse(time.RFC3339, s.TimeRfc3339)
		out = append(out, domain.TrackSample{
			Time:       ts,
			Latitude:   s.LatitudeDeg,
			Longitude:  s.LongitudeDeg,
			AltitudeKM: s.AltitudeKm,
		})
	}
	return out, nil
}

func (c *Client) FindPasses(ctx context.Context, line1, line2, name string, station domain.GroundStation, start, end time.Time, stepSec int) ([]domain.PassWindow, error) {
	ctx, cancel := context.WithTimeout(ctx, 20*time.Second)
	defer cancel()
	resp, err := c.client.FindPasses(ctx, &orbitv1.FindPassesRequest{
		Tle: &orbitv1.TLE{Line1: line1, Line2: line2, Name: name},
		Station: &orbitv1.GroundStation{
			Name:            station.Name,
			LatitudeDeg:     station.Latitude,
			LongitudeDeg:    station.Longitude,
			AltitudeM:       station.AltitudeM,
			MinElevationDeg: station.MinElevationDeg,
		},
		StartTimeRfc3339: start.UTC().Format(time.RFC3339),
		EndTimeRfc3339:   end.UTC().Format(time.RFC3339),
		StepSeconds:      int32(stepSec),
	})
	if err != nil {
		return nil, err
	}
	out := make([]domain.PassWindow, 0, len(resp.Passes))
	for _, p := range resp.Passes {
		aos, _ := time.Parse(time.RFC3339, p.AosRfc3339)
		los, _ := time.Parse(time.RFC3339, p.LosRfc3339)
		maxAt, _ := time.Parse(time.RFC3339, p.MaxElevationTimeRfc3339)
		out = append(out, domain.PassWindow{
			AOS:             aos,
			LOS:             los,
			MaxElevationDeg: p.MaxElevationDeg,
			MaxElevationAt:  maxAt,
		})
	}
	return out, nil
}
