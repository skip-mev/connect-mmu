package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "sentry",
	Short: "health report via Connect log ingestion",
	Long: "Sentry watches a Connect instance by ingesting logs and tallying successful and failed price updates. " +
		"Sentry will output a health.json at the end of the run.",
	RunE: func(cmd *cobra.Command, _ []string) error {
		return cmd.Help()
	},
}

func main() {
	rootCmd.AddCommand(runCommand())
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
