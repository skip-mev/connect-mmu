package binance_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/skip-mev/connect-mmu/market-indexer/ingesters/binance"
	"github.com/skip-mev/connect-mmu/market-indexer/ingesters/binance/mocks"
)

// Test that if binance ingester's products endpoint returns an error, the
// GetProviderMarkets method should return an error.
func TestIngesterReturnsErrorOnInstrumentsError(t *testing.T) {
	client := mocks.NewClient(t)
	ingester := binance.NewWithClient(zap.NewNop(), client)

	ctx := context.Background()
	client.On("Tickers", ctx).Return(nil, fmt.Errorf("error"))

	_, err := ingester.GetProviderMarkets(ctx)
	require.Error(t, err)
}

func TestIngesterIgnoresSpunDownMarkets(t *testing.T) {
	client := mocks.NewClient(t)
	ingester := binance.NewWithClient(zap.NewNop(), client)

	ctx := context.Background()

	// mock an empty response from the volumes endpoint
	client.On("Tickers", ctx).Return([]binance.TickerData{
		{
			Symbol:      "BTCUSDT",
			HighPrice:   "",
			LowPrice:    "",
			Volume:      "",
			QuoteVolume: "100000",
			LastPrice:   "12.32",
			FirstID:     -1,
			LastID:      -1,
		},
		{
			Symbol:      "ETHUSDT",
			HighPrice:   "",
			LowPrice:    "",
			Volume:      "",
			QuoteVolume: "50000",
			LastPrice:   "12.55",
		},
	}, nil)

	markets, err := ingester.GetProviderMarkets(ctx)
	if err != nil {
		require.NoError(t, err)
	}

	require.Len(t, markets, 1)
	require.Equal(t, "ETH", markets[0].Create.TargetBase)
	require.Equal(t, "USDT", markets[0].Create.TargetQuote)
}

// Test that the ingester only returns markets that are Trading.
func TestIngesterGetsValidMarkets(t *testing.T) {
	client := mocks.NewClient(t)
	ingester := binance.NewWithClient(zap.NewNop(), client)

	ctx := context.Background()

	// mock an empty response from the volumes endpoint
	client.On("Tickers", ctx).Return([]binance.TickerData{
		{
			Symbol:      "BTCUSDT",
			HighPrice:   "",
			LowPrice:    "",
			Volume:      "",
			QuoteVolume: "100000",
			LastPrice:   "12.32",
		},
		{
			Symbol:      "ETHUSDT",
			HighPrice:   "",
			LowPrice:    "",
			Volume:      "",
			QuoteVolume: "50000",
			LastPrice:   "12.55",
		},
	}, nil)

	markets, err := ingester.GetProviderMarkets(ctx)
	if err != nil {
		require.NoError(t, err)
	}

	require.Len(t, markets, 2)
	require.Equal(t, "BTC", markets[0].Create.TargetBase)
	require.Equal(t, "USDT", markets[0].Create.TargetQuote)
	require.Equal(t, float64(100000), markets[0].Create.QuoteVolume)
	require.Equal(t, 12.32, markets[0].Create.ReferencePrice)

	require.Equal(t, "ETH", markets[1].Create.TargetBase)
	require.Equal(t, "USDT", markets[1].Create.TargetQuote)
	require.Equal(t, float64(50000), markets[1].Create.QuoteVolume)
	require.Equal(t, 12.55, markets[1].Create.ReferencePrice)
}
