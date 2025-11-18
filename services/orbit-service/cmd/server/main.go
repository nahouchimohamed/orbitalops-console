package main

import (
	"context"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/nahouchimohamed/orbital-ops-console/services/orbit-service/internal/adapter/grpcserver"
	"github.com/nahouchimohamed/orbital-ops-console/services/orbit-service/internal/adapter/httpserver"
	"github.com/nahouchimohamed/orbital-ops-console/services/orbit-service/internal/domain"
	orbitv1 "github.com/nahouchimohamed/orbital-ops-console/services/orbit-service/api/orbit/v1"
	"github.com/nahouchimohamed/orbital-ops-console/services/orbit-service/internal/usecase"
	"google.golang.org/grpc"
)

func main() {
	level := slog.LevelInfo
	if os.Getenv("LOG_LEVEL") == "debug" {
		level = slog.LevelDebug
	}
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: level}))

	propagator := domain.NewKeplerPropagator()
	svc := usecase.NewOrbitService(propagator, logger)

	grpcAddr := envOr("ORBIT_GRPC_ADDR", ":50051")
	httpAddr := envOr("ORBIT_HTTP_ADDR", ":8082")

	grpcServer := grpc.NewServer()
	orbitv1.RegisterOrbitServiceServer(grpcServer, grpcserver.New(svc))

	lis, err := net.Listen("tcp", grpcAddr)
	if err != nil {
		logger.Error("failed to listen grpc", "err", err)
		os.Exit(1)
	}

	httpMux := httpserver.NewMux(logger)
	httpServer := &http.Server{Addr: httpAddr, Handler: httpMux}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go func() {
		logger.Info("orbit-service gRPC listening", "addr", grpcAddr)
		if err := grpcServer.Serve(lis); err != nil {
			logger.Error("grpc serve error", "err", err)
			stop()
		}
	}()

	go func() {
		logger.Info("orbit-service HTTP health listening", "addr", httpAddr)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("http serve error", "err", err)
			stop()
		}
	}()

	<-ctx.Done()
	logger.Info("shutting down")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_ = httpServer.Shutdown(shutdownCtx)
	grpcServer.GracefulStop()
}

func envOr(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}
