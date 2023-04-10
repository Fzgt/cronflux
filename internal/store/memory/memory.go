// Package memory provides an in-memory implementation of store.Store. It is
// safe for concurrent use and is cronflux's default backend: handy for tests
// and single-node runs that do not need durability across restarts.
package memory

import (
	"context"
	"sort"
	"sync"

	"github.com/Fzgt/cronflux/internal/job"
	"github.com/Fzgt/cronflux/internal/store"
)

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
