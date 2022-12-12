// Package config holds the runtime configuration for the cronflux server and
// the helpers that load it from flags, the environment and job files.
package config

import (
	"fmt"
	"time"
)

// Backend names a storage implementation.
type Backend string

// Supported storage backends.
const (
	BackendMemory   Backend = "memory"
	BackendPostgres Backend = "postgres"
)

// Config is the fully resolved runtime configuration.
type Config struct {
	ListenAddr    string
	Backend       Backend
	DatabaseURL   string
	TickInterval  time.Duration
	Workers       int
	LeaseDuration time.Duration
	JobsFile      string
	LogLevel      string
}

// Default returns the configuration used when nothing is overridden.
func Default() Config {
	return Config{
		ListenAddr:    ":8080",
		Backend:       BackendMemory,
		TickInterval:  time.Second,
		Workers:       4,
		LeaseDuration: 30 * time.Second,
		LogLevel:      "info",
	}
}

// Validate reports the first configuration problem it finds, if any.
func (c Config) Validate() error {
	switch c.Backend {
	case BackendMemory:
	case BackendPostgres:
		if c.DatabaseURL == "" {
			return fmt.Errorf("config: database URL is required for the %q backend", c.Backend)
		}
	default:
		return fmt.Errorf("config: unknown backend %q", c.Backend)
	}
	if c.Workers < 1 {
		return fmt.Errorf("config: workers must be >= 1, got %d", c.Workers)
	}
	if c.TickInterval <= 0 {
		return fmt.Errorf("config: tick interval must be positive, got %s", c.TickInterval)
	}
	if c.LeaseDuration <= 0 {
		return fmt.Errorf("config: lease duration must be positive, got %s", c.LeaseDuration)
	}
	return nil
}
