package backoff_test

import (
	"testing"
	"time"

	"github.com/Fzgt/cronflux/backoff"
)

func TestExponentialDelay(t *testing.T) {
	e := backoff.Exponential{Base: 100 * time.Millisecond, Factor: 2, Max: time.Second}
	tests := []struct {
		attempt int
		want    time.Duration
	}{
		{0, 100 * time.Millisecond},
		{1, 200 * time.Millisecond},
		{2, 400 * time.Millisecond},
		{3, 800 * time.Millisecond},
		{4, time.Second},  // capped
		{20, time.Second}, // still capped
	}
	for _, tt := range tests {
		if got := e.Delay(tt.attempt); got != tt.want {
			t.Errorf("Delay(%d) = %s, want %s", tt.attempt, got, tt.want)
		}
	}
}

func TestDelayUsesDefaults(t *testing.T) {
	var e backoff.Exponential // zero value
	if got := e.Delay(0); got != 100*time.Millisecond {
		t.Errorf("default Delay(0) = %s, want 100ms", got)
	}
	if got := e.Delay(1); got != 200*time.Millisecond {
		t.Errorf("default Delay(1) = %s, want 200ms", got)
	}
}

func TestNegativeAttemptTreatedAsZero(t *testing.T) {
	e := backoff.Exponential{Base: time.Second, Factor: 2}
	if got := e.Delay(-3); got != time.Second {
		t.Errorf("Delay(-3) = %s, want 1s", got)
	}
}

func TestJitterStaysWithinBounds(t *testing.T) {
	e := backoff.Exponential{Base: time.Second, Factor: 2, Jitter: 0.5}
	const nominal = 8 * time.Second // e.Delay(3) without jitter
	for i := 0; i < 1000; i++ {
		d := e.Delay(3)
		if d < nominal/2 || d > nominal {
			t.Fatalf("jittered delay %s outside [4s, 8s]", d)
		}
	}
}

func TestDelayDoesNotOverflow(t *testing.T) {
	// A very large attempt must saturate to a positive duration rather than
	// wrapping around to a negative value.
	e := backoff.Exponential{Base: time.Hour, Factor: 10}
	if d := e.Delay(1000); d <= 0 {
		t.Fatalf("uncapped delay overflowed to %s", d)
	}
	// With a cap the delay never exceeds it, no matter how large the attempt.
	capped := backoff.Exponential{Base: time.Second, Factor: 2, Max: time.Minute}
	if got := capped.Delay(1000); got != time.Minute {
		t.Fatalf("capped delay = %s, want 1m", got)
	}
}
