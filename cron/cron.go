// Package cron parses cron expressions and reports the times at which they
// next activate.
//
// It understands the standard five-field syntax
//
//	┌─────────── minute        (0-59)
//	│ ┌───────── hour          (0-23)
//	│ │ ┌─────── day of month  (1-31)
//	│ │ │ ┌───── month         (1-12 or JAN-DEC)
//	│ │ │ │ ┌─── day of week   (0-6 or SUN-SAT)
//	│ │ │ │ │
//	* * * * *
//
// an optional leading seconds field, the usual @-descriptors (@yearly,
// @monthly, @weekly, @daily, @hourly) and "@every <duration>".
package cron

import "time"

// Schedule describes a recurring point in time.
type Schedule interface {
	// Next returns the closest instant strictly after t at which the
	// schedule activates. If the schedule never activates again it returns
	// the zero time.Time.
	Next(t time.Time) time.Time
}
