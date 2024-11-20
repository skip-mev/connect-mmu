package gate_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/skip-mev/connect-mmu/market-indexer/ingesters/gate"
	"github.com/skip-mev/connect-mmu/market-indexer/ingesters/gate/mocks"
)

// Test that if gate.io ingester's products endpoint returns an error, the
// GetProviderMarkets method should return an error.
func TestIngesterReturnsErrorOnTickersError(t *testing.T) {
	client := mocks.NewClient(t)
	ingester := gate.NewWithClient(zap.NewNop(), client)

	ctx := context.Background()
	client.On("Tickers", ctx).Return(nil, fmt.Errorf("error"))

	_, err := ingester.GetProviderMarkets(ctx)
	require.Error(t, err)
}

// Test that the ingester returns tickers from the response.
// Ignore leverage perp markets.
func TestIngesterTickers(t *testing.T) {
	client := mocks.NewClient(t)
	ingester := gate.NewWithClient(zap.NewNop(), client)

	ctx := context.Background()
	client.On("Tickers", ctx).Return([]gate.TickerData{
		{
			CurrencyPair: "BTC_USDT",
			QuoteVolume:  "10000000",
			LastPrice:    "15.32",
		},
		{
			CurrencyPair: "BTC_ETH",
			QuoteVolume:  "60000000",
			LastPrice:    "103.235",
		},
		{
			CurrencyPair: "BTC3L_ETH",
			QuoteVolume:  "60000000",
			EtfNetValue:  "10",
			LastPrice:    "0.0235",
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
	require.Equal(t, 15.32, markets[0].Create.ReferencePrice)

	require.Equal(t, "BTC", markets[1].Create.TargetBase)
	require.Equal(t, "ETH", markets[1].Create.TargetQuote)
	require.Equal(t, float64(60000000), markets[1].Create.QuoteVolume)
	require.Equal(t, 103.235, markets[1].Create.ReferencePrice)
}
