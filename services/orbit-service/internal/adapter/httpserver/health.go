package httpserver

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"time"
)

func NewMux(logger *slog.Logger) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, map[string]any{"status": "ok", "service": "orbit-service", "ts": time.Now().UTC()})
	})
	mux.HandleFunc("/readyz", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, map[string]any{"status": "ready"})
	})
	return requestIDMiddleware(logger, mux)
}

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(v)
}

func requestIDMiddleware(logger *slog.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.Header.Get("X-Request-Id")
		if id == "" {
			id = time.Now().UTC().Format("20060102T150405.000000000")
		}
		w.Header().Set("X-Request-Id", id)
		logger.Debug("http", "path", r.URL.Path, "request_id", id)
		next.ServeHTTP(w, r)
	})
}
