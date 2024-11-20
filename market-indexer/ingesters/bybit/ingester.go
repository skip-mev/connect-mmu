package bybit

import (
	"context"
	"fmt"

	"go.uber.org/zap"

	"github.com/skip-mev/connect-mmu/market-indexer/ingesters"
	"github.com/skip-mev/connect-mmu/market-indexer/ingesters/types"
	"github.com/skip-mev/connect-mmu/store/provider"
)

const (
	Name         = "bybit"
	ProviderName = Name + types.ProviderNameSuffixWS

	StatusTrading = "Trading"
)

var _ ingesters.Ingester = &Ingester{}

// Ingester is the bybit implementation of a market data Ingester.
type Ingester struct {
	logger *zap.Logger

	client Client
}

// New creates a new Bybit ingester.
func New(logger *zap.Logger) *Ingester {
	if logger == nil {
		panic("cannot set nil logger")
	}

	return &Ingester{
		logger: logger.With(zap.String("ingester", Name)),
		client: NewHTTPClient(),
	}
}

// NewWithClient creates a new Bybit ingester.
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
		return nil, err
	}

	tickers, err := ig.client.Tickers(ctx)
	if err != nil {
		return nil, err
	}

	tickerMap := tickers.toMap()

	pms := make([]provider.CreateProviderMarket, 0, len(instruments.Result.List))
	for _, item := range instruments.Result.List {
		if item.Status != StatusTrading {
			continue
		}

		ticker, found := tickerMap[item.Symbol]
		if !found {
			return nil, fmt.Errorf("ticker not found for symbol %s", item.Symbol)
		}

		ig.logger.Debug("ticker", zap.Any("data", ticker))

		pm, err := item.toProviderMarket(ticker)
		if err != nil {
			return nil, err
		}
		pms = append(pms, pm)
	}
	return pms, nil
}

// Name returns the Ingester's human-readable name.
func (ig *Ingester) Name() string {
	return Name
}
