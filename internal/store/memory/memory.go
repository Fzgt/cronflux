// Package memory provides an in-memory implementation of store.Store. It is
// safe for concurrent use and is cronflux's default backend: handy for tests
// and single-node runs that do not need durability across restarts.
package memory

import (
	"context"
	"sort"
	"sync"
	"time"

	"github.com/Fzgt/cronflux/internal/job"
	"github.com/Fzgt/cronflux/internal/store"
)

// Store implements store.Store entirely in memory.
var _ store.Store = (*Store)(nil)

// Store keeps jobs and runs in memory behind a single mutex.
type Store struct {
	mu   sync.Mutex
	jobs map[string]job.Job
	runs map[string]job.Run
}

// New returns an empty in-memory store.
func New() *Store {
	return &Store{
		jobs: make(map[string]job.Job),
		runs: make(map[string]job.Run),
	}
}

// PutJob inserts or replaces a job.
func (s *Store) PutJob(_ context.Context, j job.Job) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.jobs[j.ID] = j.Clone()
	return nil
}

// GetJob returns a job by ID.
func (s *Store) GetJob(_ context.Context, id string) (job.Job, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	j, ok := s.jobs[id]
	if !ok {
		return job.Job{}, store.ErrNotFound
	}
	return j.Clone(), nil
}

// ListJobs returns every stored job ordered by ID.
func (s *Store) ListJobs(_ context.Context) ([]job.Job, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]job.Job, 0, len(s.jobs))
	for _, j := range s.jobs {
		out = append(out, j.Clone())
	}
	sort.Slice(out, func(i, k int) bool { return out[i].ID < out[k].ID })
	return out, nil
}

// DeleteJob removes a job by ID.
func (s *Store) DeleteJob(_ context.Context, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.jobs[id]; !ok {
		return store.ErrNotFound
	}
	delete(s.jobs, id)
	return nil
}

// Close is a no-op for the in-memory store.
func (s *Store) Close() error { return nil }

// CreateRun stores a new run.
func (s *Store) CreateRun(_ context.Context, r job.Run) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.runs[r.ID] = r.Clone()
	return nil
}

// GetRun returns a run by ID.
func (s *Store) GetRun(_ context.Context, id string) (job.Run, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	r, ok := s.runs[id]
	if !ok {
		return job.Run{}, store.ErrNotFound
	}
	return r.Clone(), nil
}

// UpdateRun persists changes to an existing run.
func (s *Store) UpdateRun(_ context.Context, r job.Run) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.runs[r.ID]; !ok {
		return store.ErrNotFound
	}
	s.runs[r.ID] = r.Clone()
	return nil
}

// ListRuns returns runs matching the filter, newest first.
func (s *Store) ListRuns(_ context.Context, f store.RunFilter) ([]job.Run, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	out := make([]job.Run, 0, len(s.runs))
	for _, r := range s.runs {
		if f.JobID != "" && r.JobID != f.JobID {
			continue
		}
		if f.BatchID != "" && r.BatchID != f.BatchID {
			continue
		}
		if f.State != "" && r.State != f.State {
			continue
		}
		out = append(out, r.Clone())
	}
	sort.Slice(out, func(i, k int) bool {
		if !out[i].CreatedAt.Equal(out[k].CreatedAt) {
			return out[i].CreatedAt.After(out[k].CreatedAt)
		}
		return out[i].ID > out[k].ID
	})

	if f.Offset > 0 {
		if f.Offset >= len(out) {
			return []job.Run{}, nil
		}
		out = out[f.Offset:]
	}
	if f.Limit > 0 && len(out) > f.Limit {
		out = out[:f.Limit]
	}
	return out, nil
}

// ClaimDue leases up to limit ready runs: pending runs that are due, plus
// running runs whose lease has expired (the at-least-once redelivery path).
func (s *Store) ClaimDue(_ context.Context, now time.Time, worker string, lease time.Duration, limit int) ([]job.Run, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	var ready []job.Run
	for _, r := range s.runs {
		switch {
		case r.State == job.StatePending && !r.ScheduledFor.After(now):
			ready = append(ready, r)
		case r.State == job.StateRunning && r.LeaseExpiry.Before(now):
			ready = append(ready, r)
		}
	}
	sort.Slice(ready, func(i, k int) bool {
		if !ready[i].ScheduledFor.Equal(ready[k].ScheduledFor) {
			return ready[i].ScheduledFor.Before(ready[k].ScheduledFor)
		}
		return ready[i].ID < ready[k].ID
	})
	if limit > 0 && len(ready) > limit {
		ready = ready[:limit]
	}

	claimed := make([]job.Run, 0, len(ready))
	for _, r := range ready {
		r.State = job.StateRunning
		r.Worker = worker
		r.LeaseExpiry = now.Add(lease)
		if r.StartedAt == nil {
			started := now
			r.StartedAt = &started
		}
		r.UpdatedAt = now
		s.runs[r.ID] = r
		claimed = append(claimed, r.Clone())
	}
	return claimed, nil
}
