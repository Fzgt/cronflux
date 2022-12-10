// Package store defines the persistence contract for cronflux together with
// the errors and query types shared by every backend.
package store

import (
	"context"
	"errors"
	"time"

	"github.com/Fzgt/cronflux/internal/job"
)

// ErrNotFound is returned when a requested job or run does not exist.
var ErrNotFound = errors.New("store: not found")

// RunFilter narrows a ListRuns query. Zero-value fields are ignored, so an
// empty filter lists every run newest-first.
type RunFilter struct {
	JobID   string
	BatchID string
	State   job.RunState
	Limit   int
	Offset  int
}

// Store persists jobs and their runs and hands due work to workers. Every
// method must be safe for concurrent use by multiple goroutines.
type Store interface {
	// PutJob inserts or replaces a job.
	PutJob(ctx context.Context, j job.Job) error
	// GetJob returns a job by ID or ErrNotFound.
	GetJob(ctx context.Context, id string) (job.Job, error)
	// ListJobs returns every stored job.
	ListJobs(ctx context.Context) ([]job.Job, error)
	// DeleteJob removes a job by ID.
	DeleteJob(ctx context.Context, id string) error

	// CreateRun stores a new run.
	CreateRun(ctx context.Context, r job.Run) error
	// GetRun returns a run by ID or ErrNotFound.
	GetRun(ctx context.Context, id string) (job.Run, error)
	// UpdateRun persists changes to an existing run.
	UpdateRun(ctx context.Context, r job.Run) error
	// ListRuns returns runs matching the filter, newest first.
	ListRuns(ctx context.Context, f RunFilter) ([]job.Run, error)

	// ClaimDue atomically leases up to limit runs that are ready to execute:
	// pending runs whose scheduled time has passed, plus running runs whose
	// lease has expired (the redelivery path that makes delivery
	// at-least-once). Claimed runs are moved to Running and stamped with the
	// worker ID and a fresh lease expiry.
	ClaimDue(ctx context.Context, now time.Time, worker string, lease time.Duration, limit int) ([]job.Run, error)

	// Close releases resources held by the store.
	Close() error
}
