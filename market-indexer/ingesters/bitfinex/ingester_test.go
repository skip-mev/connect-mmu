package bitfinex_test

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/skip-mev/connect-mmu/market-indexer/ingesters/bitfinex"
	"github.com/skip-mev/connect-mmu/market-indexer/ingesters/bitfinex/mocks"
)

var (
	rawResp = []byte(`[["tBTCUSD",62668,7.66307317,62677,5.54677145,-1190,-0.01862868,62690,770.47964675,64068,62020],["tLTCUSD",83.614,1417.08762132,83.615,1124.75543941,-1.348,-0.01588499,83.512,2211.76817607,85.76,81.914],["fLTCUSD",83.614,1417.08762132,83.615,1124.75543941,-1.348,-0.01588499,83.512,2211.76817607,85.76,81.914]]`)
	respI   [][]interface{}
)

// Test that if bitfinex ingester's products endpoint returns an error, the
// GetProviderMarkets method should return an error.
func TestIngesterReturnsErrorOnTickersError(t *testing.T) {
	client := mocks.NewClient(t)
	ingester := bitfinex.NewWithClient(zap.NewNop(), client)

	ctx := context.Background()
	client.On("Tickers", ctx).Return(nil, fmt.Errorf("error"))

	_, err := ingester.GetProviderMarkets(ctx)
	require.Error(t, err)
}

// Test that the ingester only returns valid markets that are trading.
func TestIngesterGetsValidMarkets(t *testing.T) {
	client := mocks.NewClient(t)
	ingester := bitfinex.NewWithClient(zap.NewNop(), client)

	require.NoError(t, json.Unmarshal(rawResp, &respI))

	ctx := context.Background()
	// Mock the products endpoint to return a mix of trading-enabled and disabled
	// tickers
	client.On("Tickers", ctx).Return(respI, nil)

	markets, err := ingester.GetProviderMarkets(ctx)
	if err != nil {
		require.NoError(t, err)
	}

	require.Len(t, markets, 2)
	require.Equal(t, "BTC", markets[0].Create.TargetBase)
	require.Equal(t, "USD", markets[0].Create.TargetQuote)
	require.Equal(t, 4.8574118849707e+07, markets[0].Create.QuoteVolume)

	require.Equal(t, "LTC", markets[1].Create.TargetBase)
	require.Equal(t, "USD", markets[1].Create.TargetQuote)
	require.Equal(t, 185428.00857718062, markets[1].Create.QuoteVolume)
}
