package api

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/Fzgt/cronflux/internal/job"
)

func TestListRunsFilters(t *testing.T) {
	s, st := newTestServer(t)
	ctx := context.Background()
	base := time.Date(2026, 5, 1, 10, 0, 0, 0, time.UTC)

	seed := []job.Run{
		{ID: "r1", JobID: "j1", State: job.StateSucceeded, CreatedAt: base},
		{ID: "r2", JobID: "j1", State: job.StateFailed, CreatedAt: base.Add(time.Second)},
		{ID: "r3", JobID: "j2", State: job.StateSucceeded, CreatedAt: base.Add(2 * time.Second)},
	}
	for _, r := range seed {
		if err := st.CreateRun(ctx, r); err != nil {
			t.Fatalf("CreateRun: %v", err)
		}
	}

	list := func(query string) []job.Run {
		rec := do(s, http.MethodGet, "/api/runs"+query, nil)
		if rec.Code != http.StatusOK {
			t.Fatalf("GET /api/runs%s = %d", query, rec.Code)
		}
		var runs []job.Run
		if err := json.Unmarshal(rec.Body.Bytes(), &runs); err != nil {
			t.Fatalf("decode: %v", err)
		}
		return runs
	}

	if got := list(""); len(got) != 3 {
		t.Errorf("all runs = %d, want 3", len(got))
	}
	if got := list("?job=j1"); len(got) != 2 {
		t.Errorf("job filter = %d, want 2", len(got))
	}
	if got := list("?state=succeeded"); len(got) != 2 {
		t.Errorf("state filter = %d, want 2", len(got))
	}
	if got := list("?limit=1"); len(got) != 1 {
		t.Errorf("limit = %d, want 1", len(got))
	}
}

func TestGetRunByID(t *testing.T) {
	s, st := newTestServer(t)
	ctx := context.Background()
	if err := st.CreateRun(ctx, job.Run{ID: "run-xyz", JobID: "j", State: job.StateRunning}); err != nil {
		t.Fatalf("CreateRun: %v", err)
	}

	rec := do(s, http.MethodGet, "/api/runs/run-xyz", nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("get run = %d", rec.Code)
	}
	var r job.Run
	if err := json.Unmarshal(rec.Body.Bytes(), &r); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if r.JobID != "j" || r.State != job.StateRunning {
		t.Errorf("unexpected run: %+v", r)
	}
}
