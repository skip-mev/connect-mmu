package coinbase_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/skip-mev/connect-mmu/market-indexer/ingesters/coinbase"
	coinbasemocks "github.com/skip-mev/connect-mmu/market-indexer/ingesters/coinbase/mocks"
)

// Test that if coinbase ingester's products endpoint returns an error, the
// GetProviderMarkets method should return an error.
func TestCoinbaseIngestorReturnsErrorOnProductsError(t *testing.T) {
	client := coinbasemocks.NewClient(t)
	ingester := coinbase.NewWithClient(zap.NewNop(), client)

	ctx := context.Background()
	client.On("Products", ctx).Return(coinbase.Products{}, fmt.Errorf("error"))

	_, err := ingester.GetProviderMarkets(ctx)
	require.Error(t, err)
}

// Test that the coinbase ingester only returns markets with trading enabled
// and status == "online".
func TestCoinbaseIngestorIgnoresNonTradingMarkets(t *testing.T) {
	client := coinbasemocks.NewClient(t)
	ingester := coinbase.NewWithClient(zap.NewNop(), client)

	ctx := context.Background()
	// Mock the products endpoint to return a mix of trading-enabled and disabled
	// tickers
	client.On("Products", ctx).Return(coinbase.Products{
		{
			Base:            "BTC",
			Quote:           "USD",
			ID:              "BTC-USD",
			Status:          "offline", // status is offline
			TradingDisabled: false,
		},
		{
			Base:            "ETH",
			Quote:           "USD",
			ID:              "ETH-USD",
			Status:          "online",
			TradingDisabled: true, // trading is disabled
		},
		{
			Base:            "ETH",
			Quote:           "BTC",
			ID:              "ETH-BTC",
			Status:          "online",
			TradingDisabled: false,
		},
	}, nil)

	// mock an empty response from the volumes endpoint
	client.On("Stats", ctx).Return(coinbase.Stats{}, nil)

	markets, err := ingester.GetProviderMarkets(ctx)
	if err != nil {
		require.NoError(t, err)
	}

	require.Len(t, markets, 1)
	require.Equal(t, "ETH", markets[0].Create.TargetBase)
	require.Equal(t, "BTC", markets[0].Create.TargetQuote)
}

func TestCoinbaseIngestorErrorsOnStatsError(t *testing.T) {
	client := coinbasemocks.NewClient(t)
	ingester := coinbase.NewWithClient(zap.NewNop(), client)

	ctx := context.Background()
	client.On("Products", ctx).Return(coinbase.Products{}, nil)

	client.On("Stats", ctx).Return(coinbase.Stats{}, fmt.Errorf("error"))

	_, err := ingester.GetProviderMarkets(ctx)
	require.Error(t, err)
}

func TestCoinbaseIngestorUpdatesVolumes(t *testing.T) {
	client := coinbasemocks.NewClient(t)
	ingester := coinbase.NewWithClient(zap.NewNop(), client)

	ctx := context.Background()
	// Mock the products endpoint to return a mix of trading-enabled and disabled
	// tickers
	client.On("Products", ctx).Return(coinbase.Products{
		{
			Base:            "BTC",
			Quote:           "USD",
			ID:              "BTC-USD",
			Status:          "offline", // status is offline
			TradingDisabled: false,
		},
		{
			Base:            "ETH",
			Quote:           "USD",
			ID:              "ETH-USD",
			Status:          "online",
			TradingDisabled: true, // trading is disabled
		},
		{
			Base:            "ETH",
			Quote:           "BTC",
			ID:              "ETH-BTC",
			Status:          "online",
			TradingDisabled: false,
		},
	}, nil)

	// mock an empty response from the volumes endpoint
	client.On("Stats", ctx).Return(coinbase.Stats{
		"ETH-BTC": coinbase.StatsPerMarket{
			Stats24Hour: coinbase.Stats24Hour{
				Volume: "1970.38880423",
				High:   "0.04962",
				Low:    "0.04869",
				Last:   "0.0235",
			},
		},
	}, nil)

	markets, err := ingester.GetProviderMarkets(ctx)
	if err != nil {
		require.NoError(t, err)
	}

	require.Len(t, markets, 1)
	require.Equal(t, "ETH", markets[0].Create.TargetBase)
	require.Equal(t, "BTC", markets[0].Create.TargetQuote)
	require.Equal(t, int64(96), int64(markets[0].Create.QuoteVolume))
	require.Equal(t, 0.0235, markets[0].Create.ReferencePrice)
}
