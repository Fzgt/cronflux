package job

import "time"

// Run is a single scheduled execution of a Job. Because cronflux guarantees
// at-least-once delivery a job may produce more than one Run for the same
// slot if a worker's lease expires before it acknowledges completion.
type Run struct {
	ID           string     `json:"id"`
	JobID        string     `json:"job_id"`
	BatchID      string     `json:"batch_id"`
	State        RunState   `json:"state"`
	Attempt      int        `json:"attempt"`
	ScheduledFor time.Time  `json:"scheduled_for"`
	StartedAt    *time.Time `json:"started_at,omitempty"`
	FinishedAt   *time.Time `json:"finished_at,omitempty"`
	LeaseExpiry  time.Time  `json:"lease_expiry"`
	Worker       string     `json:"worker,omitempty"`
	Error        string     `json:"error,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

// Clone returns a copy of the run. Run holds only value fields and pointers to
// immutable time values, so a shallow copy is sufficient.
func (r Run) Clone() Run { return r }
