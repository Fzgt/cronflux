package clock_test

import (
	"testing"
	"time"

	"github.com/Fzgt/cronflux/internal/clock"
)

func TestRealClockAdvances(t *testing.T) {
	c := clock.Real{}
	first := c.Now()
	if c.Now().Before(first) {
		t.Error("real clock went backwards")
	}
}

func TestMockClock(t *testing.T) {
	start := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	m := clock.NewMock(start)
	if !m.Now().Equal(start) {
		t.Fatalf("Now = %s, want %s", m.Now(), start)
	}
	m.Advance(90 * time.Minute)
	if want := start.Add(90 * time.Minute); !m.Now().Equal(want) {
		t.Errorf("after advance Now = %s, want %s", m.Now(), want)
	}
	m.Set(start)
	if !m.Now().Equal(start) {
		t.Errorf("after set Now = %s, want %s", m.Now(), start)
	}
}

func TestFromFunc(t *testing.T) {
	fixed := time.Date(2026, 6, 1, 12, 0, 0, 0, time.UTC)
	c := clock.FromFunc(func() time.Time { return fixed })
	if !c.Now().Equal(fixed) {
		t.Errorf("FromFunc Now = %s, want %s", c.Now(), fixed)
	}
}
