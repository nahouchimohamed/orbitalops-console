package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/nahouchimohamed/orbital-ops-console/services/planning-service/internal/adapter/httpapi"
	"github.com/nahouchimohamed/orbital-ops-console/services/planning-service/internal/adapter/orbitclient"
	"github.com/nahouchimohamed/orbital-ops-console/services/planning-service/internal/adapter/postgres"
	"github.com/nahouchimohamed/orbital-ops-console/services/planning-service/internal/usecase"
)

func main() {
	level := slog.LevelInfo
	if os.Getenv("LOG_LEVEL") == "debug" {
		level = slog.LevelDebug
	}
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: level}))

	dbURL := envOr("DATABASE_URL", "postgres://orbital:orbital@localhost:5432/orbitalops?sslmode=disable")
	orbitAddr := envOr("ORBIT_GRPC_ADDR", "localhost:50051")
	httpAddr := envOr("HTTP_ADDR", ":8081")

	ctx := context.Background()
	pool, err := postgres.NewPool(ctx, dbURL)
	if err != nil {
		logger.Error("db connect failed", "err", err)
		os.Exit(1)
	}
	defer pool.Close()

	orbit, err := orbitclient.New(orbitAddr)
	if err != nil {
		logger.Error("orbit client failed", "err", err)
		os.Exit(1)
	}
	defer orbit.Close()

	repo := postgres.NewRepository(pool)
	svc := usecase.NewPlanningService(repo, orbit, logger)
	handler := httpapi.New(svc, logger)

	server := &http.Server{Addr: httpAddr, Handler: handler}

	runCtx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go func() {
		logger.Info("planning-service listening", "addr", httpAddr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("http error", "err", err)
			stop()
		}
	}()

	<-runCtx.Done()
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_ = server.Shutdown(shutdownCtx)
}

func envOr(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}
