package huobi_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/skip-mev/connect-mmu/market-indexer/ingesters/huobi"
	"github.com/skip-mev/connect-mmu/market-indexer/ingesters/huobi/mocks"
)

// Test that if huobi ingester's products endpoint returns an error, the
// GetProviderMarkets method should return an error.
func TestIngesterReturnsErrorOnInstrumentsError(t *testing.T) {
	client := mocks.NewClient(t)
	ingester := huobi.NewWithClient(zap.NewNop(), client)

	ctx := context.Background()
	client.On("Tickers", ctx).Return(huobi.TickersResponse{}, fmt.Errorf("error"))

	_, err := ingester.GetProviderMarkets(ctx)
	require.Error(t, err)
}

// Test that the ingester returns tickers from the response.
func TestIngesterTickers(t *testing.T) {
	client := mocks.NewClient(t)
	ingester := huobi.NewWithClient(zap.NewNop(), client)

	ctx := context.Background()
	// Mock the products endpoint to return a mix of trading-enabled and disabled
	// tickers
	client.On("Tickers", ctx).Return(huobi.TickersResponse{
		Status: huobi.StatusOK,
		Data: []huobi.TickerData{
			{
				Symbol: "btcusdt",
				High:   0,
				Low:    0,
				Vol:    10000000,
			},
			{
				Symbol: "btceth",
				High:   0,
				Low:    0,
				Vol:    60000000,
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
	require.Equal(t, "BTC", markets[1].Create.TargetBase)
	require.Equal(t, "ETH", markets[1].Create.TargetQuote)
	require.Equal(t, float64(60000000), markets[1].Create.QuoteVolume)
}
