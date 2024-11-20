package okx

import (
	"context"
	"fmt"

	"go.uber.org/zap"

	"github.com/skip-mev/connect-mmu/market-indexer/ingesters"
	"github.com/skip-mev/connect-mmu/market-indexer/ingesters/types"
	"github.com/skip-mev/connect-mmu/store/provider"
)

const (
	Name         = "okx"
	ProviderName = Name + types.ProviderNameSuffixWS
)

var _ ingesters.Ingester = &Ingester{}

// Ingester is the okx implementation of a market data Ingester.
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

func (ig *Ingester) GetProviderMarkets(ctx context.Context) ([]provider.CreateProviderMarket, error) {
	ig.logger.Info("fetching data")

	instruments, err := ig.client.Instruments(ctx)
	if err != nil {
		ig.logger.Error("failed to fetch instruments", zap.Error(err))
		return nil, err
	}

	tickers, err := ig.client.Tickers(ctx)
	if err != nil {
		ig.logger.Error("failed to fetch tickers", zap.Error(err))
		return nil, err
	}

	ig.logger.Info("fetched data", zap.Int("count", len(tickers.Data)))

	pms := make([]provider.CreateProviderMarket, 0, len(instruments.Data))
	tickerMap := tickers.toMap()
	for _, data := range instruments.Data {
		if data.State != StateLive {
			continue
		}

		ticker, found := tickerMap[data.InstID]
		if !found {
			return nil, fmt.Errorf("ticker %s not found in ticker map", data.InstID)
		}

		pm, err := data.toProviderMarket(ticker)
		if err != nil {
			return nil, err
		}

		pms = append(pms, pm)
	}

	ig.logger.Info("creates", zap.Int("count", len(pms)))

	return pms, nil
}

// Name returns the Ingester's human-readable name.
func (ig *Ingester) Name() string {
	return Name
}
