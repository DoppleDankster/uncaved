package main

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/DoppleDankster/uncaved/internal/version"
)

func versionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print build version information",
		Run: func(cmd *cobra.Command, _ []string) {
			fmt.Fprintln(cmd.OutOrStdout(), version.Info())
		},
	}
}
