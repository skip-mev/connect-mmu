package kucoin_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/skip-mev/connect-mmu/market-indexer/ingesters/kucoin"
	"github.com/skip-mev/connect-mmu/market-indexer/ingesters/kucoin/mocks"
)

// Test that if kucoin ingester's products endpoint returns an error, the
// GetProviderMarkets method should return an error.
func TestIngesterReturnsErrorOnTickersError(t *testing.T) {
	client := mocks.NewClient(t)
	ingester := kucoin.NewWithClient(zap.NewNop(), client)

	ctx := context.Background()
	client.On("Tickers", ctx).Return(kucoin.TickersResponse{}, fmt.Errorf("error"))

	_, err := ingester.GetProviderMarkets(ctx)
	require.Error(t, err)
}

// Test that the ingester returns tickers from the response.
func TestIngesterTickers(t *testing.T) {
	client := mocks.NewClient(t)
	ingester := kucoin.NewWithClient(zap.NewNop(), client)

	ctx := context.Background()
	client.On("Tickers", ctx).Return(kucoin.TickersResponse{
		Data: kucoin.TickerData{
			Tickers: []kucoin.Ticker{
				{
					Symbol:       "BTC-USDT",
					SymbolName:   "BTC-USDT",
					VolValue:     "10000000",
					AveragePrice: "67000.32",
				},
				{
					Symbol:       "ETH-USDT",
					SymbolName:   "ETH-USDT",
					VolValue:     "60000000",
					AveragePrice: "2309.2390",
				},
			},
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
	require.Equal(t, 67000.32, markets[0].Create.ReferencePrice)

	require.Equal(t, "ETH", markets[1].Create.TargetBase)
	require.Equal(t, "USDT", markets[1].Create.TargetQuote)
	require.Equal(t, float64(60000000), markets[1].Create.QuoteVolume)
	require.Equal(t, 2309.2390, markets[1].Create.ReferencePrice)
}
