package main

import (
	"context"
	"log/slog"
	"time"

	"github.com/spf13/cobra"

	"github.com/DoppleDankster/uncaved/internal/config"
	"github.com/DoppleDankster/uncaved/internal/event"
	"github.com/DoppleDankster/uncaved/internal/instrumentation"
	"github.com/DoppleDankster/uncaved/internal/server"
	"github.com/DoppleDankster/uncaved/internal/store"
)

func serveCmd(a *app) *cobra.Command {
	cmd := &cobra.Command{
		Use:               "serve",
		Short:             "Run the API server",
		PersistentPreRunE: loadConfigPreRunE(a),
		RunE: func(cmd *cobra.Command, _ []string) error {
			// The flag overrides config/env, but only when explicitly passed —
			// otherwise its zero value would stomp a config-driven default.
			if cmd.Flags().Changed("auto-migrate") {
				v, err := cmd.Flags().GetBool("auto-migrate")
				if err != nil {
					return err
				}
				a.cfg.DB.AutoMigrate = v
			}
			return runServe(a.cfg)
		},
	}
	cmd.Flags().Bool("auto-migrate", true, "apply pending DB migrations at startup")
	return cmd
}

// runServe wires instrumentation and the server from the already-loaded config.
func runServe(cfg config.Config) error {
	ctx := context.Background()
	shutdown, err := instrumentation.Setup(ctx, cfg.Instrumentation)
	if err != nil {
		return err
	}
	defer func() {
		sctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := shutdown(sctx); err != nil {
			slog.Error("instrumentation shutdown", "error", err)
		}
	}()

	if cfg.DB.AutoMigrate {
		if err := migrateUp(ctx, cfg.DB); err != nil {
			return err
		}
	}

	st, err := store.Open(ctx, cfg.DB)
	if err != nil {
		return err
	}
	defer st.Close()

	slog.Info(
		"starting server",
		"port",
		cfg.Server.Port,
		"environment",
		cfg.Instrumentation.Environment,
	)

	ws := server.NewWebservice(cfg.Server, event.NewHandler(st))
	return ws.Run()
}

// migrateUp applies pending migrations at boot with a short-lived Migrator. Its
// database/sql connection is closed before the server starts — the runtime
// serves off the pgx pool, not this connection.
func migrateUp(ctx context.Context, cfg store.Config) error {
	m, err := store.NewMigrator(cfg)
	if err != nil {
		return err
	}
	defer m.Close()

	slog.InfoContext(ctx, "applying database migrations")
	return m.Up(ctx)
}
