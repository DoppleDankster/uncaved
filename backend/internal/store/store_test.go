package store

import (
	"context"
	"net"
	"strconv"
	"testing"
	"time"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

// newTestStore spins up a throwaway Postgres container, runs every migration
// against it, and returns a connected *Store. Container and pool are torn down
// via t.Cleanup, so each caller gets an isolated, fully-migrated database.
//
// Requires a running Docker daemon; `go test -short` skips these tests so the
// unit suite stays Docker-free.
func newTestStore(t *testing.T) *Store {
	t.Helper()
	if testing.Short() {
		t.Skip("skipping: requires Docker (testcontainers)")
	}
	ctx := context.Background()

	container, err := postgres.Run(ctx,
		"postgres:17-alpine",
		postgres.WithDatabase("uncaved_test"),
		postgres.WithUsername("test"),
		postgres.WithPassword("test"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(60*time.Second),
		),
	)
	if err != nil {
		t.Fatalf("start postgres container: %v", err)
	}
	t.Cleanup(func() {
		if err := container.Terminate(ctx); err != nil {
			t.Logf("terminate container: %v", err)
		}
	})

	// Endpoint returns "host:port" for the mapped 5432 — split it rather than
	// depend on the container port type's method set (which shifts across
	// testcontainers versions).
	endpoint, err := container.Endpoint(ctx, "")
	if err != nil {
		t.Fatalf("container endpoint: %v", err)
	}
	host, portStr, err := net.SplitHostPort(endpoint)
	if err != nil {
		t.Fatalf("split endpoint %q: %v", endpoint, err)
	}
	port, err := strconv.Atoi(portStr)
	if err != nil {
		t.Fatalf("parse port %q: %v", portStr, err)
	}

	cfg := Config{
		Host:     host,
		Port:     port,
		Username: "test",
		Password: "test",
		Database: "uncaved_test",
	}

	// Migrate with the real Migrator (its own database/sql conn), then close it —
	// the Store serves off its own pgx pool.
	migrator, err := NewMigrator(cfg)
	if err != nil {
		t.Fatalf("new migrator: %v", err)
	}
	if err := migrator.Up(ctx); err != nil {
		t.Fatalf("migrate up: %v", err)
	}
	if err := migrator.Close(); err != nil {
		t.Fatalf("close migrator: %v", err)
	}

	st, err := Open(ctx, cfg)
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	t.Cleanup(st.Close)

	return st
}
