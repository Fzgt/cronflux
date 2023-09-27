package scheduler

import (
	"context"
	"time"

	"github.com/Fzgt/cronflux/cron"
	"github.com/Fzgt/cronflux/internal/job"
	"github.com/Fzgt/cronflux/internal/store"
)

// dispatcher materialises pending runs. For cron-scheduled jobs it enqueues a
// run each time a fire time elapses; for DAG jobs it enqueues dependents once
// their upstreams have succeeded within the same batch.
//
// A dispatcher is driven from the single scheduler loop goroutine, so its
// caches need no locking.
type dispatcher struct {
	store store.Store
	newID func() string

	nextFire map[string]time.Time
	parsed   map[string]cron.Schedule
}

func newDispatcher(s store.Store, newID func() string) *dispatcher {
	return &dispatcher{
		store:    s,
		newID:    newID,
		nextFire: make(map[string]time.Time),
		parsed:   make(map[string]cron.Schedule),
	}
}

// dispatch enqueues a pending run for every enabled cron job whose fire time is
// at or before now, returning how many runs were created. The first time a job
// is seen its next fire is scheduled in the future so startup does not trigger
// a spurious immediate run.
func (d *dispatcher) dispatch(ctx context.Context, now time.Time) (int, error) {
	jobs, err := d.store.ListJobs(ctx)
	if err != nil {
		return 0, err
	}

	created := 0
	for _, j := range jobs {
		if !j.Enabled || j.Spec == "" {
			continue
		}
		sched, err := d.schedule(j)
		if err != nil {
			// Specs are validated on ingest; skip anything malformed here.
			continue
		}

		next, ok := d.nextFire[j.ID]
		if !ok {
			d.nextFire[j.ID] = sched.Next(now)
			continue
		}

		for !next.After(now) {
			if err := d.enqueueRoot(ctx, j, next); err != nil {
				return created, err
			}
			created++
			next = sched.Next(next)
		}
		d.nextFire[j.ID] = next
	}
	return created, nil
}

// schedule returns the parsed schedule for a job, caching the result.
func (d *dispatcher) schedule(j job.Job) (cron.Schedule, error) {
	if s, ok := d.parsed[j.ID]; ok {
		return s, nil
	}
	s, err := cron.Parse(j.Spec)
	if err != nil {
		return nil, err
	}
	d.parsed[j.ID] = s
	return s, nil
}

// enqueueRoot creates the pending run that starts a new batch for a cron fire.
// It is a no-op if a run already exists for the same job and slot, which keeps
// materialisation idempotent across scheduler restarts that re-seed the
// next-fire cache.
func (d *dispatcher) enqueueRoot(ctx context.Context, j job.Job, fireAt time.Time) error {
	existing, err := d.store.ListRuns(ctx, store.RunFilter{JobID: j.ID})
	if err != nil {
		return err
	}
	for _, r := range existing {
		if r.ScheduledFor.Equal(fireAt) {
			return nil
		}
	}

	run := job.Run{
		ID:           d.newID(),
		JobID:        j.ID,
		BatchID:      d.newID(),
		State:        job.StatePending,
		ScheduledFor: fireAt,
		CreatedAt:    fireAt,
		UpdatedAt:    fireAt,
	}
	return d.store.CreateRun(ctx, run)
}

// enqueueDependents creates pending runs for jobs that depend on the completed
// job, once every one of their dependencies has succeeded in the same batch. It
// is idempotent: a dependent already present in the batch is left untouched.
func (d *dispatcher) enqueueDependents(ctx context.Context, completed job.Run, now time.Time) error {
	jobs, err := d.store.ListJobs(ctx)
	if err != nil {
		return err
	}
	for _, j := range jobs {
		if !contains(j.DependsOn, completed.JobID) {
			continue
		}
		ready, err := d.depsSatisfied(ctx, j, completed.BatchID)
		if err != nil {
			return err
		}
		if !ready {
			continue
		}
		existing, err := d.store.ListRuns(ctx, store.RunFilter{JobID: j.ID, BatchID: completed.BatchID})
		if err != nil {
			return err
		}
		if len(existing) > 0 {
			continue
		}
		run := job.Run{
			ID:           d.newID(),
			JobID:        j.ID,
			BatchID:      completed.BatchID,
			State:        job.StatePending,
			ScheduledFor: now,
			CreatedAt:    now,
			UpdatedAt:    now,
		}
		if err := d.store.CreateRun(ctx, run); err != nil {
			return err
		}
	}
	return nil
}

// depsSatisfied reports whether every dependency of j has a succeeded run in
// the given batch.
func (d *dispatcher) depsSatisfied(ctx context.Context, j job.Job, batchID string) (bool, error) {
	for _, dep := range j.DependsOn {
		runs, err := d.store.ListRuns(ctx, store.RunFilter{JobID: dep, BatchID: batchID})
		if err != nil {
			return false, err
		}
		if !hasSucceeded(runs) {
			return false, nil
		}
	}
	return true, nil
}

func hasSucceeded(runs []job.Run) bool {
	for _, r := range runs {
		if r.State == job.StateSucceeded {
			return true
		}
	}
	return false
}

func contains(xs []string, want string) bool {
	for _, x := range xs {
		if x == want {
			return true
		}
	}
	return false
}
