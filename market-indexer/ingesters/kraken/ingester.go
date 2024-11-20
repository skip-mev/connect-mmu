package kraken

import (
	"context"
	"fmt"

	"go.uber.org/zap"

	"github.com/skip-mev/connect-mmu/market-indexer/ingesters"
	"github.com/skip-mev/connect-mmu/market-indexer/ingesters/types"
	"github.com/skip-mev/connect-mmu/store/provider"
)

const (
	Name         = "kraken"
	ProviderName = Name + types.ProviderNameSuffixAPI
)

var _ ingesters.Ingester = &Ingester{}

// Ingester is the kraken implementation of a market data Ingester.
type Ingester struct {
	logger *zap.Logger

	client Client
}

// New creates a new kraken Ingester.
func New(logger *zap.Logger) *Ingester {
	if logger == nil {
		panic("cannot set nil logger")
	}

	return &Ingester{
		logger: logger.With(zap.String("ingester", Name)),
		client: NewHTTPClient(),
	}
}

// NewWithClient creates a new kraken Ingester with the given Client.
func NewWithClient(logger *zap.Logger, client Client) *Ingester {
	if logger == nil {
		panic("cannot set nil logger")
	}

	return &Ingester{
		logger: logger.With(zap.String("ingester", Name)),
		client: client,
	}
}

func (ig *Ingester) GetProviderMarkets(ctx context.Context) ([]provider.CreateProviderMarket, error) {
	ig.logger.Info("fetching data")

	assets, err := ig.client.AssetPairs(ctx)
	if err != nil {
		return nil, err
	}

	tickers, err := ig.client.Tickers(ctx)
	if err != nil {
		return nil, err
	}

	pms := make([]provider.CreateProviderMarket, 0, len(assets.Result))
	for offChainTicker, result := range assets.Result {
		if result.Status != StatusOnline {
			continue
		}

		data, found := tickers.Result[offChainTicker]
		if !found {
			return nil, fmt.Errorf("ticker %s not found", offChainTicker)
		}

		ig.logger.Debug("ticker", zap.String("offchain ticker", offChainTicker), zap.Any("data", data))

		pm, err := result.toProviderMarket(offChainTicker, data)
		if err != nil {
			return nil, err
		}

		ig.logger.Debug("ticker", zap.String("offchain ticker", offChainTicker), zap.Any("pm", pm))

		pms = append(pms, pm)
	}

	return pms, nil
}

// Name returns the Ingester's human-readable name.
func (ig *Ingester) Name() string {
	return Name
}
