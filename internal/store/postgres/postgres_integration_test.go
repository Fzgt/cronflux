//go:build integration

package postgres_test

import (
	"context"
	"os"
	"testing"

	"github.com/Fzgt/cronflux/internal/store"
	"github.com/Fzgt/cronflux/internal/store/postgres"
	"github.com/Fzgt/cronflux/internal/store/storetest"
)

// TestPostgresConformance runs the shared store suite against a real database.
// It is guarded by the "integration" build tag and skipped unless
// CRONFLUX_TEST_DATABASE_URL points at a reachable PostgreSQL instance.
func TestPostgresConformance(t *testing.T) {
	dsn := os.Getenv("CRONFLUX_TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("set CRONFLUX_TEST_DATABASE_URL to run PostgreSQL integration tests")
	}

	storetest.Run(t, func(t *testing.T) store.Store {
		t.Helper()
		s, err := postgres.Open(context.Background(), dsn)
		if err != nil {
			t.Fatalf("open postgres: %v", err)
		}
		if err := s.Reset(context.Background()); err != nil {
			t.Fatalf("reset: %v", err)
		}
		t.Cleanup(func() { _ = s.Close() })
		return s
	})
}
