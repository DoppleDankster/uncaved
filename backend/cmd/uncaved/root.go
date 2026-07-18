package main

import (
	"os"

	"github.com/spf13/cobra"
)

const defaultConfigPath = "/etc/uncaved/config.toml"

var rootCmd = &cobra.Command{
	Use:          "uncaved",
	Short:        "uncave go backend",
	SilenceUsage: true,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().String("config", defaultConfigPath, "path to TOML config file")
}
