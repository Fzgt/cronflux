package config_test

import (
	"testing"
	"time"

	"github.com/Fzgt/cronflux/internal/config"
)

func TestDefaultIsValid(t *testing.T) {
	if err := config.Default().Validate(); err != nil {
		t.Fatalf("default config invalid: %v", err)
	}
}

func TestValidateErrors(t *testing.T) {
	pg := config.Default()
	pg.Backend = config.BackendPostgres
	if err := pg.Validate(); err == nil {
		t.Error("postgres backend without a database URL should be invalid")
	}

	bad := config.Default()
	bad.Workers = 0
	if err := bad.Validate(); err == nil {
		t.Error("zero workers should be invalid")
	}
}

func TestFromEnvOverrides(t *testing.T) {
	t.Setenv("CRONFLUX_ADDR", ":7000")
	t.Setenv("CRONFLUX_WORKERS", "9")
	t.Setenv("CRONFLUX_TICK", "250ms")

	cfg := config.FromEnv(config.Default())
	if cfg.ListenAddr != ":7000" {
		t.Errorf("addr = %q, want :7000", cfg.ListenAddr)
	}
	if cfg.Workers != 9 {
		t.Errorf("workers = %d, want 9", cfg.Workers)
	}
	if cfg.TickInterval != 250*time.Millisecond {
		t.Errorf("tick = %s, want 250ms", cfg.TickInterval)
	}
}

func TestLoadFlagsBeatEnv(t *testing.T) {
	t.Setenv("CRONFLUX_ADDR", ":7000")

	// With no flags the env value wins over the default.
	cfg, err := config.Load(nil)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.ListenAddr != ":7000" {
		t.Errorf("env addr = %q, want :7000", cfg.ListenAddr)
	}

	// An explicit flag beats the environment.
	cfg, err = config.Load([]string{"-addr", ":9000"})
	if err != nil {
		t.Fatalf("Load with flag: %v", err)
	}
	if cfg.ListenAddr != ":9000" {
		t.Errorf("flag addr = %q, want :9000", cfg.ListenAddr)
	}
}
