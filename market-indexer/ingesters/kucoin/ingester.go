package kucoin

import (
	"context"

	"go.uber.org/zap"

	"github.com/skip-mev/connect-mmu/market-indexer/ingesters"
	"github.com/skip-mev/connect-mmu/market-indexer/ingesters/types"
	"github.com/skip-mev/connect-mmu/store/provider"
)

const (
	Name         = "kucoin"
	ProviderName = Name + types.ProviderNameSuffixWS
)

var _ ingesters.Ingester = &Ingester{}

// Ingester is the kucoin implementation of a market data Ingester.
type Ingester struct {
	logger *zap.Logger

	client Client
}

// New creates a new okx Ingester.
func New(logger *zap.Logger) *Ingester {
	if logger == nil {
		panic("cannot set nil logger")
	}

	return &Ingester{
		logger: logger.With(zap.String("ingester", Name)),
		client: NewClient(),
	}
}

// NewWithClient creates a new okx Ingester with the given Client.
func NewWithClient(logger *zap.Logger, client Client) *Ingester {
	if logger == nil {
		panic("cannot set nil logger")
	}

	return &Ingester{
		logger: logger.With(zap.String("ingester", Name)),
		client: client,
	}
}

func (i *Ingester) GetProviderMarkets(ctx context.Context) ([]provider.CreateProviderMarket, error) {
	i.logger.Info("fetching data")

	tickersResp, err := i.client.Tickers(ctx)
	if err != nil {
		return nil, err
	}

	i.logger.Debug("fetched data", zap.Int("num tickers", len(tickersResp.Data.Tickers)))

	pms := make([]provider.CreateProviderMarket, 0, len(tickersResp.Data.Tickers))
	for _, ticker := range tickersResp.Data.Tickers {
		i.logger.Debug("parsing", zap.Any("ticker", ticker))

		pm, err := ticker.toProviderMarket()
		if err != nil {
			i.logger.Error("failed to convert ticker to providerMarket", zap.Error(err))
			continue
		}

		pms = append(pms, pm)
	}

	return pms, nil
}

// Name returns the Ingester's human-readable name.
func (i *Ingester) Name() string {
	return Name
}
