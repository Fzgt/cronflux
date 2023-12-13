// Package storetest provides a reusable conformance suite that any
// store.Store implementation can run to prove it honors the interface
// contract. Both the in-memory and PostgreSQL backends share it so their
// behavior cannot drift apart.
package storetest

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Fzgt/cronflux/internal/job"
	"github.com/Fzgt/cronflux/internal/store"
)

// Factory builds a fresh, empty store for a single sub-test.
type Factory func(t *testing.T) store.Store

// Run executes the full conformance suite against stores produced by newStore.
func Run(t *testing.T, newStore Factory) {
	t.Run("JobCRUD", func(t *testing.T) { testJobCRUD(t, newStore(t)) })
	t.Run("RunLifecycle", func(t *testing.T) { testRunLifecycle(t, newStore(t)) })
	t.Run("ClaimDue", func(t *testing.T) { testClaimDue(t, newStore(t)) })
	t.Run("LeaseRedelivery", func(t *testing.T) { testLeaseRedelivery(t, newStore(t)) })
	t.Run("ListRunsFilter", func(t *testing.T) { testListRunsFilter(t, newStore(t)) })
}

func testJobCRUD(t *testing.T, s store.Store) {
	ctx := context.Background()
	j := job.Job{ID: "build", Name: "Build", Spec: "@hourly", Command: []string{"make"}}

	if err := s.PutJob(ctx, j); err != nil {
		t.Fatalf("PutJob: %v", err)
	}
	got, err := s.GetJob(ctx, "build")
	if err != nil {
		t.Fatalf("GetJob: %v", err)
	}
	if got.Name != "Build" || got.Spec != "@hourly" {
		t.Errorf("GetJob returned %+v", got)
	}

	jobs, err := s.ListJobs(ctx)
	if err != nil {
		t.Fatalf("ListJobs: %v", err)
	}
	if len(jobs) != 1 {
		t.Fatalf("ListJobs len = %d, want 1", len(jobs))
	}

	if err := s.DeleteJob(ctx, "build"); err != nil {
		t.Fatalf("DeleteJob: %v", err)
	}
	if _, err := s.GetJob(ctx, "build"); !errors.Is(err, store.ErrNotFound) {
		t.Errorf("GetJob after delete = %v, want ErrNotFound", err)
	}
}

func testRunLifecycle(t *testing.T, s store.Store) {
	ctx := context.Background()
	now := time.Date(2026, 5, 1, 10, 0, 0, 0, time.UTC)
	r := job.Run{ID: "r1", JobID: "build", State: job.StatePending, ScheduledFor: now, CreatedAt: now}

	if err := s.CreateRun(ctx, r); err != nil {
		t.Fatalf("CreateRun: %v", err)
	}
	r.State = job.StateSucceeded
	if err := s.UpdateRun(ctx, r); err != nil {
		t.Fatalf("UpdateRun: %v", err)
	}
	got, err := s.GetRun(ctx, "r1")
	if err != nil {
		t.Fatalf("GetRun: %v", err)
	}
	if got.State != job.StateSucceeded {
		t.Errorf("run state = %q, want succeeded", got.State)
	}
	if _, err := s.GetRun(ctx, "missing"); !errors.Is(err, store.ErrNotFound) {
		t.Errorf("GetRun(missing) = %v, want ErrNotFound", err)
	}
}

func testClaimDue(t *testing.T, s store.Store) {
	ctx := context.Background()
	now := time.Date(2026, 5, 1, 10, 0, 0, 0, time.UTC)

	due := job.Run{ID: "due", JobID: "j", State: job.StatePending, ScheduledFor: now.Add(-time.Minute), CreatedAt: now}
	future := job.Run{ID: "future", JobID: "j", State: job.StatePending, ScheduledFor: now.Add(time.Hour), CreatedAt: now}
	mustCreate(t, s, due, future)

	claimed, err := s.ClaimDue(ctx, now, "worker-1", 30*time.Second, 10)
	if err != nil {
		t.Fatalf("ClaimDue: %v", err)
	}
	if len(claimed) != 1 || claimed[0].ID != "due" {
		t.Fatalf("claimed = %v, want [due]", ids(claimed))
	}
	if claimed[0].State != job.StateRunning || claimed[0].Worker != "worker-1" {
		t.Errorf("claimed run not leased correctly: %+v", claimed[0])
	}
	if !claimed[0].LeaseExpiry.Equal(now.Add(30 * time.Second)) {
		t.Errorf("lease expiry = %s, want %s", claimed[0].LeaseExpiry, now.Add(30*time.Second))
	}

	// A second claim finds nothing: the due run is leased, the other is future.
	again, err := s.ClaimDue(ctx, now, "worker-2", 30*time.Second, 10)
	if err != nil {
		t.Fatalf("ClaimDue again: %v", err)
	}
	if len(again) != 0 {
		t.Errorf("second claim returned %v, want none", ids(again))
	}
}

func testLeaseRedelivery(t *testing.T, s store.Store) {
	ctx := context.Background()
	now := time.Date(2026, 5, 1, 10, 0, 0, 0, time.UTC)
	r := job.Run{ID: "r", JobID: "j", State: job.StatePending, ScheduledFor: now.Add(-time.Minute), CreatedAt: now}
	mustCreate(t, s, r)

	if _, err := s.ClaimDue(ctx, now, "worker-1", 30*time.Second, 10); err != nil {
		t.Fatalf("first claim: %v", err)
	}
	// Before the lease expires nobody else may claim it.
	if got, _ := s.ClaimDue(ctx, now.Add(10*time.Second), "worker-2", 30*time.Second, 10); len(got) != 0 {
		t.Fatalf("claimed a still-leased run: %v", ids(got))
	}
	// After the lease expires the run is redelivered.
	got, err := s.ClaimDue(ctx, now.Add(time.Minute), "worker-2", 30*time.Second, 10)
	if err != nil {
		t.Fatalf("redelivery claim: %v", err)
	}
	if len(got) != 1 || got[0].Worker != "worker-2" {
		t.Fatalf("expected redelivery to worker-2, got %v", ids(got))
	}
}

func testListRunsFilter(t *testing.T, s store.Store) {
	ctx := context.Background()
	base := time.Date(2026, 5, 1, 10, 0, 0, 0, time.UTC)
	mustCreate(t, s,
		job.Run{ID: "a", JobID: "j1", State: job.StateSucceeded, CreatedAt: base},
		job.Run{ID: "b", JobID: "j1", State: job.StateFailed, CreatedAt: base.Add(time.Second)},
		job.Run{ID: "c", JobID: "j2", State: job.StateSucceeded, CreatedAt: base.Add(2 * time.Second)},
	)

	byJob, err := s.ListRuns(ctx, store.RunFilter{JobID: "j1"})
	if err != nil {
		t.Fatalf("ListRuns: %v", err)
	}
	if len(byJob) != 2 {
		t.Errorf("JobID filter returned %v, want 2", ids(byJob))
	}
	// Newest first.
	if len(byJob) == 2 && byJob[0].ID != "b" {
		t.Errorf("expected newest run first, got %s", byJob[0].ID)
	}

	byState, err := s.ListRuns(ctx, store.RunFilter{State: job.StateSucceeded})
	if err != nil {
		t.Fatalf("ListRuns state: %v", err)
	}
	if len(byState) != 2 {
		t.Errorf("state filter returned %v, want 2", ids(byState))
	}

	limited, err := s.ListRuns(ctx, store.RunFilter{Limit: 1})
	if err != nil {
		t.Fatalf("ListRuns limit: %v", err)
	}
	if len(limited) != 1 {
		t.Errorf("limit returned %d rows, want 1", len(limited))
	}
}

func mustCreate(t *testing.T, s store.Store, runs ...job.Run) {
	t.Helper()
	for _, r := range runs {
		if err := s.CreateRun(context.Background(), r); err != nil {
			t.Fatalf("CreateRun(%s): %v", r.ID, err)
		}
	}
}

func ids(runs []job.Run) []string {
	out := make([]string, len(runs))
	for i, r := range runs {
		out[i] = r.ID
	}
	return out
}
