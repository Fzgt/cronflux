package api

import (
	"net/http"
	"strconv"

	"github.com/Fzgt/cronflux/internal/job"
	"github.com/Fzgt/cronflux/internal/store"
)

// registerRunRoutes wires the /api/runs endpoints onto the mux.
func (s *Server) registerRunRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/runs", s.handleListRuns)
	mux.HandleFunc("GET /api/runs/{id}", s.handleGetRun)
}

func (s *Server) handleListRuns(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	f := store.RunFilter{
		JobID:   q.Get("job"),
		BatchID: q.Get("batch"),
		State:   job.RunState(q.Get("state")),
		Limit:   atoiDefault(q.Get("limit"), 100),
		Offset:  atoiDefault(q.Get("offset"), 0),
	}
	runs, err := s.store.ListRuns(r.Context(), f)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, runs)
}

func (s *Server) handleGetRun(w http.ResponseWriter, r *http.Request) {
	run, err := s.store.GetRun(r.Context(), r.PathValue("id"))
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, run)
}

// atoiDefault parses s as an int, falling back to def on any problem.
func atoiDefault(s string, def int) int {
	if s == "" {
		return def
	}
	n, err := strconv.Atoi(s)
	if err != nil || n < 0 {
		return def
	}
	return n
}
