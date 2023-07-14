package config

import (
	"flag"
	"os"
	"strconv"
	"time"
)

// FromEnv overlays CRONFLUX_* environment variables on top of base and returns
// the result. Unset or empty variables leave the base value untouched.
func FromEnv(base Config) Config {
	if v := os.Getenv("CRONFLUX_ADDR"); v != "" {
		base.ListenAddr = v
	}
	if v := os.Getenv("CRONFLUX_STORE"); v != "" {
		base.Backend = Backend(v)
	}
	if v := os.Getenv("CRONFLUX_DATABASE_URL"); v != "" {
		base.DatabaseURL = v
	}
	if v := os.Getenv("CRONFLUX_TICK"); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			base.TickInterval = d
		}
	}
	if v := os.Getenv("CRONFLUX_WORKERS"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			base.Workers = n
		}
	}
	if v := os.Getenv("CRONFLUX_LEASE"); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			base.LeaseDuration = d
		}
	}
	if v := os.Getenv("CRONFLUX_JOBS"); v != "" {
		base.JobsFile = v
	}
	if v := os.Getenv("CRONFLUX_LOG_LEVEL"); v != "" {
		base.LogLevel = v
	}
	return base
}

// Load resolves configuration from defaults, then the environment, then the
// given command-line arguments, in increasing order of precedence. It returns
// an error if the resulting configuration is invalid or the flags cannot be
// parsed.
func Load(args []string) (Config, error) {
	cfg := FromEnv(Default())

	fs := flag.NewFlagSet("cronflux", flag.ContinueOnError)
	addr := fs.String("addr", cfg.ListenAddr, "HTTP listen address")
	backend := fs.String("store", string(cfg.Backend), "storage backend: memory|postgres")
	dbURL := fs.String("database-url", cfg.DatabaseURL, "PostgreSQL connection URL")
	tick := fs.Duration("tick", cfg.TickInterval, "scheduler tick interval")
	workers := fs.Int("workers", cfg.Workers, "number of worker goroutines")
	lease := fs.Duration("lease", cfg.LeaseDuration, "run lease duration")
	jobsFile := fs.String("jobs", cfg.JobsFile, "path to a job definitions file")
	logLevel := fs.String("log-level", cfg.LogLevel, "log level: debug|info|warn|error")

	if err := fs.Parse(args); err != nil {
		return Config{}, err
	}

	cfg.ListenAddr = *addr
	cfg.Backend = Backend(*backend)
	cfg.DatabaseURL = *dbURL
	cfg.TickInterval = *tick
	cfg.Workers = *workers
	cfg.LeaseDuration = *lease
	cfg.JobsFile = *jobsFile
	cfg.LogLevel = *logLevel

	if err := cfg.Validate(); err != nil {
		return Config{}, err
	}
	return cfg, nil
}
