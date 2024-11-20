package kraken_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/skip-mev/connect-mmu/market-indexer/ingesters/kraken"
	"github.com/skip-mev/connect-mmu/market-indexer/ingesters/kraken/mocks"
)

// Test that if kraken ingester's asset pairs endpoint returns an error, the
// GetProviderMarkets method should return an error.
func TestIngesterReturnsErrorOnAssetPairsError(t *testing.T) {
	client := mocks.NewClient(t)
	ingester := kraken.NewWithClient(zap.NewNop(), client)

	ctx := context.Background()
	client.On("AssetPairs", ctx).Return(kraken.AssetPairsResponse{}, fmt.Errorf("error"))

	_, err := ingester.GetProviderMarkets(ctx)
	require.Error(t, err)
}

// Test that the ingester only returns markets that are enabled.
func TestIngesterGetsValidMarkets(t *testing.T) {
	client := mocks.NewClient(t)
	ingester := kraken.NewWithClient(zap.NewNop(), client)

	ctx := context.Background()
	// Mock the products endpoint to return a mix of trading-enabled and disabled
	// tickers
	client.On("AssetPairs", ctx).Return(kraken.AssetPairsResponse{
		Errors: nil,
		Result: map[string]kraken.AssetData{
			"XXBTZUSD": {
				Wsname: "BTC/USD",
				Base:   "BTC",
				Quote:  "USD",
				Status: kraken.StatusOnline,
			},
			"XLTCZUSD": {
				Wsname: "LTC/USD",
				Base:   "LTC",
				Quote:  "USD",
				Status: "",
			},
		},
	}, nil)

	// mock an empty response from the volumes endpoint
	client.On("Tickers", ctx).Return(kraken.TickersResponse{
		Error: nil,
		Result: map[string]kraken.TickerData{
			"XXBTZUSD": {
				V: []string{"1108.13627835", "1350.13382379"},
				L: []string{"61774.60000", "61774.60000"},
				H: []string{"63300.00000", "63908.00000"},
				C: []string{"10.23", "0.23"},
			},
			"XLTCZUSD": {
				V: []string{"14967.33246548", "18641.22923157"},
				L: []string{"81.81000", "81.81000"},
				H: []string{"85.55000", "85.55000"},
				C: []string{"10.23", "0.23"},
			},
		},
	}, nil)

	markets, err := ingester.GetProviderMarkets(ctx)
	if err != nil {
		require.NoError(t, err)
	}

	require.Len(t, markets, 1)
	require.Equal(t, "BTC", markets[0].Create.TargetBase)
	require.Equal(t, "USD", markets[0].Create.TargetQuote)
	require.Equal(t, 8.484416466093452e+07, markets[0].Create.QuoteVolume)
	require.Equal(t, 10.23, markets[0].Create.ReferencePrice)
}

func TestIngesterErrorsOnTickersError(t *testing.T) {
	client := mocks.NewClient(t)
	ingester := kraken.NewWithClient(zap.NewNop(), client)

	ctx := context.Background()
	client.On("AssetPairs", ctx).Return(kraken.AssetPairsResponse{}, nil)

	client.On("Tickers", ctx).Return(kraken.TickersResponse{}, fmt.Errorf("error"))

	_, err := ingester.GetProviderMarkets(ctx)
	require.Error(t, err)
}
