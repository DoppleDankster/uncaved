package main

import (
	"context"
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/DoppleDankster/uncaved/internal/store"
)

// withMigrator builds a Migrator from the loaded config, runs fn with it, and
// always closes it — so every migrate subcommand shares one lifecycle and none
// can leak the underlying sql.DB connection.
func withMigrator(a *app, cmd *cobra.Command, fn func(context.Context, *store.Migrator) error) error {
	m, err := store.NewMigrator(a.cfg.DB)
	if err != nil {
		return err
	}
	defer m.Close()
	return fn(cmd.Context(), m)
}

func migrateStatusCmd(a *app) *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Show migration status",
		RunE: func(cmd *cobra.Command, _ []string) error {
			return withMigrator(a, cmd, func(ctx context.Context, m *store.Migrator) error {
				status, err := m.Status(ctx)
				if err != nil {
					return err
				}
				w := cmd.OutOrStdout()
				fmt.Fprintln(w, "VERSION\tNAME\tAPPLIED")
				for _, s := range status {
					applied := "pending"
					if s.Applied {
						applied = s.AppliedAt.Format(time.RFC3339)
					}
					fmt.Fprintf(w, "%d\t%s\t%s\n", s.Version, s.Name, applied)
				}
				return nil
			})
		},
	}
}

func migrateUpCmd(a *app) *cobra.Command {
	return &cobra.Command{
		Use:   "up",
		Short: "Apply all pending migrations",
		RunE: func(cmd *cobra.Command, _ []string) error {
			return withMigrator(a, cmd, func(ctx context.Context, m *store.Migrator) error {
				return m.Up(ctx)
			})
		},
	}
}

func migrateDownCmd(a *app) *cobra.Command {
	return &cobra.Command{
		Use:   "down",
		Short: "Roll back the most recent migration",
		RunE: func(cmd *cobra.Command, _ []string) error {
			return withMigrator(a, cmd, func(ctx context.Context, m *store.Migrator) error {
				return m.Down(ctx)
			})
		},
	}
}

func migrateCmd(a *app) *cobra.Command {
	cmd := &cobra.Command{
		Use:               "migrate",
		Short:             "Manage database migrations",
		PersistentPreRunE: loadConfigPreRunE(a),
	}
	cmd.AddCommand(
		migrateUpCmd(a),
		migrateDownCmd(a),
		migrateStatusCmd(a),
	)
	return cmd
}
