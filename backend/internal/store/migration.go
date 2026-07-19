package store

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"io/fs"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
)

//go:embed migrations/*.sql
var migrationFS embed.FS

type Migrator struct {
	p *goose.Provider
}

type MigrationStatus struct {
	Version   int64
	Name      string
	Applied   bool
	AppliedAt time.Time
}

func NewMigrator(cfg Config) (*Migrator, error) {
	db, err := sql.Open("pgx", cfg.dsn())
	if err != nil {
		return nil, fmt.Errorf("store: open conn: %w", err)
	}
	embedFS, err := fs.Sub(migrationFS, "migrations")
	if err != nil {
		return nil, fmt.Errorf("store: set fs: %w", err)
	}
	p, err := goose.NewProvider(
		"postgres",
		db,
		embedFS,
	)
	if err != nil {
		return nil, fmt.Errorf("store: create migration provider: %w", err)
	}
	return &Migrator{p: p}, nil
}

func (m *Migrator) Up(ctx context.Context) error {
	_, err := m.p.Up(ctx)
	if err != nil {
		return fmt.Errorf("store: migrate up: %w", err)
	}
	return nil
}

func (m *Migrator) Down(ctx context.Context) error {
	_, err := m.p.Down(ctx)
	if err != nil {
		return fmt.Errorf("store: migrate down: %w", err)
	}
	return nil
}

func (m *Migrator) Status(ctx context.Context) ([]MigrationStatus, error) {
	status, err := m.p.Status(ctx)
	if err != nil {
		return nil, fmt.Errorf("store: migration status: %w", err)
	}
	out := make([]MigrationStatus, 0, len(status))
	for _, s := range status {
		out = append(out, MigrationStatus{
			Version:   s.Source.Version,
			Name:      s.Source.Path,
			Applied:   s.State == goose.StateApplied,
			AppliedAt: s.AppliedAt,
		})
	}
	return out, nil
}

func (m *Migrator) Close() error {
	err := m.p.Close()
	if err != nil {
		return fmt.Errorf("store: closing migrator: %w", err)
	}
	return nil
}
