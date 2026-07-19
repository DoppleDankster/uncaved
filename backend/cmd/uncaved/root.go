package main

import (
	"os"

	"github.com/DoppleDankster/uncaved/internal/config"
	"github.com/spf13/cobra"
)

const defaultConfigPath = "/etc/uncaved/config.toml"

// newRootCmd builds the command tree: the root, its persistent --config flag,
// and every subcommand. Constructing commands with functions instead of
// package-level vars + init() keeps the wiring explicit and each command
// independently testable.
func newRootCmd() *cobra.Command {
	a := &app{}
	// No PersistentPreRunE here: config loading (and DB validation) is scoped to
	// the commands that need a database — serve and migrate — so `version` and
	// `help` run without requiring a valid DB config.
	cmd := &cobra.Command{
		Use:          "uncaved",
		Short:        "uncave go backend",
		SilenceUsage: true,
	}

	cmd.PersistentFlags().String("config", defaultConfigPath, "path to TOML config file")

	cmd.AddCommand(
		serveCmd(a),
		versionCmd(),
		migrateCmd(a),
	)

	return cmd
}

func Execute() {
	if err := newRootCmd().Execute(); err != nil {
		os.Exit(1)
	}
}

type app struct {
	cfg config.Config
}

func loadConfigPreRunE(a *app) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		path, err := cmd.Flags().GetString("config")
		if err != nil {
			return err
		}
		a.cfg, err = config.Load(path)
		return err
	}
}
