package cron

import (
	"testing"
	"time"
)

// buildSpec assembles a SpecSchedule from raw field strings, failing the test
// on any parse error. It lets the Next tests stay readable.
func buildSpec(t *testing.T, sec, min, hour, dom, month, dow string) *SpecSchedule {
	t.Helper()
	s := &SpecSchedule{Location: time.UTC}
	var err error
	if s.Second, err = parseField(sec, secondsBound); err != nil {
		t.Fatalf("second %q: %v", sec, err)
	}
	if s.Minute, err = parseField(min, minutesBound); err != nil {
		t.Fatalf("minute %q: %v", min, err)
	}
	if s.Hour, err = parseField(hour, hoursBound); err != nil {
		t.Fatalf("hour %q: %v", hour, err)
	}
	if s.Dom, err = parseField(dom, domBound); err != nil {
		t.Fatalf("dom %q: %v", dom, err)
	}
	if s.Month, err = parseField(month, monthBound); err != nil {
		t.Fatalf("month %q: %v", month, err)
	}
	if s.Dow, err = parseField(dow, dowBound); err != nil {
		t.Fatalf("dow %q: %v", dow, err)
	}
	return s
}

func mustTime(t *testing.T, v string) time.Time {
	t.Helper()
	ts, err := time.Parse(time.RFC3339, v)
	if err != nil {
		t.Fatalf("parse time %q: %v", v, err)
	}
	return ts
}

func TestSpecNext(t *testing.T) {
	tests := []struct {
		name                            string
		sec, min, hour, dom, month, dow string
		from, want                      string
	}{
		{"every minute", "0", "*", "*", "*", "*", "*", "2026-01-01T00:00:30Z", "2026-01-01T00:01:00Z"},
		{"daily midnight", "0", "0", "0", "*", "*", "*", "2026-01-01T12:00:00Z", "2026-01-02T00:00:00Z"},
		{"monday 9:30", "0", "30", "9", "*", "*", "1", "2026-01-01T00:00:00Z", "2026-01-05T09:30:00Z"},
		{"yearly jan 1", "0", "0", "0", "1", "1", "*", "2026-06-01T00:00:00Z", "2027-01-01T00:00:00Z"},
		{"dom or dow", "0", "0", "0", "13", "*", "5", "2026-02-01T00:00:00Z", "2026-02-06T00:00:00Z"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := buildSpec(t, tt.sec, tt.min, tt.hour, tt.dom, tt.month, tt.dow)
			got := s.Next(mustTime(t, tt.from))
			want := mustTime(t, tt.want)
			if !got.Equal(want) {
				t.Errorf("Next(%s) = %s, want %s", tt.from, got.Format(time.RFC3339), tt.want)
			}
		})
	}
}

func TestSpecNextImpossibleReturnsZero(t *testing.T) {
	// 30th of February never happens.
	s := buildSpec(t, "0", "0", "0", "30", "2", "*")
	if got := s.Next(mustTime(t, "2026-01-01T00:00:00Z")); !got.IsZero() {
		t.Errorf("expected zero time for impossible spec, got %s", got)
	}
}
