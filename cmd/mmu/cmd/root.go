package cmd

import (
	"github.com/spf13/cobra"

	"github.com/skip-mev/connect-mmu/cmd/mmu/cmd/basic"
	"github.com/skip-mev/connect-mmu/cmd/mmu/cmd/composite"
	"github.com/skip-mev/connect-mmu/cmd/mmu/cmd/utils"
	"github.com/skip-mev/connect-mmu/cmd/mmu/logging"
	"github.com/skip-mev/connect-mmu/signing"
)

func RootCmd(registry *signing.Registry) *cobra.Command {
	var logLevel string
	rootCmd := &cobra.Command{
		Use:   "mmu",
		Short: "Market Map Updater allows for indexing of asset markets and generation of market maps.",
		RunE: func(cmd *cobra.Command, _ []string) error {
			return cmd.Help()
		},
		PersistentPreRun: func(cmd *cobra.Command, _ []string) {
			logging.ConfigureLogger(logLevel)
			cmd.SetContext(logging.LoggerContext(cmd.Context()))
		},
	}

	// Basic Commands
	rootCmd.AddCommand(
		basic.IndexCmd(),
		basic.GenerateCmd(),
		basic.OverrideCmd(),
		basic.UpsertsCmd(),
		basic.DispatchCmd(registry),
	)

	// Utility Commands
	rootCmd.AddCommand(
		utils.ConfigInitCmd(),
		utils.DiffCmd(),
		utils.ValidateCmd(),
	)

	// Composite Commands
	rootCmd.AddCommand(
		composite.GenerateUpsertsCmd(),
	)

	rootCmd.PersistentFlags().StringVar(&logLevel, "log-level", "info", "Set the logging level (debug, info, warn, error, dpanic, panic, fatal)")

	return rootCmd
}
