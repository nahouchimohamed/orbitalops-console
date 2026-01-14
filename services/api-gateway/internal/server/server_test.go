package server

import (
	"log/slog"
	"net"
	"os"
	"testing"

	"google.golang.org/grpc"
)

func TestNewGatewayWithFakeOrbit(t *testing.T) {
	lis, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	gs := grpc.NewServer()
	go gs.Serve(lis)
	defer gs.Stop()

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	h, cleanup, err := New(Config{
		HTTPAddr:    ":0",
		PlanningURL: "http://127.0.0.1:9",
		OrbitGRPC:   lis.Addr().String(),
		CORSOrigins: []string{"http://localhost:5173"},
	}, logger)
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup()
	if h == nil {
		t.Fatal("expected handler")
	}
}
