package cron

import "testing"

func TestParseFieldWildcard(t *testing.T) {
	got, err := parseField("*", minutesBound)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got&starBit == 0 {
		t.Errorf("wildcard should set the star bit")
	}
	for i := uint(0); i <= 59; i++ {
		if got&(1<<i) == 0 {
			t.Errorf("minute %d should be allowed by wildcard", i)
		}
	}
}

func TestParseFieldSingle(t *testing.T) {
	tests := []struct {
		field string
		bit   uint
	}{
		{"0", 0},
		{"5", 5},
		{"59", 59},
	}
	for _, tt := range tests {
		t.Run(tt.field, func(t *testing.T) {
			got, err := parseField(tt.field, minutesBound)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != 1<<tt.bit {
				t.Errorf("parseField(%q) = %b, want bit %d", tt.field, got, tt.bit)
			}
		})
	}
}

func TestParseFieldInvalid(t *testing.T) {
	if _, err := parseField("abc", minutesBound); err == nil {
		t.Error("expected error for non-numeric field")
	}
}

func TestParseFieldOutOfRange(t *testing.T) {
	cases := []struct {
		field string
		b     bounds
	}{
		{"60", minutesBound},
		{"24", hoursBound},
		{"0", domBound},   // day-of-month starts at 1
		{"13", monthBound},
		{"1-70", minutesBound},
		{"5-1", hoursBound}, // inverted range
	}
	for _, tc := range cases {
		if _, err := parseField(tc.field, tc.b); err == nil {
			t.Errorf("parseField(%q) should have failed", tc.field)
		}
	}
}
