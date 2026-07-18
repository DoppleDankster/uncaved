package main

import (
	"context"
	"log/slog"
	"time"

	"github.com/spf13/cobra"

	"github.com/DoppleDankster/uncaved/internal/config"
	"github.com/DoppleDankster/uncaved/internal/instrumentation"
	"github.com/DoppleDankster/uncaved/internal/server"
)

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Run the API server",
	RunE:  runServe,
}

func init() {
	rootCmd.AddCommand(serveCmd)
}

// runServe wires config, instrumentation and the server.
func runServe(cmd *cobra.Command, _ []string) error {
	path, err := cmd.Flags().GetString("config")
	if err != nil {
		return err
	}

	cfg, err := config.Load(path)
	if err != nil {
		return err
	}

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

	slog.Info(
		"starting server",
		"port",
		cfg.Server.Port,
		"environment",
		cfg.Instrumentation.Environment,
	)

	ws := server.NewWerservice(cfg.Server)
	return ws.Run()
}
