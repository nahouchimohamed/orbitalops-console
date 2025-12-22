package httpapi

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/google/uuid"
	"github.com/nahouchimohamed/orbital-ops-console/services/planning-service/internal/domain"
	"github.com/nahouchimohamed/orbital-ops-console/services/planning-service/internal/usecase"
)

type Handler struct {
	svc    *usecase.PlanningService
	logger *slog.Logger
}

func New(svc *usecase.PlanningService, logger *slog.Logger) http.Handler {
	h := &Handler{svc: svc, logger: logger}
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(30 * time.Second))
	r.Use(h.logRequests)

	r.Get("/healthz", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]any{"status": "ok", "service": "planning-service"})
	})
	r.Get("/readyz", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]any{"status": "ready"})
	})

	r.Get("/api/v1/satellites", h.listSatellites)
	r.Get("/api/v1/ground-stations", h.listStations)
	r.Get("/api/v1/satellites/{id}/ground-track", h.groundTrack)
	r.Get("/api/v1/ground-stations/{id}/passes", h.passes)
	r.Get("/api/v1/requests", h.listRequests)
	r.Post("/api/v1/requests", h.createRequest)
	r.Get("/api/v1/plans", h.listPlans)
	return r
}

func (h *Handler) logRequests(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
		next.ServeHTTP(ww, r)
		h.logger.Info("request",
			"method", r.Method,
			"path", r.URL.Path,
			"status", ww.Status(),
			"duration_ms", time.Since(start).Milliseconds(),
			"request_id", middleware.GetReqID(r.Context()),
		)
	})
}

func (h *Handler) listSatellites(w http.ResponseWriter, r *http.Request) {
	items, err := h.svc.ListSatellites(r.Context())
	if err != nil {
		writeErr(w, err)
		return
	}
	if items == nil {
		items = []domain.Satellite{}
	}
	writeJSON(w, http.StatusOK, items)
}

func (h *Handler) listStations(w http.ResponseWriter, r *http.Request) {
	items, err := h.svc.ListGroundStations(r.Context())
	if err != nil {
		writeErr(w, err)
		return
	}
	if items == nil {
		items = []domain.GroundStation{}
	}
	writeJSON(w, http.StatusOK, items)
}

func (h *Handler) groundTrack(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid id"})
		return
	}
	hours, _ := strconv.ParseFloat(r.URL.Query().Get("hours"), 64)
	step, _ := strconv.Atoi(r.URL.Query().Get("step_seconds"))
	track, err := h.svc.GroundTrack(r.Context(), id, hours, step)
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, track)
}

func (h *Handler) passes(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid id"})
		return
	}
	hours, _ := strconv.ParseFloat(r.URL.Query().Get("hours"), 64)
	passes, err := h.svc.FindPasses(r.Context(), id, hours)
	if err != nil {
		writeErr(w, err)
		return
	}
	if passes == nil {
		passes = []domain.PassWindow{}
	}
	writeJSON(w, http.StatusOK, passes)
}

func (h *Handler) listRequests(w http.ResponseWriter, r *http.Request) {
	items, err := h.svc.ListRequests(r.Context())
	if err != nil {
		writeErr(w, err)
		return
	}
	if items == nil {
		items = []domain.ObservationRequest{}
	}
	writeJSON(w, http.StatusOK, items)
}

func (h *Handler) createRequest(w http.ResponseWriter, r *http.Request) {
	var in usecase.CreateRequestInput
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid json"})
		return
	}
	req, plans, err := h.svc.CreateRequestAndPlan(r.Context(), in)
	if err != nil {
		writeErr(w, err)
		return
	}
	if plans == nil {
		plans = []domain.ImagingPlan{}
	}
	writeJSON(w, http.StatusCreated, map[string]any{"request": req, "plans": plans})
}

func (h *Handler) listPlans(w http.ResponseWriter, r *http.Request) {
	items, err := h.svc.ListPlans(r.Context())
	if err != nil {
		writeErr(w, err)
		return
	}
	if items == nil {
		items = []domain.ImagingPlan{}
	}
	writeJSON(w, http.StatusOK, items)
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeErr(w http.ResponseWriter, err error) {
	writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
}
