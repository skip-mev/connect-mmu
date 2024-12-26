package querier

import (
	"context"
	"fmt"

	connecttypes "github.com/skip-mev/connect/v2/pkg/types"
	mmtypes "github.com/skip-mev/connect/v2/x/marketmap/types"
	"go.uber.org/zap"
	"golang.org/x/exp/maps"

	"github.com/skip-mev/connect-mmu/config"
	"github.com/skip-mev/connect-mmu/generator/types"
	"github.com/skip-mev/connect-mmu/store/provider"
	mmutypes "github.com/skip-mev/connect-mmu/types"
)

type Querier struct {
	logger        *zap.Logger
	providerStore provider.Store
}

// New creates a new Querier to read in indexed data to a MemoryStore
func New(logger *zap.Logger, providerStore provider.Store) Querier {
	return Querier{
		logger:        logger.With(zap.String("service", "querier")),
		providerStore: providerStore,
	}
}

func (q *Querier) logConfig(cfg config.GenerateConfig) {
	q.logger.Info("filter", zap.Any("provider", cfg.Providers))
	q.logger.Info("filter", zap.Any("quotes", cfg.Quotes))
	targetQuotes := maps.Keys(cfg.Quotes)
	if len(targetQuotes) != 0 {
		q.logger.Info("filter", zap.Any("target quotes", targetQuotes))
	}
}

func (q *Querier) Feeds(ctx context.Context, cfg config.GenerateConfig) (types.Feeds, error) {
	q.logConfig(cfg)

	args := provider.GetFilteredProviderMarketsParams{
		ProviderNames: maps.Keys(cfg.Providers),
	}
	q.logger.Info("query", zap.Any("args", args))
	rows, err := q.providerStore.GetProviderMarkets(ctx, args)
	if err != nil {
		return nil, fmt.Errorf("query failed: %w", err)
	}

	feeds := make(types.Feeds, 0, len(rows))
	for _, row := range rows {
		feed, err := toFeed(row, cfg)
		if err != nil {
			q.logger.Error("failed to convert row to feed", zap.Error(err), zap.Any("row", row))
			return nil, fmt.Errorf("failed to convert row to feed: %w", err)
		}

		feeds = append(feeds, feed)
	}

	return feeds, nil
}

func toFeed(pm provider.GetFilteredProviderMarketsRow, cfg config.GenerateConfig) (types.Feed, error) {
	// use provider name -> isCex or isDex -> minProviderCount
	var minProviderCount uint64
	if cfg.IsProviderDefi(pm.ProviderName) {
		minProviderCount = cfg.MinDexProviderCount
	} else {
		minProviderCount = cfg.MinCexProviderCount
	}

	ticker := mmtypes.Ticker{
		CurrencyPair:     connecttypes.NewCurrencyPair(pm.TargetBase, pm.TargetQuote),
		Decimals:         8,
		MinProviderCount: minProviderCount,
		Enabled:          false,
		Metadata_JSON:    "",
	}

	providerConfig := mmtypes.ProviderConfig{
		Name:            pm.ProviderName,
		OffChainTicker:  pm.OffChainTicker,
		NormalizeByPair: nil,
		Invert:          false,
		Metadata_JSON:   string(pm.MetadataJSON),
	}

	cmcInfo := mmutypes.NewCoinMarketCapInfo(pm.BaseCmcID, pm.QuoteCmcID, pm.BaseRank, pm.QuoteRank)

	liquidityInfo := mmutypes.LiquidityInfo{
		NegativeDepthTwo: pm.NegativeDepthTwo,
		PositiveDepthTwo: pm.PositiveDepthTwo,
	}

	return types.NewFeed(
		ticker,
		providerConfig,
		pm.QuoteVolume,
		pm.UsdVolume,
		pm.ReferencePrice,
		liquidityInfo,
		cmcInfo,
	), nil
}
