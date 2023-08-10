package scheduler

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Fzgt/cronflux/internal/job"
	"github.com/Fzgt/cronflux/internal/store"
	"github.com/Fzgt/cronflux/internal/store/memory"
)

func TestDagEnqueuesDependentOnSuccess(t *testing.T) {
	st := memory.New()
	rec := &recorder{}
	now := time.Date(2026, 5, 1, 10, 0, 0, 0, time.UTC)
	s := newTestScheduler(st, rec, &now)
	ctx := context.Background()

	mustPut(t, st, job.Job{ID: "a", Enabled: true})
	mustPut(t, st, job.Job{ID: "b", Enabled: true, DependsOn: []string{"a"}})

	// Trigger the root; the first step runs "a" and enqueues "b".
	if _, err := s.Trigger(ctx, "a"); err != nil {
		t.Fatalf("Trigger: %v", err)
	}
	s.step(ctx, now)
	if got := rec.calls; len(got) != 1 || got[0] != "a" {
		t.Fatalf("after first step ran %v, want [a]", got)
	}

	// The second step picks up the freshly enqueued "b".
	now = now.Add(time.Second)
	s.step(ctx, now)
	if got := rec.calls; len(got) != 2 || got[1] != "b" {
		t.Fatalf("after second step ran %v, want [a b]", got)
	}

	// "b" ran in the same batch as "a".
	runsA, _ := st.ListRuns(ctx, store.RunFilter{JobID: "a"})
	runsB, _ := st.ListRuns(ctx, store.RunFilter{JobID: "b"})
	if len(runsA) != 1 || len(runsB) != 1 {
		t.Fatalf("expected one run each, got a=%d b=%d", len(runsA), len(runsB))
	}
	if runsA[0].BatchID != runsB[0].BatchID {
		t.Errorf("dependent ran in batch %q, want %q", runsB[0].BatchID, runsA[0].BatchID)
	}
}

func TestDagSkipsDependentWhenParentFails(t *testing.T) {
	st := memory.New()
	now := time.Date(2026, 5, 1, 10, 0, 0, 0, time.UTC)
	failing := ExecutorFunc(func(_ context.Context, j job.Job, _ job.Run) error {
		if j.ID == "a" {
			return errors.New("boom")
		}
		return nil
	})
	s := newTestScheduler(st, failing, &now)
	ctx := context.Background()

	mustPut(t, st, job.Job{ID: "a", Enabled: true})
	mustPut(t, st, job.Job{ID: "b", Enabled: true, DependsOn: []string{"a"}})

	if _, err := s.Trigger(ctx, "a"); err != nil {
		t.Fatalf("Trigger: %v", err)
	}
	s.step(ctx, now)
	now = now.Add(time.Second)
	s.step(ctx, now)

	runsB, _ := st.ListRuns(ctx, store.RunFilter{JobID: "b"})
	if len(runsB) != 0 {
		t.Fatalf("dependent should not run when parent fails, got %d runs", len(runsB))
	}
}

func mustPut(t *testing.T, st store.Store, j job.Job) {
	t.Helper()
	if err := st.PutJob(context.Background(), j); err != nil {
		t.Fatalf("PutJob(%s): %v", j.ID, err)
	}
}
