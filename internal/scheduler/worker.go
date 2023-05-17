package scheduler

import (
	"context"
	"log/slog"
	"time"

	"github.com/Fzgt/cronflux/internal/job"
	"github.com/Fzgt/cronflux/internal/metrics"
	"github.com/Fzgt/cronflux/internal/store"
)

// worker executes a single claimed run and records its outcome in the store.
type worker struct {
	store   store.Store
	exec    Executor
	metrics *metrics.Metrics
	log     *slog.Logger
	newID   func() string
	now     func() time.Time
}

// process runs r to completion and returns the run in its resulting state.
func (w *worker) process(ctx context.Context, r job.Run) (job.Run, error) {
	j, err := w.store.GetJob(ctx, r.JobID)
	if err != nil {
		return r, err
	}

	runCtx := ctx
	if j.Timeout > 0 {
		var cancel context.CancelFunc
		runCtx, cancel = context.WithTimeout(ctx, j.Timeout)
		defer cancel()
	}

	start := w.now()
	execErr := w.exec.Execute(runCtx, j, r)
	finish := w.now()
	dur := finish.Sub(start)

	finished := finish
	r.FinishedAt = &finished
	r.UpdatedAt = finish

	if execErr == nil {
		r.State = job.StateSucceeded
		r.Error = ""
		w.metrics.ObserveRun(j.ID, r.State, dur)
		return r, w.store.UpdateRun(ctx, r)
	}

	// Retry and dead-letter handling is layered on separately.
	r.State = job.StateFailed
	r.Error = execErr.Error()
	w.metrics.ObserveRun(j.ID, r.State, dur)
	return r, w.store.UpdateRun(ctx, r)
}
