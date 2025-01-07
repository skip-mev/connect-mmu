package basic

import (
	"context"
	"errors"
	"fmt"

	"github.com/spf13/cobra"
	"go.uber.org/zap"

	mmtypes "github.com/skip-mev/connect/v2/x/marketmap/types"

	"github.com/skip-mev/connect-mmu/cmd/mmu/logging"
	"github.com/skip-mev/connect-mmu/config"
	"github.com/skip-mev/connect-mmu/diffs"
	"github.com/skip-mev/connect-mmu/generator"
	"github.com/skip-mev/connect-mmu/generator/types"
	"github.com/skip-mev/connect-mmu/lib/file"
	"github.com/skip-mev/connect-mmu/store/provider"
)

func GenerateCmd() *cobra.Command {
	var flags generateCmdFlags

	cmd := &cobra.Command{
		Use:     "generate",
		Short:   "generate market map from market providers",
		Example: "mmu generate --config config.json --provider-data provider-data.json --generated-market-map-out market-map.json --generated-market-map-exclusions-out market-map-exclusions.json",
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			ctx := cmd.Context()

			logger := logging.Logger(ctx)
			defer logger.Sync()

			cfg, err := config.ReadConfig(flags.configPath)
			if err != nil {
				return fmt.Errorf("failed to read in config at %s: %w", flags.configPath, err)
			}

			if cfg.Generate == nil {
				return errors.New("generate configuration missing from mmu config")
			}

			logger.Info("successfully read config", zap.String("path", flags.configPath))

			mm, exclusionReasons, err := GenerateFromConfig(ctx, logger, *cfg.Generate, flags.providerDataPath)
			if err != nil {
				logger.Error("failed to generate marketmap", zap.Error(err))
				return err
			}

			if flags.marketMapOutPath != "" {
				logger.Info("writing markets", zap.String("file", flags.marketMapOutPath))
				if err := file.WriteMarketMapToFile(flags.marketMapOutPath, mm); err != nil {
					return err
				}
			}

			if flags.marketMapExclusionsOutPath != "" {
				logger.Info("writing exclusion reasons", zap.String("file", flags.marketMapExclusionsOutPath))
				// TODO(zrbecker): this file name should be set by a flag.
				if err := diffs.WriteExclusionReasonsToFile(flags.marketMapExclusionsOutPath, exclusionReasons); err != nil {
					return fmt.Errorf("failed to write exclusions to file: %w", err)
				}
			}

			return nil
		},
	}

	generateCmdConfigureFlags(cmd, &flags)

	return cmd
}

type generateCmdFlags struct {
	configPath                 string
	providerDataPath           string
	marketMapOutPath           string
	marketMapExclusionsOutPath string
}

func generateCmdConfigureFlags(cmd *cobra.Command, flags *generateCmdFlags) {
	cmd.Flags().StringVar(&flags.configPath, ConfigPathFlag, ConfigPathDefault, ConfigPathDescription)
	cmd.Flags().StringVar(&flags.providerDataPath, ProviderDataPathFlag, ProviderDataPathDefault, ProviderDataPathDescription)

	cmd.Flags().StringVar(&flags.marketMapOutPath, MarketMapOutPathGeneratedFlag, MarketMapOutPathGeneratedDefault, MarketMapOutPathGenderatedDescription)
	cmd.Flags().StringVar(&flags.marketMapExclusionsOutPath, MarketMapExclusionsOutPathFlag, MarketMapExclusionsOutPathDefault, MarketMapExclusionsOutPathDescription)
}

func GenerateFromConfig(
	ctx context.Context,
	logger *zap.Logger,
	cfg config.GenerateConfig,
	providerPath string,
) (mmtypes.MarketMap, types.ExclusionReasons, error) {
	providerStore, err := provider.NewMemoryStoreFromFile(providerPath)
	if err != nil {
		return mmtypes.MarketMap{}, nil, err
	}

	g := generator.New(logger, providerStore)
	mm, exclusionReasons, err := g.GenerateMarketMap(ctx, cfg)
	if err != nil {
		return mmtypes.MarketMap{}, nil, err
	}

	return mm, exclusionReasons, nil
}
