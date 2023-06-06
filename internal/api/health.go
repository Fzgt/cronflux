package api

import (
	"net/http"

	"github.com/Fzgt/cronflux/internal/buildinfo"
)

// registerHealthRoutes wires the liveness and readiness probes onto the mux.
func (s *Server) registerHealthRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /healthz", s.handleHealth)
	mux.HandleFunc("GET /readyz", s.handleReady)
}

// handleHealth is a liveness probe: it returns as long as the process is up.
func (s *Server) handleHealth(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"status":  "ok",
		"version": buildinfo.Version,
	})
}

// handleReady is a readiness probe: it reports unavailable until the store can
// be reached.
func (s *Server) handleReady(w http.ResponseWriter, r *http.Request) {
	if _, err := s.store.ListJobs(r.Context()); err != nil {
		writeError(w, http.StatusServiceUnavailable, "store unavailable: "+err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ready"})
}
