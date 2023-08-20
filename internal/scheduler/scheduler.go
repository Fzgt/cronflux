package scheduler

import (
	"context"
	crand "crypto/rand"
	"encoding/hex"
	"log/slog"
	"sync"
	"time"

	"github.com/Fzgt/cronflux/internal/clock"
	"github.com/Fzgt/cronflux/internal/job"
	"github.com/Fzgt/cronflux/internal/metrics"
	"github.com/Fzgt/cronflux/internal/store"
)

// Options configures a Scheduler. Only Store is required; every other field
// has a sensible default.
type Options struct {
	Store        store.Store
	Executor     Executor
	Metrics      *metrics.Metrics
	TickInterval time.Duration
	Workers      int
	Lease        time.Duration
	Logger       *slog.Logger
	NewID        func() string
	// Clock overrides the time source. If nil, Now is used, and if that is
	// also nil the system clock is used.
	Clock clock.Clock
	// Now is a convenience shortcut for a clock built from a function.
	Now func() time.Time
}

// Scheduler materialises runs, dispatches them to workers and advances DAG
// dependencies. It is safe to Trigger concurrently with Run.
type Scheduler struct {
	store   store.Store
	exec    Executor
	metrics *metrics.Metrics
	disp    *dispatcher

	tick    time.Duration
	workers int
	lease   time.Duration
	log     *slog.Logger
	newID   func() string
	clock   clock.Clock

	mu sync.Mutex // serialises dependent enqueueing across worker goroutines
}

// New builds a Scheduler, filling in defaults for any unset option.
func New(opts Options) *Scheduler {
	if opts.TickInterval <= 0 {
		opts.TickInterval = time.Second
	}
	if opts.Workers < 1 {
		opts.Workers = 4
	}
	if opts.Lease <= 0 {
		opts.Lease = 30 * time.Second
	}
	if opts.Logger == nil {
		opts.Logger = slog.Default()
	}
	if opts.NewID == nil {
		opts.NewID = defaultID
	}
	if opts.Metrics == nil {
		opts.Metrics = metrics.New()
	}
	if opts.Executor == nil {
		opts.Executor = NoopExecutor{}
	}
	clk := opts.Clock
	if clk == nil {
		if opts.Now != nil {
			clk = clock.FromFunc(opts.Now)
		} else {
			clk = clock.Real{}
		}
	}
	return &Scheduler{
		store:   opts.Store,
		exec:    opts.Executor,
		metrics: opts.Metrics,
		disp:    newDispatcher(opts.Store, opts.NewID),
		tick:    opts.TickInterval,
		workers: opts.Workers,
		lease:   opts.Lease,
		log:     opts.Logger,
		newID:   opts.NewID,
		clock:   clk,
	}
}

// Metrics returns the scheduler's metrics so the HTTP layer can expose them.
func (s *Scheduler) Metrics() *metrics.Metrics { return s.metrics }

// Run drives the scheduler until ctx is cancelled, at which point it returns
// nil after the in-flight tick completes.
func (s *Scheduler) Run(ctx context.Context) error {
	ticker := time.NewTicker(s.tick)
	defer ticker.Stop()
	s.log.Info("scheduler started", "workers", s.workers, "tick", s.tick.String())
	for {
		select {
		case <-ctx.Done():
			s.log.Info("scheduler stopping")
			return nil
		case <-ticker.C:
			s.step(ctx, s.clock.Now())
		}
	}
}

// step performs one scheduling pass. It is called from Run's loop and directly
// from tests, which makes the scheduler deterministic under a mock clock.
func (s *Scheduler) step(ctx context.Context, now time.Time) {
	s.metrics.Tick()
	if _, err := s.disp.dispatch(ctx, now); err != nil {
		s.log.Error("dispatch failed", "err", err)
	}
	s.drain(ctx, now)
	s.publishBacklog(ctx, now)
}

// drain claims every ready run and processes it across the worker pool.
func (s *Scheduler) drain(ctx context.Context, now time.Time) {
	limit := minInt(s.workers*4, 256)
	claimed, err := s.store.ClaimDue(ctx, now, s.newID(), s.lease, limit)
	if err != nil {
		s.log.Error("claim failed", "err", err)
		return
	}
	if len(claimed) == 0 {
		return
	}

	sem := make(chan struct{}, s.workers)
	var wg sync.WaitGroup
	for _, r := range claimed {
		wg.Add(1)
		sem <- struct{}{}
		go func(run job.Run) {
			defer wg.Done()
			defer func() { <-sem }()
			w := &worker{
				store:   s.store,
				exec:    s.exec,
				metrics: s.metrics,
				log:     s.log,
				newID:   s.newID,
				now:     s.clock.Now,
			}
			result, err := w.process(ctx, run)
			if err != nil {
				s.log.Error("run processing failed", "run", run.ID, "err", err)
				return
			}
			if result.State == job.StateSucceeded {
				s.mu.Lock()
				err := s.disp.enqueueDependents(ctx, result, s.clock.Now())
				s.mu.Unlock()
				if err != nil {
					s.log.Error("enqueue dependents failed", "run", run.ID, "err", err)
				}
			}
		}(r)
	}
	wg.Wait()
}

// Trigger enqueues an immediate run for a job in a fresh batch. It backs the
// manual "run now" action in the HTTP API.
func (s *Scheduler) Trigger(ctx context.Context, jobID string) (job.Run, error) {
	j, err := s.store.GetJob(ctx, jobID)
	if err != nil {
		return job.Run{}, err
	}
	now := s.clock.Now()
	run := job.Run{
		ID:           s.newID(),
		JobID:        j.ID,
		BatchID:      s.newID(),
		State:        job.StatePending,
		ScheduledFor: now,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	if err := s.store.CreateRun(ctx, run); err != nil {
		return job.Run{}, err
	}
	return run, nil
}

// publishBacklog updates the pending-run and lag gauges.
func (s *Scheduler) publishBacklog(ctx context.Context, now time.Time) {
	pending, err := s.store.ListRuns(ctx, store.RunFilter{State: job.StatePending})
	if err != nil {
		return
	}
	s.metrics.SetPending(len(pending))

	var oldest time.Time
	for _, r := range pending {
		if r.ScheduledFor.After(now) {
			continue
		}
		if oldest.IsZero() || r.ScheduledFor.Before(oldest) {
			oldest = r.ScheduledFor
		}
	}
	if oldest.IsZero() {
		s.metrics.SetLag(0)
		return
	}
	s.metrics.SetLag(now.Sub(oldest))
}

// defaultID returns a random 96-bit hex identifier.
func defaultID() string {
	var b [12]byte
	if _, err := crand.Read(b[:]); err != nil {
		return hex.EncodeToString([]byte(time.Now().Format(time.RFC3339Nano)))
	}
	return hex.EncodeToString(b[:])
}
