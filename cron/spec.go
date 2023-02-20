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

// maxSearchYears bounds how far Next will scan before giving up on an
// impossible expression such as "0 0 30 2 *" (February 30th).
const maxSearchYears = 5

// Next returns the closest instant strictly after t at which the schedule
// fires. If no such instant exists within maxSearchYears it returns the zero
// time.Time.
//
// The algorithm walks the calendar field by field, from the most significant
// (month) to the least (second), advancing to the next candidate whenever a
// field does not match and restarting the scan when a field wraps.
func (s *SpecSchedule) Next(t time.Time) time.Time {
	loc := s.loc()
	origLoc := t.Location()
	t = t.In(loc)

	// Start one second after the current instant, on a second boundary.
	t = t.Truncate(time.Second).Add(time.Second)

	yearLimit := t.Year() + maxSearchYears
	added := false

WRAP:
	if t.Year() > yearLimit {
		return time.Time{}
	}

	for 1<<uint(t.Month())&s.Month == 0 {
		if !added {
			added = true
			t = time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, loc)
		}
		t = t.AddDate(0, 1, 0)
		if t.Month() == time.January {
			goto WRAP
		}
	}

	for !s.dayMatches(t) {
		if !added {
			added = true
			t = time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, loc)
		}
		t = t.AddDate(0, 0, 1)
		if t.Day() == 1 {
			goto WRAP
		}
	}

	for 1<<uint(t.Hour())&s.Hour == 0 {
		if !added {
			added = true
			t = time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), 0, 0, 0, loc)
		}
		t = t.Add(time.Hour)
		if t.Hour() == 0 {
			goto WRAP
		}
	}

	for 1<<uint(t.Minute())&s.Minute == 0 {
		if !added {
			added = true
			t = t.Truncate(time.Minute)
		}
		t = t.Add(time.Minute)
		if t.Minute() == 0 {
			goto WRAP
		}
	}

	for 1<<uint(t.Second())&s.Second == 0 {
		if !added {
			added = true
			t = t.Truncate(time.Second)
		}
		t = t.Add(time.Second)
		if t.Second() == 0 {
			goto WRAP
		}
	}

	return t.In(origLoc)
}

// dayMatches reports whether t's day satisfies the schedule. When both the
// day-of-month and day-of-week fields are restricted a day matches if either
// does (the classic Vixie-cron OR rule); otherwise both must match.
func (s *SpecSchedule) dayMatches(t time.Time) bool {
	domMatch := 1<<uint(t.Day())&s.Dom > 0
	dowMatch := 1<<uint(t.Weekday())&s.Dow > 0
	if s.Dom&starBit > 0 || s.Dow&starBit > 0 {
		return domMatch && dowMatch
	}
	return domMatch || dowMatch
}
