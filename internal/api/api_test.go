package api

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Fzgt/cronflux/internal/job"
	"github.com/Fzgt/cronflux/internal/metrics"
	"github.com/Fzgt/cronflux/internal/scheduler"
	"github.com/Fzgt/cronflux/internal/store"
	"github.com/Fzgt/cronflux/internal/store/memory"
)

func discardLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

func newTestServer(t *testing.T) (*Server, store.Store) {
	t.Helper()
	st := memory.New()
	m := metrics.New()
	sch := scheduler.New(scheduler.Options{Store: st, Metrics: m, Logger: discardLogger()})
	s := NewServer(Config{
		Addr:      ":0",
		Store:     st,
		Scheduler: sch,
		Gatherer:  m.Registry(),
		Logger:    discardLogger(),
	})
	return s, st
}

func do(s *Server, method, path string, body any) *httptest.ResponseRecorder {
	var reader io.Reader
	if body != nil {
		b, _ := json.Marshal(body)
		reader = bytes.NewReader(b)
	}
	req := httptest.NewRequest(method, path, reader)
	rec := httptest.NewRecorder()
	s.http.Handler.ServeHTTP(rec, req)
	return rec
}

func TestJobLifecycleOverHTTP(t *testing.T) {
	s, _ := newTestServer(t)

	j := job.Job{ID: "nightly", Name: "Nightly", Spec: "@daily", Enabled: true}
	if rec := do(s, http.MethodPost, "/api/jobs", j); rec.Code != http.StatusCreated {
		t.Fatalf("create job = %d, body %s", rec.Code, rec.Body)
	}

	rec := do(s, http.MethodGet, "/api/jobs", nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("list jobs = %d", rec.Code)
	}
	var jobs []job.Job
	if err := json.Unmarshal(rec.Body.Bytes(), &jobs); err != nil {
		t.Fatalf("decode jobs: %v", err)
	}
	if len(jobs) != 1 || jobs[0].ID != "nightly" {
		t.Fatalf("unexpected jobs: %+v", jobs)
	}

	if rec := do(s, http.MethodGet, "/api/jobs/nightly", nil); rec.Code != http.StatusOK {
		t.Fatalf("get job = %d", rec.Code)
	}
}

func TestCreateJobRejectsBadSpec(t *testing.T) {
	s, _ := newTestServer(t)
	j := job.Job{ID: "broken", Spec: "not a cron"}
	if rec := do(s, http.MethodPost, "/api/jobs", j); rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for bad spec, got %d", rec.Code)
	}
}

func TestTriggerCreatesRun(t *testing.T) {
	s, st := newTestServer(t)
	j := job.Job{ID: "adhoc", Enabled: true}
	do(s, http.MethodPost, "/api/jobs", j)

	rec := do(s, http.MethodPost, "/api/jobs/adhoc/trigger", nil)
	if rec.Code != http.StatusAccepted {
		t.Fatalf("trigger = %d, body %s", rec.Code, rec.Body)
	}
	runs, err := st.ListRuns(context.Background(), store.RunFilter{JobID: "adhoc"})
	if err != nil {
		t.Fatalf("ListRuns: %v", err)
	}
	if len(runs) != 1 {
		t.Fatalf("expected 1 run after trigger, got %d", len(runs))
	}
}

func TestHealthAndMetrics(t *testing.T) {
	s, _ := newTestServer(t)

	if rec := do(s, http.MethodGet, "/healthz", nil); rec.Code != http.StatusOK {
		t.Errorf("healthz = %d", rec.Code)
	}
	if rec := do(s, http.MethodGet, "/readyz", nil); rec.Code != http.StatusOK {
		t.Errorf("readyz = %d", rec.Code)
	}
	rec := do(s, http.MethodGet, "/metrics", nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("metrics = %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "cronflux_") {
		t.Errorf("metrics output missing cronflux_ series")
	}
}

func TestDashboardServesIndex(t *testing.T) {
	s, _ := newTestServer(t)
	rec := do(s, http.MethodGet, "/", nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("dashboard = %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "<title>cronflux</title>") {
		t.Errorf("dashboard did not serve index.html")
	}
}
