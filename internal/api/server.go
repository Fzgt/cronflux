// Package api serves cronflux's HTTP interface: a small REST API over jobs and
// runs, health probes, Prometheus metrics and an embedded web dashboard.
package api

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/Fzgt/cronflux/internal/scheduler"
	"github.com/Fzgt/cronflux/internal/store"
)

// Config configures the HTTP server.
type Config struct {
	Addr      string
	Store     store.Store
	Scheduler *scheduler.Scheduler
	Gatherer  prometheus.Gatherer
	Logger    *slog.Logger
}

// Server hosts the cronflux HTTP API.
type Server struct {
	store     store.Store
	scheduler *scheduler.Scheduler
	gatherer  prometheus.Gatherer
	log       *slog.Logger
	http      *http.Server
}

// NewServer builds a Server ready to be started.
func NewServer(cfg Config) *Server {
	log := cfg.Logger
	if log == nil {
		log = slog.Default()
	}
	s := &Server{
		store:     cfg.Store,
		scheduler: cfg.Scheduler,
		gatherer:  cfg.Gatherer,
		log:       log,
	}
	mux := http.NewServeMux()
	s.routes(mux)
	s.http = &http.Server{
		Addr:              cfg.Addr,
		Handler:           s.withLogging(mux),
		ReadHeaderTimeout: 5 * time.Second,
	}
	return s
}

// routes registers every handler on the mux. It grows as features are added.
func (s *Server) routes(mux *http.ServeMux) {
	s.registerJobRoutes(mux)
	s.registerRunRoutes(mux)
	s.registerHealthRoutes(mux)
	mux.HandleFunc("GET /api", s.serviceInfo)
	if s.gatherer != nil {
		mux.Handle("GET /metrics", promhttp.HandlerFor(s.gatherer, promhttp.HandlerOpts{}))
	}
	s.registerDashboard(mux)
	mux.HandleFunc("GET /", s.handleRoot)
}

// handleRoot answers unmatched paths with a 404.
func (s *Server) handleRoot(w http.ResponseWriter, _ *http.Request) {
	writeError(w, http.StatusNotFound, "not found")
}

// Start begins serving and blocks until the server is shut down.
func (s *Server) Start() error {
	s.log.Info("http server listening", "addr", s.http.Addr)
	return s.http.ListenAndServe()
}

// Shutdown gracefully drains in-flight requests, giving them until ctx's
// deadline before forcing connections closed.
func (s *Server) Shutdown(ctx context.Context) error {
	s.log.Info("http server shutting down")
	return s.http.Shutdown(ctx)
}

// ListenAndServe runs the server until ctx is cancelled, then drains it with
// the given grace period. It returns nil on a clean shutdown.
func (s *Server) ListenAndServe(ctx context.Context, grace time.Duration) error {
	errCh := make(chan error, 1)
	go func() {
		if err := s.Start(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
			return
		}
		errCh <- nil
	}()

	select {
	case err := <-errCh:
		return err
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), grace)
		defer cancel()
		return s.Shutdown(shutdownCtx)
	}
}

// handleRoot reports basic service information.
func (s *Server) serviceInfo(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{
		"service": "cronflux",
		"api":     "/api",
		"metrics": "/metrics",
	})
}

// withLogging wraps a handler with structured request logging.
func (s *Server) withLogging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(rec, r)
		s.log.Debug("request",
			"method", r.Method,
			"path", r.URL.Path,
			"status", rec.status,
			"dur", time.Since(start).String(),
		)
	})
}

// statusRecorder captures the response status code for logging.
type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(code int) {
	r.status = code
	r.ResponseWriter.WriteHeader(code)
}

// writeJSON writes v as an indented JSON response with the given status.
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	_ = enc.Encode(v)
}

// writeError writes a JSON error envelope.
func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

// writeStoreError maps a store error to an HTTP response, translating
// store.ErrNotFound into a 404 and everything else into a 500.
func writeStoreError(w http.ResponseWriter, err error) {
	if errors.Is(err, store.ErrNotFound) {
		writeError(w, http.StatusNotFound, "not found")
		return
	}
	writeError(w, http.StatusInternalServerError, err.Error())
}
