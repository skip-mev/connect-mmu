package generator

import (
	"context"

	mmtypes "github.com/skip-mev/connect/v2/x/marketmap/types"
	"go.uber.org/zap"

	"github.com/skip-mev/connect-mmu/config"
	"github.com/skip-mev/connect-mmu/generator/querier"
	"github.com/skip-mev/connect-mmu/generator/transformer"
	"github.com/skip-mev/connect-mmu/generator/types"
	"github.com/skip-mev/connect-mmu/store/provider"
)

type Generator struct {
	logger *zap.Logger

	q querier.Querier
	t transformer.Transformer
}

func New(logger *zap.Logger, providerStore provider.Store) Generator {
	return Generator{
		logger: logger.With(zap.String("service", "generator")),
		q:      querier.New(logger, providerStore),
		t:      transformer.New(logger),
	}
}

func (g *Generator) GenerateMarketMap(
	ctx context.Context,
	cfg config.GenerateConfig,
) (mmtypes.MarketMap, types.ExclusionReasons, error) {
	feeds, err := g.q.Feeds(ctx, cfg)
	if err != nil {
		g.logger.Error("Unable to query", zap.Error(err))
		return mmtypes.MarketMap{}, nil, err
	}

	g.logger.Info("queried", zap.Int("feeds", len(feeds)))

	transformed, dropped, err := g.t.TransformFeeds(ctx, cfg, feeds)
	if err != nil {
		g.logger.Error("Unable to transform feeds", zap.Error(err))
		return mmtypes.MarketMap{}, nil, err
	}

	g.logger.Info("feed transforms complete", zap.Int("remaining feeds", len(transformed)))

	mm, err := transformed.ToMarketMap()
	if err != nil {
		g.logger.Error("Unable to transform feeds to a MarketMap", zap.Error(err))
		return mmtypes.MarketMap{}, nil, err
	}

	mm, droppedMarkets, err := g.t.TransformMarketMap(ctx, cfg, mm)
	if err != nil {
		g.logger.Error("Unable to transform market map", zap.Error(err))
		return mm, nil, err
	}
	dropped.Merge(droppedMarkets)

	g.logger.Info("final market", zap.Int("size", len(mm.Markets)))

	return mm, dropped, nil
}
