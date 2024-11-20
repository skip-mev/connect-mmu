package mexc_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/skip-mev/connect-mmu/market-indexer/ingesters/mexc"
	"github.com/skip-mev/connect-mmu/market-indexer/ingesters/mexc/mocks"
)

// Test that if gate.io ingester's products endpoint returns an error, the
// GetProviderMarkets method should return an error.
func TestIngesterReturnsErrorOnTickersError(t *testing.T) {
	client := mocks.NewClient(t)
	ingester := mexc.NewWithClient(zap.NewNop(), client)

	ctx := context.Background()
	client.On("Tickers", ctx).Return(nil, fmt.Errorf("error"))

	_, err := ingester.GetProviderMarkets(ctx)
	require.Error(t, err)
}

// Test that the ingester returns tickers from the response.
func TestIngesterTickers(t *testing.T) {
	client := mocks.NewClient(t)
	ingester := mexc.NewWithClient(zap.NewNop(), client)

	ctx := context.Background()
	client.On("Tickers", ctx).Return([]mexc.TickerData{
		{
			Symbol:      "BTCUSDT",
			QuoteVolume: "10000000",
			OpenPrice:   "68000.32",
		},
		{
			Symbol:      "BTCETH",
			QuoteVolume: "60000000",
			OpenPrice:   "0.023",
		},
	}, nil)

	markets, err := ingester.GetProviderMarkets(ctx)
	if err != nil {
		require.NoError(t, err)
	}

	require.Len(t, markets, 2)
	require.Equal(t, "BTC", markets[0].Create.TargetBase)
	require.Equal(t, "USDT", markets[0].Create.TargetQuote)
	require.Equal(t, float64(10000000), markets[0].Create.QuoteVolume)
	require.Equal(t, 68000.32, markets[0].Create.ReferencePrice)

	require.Equal(t, "BTC", markets[1].Create.TargetBase)
	require.Equal(t, "ETH", markets[1].Create.TargetQuote)
	require.Equal(t, float64(60000000), markets[1].Create.QuoteVolume)
	require.Equal(t, 0.023, markets[1].Create.ReferencePrice)
}
