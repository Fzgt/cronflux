package cron

import (
	"testing"
	"time"
)

func TestParseNext(t *testing.T) {
	tests := []struct {
		spec string
		from string
		want string
	}{
		{"@hourly", "2026-03-10T10:15:00Z", "2026-03-10T11:00:00Z"},
		{"@daily", "2026-03-10T10:15:00Z", "2026-03-11T00:00:00Z"},
		{"*/15 * * * *", "2026-03-10T10:04:00Z", "2026-03-10T10:15:00Z"},
		{"0 9 * * 1-5", "2026-03-14T12:00:00Z", "2026-03-16T09:00:00Z"}, // Sat -> Mon
		{"0 0 1 JAN *", "2026-06-01T00:00:00Z", "2027-01-01T00:00:00Z"},
		{"@every 90m", "2026-03-10T10:00:00Z", "2026-03-10T11:30:00Z"},
	}
	for _, tt := range tests {
		t.Run(tt.spec, func(t *testing.T) {
			sched, err := Parse(tt.spec)
			if err != nil {
				t.Fatalf("Parse(%q): %v", tt.spec, err)
			}
			from, _ := time.Parse(time.RFC3339, tt.from)
			want, _ := time.Parse(time.RFC3339, tt.want)
			if got := sched.Next(from); !got.Equal(want) {
				t.Errorf("Next = %s, want %s", got.Format(time.RFC3339), tt.want)
			}
		})
	}
}

func TestParseErrors(t *testing.T) {
	bad := []string{
		"",
		"* * *",         // too few fields
		"* * * * * * *", // too many fields
		"@bogus",        // unknown descriptor
		"@every notaduration",
	}
	for _, spec := range bad {
		if _, err := Parse(spec); err == nil {
			t.Errorf("Parse(%q) expected error", spec)
		}
	}
}
