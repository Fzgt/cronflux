package cron

import (
	"fmt"
	"strings"
	"time"
)

// ConstantDelaySchedule fires at a fixed interval measured from the previous
// activation rather than from the wall clock. It backs "@every <duration>".
type ConstantDelaySchedule struct {
	Delay time.Duration
}

// Every returns a ConstantDelaySchedule that fires every d. The delay is
// rounded down to whole seconds and clamped to a one-second minimum to match
// cron's granularity.
func Every(d time.Duration) ConstantDelaySchedule {
	if d < time.Second {
		d = time.Second
	}
	d -= d % time.Second
	return ConstantDelaySchedule{Delay: d}
}

// Next returns the next activation, Delay after t, truncated to the second.
func (s ConstantDelaySchedule) Next(t time.Time) time.Time {
	return t.Add(s.Delay - time.Duration(t.Nanosecond()))
}

// parseEvery parses an "@every <duration>" descriptor such as "@every 90s".
func parseEvery(spec string) (Schedule, error) {
	raw := strings.TrimSpace(strings.TrimPrefix(spec, "@every "))
	d, err := time.ParseDuration(raw)
	if err != nil {
		return nil, fmt.Errorf("cron: invalid @every duration %q: %w", raw, err)
	}
	if d <= 0 {
		return nil, fmt.Errorf("cron: @every duration must be positive, got %s", d)
	}
	return Every(d), nil
}
