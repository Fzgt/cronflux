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
