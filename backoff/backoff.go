// Package backoff computes how long to wait between retry attempts.
//
// The zero value of Exponential is not useful; use New or set the fields
// explicitly. Delays are deterministic unless a non-zero Jitter is set.
package backoff

import "time"

// Strategy returns the delay to wait before a given (zero-based) retry
// attempt. Attempt 0 is the delay before the first retry.
type Strategy interface {
	Delay(attempt int) time.Duration
}

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
