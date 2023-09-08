package scheduler

import (
	"context"
	"testing"
	"time"

	"github.com/Fzgt/cronflux/internal/job"
	"github.com/Fzgt/cronflux/internal/store"
	"github.com/Fzgt/cronflux/internal/store/memory"
)

func TestRetryThenDeadLetter(t *testing.T) {
	st := memory.New()
	rec := &recorder{fail: true}
	now := time.Date(2026, 5, 1, 10, 0, 0, 0, time.UTC)
	s := New(Options{
		Store:    st,
		Executor: rec,
		Workers:  1,
		Lease:    time.Minute,
		NewID:    seqID(),
		Now:      func() time.Time { return now },
		Logger:   quietLogger(),
	})
	ctx := context.Background()

	mustPut(t, st, job.Job{
		ID:         "flaky",
		Enabled:    true,
		MaxRetries: 2,
		Backoff:    job.RetryPolicy{Base: time.Second, Factor: 2},
	})

	if _, err := s.Trigger(ctx, "flaky"); err != nil {
		t.Fatalf("Trigger: %v", err)
	}

	// Advance well past each backoff so every attempt is claimed in turn.
	for i := 0; i < 5; i++ {
		s.step(ctx, now)
		now = now.Add(10 * time.Second)
	}

	if got := rec.count(); got != 3 {
		t.Fatalf("executor ran %d times, want 3 (initial + 2 retries)", got)
	}

	runs, _ := st.ListRuns(ctx, store.RunFilter{JobID: "flaky"})
	var dead int
	for _, r := range runs {
		if r.State == job.StateDead {
			dead++
		}
	}
	if dead != 1 {
		t.Fatalf("expected exactly one dead run, got %d among %d runs", dead, len(runs))
	}
}
