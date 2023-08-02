package job_test

import (
	"testing"

	"github.com/Fzgt/cronflux/internal/job"
)

func TestRunStateValid(t *testing.T) {
	valid := []job.RunState{
		job.StatePending, job.StateRunning, job.StateSucceeded,
		job.StateFailed, job.StateDead, job.StateSkipped,
	}
	for _, s := range valid {
		if !s.Valid() {
			t.Errorf("%q should be valid", s)
		}
	}
	if job.RunState("bogus").Valid() {
		t.Error("bogus state reported valid")
	}
}

func TestRunStateTerminal(t *testing.T) {
	tests := map[job.RunState]bool{
		job.StatePending:   false,
		job.StateRunning:   false,
		job.StateFailed:    false,
		job.StateSucceeded: true,
		job.StateDead:      true,
		job.StateSkipped:   true,
	}
	for state, want := range tests {
		if got := state.Terminal(); got != want {
			t.Errorf("%q.Terminal() = %v, want %v", state, got, want)
		}
	}
}
