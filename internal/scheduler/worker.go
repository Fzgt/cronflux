package scheduler

import (
	"context"
	"log/slog"
	"time"

	"github.com/Fzgt/cronflux/backoff"
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

	// The execution failed. Retry with backoff until MaxRetries is spent, then
	// move the run to the dead-letter state.
	r.Error = execErr.Error()
	if r.Attempt < j.MaxRetries {
		r.State = job.StateFailed
		if err := w.store.UpdateRun(ctx, r); err != nil {
			return r, err
		}
		if err := w.scheduleRetry(ctx, j, r, finish); err != nil {
			return r, err
		}
		w.metrics.ObserveRun(j.ID, job.StateFailed, dur)
		w.metrics.IncRetry(j.ID)
		return r, nil
	}

	r.State = job.StateDead
	w.metrics.ObserveRun(j.ID, job.StateDead, dur)
	return r, w.store.UpdateRun(ctx, r)
}

// scheduleRetry enqueues a fresh pending run for the next attempt, delayed by
// the job's backoff policy.
func (w *worker) scheduleRetry(ctx context.Context, j job.Job, failed job.Run, at time.Time) error {
	delay := backoffFor(j.Backoff).Delay(failed.Attempt)
	retry := job.Run{
		ID:           w.newID(),
		JobID:        failed.JobID,
		BatchID:      failed.BatchID,
		State:        job.StatePending,
		Attempt:      failed.Attempt + 1,
		ScheduledFor: at.Add(delay),
		CreatedAt:    at,
		UpdatedAt:    at,
	}
	if w.log != nil {
		w.log.Debug("scheduling retry", "job", j.ID, "attempt", retry.Attempt, "delay", delay)
	}
	return w.store.CreateRun(ctx, retry)
}

// backoffFor builds an exponential backoff from a job's retry policy.
func backoffFor(p job.RetryPolicy) backoff.Exponential {
	return backoff.Exponential{Base: p.Base, Max: p.Max, Factor: p.Factor, Jitter: p.Jitter}
}
