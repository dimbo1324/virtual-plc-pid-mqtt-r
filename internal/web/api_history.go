package web

import (
	"net/http"

	"github.com/dimbo1324/virtual-plc-pid-mqtt-r/internal/storage"
)

func (s *Server) handleRecentEvents(w http.ResponseWriter, r *http.Request) {
	if s.deps.Store == nil {
		jsonError(w, http.StatusServiceUnavailable, "storage not available")
		return
	}
	limit := queryInt(r, "limit", 100, 1, 500)
	events, err := s.deps.Store.RecentEvents(r.Context(), limit)
	if err != nil {
		jsonError(w, http.StatusInternalServerError, "failed to fetch events")
		return
	}
	if events == nil {
		events = []storage.EventRecord{}
	}
	jsonOK(w, events)
}

func (s *Server) handleRecentTelemetry(w http.ResponseWriter, r *http.Request) {
	if s.deps.Store == nil {
		jsonError(w, http.StatusServiceUnavailable, "storage not available")
		return
	}
	loop := r.URL.Query().Get("loop")
	if loop == "" {
		jsonError(w, http.StatusBadRequest, "loop query parameter is required")
		return
	}
	limit := queryInt(r, "limit", 300, 1, 1000)
	samples, err := s.deps.Store.RecentTelemetry(r.Context(), loop, limit)
	if err != nil {
		jsonError(w, http.StatusInternalServerError, "failed to fetch telemetry")
		return
	}
	if samples == nil {
		samples = []storage.TelemetrySample{}
	}
	jsonOK(w, samples)
}
