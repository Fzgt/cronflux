// Package buildinfo exposes version metadata that is stamped into the
// binary at build time via -ldflags. It has no dependencies so it can be
// imported from anywhere, including main, without creating import cycles.
package buildinfo

import (
	"fmt"
	"runtime"
)

// These variables are overridden at build time, e.g.:
//
//	go build -ldflags "-X github.com/Fzgt/cronflux/internal/buildinfo.Version=v0.1.0"
var (
	// Version is the semantic version or git description of the build.
	Version = "dev"
	// Commit is the short git SHA the binary was built from.
	Commit = "none"
	// Date is the RFC3339 build timestamp.
	Date = "unknown"
)

// Info is a structured snapshot of the build metadata.
type Info struct {
	Version   string `json:"version"`
	Commit    string `json:"commit"`
	Date      string `json:"date"`
	GoVersion string `json:"go_version"`
}

// Current returns the build metadata for this binary.
func Current() Info {
	return Info{
		Version:   Version,
		Commit:    Commit,
		Date:      Date,
		GoVersion: runtime.Version(),
	}
}

// String renders the build info as a single human-readable line.
func (i Info) String() string {
	return fmt.Sprintf("cronflux %s (commit %s, built %s, %s)",
		i.Version, i.Commit, i.Date, i.GoVersion)
}
