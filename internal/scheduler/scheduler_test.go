package scheduler

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"sync"
	"testing"
	"time"

	"github.com/Fzgt/cronflux/internal/job"
	"github.com/Fzgt/cronflux/internal/store"
	"github.com/Fzgt/cronflux/internal/store/memory"
)

// recorder is a test Executor that records the jobs it ran and can be told to
// fail.
type recorder struct {
	mu    sync.Mutex
	calls []string
	fail  bool
}

func (r *recorder) Execute(_ context.Context, j job.Job, _ job.Run) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.calls = append(r.calls, j.ID)
	if r.fail {
		return errors.New("boom")
	}
	return nil
}

func (r *recorder) count() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return len(r.calls)
}

func seqID() func() string {
	var mu sync.Mutex
	var n int
	return func() string {
		mu.Lock()
		defer mu.Unlock()
		n++
		return fmt.Sprintf("id-%d", n)
	}
}

func quietLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

func newTestScheduler(st store.Store, exec Executor, now *time.Time) *Scheduler {
	return New(Options{
		Store:        st,
		Executor:     exec,
		TickInterval: time.Minute,
		Workers:      1,
		Lease:        time.Minute,
		NewID:        seqID(),
		Now:          func() time.Time { return *now },
		Logger:       quietLogger(),
	})
}

func TestSchedulerRunsDueJob(t *testing.T) {
	st := memory.New()
	rec := &recorder{}
	now := time.Date(2026, 5, 1, 10, 0, 0, 0, time.UTC)
	s := newTestScheduler(st, rec, &now)
	ctx := context.Background()

	if err := st.PutJob(ctx, job.Job{ID: "beat", Spec: "* * * * *", Enabled: true}); err != nil {
		t.Fatalf("PutJob: %v", err)
	}

	// First tick just schedules the next fire in the future.
	s.step(ctx, now)
	if rec.count() != 0 {
		t.Fatalf("job ran on the priming tick")
	}

	// Advance past the next minute boundary; the job should run once.
	now = now.Add(90 * time.Second)
	s.step(ctx, now)

	if rec.count() != 1 {
		t.Fatalf("job ran %d times, want 1", rec.count())
	}
	runs, err := st.ListRuns(ctx, store.RunFilter{JobID: "beat"})
	if err != nil {
		t.Fatalf("ListRuns: %v", err)
	}
	if len(runs) != 1 || runs[0].State != job.StateSucceeded {
		t.Fatalf("unexpected runs: %+v", runs)
	}
}

func TestTriggerRunsImmediately(t *testing.T) {
	st := memory.New()
	rec := &recorder{}
	now := time.Date(2026, 5, 1, 10, 0, 0, 0, time.UTC)
	s := newTestScheduler(st, rec, &now)
	ctx := context.Background()

	if err := st.PutJob(ctx, job.Job{ID: "adhoc", Enabled: true}); err != nil {
		t.Fatalf("PutJob: %v", err)
	}
	run, err := s.Trigger(ctx, "adhoc")
	if err != nil {
		t.Fatalf("Trigger: %v", err)
	}
	if run.State != job.StatePending {
		t.Fatalf("triggered run state = %q, want pending", run.State)
	}

	s.step(ctx, now)
	if rec.count() != 1 {
		t.Fatalf("triggered job ran %d times, want 1", rec.count())
	}
}
