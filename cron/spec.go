package cron

import "time"

// SpecSchedule is a parsed cron expression. Each field is a bitmask over the
// values that field permits; the day-of-month and day-of-week masks may carry
// the starBit to mark that they were left unrestricted.
type SpecSchedule struct {
	Second uint64
	Minute uint64
	Hour   uint64
	Dom    uint64
	Month  uint64
	Dow    uint64

	// Location is the timezone the schedule is evaluated in. A nil Location
	// is treated as time.UTC.
	Location *time.Location
}

// loc returns the schedule's location, defaulting to UTC.
func (s *SpecSchedule) loc() *time.Location {
	if s.Location == nil {
		return time.UTC
	}
	return s.Location
}
