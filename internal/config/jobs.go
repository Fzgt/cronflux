package config

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/Fzgt/cronflux/internal/job"
)

// backoffSpec mirrors job.RetryPolicy but uses human-friendly duration strings.
type backoffSpec struct {
	Base   string  `json:"base"`
	Max    string  `json:"max"`
	Factor float64 `json:"factor"`
	Jitter float64 `json:"jitter"`
}

// JobSpec is the on-disk representation of a job. Durations are written as Go
// duration strings ("30s", "5m") for readability rather than nanoseconds.
type JobSpec struct {
	ID         string      `json:"id"`
	Name       string      `json:"name"`
	Spec       string      `json:"spec"`
	Command    []string    `json:"command"`
	MaxRetries int         `json:"max_retries"`
	Backoff    backoffSpec `json:"backoff"`
	DependsOn  []string    `json:"depends_on"`
	Timeout    string      `json:"timeout"`
	Enabled    *bool       `json:"enabled"`
}

// jobsFile is the top-level document shape.
type jobsFile struct {
	Jobs []JobSpec `json:"jobs"`
}

// LoadJobs reads a job-definitions file, converts it to domain jobs and checks
// that the declared dependencies form a valid DAG.
func LoadJobs(path string) ([]job.Job, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("config: read jobs file: %w", err)
	}
	var doc jobsFile
	if err := json.Unmarshal(data, &doc); err != nil {
		return nil, fmt.Errorf("config: parse jobs file: %w", err)
	}

	jobs := make([]job.Job, 0, len(doc.Jobs))
	for i, spec := range doc.Jobs {
		j, err := spec.toJob()
		if err != nil {
			return nil, fmt.Errorf("config: job %d (%q): %w", i, spec.ID, err)
		}
		jobs = append(jobs, j)
	}

	if _, err := job.BuildGraph(jobs); err != nil {
		return nil, fmt.Errorf("config: invalid job graph: %w", err)
	}
	return jobs, nil
}

// toJob converts an on-disk spec to a domain job.
func (s JobSpec) toJob() (job.Job, error) {
	if s.ID == "" {
		return job.Job{}, fmt.Errorf("id is required")
	}
	timeout, err := parseOptionalDuration(s.Timeout)
	if err != nil {
		return job.Job{}, fmt.Errorf("timeout: %w", err)
	}
	base, err := parseOptionalDuration(s.Backoff.Base)
	if err != nil {
		return job.Job{}, fmt.Errorf("backoff.base: %w", err)
	}
	maxDelay, err := parseOptionalDuration(s.Backoff.Max)
	if err != nil {
		return job.Job{}, fmt.Errorf("backoff.max: %w", err)
	}

	enabled := true
	if s.Enabled != nil {
		enabled = *s.Enabled
	}
	return job.Job{
		ID:         s.ID,
		Name:       s.Name,
		Spec:       s.Spec,
		Command:    s.Command,
		MaxRetries: s.MaxRetries,
		Backoff: job.RetryPolicy{
			Base:   base,
			Max:    maxDelay,
			Factor: s.Backoff.Factor,
			Jitter: s.Backoff.Jitter,
		},
		DependsOn: s.DependsOn,
		Timeout:   timeout,
		Enabled:   enabled,
	}, nil
}

func parseOptionalDuration(s string) (time.Duration, error) {
	if s == "" {
		return 0, nil
	}
	return time.ParseDuration(s)
}
