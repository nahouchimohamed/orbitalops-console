package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/nahouchimohamed/orbital-ops-console/services/api-gateway/internal/server"
)

func main() {
	level := slog.LevelInfo
	if os.Getenv("LOG_LEVEL") == "debug" {
		level = slog.LevelDebug
	}
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: level}))

	cfg := server.Config{
		HTTPAddr:     envOr("HTTP_ADDR", ":8080"),
		PlanningURL:  strings.TrimRight(envOr("PLANNING_URL", "http://localhost:8081"), "/"),
		OrbitGRPC:    envOr("ORBIT_GRPC_ADDR", "localhost:50051"),
		CORSOrigins:  strings.Split(envOr("CORS_ORIGINS", "http://localhost:5173"), ","),
	}

	h, cleanup, err := server.New(cfg, logger)
	if err != nil {
		logger.Error("gateway init failed", "err", err)
		os.Exit(1)
	}
	defer cleanup()

	srv := &http.Server{Addr: cfg.HTTPAddr, Handler: h}
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go func() {
		logger.Info("api-gateway listening", "addr", cfg.HTTPAddr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("http error", "err", err)
			stop()
		}
	}()

	<-ctx.Done()
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_ = srv.Shutdown(shutdownCtx)
}

func envOr(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}
