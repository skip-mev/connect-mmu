package utils

import (
	"github.com/spf13/cobra"

	"github.com/skip-mev/connect-mmu/config"
)

func ConfigInitCmd() *cobra.Command {
	var flags configInitFlags

	cmd := &cobra.Command{
		Use:     "config-init",
		Short:   "generate a config with some defaults",
		Example: "mmu config-init --config-out config.json",
		Args:    cobra.NoArgs,
		RunE: func(_ *cobra.Command, _ []string) error {
			cfg := config.DefaultConfig()

			return config.WriteConfig(cfg, flags.configOutPath)
		},
	}

	configInitConfigureFlags(cmd, &flags)

	return cmd
}

type configInitFlags struct {
	configOutPath string
}

func configInitConfigureFlags(cmd *cobra.Command, flags *configInitFlags) {
	cmd.Flags().StringVar(&flags.configOutPath, "config-out", "./tmp/config-default.json", "path to output default config")
}
