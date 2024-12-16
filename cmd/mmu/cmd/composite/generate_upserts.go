package composite

import (
	"context"
	"errors"
	"fmt"

	"github.com/spf13/cobra"
	"go.uber.org/zap"

	"github.com/skip-mev/connect-mmu/cmd/mmu/cmd/basic"
	"github.com/skip-mev/connect-mmu/cmd/mmu/logging"
	"github.com/skip-mev/connect-mmu/config"
	"github.com/skip-mev/connect-mmu/diffs"
	"github.com/skip-mev/connect-mmu/lib/file"
)

func GenerateUpsertsCmd() *cobra.Command {
	var flags generateUpsertsFlags

	cmd := &cobra.Command{
		Use:     "generate-upserts",
		Short:   "generate upserts from a set of market data",
		Example: "mmu generate-upserts --config config.json --provider-data provider-data.json --upserts-out upserts.json",
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return generateUpserts(cmd.Context(), flags)
		},
	}

	generateUpsertsConfigureFlags(cmd, &flags)

	return cmd
}

type generateUpsertsFlags struct {
	configPath               string
	providerDataPath         string
	updateEnabled            bool
	overwriteProviders       bool
	existingOnly             bool
	disableDeFiMarketMerging bool

	generatedMarketMapOutPath         string
	generatedMarketMapRemovalsOutPath string
	overrideMarketMapOutPath          string
	upsertsOutPath                    string

	writeIntermediate      bool
	warnOnInvalidMarketMap bool
}

func generateUpsertsConfigureFlags(cmd *cobra.Command, flags *generateUpsertsFlags) {
	cmd.Flags().StringVar(&flags.configPath, basic.ConfigPathFlag, basic.ConfigPathDefault, basic.ConfigPathDescription)
	cmd.Flags().StringVar(&flags.providerDataPath, basic.ProviderDataPathFlag, basic.ProviderDataPathDefault, basic.ProviderDataPathDescription)
	cmd.Flags().BoolVar(&flags.updateEnabled, basic.UpdateEnabledFlag, basic.UpdateEnabledDefault, basic.UpdateEnabledDescription)
	cmd.Flags().BoolVar(&flags.overwriteProviders, basic.OverwriteProvidersFlag, basic.OverwriteProvidersDefault, basic.OverwriteProvidersDescription)
	cmd.Flags().BoolVar(&flags.existingOnly, basic.ExistingOnlyFlag, basic.ExistingOnlyDefault, basic.ExistingOnlyDescription)
	cmd.Flags().BoolVar(&flags.warnOnInvalidMarketMap, basic.WarnOnInvalidMarketMapFlag, basic.WarnOnInvalidMarketMapDefault, basic.WarnOnInvalidMarketMapDescription)
	cmd.Flags().BoolVar(&flags.disableDeFiMarketMerging, basic.DisableDeFiMarketMerging, basic.DisableDeFiMarketMergingDefault, basic.DisableDeFiMarketMergingDescription)

	cmd.Flags().StringVar(&flags.generatedMarketMapOutPath, basic.MarketMapOutPathGeneratedFlag, basic.MarketMapOutPathGeneratedDefault, basic.MarketMapOutPathGenderatedDescription)
	cmd.Flags().StringVar(&flags.generatedMarketMapRemovalsOutPath, basic.MarketMapRemovalsOutPathFlag, basic.MarketMapRemovalsOutPathDefault, basic.MarketMapRemovalsOutPathDescription)
	cmd.Flags().StringVar(&flags.overrideMarketMapOutPath, basic.MarketMapOutPathOverrideFlag, basic.MarketMapOutPathOverrideDefault, basic.MarketMapOutPathOverrideDescription)
	cmd.Flags().StringVar(&flags.upsertsOutPath, basic.UpsertsOutPathFlag, basic.UpsertsOutPathDefault, basic.UpsertsOutPathDescription)

	cmd.Flags().BoolVar(&flags.writeIntermediate, WriteIntermediateFlag, WriteIntermediateDefault, WriteIntermediateDescription)
}

func generateUpserts(ctx context.Context, flags generateUpsertsFlags) error {
	logger := logging.Logger(ctx)
	defer logger.Sync()

	cfg, err := config.ReadConfig(flags.configPath)
	if err != nil {
		return fmt.Errorf("failed to read config at %s: %w", flags.configPath, err)
	}

	logger.Info("successfully read config", zap.String("path", flags.configPath))

	// GENERATE
	if cfg.Generate == nil {
		return errors.New("generate configuration missing from mmu config")
	}

	generated, removalReasons, err := basic.GenerateFromConfig(ctx, logger, *cfg.Generate, flags.providerDataPath)
	if err != nil {
		logger.Error("failed to generate marketmap", zap.Error(err))
		return err
	}

	if flags.writeIntermediate {
		logger.Info("writing markets", zap.String("file", flags.generatedMarketMapOutPath))
		if err := file.WriteMarketMapToFile(flags.generatedMarketMapOutPath, generated); err != nil {
			return fmt.Errorf("failed to write generated market map: %w", err)
		}

		logger.Info("writing removal reasons", zap.String("file", flags.generatedMarketMapRemovalsOutPath))
		if err := diffs.WriteRemovalReasonsToFile(flags.generatedMarketMapRemovalsOutPath, removalReasons); err != nil {
			return fmt.Errorf("failed to write removals to file: %w", err)
		}
	}

	// OVERRIDE
	if cfg.Chain == nil {
		return errors.New("chain configuration missing from mmu config")
	}

	overriddenMarketMap, err := basic.OverrideMarketsFromConfig(
		ctx,
		logger,
		*cfg.Chain,
		generated,
		flags.updateEnabled,
		flags.overwriteProviders,
		flags.existingOnly,
		flags.disableDeFiMarketMerging,
	)
	if err != nil {
		return err
	}

	logger.Info("successfully overrode market map with on-chain markets", zap.Int("num markets", len(overriddenMarketMap.Markets)))

	if flags.writeIntermediate {
		logger.Info("writing overridden market map", zap.String("file", flags.overrideMarketMapOutPath))
		err = file.WriteMarketMapToFile(flags.overrideMarketMapOutPath, overriddenMarketMap)
		if err != nil {
			logger.Error("failed to write overridden marketmap", zap.Error(err))
			return err
		}
	}

	// UPSERT
	if cfg.Upsert == nil {
		return errors.New("upsert configuration missing from mmu config")
	}

	upserts, err := basic.UpsertsFromConfigs(
		ctx,
		logger,
		overriddenMarketMap,
		*cfg.Chain,
		*cfg.Upsert,
		flags.warnOnInvalidMarketMap,
	)
	if err != nil {
		return err
	}

	err = file.WriteJSONToFile(flags.upsertsOutPath, upserts)
	if err != nil {
		return fmt.Errorf("failed to write upserts: %w", err)
	}
	logger.Info("upserts written to file", zap.String("file", flags.upsertsOutPath))

	return nil
}
