// Package clock abstracts the passage of time so components that schedule work
// can be driven deterministically in tests while using the system clock in
// production.
package clock

import (
	"sync"
	"time"
)

// Clock reports the current time.
type Clock interface {
	Now() time.Time
}

// Real is a Clock backed by the system clock.
type Real struct{}

// Now returns the current system time.
func (Real) Now() time.Time { return time.Now() }

// FromFunc adapts a now function to a Clock.
func FromFunc(f func() time.Time) Clock { return funcClock(f) }

type funcClock func() time.Time

func (f funcClock) Now() time.Time { return f() }

// Mock is a Clock whose time is advanced explicitly. It is safe for concurrent
// use.
type Mock struct {
	mu sync.Mutex
	t  time.Time
}

// NewMock returns a Mock starting at t.
func NewMock(t time.Time) *Mock { return &Mock{t: t} }

// Now returns the mock's current time.
func (m *Mock) Now() time.Time {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.t
}

// Set moves the mock clock to t.
func (m *Mock) Set(t time.Time) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.t = t
}

// Advance moves the mock clock forward by d.
func (m *Mock) Advance(d time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.t = m.t.Add(d)
}
