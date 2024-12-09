package basic

import (
	"context"
	"errors"
	"fmt"

	"github.com/spf13/cobra"
	"go.uber.org/zap"

	mmtypes "github.com/skip-mev/connect/v2/x/marketmap/types"

	"github.com/skip-mev/connect-mmu/client/dydx"
	marketmapclient "github.com/skip-mev/connect-mmu/client/marketmap"
	"github.com/skip-mev/connect-mmu/cmd/mmu/logging"
	"github.com/skip-mev/connect-mmu/config"
	"github.com/skip-mev/connect-mmu/lib/file"
	"github.com/skip-mev/connect-mmu/override"
	"github.com/skip-mev/connect-mmu/override/update"
)

func OverrideCmd() *cobra.Command {
	var flags overrideCmdFlags

	cmd := &cobra.Command{
		Use:     "override",
		Short:   "override all markets using an on-chain market map",
		Example: "mmu override --config config.json --market-map market-map.json --update-enabled false --overwrite-providers false --existing-only false",
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			ctx := cmd.Context()

			logger := logging.Logger(ctx)

			cfg, err := config.ReadConfig(flags.configPath)
			if err != nil {
				return fmt.Errorf("failed to read chain config file: %w", err)
			}

			if cfg.Chain == nil {
				return errors.New("override configuration missing from mmu config")
			}

			logger.Info("successfully read chain config", zap.String("path", flags.configPath))

			fileMarketMap, err := file.ReadMarketMapFromFile(flags.marketMapPath)
			if err != nil {
				logger.Error("failed to read marketmap", zap.Error(err))
				return err
			}

			logger.Info("successfully read marketmap", zap.String("path", flags.marketMapPath), zap.Int("num markets", len(fileMarketMap.Markets)))

			overriddenMarketMap, err := OverrideMarketsFromConfig(
				ctx,
				logger,
				*cfg.Chain,
				fileMarketMap,
				flags.updateEnabled,
				flags.overwriteProviders,
				flags.existingOnly,
			)
			if err != nil {
				return err
			}

			logger.Info("successfully overrode market map with on-chain markets", zap.Int("num markets", len(overriddenMarketMap.Markets)))

			err = file.WriteMarketMapToFile(flags.marketMapOutPath, overriddenMarketMap)
			if err != nil {
				logger.Error("failed to write marketmap", zap.Error(err))
				return err
			}

			logger.Info("successfully wrote marketmap", zap.String("path", flags.marketMapOutPath))

			return nil
		},
	}

	overrideCmdConfigureFlags(cmd, &flags)

	return cmd
}

type overrideCmdFlags struct {
	configPath         string
	marketMapPath      string
	marketMapOutPath   string
	updateEnabled      bool
	overwriteProviders bool
	existingOnly       bool
}

func overrideCmdConfigureFlags(cmd *cobra.Command, flags *overrideCmdFlags) {
	cmd.Flags().StringVar(&flags.configPath, ConfigPathFlag, ConfigPathDefault, ConfigPathDescription)
	cmd.Flags().StringVar(&flags.marketMapPath, MarketMapGeneratedFlag, MarketMapGeneratedDefault, MarketMapGeneratedDescription)
	cmd.Flags().BoolVar(&flags.updateEnabled, UpdateEnabledFlag, UpdateEnabledDefault, UpdateEnabledDescription)
	cmd.Flags().BoolVar(&flags.overwriteProviders, OverwriteProvidersFlag, OverwriteProvidersDefault, OverwriteProvidersDescription)
	cmd.Flags().BoolVar(&flags.existingOnly, ExistingOnlyFlag, ExistingOnlyDefault, ExistingOnlyDescription)

	cmd.Flags().StringVar(&flags.marketMapOutPath, MarketMapOutPathOverrideFlag, MarketMapOutPathOverrideDefault, MarketMapOutPathOverrideDescription)
}

func OverrideMarketsFromConfig(
	ctx context.Context,
	logger *zap.Logger,
	cfg config.ChainConfig,
	generated mmtypes.MarketMap,
	updateEnabled, overwriteProviders, existingOnly bool,
) (mmtypes.MarketMap, error) {
	// create client based on config
	mmClient, err := marketmapclient.NewClientFromChainConfig(logger, cfg)
	if err != nil {
		logger.Error("failed to create marketmap client", zap.Error(err))
		return mmtypes.MarketMap{}, err
	}

	onChainMarketMap, err := mmClient.GetMarketMap(ctx)
	if err != nil {
		logger.Error("failed to get marketmap from chain", zap.Error(err))
		return mmtypes.MarketMap{}, err
	}

	logger.Info("successfully got on chain marketmap", zap.Int("num markets", len(onChainMarketMap.Markets)))

	// create override method based on config
	marketOverride := override.NewCoreOverride()
	if cfg.DYDX {
		marketOverride, err = override.NewDyDxOverride(dydx.NewHTTPClient(cfg.RESTAddress))
		if err != nil {
			logger.Error("failed to create dydx override", zap.Error(err))
			return mmtypes.MarketMap{}, err
		}
	}

	overriddenMarketMap, err := marketOverride.OverrideGeneratedMarkets(
		ctx,
		logger,
		onChainMarketMap,
		generated,
		update.Options{
			UpdateEnabled:      updateEnabled,
			OverwriteProviders: overwriteProviders,
			ExistingOnly:       existingOnly,
		},
	)
	if err != nil {
		logger.Error("failed to override marketmap", zap.Error(err))
		return mmtypes.MarketMap{}, err
	}

	return overriddenMarketMap, nil
}
