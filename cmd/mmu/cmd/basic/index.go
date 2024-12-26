package basic

import (
	"errors"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/skip-mev/connect-mmu/cmd/mmu/logging"
	"github.com/skip-mev/connect-mmu/config"
	indexer "github.com/skip-mev/connect-mmu/market-indexer"
	"github.com/skip-mev/connect-mmu/store/provider"
)

func IndexCmd() *cobra.Command {
	var flags indexCmdFlags

	cmd := &cobra.Command{
		Use:     "index",
		Short:   "index markets from configured providers",
		Example: "mmu index --config config.json --provider-data-out provider-data.json",
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			ctx := cmd.Context()

			logger := logging.Logger(ctx)
			logger.Info("indexing markets...")

			cfg, err := config.ReadConfig(flags.configPath)
			if err != nil {
				return fmt.Errorf("failed to read config: %w", err)
			}

			if cfg.Index == nil {
				return errors.New("index configuration missing from mmu config")
			}

			providerStore := provider.NewMemoryStore()

			idx, err := indexer.NewIndexer(*cfg.Index, logger, providerStore, flags.archiveIntermediateSteps)
			if err != nil {
				return err
			}

			if err := idx.Index(ctx); err != nil {
				return err
			}

			if flags.providerDataOutPath != "" {
				logger.Info(fmt.Sprintf("Writing indexed markets to path: %s", flags.providerDataOutPath))
				if err := providerStore.WriteToPath(ctx, flags.providerDataOutPath); err != nil {
					return err
				}
			}

			return nil
		},
	}

	indexCmdConfigureFlags(cmd, &flags)

	return cmd
}

type indexCmdFlags struct {
	configPath               string
	providerDataOutPath      string
	archiveIntermediateSteps bool
}

func indexCmdConfigureFlags(cmd *cobra.Command, flags *indexCmdFlags) {
	cmd.Flags().StringVar(&flags.configPath, ConfigPathFlag, ConfigPathDefault, ConfigPathDescription)
	cmd.Flags().StringVar(&flags.providerDataOutPath, ProviderDataOutPathFlag, ProviderDataOutPathDefault, ProviderDataOutPathDescription)
	cmd.Flags().BoolVar(&flags.archiveIntermediateSteps, ArchiveIntermediateStepsFlag, ArchiveIntermediateStepsDefault, ArchiveIntermediateStepsDescription)
}
