// Package scheduler drives cronflux: it materializes runs from cron schedules,
// hands them to a pool of workers with at-least-once delivery, retries failures
// with exponential backoff and walks DAG dependencies between jobs.
package scheduler

import (
	"context"
	"os/exec"

	"github.com/Fzgt/cronflux/internal/job"
)

// Executor runs a single job execution.
//
// Implementations must be safe for concurrent use. Because delivery is
// at-least-once an executor may be invoked more than once for the same logical
// run, so side effects should be idempotent where correctness depends on it.
type Executor interface {
	Execute(ctx context.Context, j job.Job, r job.Run) error
}

// ExecutorFunc adapts an ordinary function to the Executor interface.
type ExecutorFunc func(ctx context.Context, j job.Job, r job.Run) error

// Execute calls f.
func (f ExecutorFunc) Execute(ctx context.Context, j job.Job, r job.Run) error {
	return f(ctx, j, r)
}

// NoopExecutor succeeds immediately without doing any work. It is useful for
// testing scheduling behavior in isolation.
type NoopExecutor struct{}

// Execute always succeeds.
func (NoopExecutor) Execute(context.Context, job.Job, job.Run) error { return nil }

// ShellExecutor runs job.Command as a subprocess, inheriting the ambient
// environment. An empty command is treated as a successful no-op.
type ShellExecutor struct{}

// Execute runs the job's command, honoring cancellation via ctx.
func (ShellExecutor) Execute(ctx context.Context, j job.Job, _ job.Run) error {
	if len(j.Command) == 0 {
		return nil
	}
	cmd := exec.CommandContext(ctx, j.Command[0], j.Command[1:]...)
	return cmd.Run()
}
