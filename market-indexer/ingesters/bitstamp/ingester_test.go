package bitstamp_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/skip-mev/connect-mmu/market-indexer/ingesters/bitstamp"
	"github.com/skip-mev/connect-mmu/market-indexer/ingesters/bitstamp/mocks"
)

// Test that if bitstamp ingester's products endpoint returns an error, the
// GetProviderMarkets method should return an error.
func TestIngesterReturnsErrorOnTickersError(t *testing.T) {
	client := mocks.NewClient(t)
	ingester := bitstamp.NewWithClient(zap.NewNop(), client)

	ctx := context.Background()
	client.On("Tickers", ctx).Return(nil, fmt.Errorf("error"))

	_, err := ingester.GetProviderMarkets(ctx)
	require.Error(t, err)
}

// Test that the ingester parses tickers properly.
func TestIngesterParse(t *testing.T) {
	client := mocks.NewClient(t)
	ingester := bitstamp.NewWithClient(zap.NewNop(), client)

	ctx := context.Background()
	client.On("Tickers", ctx).Return([]bitstamp.TickerData{
		{
			Pair:      "BTC/USD",
			Volume:    "40450",
			High:      "10000",
			Low:       "5000",
			OpenPrice: "123.01",
		},
		{
			Pair:      "ETH/USD",
			Volume:    "48022",
			High:      "1000",
			Low:       "500",
			OpenPrice: "15.23",
		},
	}, nil)

	markets, err := ingester.GetProviderMarkets(ctx)
	if err != nil {
		require.NoError(t, err)
	}

	require.Len(t, markets, 2)
	require.Equal(t, "BTC", markets[0].Create.TargetBase)
	require.Equal(t, "USD", markets[0].Create.TargetQuote)
	require.Equal(t, 3.03375e+08, markets[0].Create.QuoteVolume)
	require.Equal(t, 123.01, markets[0].Create.ReferencePrice)

	require.Equal(t, "ETH", markets[1].Create.TargetBase)
	require.Equal(t, "USD", markets[1].Create.TargetQuote)
	require.Equal(t, 3.60165e+07, markets[1].Create.QuoteVolume)
	require.Equal(t, 15.23, markets[1].Create.ReferencePrice)
}
