// Package job defines the core scheduling domain: jobs, the runs they produce
// and the dependency graph that orders them.
package job

import "time"

// RetryPolicy configures exponential backoff for a job's retries.
type RetryPolicy struct {
	Base   time.Duration `json:"base"`
	Max    time.Duration `json:"max"`
	Factor float64       `json:"factor"`
	Jitter float64       `json:"jitter"`
}

// Job is a unit of work that cronflux schedules.
//
// A job with a non-empty Spec fires on a cron schedule. A job with an empty
// Spec but a non-empty DependsOn fires only when its upstream jobs complete,
// which is how DAG-style workflows are expressed.
type Job struct {
	ID         string        `json:"id"`
	Name       string        `json:"name"`
	Spec       string        `json:"spec"`
	Command    []string      `json:"command"`
	MaxRetries int           `json:"max_retries"`
	Backoff    RetryPolicy   `json:"backoff"`
	DependsOn  []string      `json:"depends_on"`
	Timeout    time.Duration `json:"timeout"`
	Enabled    bool          `json:"enabled"`
	CreatedAt  time.Time     `json:"created_at"`
	UpdatedAt  time.Time     `json:"updated_at"`
}

// Clone returns a deep copy of the job so callers cannot mutate slices held by
// a store after handing the job over.
func (j Job) Clone() Job {
	out := j
	out.Command = append([]string(nil), j.Command...)
	out.DependsOn = append([]string(nil), j.DependsOn...)
	return out
}
