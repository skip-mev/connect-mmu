package coinbase

import (
	"context"
	"fmt"
	"strconv"

	"go.uber.org/zap"

	"github.com/skip-mev/connect-mmu/lib/symbols"
	"github.com/skip-mev/connect-mmu/market-indexer/ingesters/types"
	"github.com/skip-mev/connect-mmu/store/provider"
)

const (
	Name         = "coinbase"
	ProviderName = Name + types.ProviderNameSuffixWS
)

// Ingester is the coinbase implementation of a market data Ingester.
type Ingester struct {
	logger *zap.Logger
	client Client
}

// New creates a new coinbase Ingester.
func New(logger *zap.Logger) *Ingester {
	return &Ingester{
		logger: logger.With(zap.String("ingester", Name)),
		client: NewHTTPCoinbaseClient(),
	}
}

// NewWithClient creates a new coinbase Ingester with a custom client.
func NewWithClient(logger *zap.Logger, client Client) *Ingester {
	return &Ingester{
		logger: logger.With(zap.String("ingester", Name)),
		client: client,
	}
}

// GetProviderMarkets fetches the market data from the coinbase API. Specifically
// we query the volume per ticker on coinbase + meta-data about the ticker. All
// tickers considered must have trading-enabled + status == "online".
func (i *Ingester) GetProviderMarkets(ctx context.Context) ([]provider.CreateProviderMarket, error) {
	// query the all tickers endpoint
	products, err := i.client.Products(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get products: %w", err)
	}

	markets := make([]provider.CreateProviderMarket, 0, len(products))
	for _, product := range products {
		targetBase, err := symbols.ToTickerString(product.Base)
		if err != nil {
			i.logger.Debug("skip creating a ticker", zap.Error(err))
			continue
		}
		targetQuote, err := symbols.ToTickerString(product.Quote)
		if err != nil {
			i.logger.Debug("skip creating a ticker", zap.Error(err))
			continue
		}

		// only consider trading-enabled tickers
		if !product.TradingDisabled && product.Status == "online" {
			markets = append(markets, provider.CreateProviderMarket{
				Create: provider.CreateProviderMarketParams{
					TargetBase:     targetBase,
					TargetQuote:    targetQuote,
					OffChainTicker: product.ID,
					ProviderName:   ProviderName,
				},
			})
		}
	}

	// query the volume per ticker endpoint
	stats, err := i.client.Stats(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get stats: %w for %s-ingester", err, Name)
	}
	i.logger.Debug("fetched stats", zap.Int("num_markets", len(stats)), zap.String("ingester", Name))

	for idx := range markets {
		market := markets[idx]
		if stats, ok := stats[market.Create.OffChainTicker]; ok {
			// get 24 hour spot volume (denominated in base currency)
			volumeInBase, err := strconv.ParseFloat(stats.Stats24Hour.Volume, 64)
			if err != nil {
				return nil, fmt.Errorf("failed to parse volume for %s: %w", market.Create.OffChainTicker, err)
			}

			// get the high / low price for the last 24 hours
			high, err := strconv.ParseFloat(stats.Stats24Hour.High, 64)
			if err != nil {
				return nil, fmt.Errorf("failed to parse high for %s: %w", market.Create.OffChainTicker, err)
			}

			low, err := strconv.ParseFloat(stats.Stats24Hour.Low, 64)
			if err != nil {
				return nil, fmt.Errorf("failed to parse low for %s: %w", market.Create.OffChainTicker, err)
			}

			// scale the volume by the avg(high, low) price
			// this is a rough estimate of the volume in the quote currency
			volumeInQuote := volumeInBase * (high + low) / 2

			// set the volume in the quote currency
			markets[idx].Create.QuoteVolume = volumeInQuote

			refPrice, err := strconv.ParseFloat(stats.Stats24Hour.Last, 64)
			if err != nil {
				return nil, fmt.Errorf("failed to convert Stats24Hour.Last: %w", err)
			}
			markets[idx].Create.ReferencePrice = refPrice

			i.logger.Debug("fetched volume for market", zap.Float64("volume", volumeInQuote), zap.String("market", market.Create.OffChainTicker), zap.String("ingester", Name))

			if err := markets[idx].ValidateBasic(); err != nil {
				return nil, fmt.Errorf("failed to validate market %s: %w", market.Create.OffChainTicker, err)
			}
		}
	}

	return markets, nil
}

// Name returns the Ingester's human-readable name.
func (i *Ingester) Name() string {
	return Name
}
