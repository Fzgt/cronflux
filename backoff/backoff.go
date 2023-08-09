// Package backoff computes how long to wait between retry attempts.
//
// The zero value of Exponential is not useful; use New or set the fields
// explicitly. Delays are deterministic unless a non-zero Jitter is set.
package backoff

import (
	"math"
	"math/rand/v2"
	"time"
)

// Strategy returns the delay to wait before a given (zero-based) retry
// attempt. Attempt 0 is the delay before the first retry.
type Strategy interface {
	Delay(attempt int) time.Duration
}

// Default values applied when a field is left at its zero value.
const (
	defaultBase   = 100 * time.Millisecond
	defaultFactor = 2.0
)

// Exponential is a capped exponential backoff with optional jitter.
type Exponential struct {
	// Base is the delay before the first retry (attempt 0).
	Base time.Duration
	// Max caps the delay; a non-positive value means "no cap".
	Max time.Duration
	// Factor is the multiplier applied per attempt. Values <= 1 fall back
	// to the default of 2.
	Factor float64
	// Jitter is the fraction of the computed delay that is randomised, in
	// the range [0, 1]. Zero means fully deterministic.
	Jitter float64
}

// Delay returns the backoff duration for the given zero-based attempt. The
// growth is base * factor^attempt, capped at Max and clamped so it can never
// overflow time.Duration.
func (e Exponential) Delay(attempt int) time.Duration {
	if attempt < 0 {
		attempt = 0
	}
	base := e.Base
	if base <= 0 {
		base = defaultBase
	}
	factor := e.Factor
	if factor <= 1 {
		factor = defaultFactor
	}

	d := float64(base) * math.Pow(factor, float64(attempt))
	if e.Max > 0 && d > float64(e.Max) {
		d = float64(e.Max)
	}
	if j := e.Jitter; j > 0 {
		if j > 1 {
			j = 1
		}
		// Randomise within [d*(1-j), d] so jitter only ever shortens the
		// delay and can never exceed the cap computed above.
		d = d*(1-j) + rand.Float64()*d*j
	}
	// Saturate rather than converting a float that would overflow int64, which
	// would wrap around to a negative duration.
	if d >= float64(math.MaxInt64) {
		return time.Duration(math.MaxInt64)
	}
	return time.Duration(d)
}
