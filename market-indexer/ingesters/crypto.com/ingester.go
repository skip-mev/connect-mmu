//nolint:revive
package crypto_com

import (
	"context"
	"fmt"

	"go.uber.org/zap"

	"github.com/skip-mev/connect-mmu/market-indexer/ingesters"
	"github.com/skip-mev/connect-mmu/market-indexer/ingesters/types"
	"github.com/skip-mev/connect-mmu/store/provider"
)

const (
	Name         = "crypto_dot_com"
	ProviderName = Name + types.ProviderNameSuffixWS

	InstrumentTypePerpetual = "PERPETUAL_SWAP"
	InstrumentTypeCCYPair   = "CCY_PAIR"
)

var _ ingesters.Ingester = &Ingester{}

// Ingester is the crypto.com implementation of a market data Ingester.
type Ingester struct {
	logger *zap.Logger

	client Client
}

// New creates a new crypto.com Ingester.
func New(logger *zap.Logger) *Ingester {
	if logger == nil {
		panic("cannot set nil logger")
	}

	return &Ingester{
		logger: logger.With(zap.String("ingester", Name)),
		client: NewHTTPClient(),
	}
}

// NewWithClient creates a new crypto.com Ingester with the given Client.
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

	pms := make([]provider.CreateProviderMarket, 0, len(instruments.Result.Data))
	for _, result := range instruments.Result.Data {
		if result.InstType != InstrumentTypeCCYPair || !result.Tradable {
			continue
		}

		ticker, found := tickerMap[result.Symbol]
		if !found {
			return nil, fmt.Errorf("ticker not found for symbol %s", result.Symbol)
		}

		pm, err := result.toProviderMarket(ticker)
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
