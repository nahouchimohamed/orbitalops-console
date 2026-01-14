package server

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httputil"
	"net/url"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	orbitv1 "github.com/nahouchimohamed/orbital-ops-console/services/orbit-service/api/orbit/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Config struct {
	HTTPAddr    string
	PlanningURL string
	OrbitGRPC   string
	CORSOrigins []string
}

func New(cfg Config, logger *slog.Logger) (http.Handler, func(), error) {
	planningURL, err := url.Parse(cfg.PlanningURL)
	if err != nil {
		return nil, nil, err
	}
	proxy := httputil.NewSingleHostReverseProxy(planningURL)

	conn, err := grpc.NewClient(cfg.OrbitGRPC, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, nil, err
	}
	orbit := orbitv1.NewOrbitServiceClient(conn)

	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(45 * time.Second))
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   cfg.CORSOrigins,
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-Request-Id"},
		ExposedHeaders:   []string{"X-Request-Id"},
		AllowCredentials: false,
		MaxAge:           300,
	}))

	r.Get("/healthz", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]any{
			"status":  "ok",
			"service": "api-gateway",
			"deps": map[string]string{
				"planning": cfg.PlanningURL,
				"orbit":    cfg.OrbitGRPC,
			},
		})
	})

	r.Get("/readyz", func(w http.ResponseWriter, r *http.Request) {
		// Best-effort planning health probe
		client := &http.Client{Timeout: 2 * time.Second}
		resp, err := client.Get(cfg.PlanningURL + "/readyz")
		if err != nil {
			writeJSON(w, http.StatusServiceUnavailable, map[string]string{"status": "not_ready", "error": err.Error()})
			return
		}
		defer resp.Body.Close()
		writeJSON(w, http.StatusOK, map[string]string{"status": "ready"})
	})

	// Aggregate health for ops dashboard
	r.Get("/api/v1/system/status", func(w http.ResponseWriter, r *http.Request) {
		client := &http.Client{Timeout: 2 * time.Second}
		planningOK := false
		if resp, err := client.Get(cfg.PlanningURL + "/healthz"); err == nil {
			planningOK = resp.StatusCode == 200
			_ = resp.Body.Close()
		}
		writeJSON(w, http.StatusOK, map[string]any{
			"gateway":  "ok",
			"planning": planningOK,
			"orbit":    conn.GetState().String(),
			"time":     time.Now().UTC(),
		})
	})

	// Proxy planning REST under /api/v1/*
	r.Handle("/api/v1/*", http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		logger.Debug("proxy planning", "path", req.URL.Path, "request_id", middleware.GetReqID(req.Context()))
		proxy.ServeHTTP(w, req)
	}))

	// Direct orbit echo for debugging (optional)
	r.Get("/api/v1/orbit/ping", func(w http.ResponseWriter, r *http.Request) {
		_ = orbit
		writeJSON(w, http.StatusOK, map[string]string{"orbit": "reachable", "state": conn.GetState().String()})
	})

	cleanup := func() { _ = conn.Close() }
	return r, cleanup, nil
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

// Drain helper kept for future BFF aggregation endpoints.
func drain(body io.ReadCloser) {
	_, _ = io.Copy(io.Discard, body)
	_ = body.Close()
}
