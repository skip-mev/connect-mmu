package transformer

import (
	"context"
	"fmt"

	mmtypes "github.com/skip-mev/connect/v2/x/marketmap/types"
	"go.uber.org/zap"

	"github.com/skip-mev/connect-mmu/config"
	"github.com/skip-mev/connect-mmu/generator/types"
)

type Transformer struct {
	logger         *zap.Logger
	feedTransforms []TransformFeed
	mmTransforms   []TransformMarketMap
}

// New creates a new Transformer.
//
// It performs the following chain of transforms:
//  1. Add all NormalizeByPairs
//  2. Resolve any conflicts that may have arisen from prior transformations.
func New(logger *zap.Logger) Transformer {
	return Transformer{
		logger: logger.With(zap.String("service", "transformer")),
		feedTransforms: []TransformFeed{
			InvertOrDrop(), // must invert before normalize
			PruneByLiquidity(),
			PruneByQuoteVolume(),
			PruneByProviderLiquidity(),
			PruneByProviderUsdVolume(),
			ResolveNamingAliases(),
			NormalizeBy(),
			DropFeedsWithoutAggregatorIDs(),
			ResolveCMCConflictsForMarket(),
			ResolveConflictsForProvider(),
			TopFeedsForProvider(),
		},
		mmTransforms: []TransformMarketMap{
			PruneMarkets(),
			ExcludeDisabledProviders(),
			EnableMarkets(),
			ProcessDefiMarkets(),
			PruneInsufficientlyProvidedMarkets(),
			OverrideMinProviderCount(),
			// always override after transforms so they are not overwritten
			OverrideMarkets(),
		},
	}
}

// TransformFeeds runs all feed transformers that are assigned to the Transformer.
func (d *Transformer) TransformFeeds(ctx context.Context, cfg config.GenerateConfig, feeds types.Feeds) (types.Feeds, types.ExclusionReasons, error) {
	dropped := types.NewExclusionReasons()

	for _, t := range d.feedTransforms {
		transformFeeds, transformDrops, err := t(ctx, d.logger, cfg, feeds)
		if err != nil {
			return nil, nil, err
		}
		feeds = transformFeeds
		dropped.Merge(transformDrops)
	}

	return feeds, dropped, nil
}

// TransformMarketMap runs all market map transformers that are assigned to the Transformer.
func (d *Transformer) TransformMarketMap(ctx context.Context, cfg config.GenerateConfig, marketMap mmtypes.MarketMap) (mmtypes.MarketMap, types.ExclusionReasons, error) {
	if marketMap.Markets == nil {
		return mmtypes.MarketMap{}, nil, fmt.Errorf("markets cannot be nil")
	}

	dropped := types.NewExclusionReasons()
	for _, t := range d.mmTransforms {
		transformMM, transformDrops, err := t(ctx, d.logger, cfg, marketMap)
		if err != nil {
			return mmtypes.MarketMap{}, nil, err
		}
		marketMap = transformMM
		dropped.Merge(transformDrops)
	}

	// validate final transform
	return marketMap, dropped, marketMap.ValidateBasic()
}
