package cron

import (
	"fmt"
	"strings"
	"time"
)

// Parse parses a cron specification and returns a Schedule evaluated in UTC.
// It accepts the standard five-field form (minute hour dom month dow), the
// six-field form with a leading seconds field, and the @-descriptors.
func Parse(spec string) (Schedule, error) {
	return ParseInLocation(spec, time.UTC)
}

// ParseInLocation is like Parse but evaluates the schedule in loc.
func ParseInLocation(spec string, loc *time.Location) (Schedule, error) {
	spec = strings.TrimSpace(spec)
	if spec == "" {
		return nil, fmt.Errorf("cron: empty spec")
	}
	if strings.HasPrefix(spec, "@") {
		return parseDescriptor(spec, loc)
	}
	return parseFields(strings.Fields(spec), loc)
}

// descriptors maps the common @-shortcuts to their equivalent six-field spec.
var descriptors = map[string]string{
	"@yearly":   "0 0 0 1 1 *",
	"@annually": "0 0 0 1 1 *",
	"@monthly":  "0 0 0 1 * *",
	"@weekly":   "0 0 0 * * 0",
	"@daily":    "0 0 0 * * *",
	"@midnight": "0 0 0 * * *",
	"@hourly":   "0 0 * * * *",
}

// parseDescriptor resolves an @-prefixed schedule.
func parseDescriptor(spec string, loc *time.Location) (Schedule, error) {
	if strings.HasPrefix(spec, "@every ") {
		return parseEvery(spec)
	}
	if expanded, ok := descriptors[spec]; ok {
		return parseFields(strings.Fields(expanded), loc)
	}
	return nil, fmt.Errorf("cron: unrecognised descriptor %q", spec)
}

// parseFields builds a SpecSchedule from an already-split field list.
func parseFields(fields []string, loc *time.Location) (*SpecSchedule, error) {
	var sec, min, hour, dom, month, dow string
	switch len(fields) {
	case 5:
		sec = "0"
		min, hour, dom, month, dow = fields[0], fields[1], fields[2], fields[3], fields[4]
	case 6:
		sec, min, hour, dom, month, dow = fields[0], fields[1], fields[2], fields[3], fields[4], fields[5]
	default:
		return nil, fmt.Errorf("cron: expected 5 or 6 fields, got %d in %q", len(fields), strings.Join(fields, " "))
	}

	s := &SpecSchedule{Location: loc}
	var err error
	if s.Second, err = parseField(sec, secondsBound); err != nil {
		return nil, err
	}
	if s.Minute, err = parseField(min, minutesBound); err != nil {
		return nil, err
	}
	if s.Hour, err = parseField(hour, hoursBound); err != nil {
		return nil, err
	}
	if s.Dom, err = parseField(dom, domBound); err != nil {
		return nil, err
	}
	if s.Month, err = parseField(month, monthBound); err != nil {
		return nil, err
	}
	if s.Dow, err = parseField(dow, dowBound); err != nil {
		return nil, err
	}
	s.Dow = normalizeDow(s.Dow)
	return s, nil
}

// normalizeDow folds the "Sunday is 7" convention onto "Sunday is 0" so both
// spellings behave identically.
func normalizeDow(mask uint64) uint64 {
	if mask&(1<<7) != 0 {
		mask = (mask &^ (1 << 7)) | 1
	}
	return mask
}
