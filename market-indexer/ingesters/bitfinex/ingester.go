package bitfinex

import (
	"context"
	"fmt"

	"go.uber.org/zap"

	"github.com/skip-mev/connect-mmu/lib/symbols"
	"github.com/skip-mev/connect-mmu/market-indexer/ingesters"
	"github.com/skip-mev/connect-mmu/market-indexer/ingesters/types"
	"github.com/skip-mev/connect-mmu/store/provider"
)

const (
	Name         = "bitfinex"
	ProviderName = Name + types.ProviderNameSuffixWS
)

var _ ingesters.Ingester = &Ingester{}

// Ingester is the bitfinex implementation of a market data Ingester.
type Ingester struct {
	logger *zap.Logger

	client Client
}

// New creates a new bitfinex Ingester.
func New(logger *zap.Logger) *Ingester {
	if logger == nil {
		panic("cannot set nil logger")
	}

	return &Ingester{
		logger: logger.With(zap.String("ingester", Name)),
		client: NewHTTPClient(),
	}
}

// NewWithClient creates a new bitfinex Ingester with the given client.
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

	tickers, err := i.client.Tickers(ctx)
	if err != nil {
		return nil, err
	}

	pms := make([]provider.CreateProviderMarket, 0, len(tickers))
	for _, data := range tickers {
		ticker, ok := data[indexSymbol].(string)
		if !ok {
			return nil, fmt.Errorf("received non-string type in response symbol %v", data)
		}

		quoteVolume, err := getVolume(data)
		if err != nil {
			return nil, err
		}

		trimmed, err := checkAndTrimSymbol(ticker)
		if err != nil {
			// skip all non-trading symbols
			continue
		}

		base, quote, err := decodeSymbol(trimmed)
		if err != nil {
			i.logger.Debug("unable to decode symbol", zap.Error(err))
			continue
		}

		lastPrice, ok := data[indexLastPrice].(float64)
		if !ok {
			return nil, fmt.Errorf("received non-float64 type in response lastPrice: %v", data[indexLastPrice])
		}

		targetBase, err := symbols.ToTickerString(replaceAliases(base))
		if err != nil {
			i.logger.Debug("unable to replace base symbol", zap.Error(err))
			continue
		}
		targetQuote, err := symbols.ToTickerString(replaceAliases(quote))
		if err != nil {
			i.logger.Debug("unable to replace quote symbol", zap.Error(err))
			continue
		}

		if ticker[0] == symbolPrefixTrading {
			pm := provider.CreateProviderMarket{
				Create: provider.CreateProviderMarketParams{
					TargetBase:     targetBase,
					TargetQuote:    targetQuote,
					OffChainTicker: trimmed,
					ProviderName:   ProviderName,
					QuoteVolume:    quoteVolume,
					ReferencePrice: lastPrice,
				},
			}
			if err := pm.ValidateBasic(); err != nil {
				return nil, err
			}

			pms = append(pms, pm)
		}

	}

	return pms, nil
}

// Name returns the Ingester's human-readable name.
func (i *Ingester) Name() string {
	return Name
}
