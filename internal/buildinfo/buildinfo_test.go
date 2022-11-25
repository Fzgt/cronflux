package buildinfo

import (
	"runtime"
	"strings"
	"testing"
)

func TestCurrentUsesRuntimeGoVersion(t *testing.T) {
	got := Current()
	if got.GoVersion != runtime.Version() {
		t.Fatalf("GoVersion = %q, want %q", got.GoVersion, runtime.Version())
	}
}

func TestInfoStringContainsFields(t *testing.T) {
	info := Info{Version: "v1.2.3", Commit: "abc123", Date: "2026-01-01", GoVersion: "go1.23"}
	s := info.String()
	for _, want := range []string{"v1.2.3", "abc123", "2026-01-01", "go1.23"} {
		if !strings.Contains(s, want) {
			t.Errorf("String() = %q, missing %q", s, want)
		}
	}
}

func TestDefaultsAreSet(t *testing.T) {
	if Version == "" || Commit == "" || Date == "" {
		t.Fatal("build info defaults must not be empty")
	}
}
