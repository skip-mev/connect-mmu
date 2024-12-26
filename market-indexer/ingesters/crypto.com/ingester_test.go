package crypto_com_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	crypto_com "github.com/skip-mev/connect-mmu/market-indexer/ingesters/crypto.com"
	crypto_com_mocks "github.com/skip-mev/connect-mmu/market-indexer/ingesters/crypto.com/mocks"
)

// Test that if coinbase ingester's products endpoint returns an error, the
// GetProviderMarkets method should return an error.
func TestIngesterReturnsErrorOnInstrumentsError(t *testing.T) {
	client := crypto_com_mocks.NewClient(t)
	ingester := crypto_com.NewWithClient(zap.NewNop(), client)

	ctx := context.Background()
	client.On("Instruments", ctx).Return(crypto_com.InstrumentsResponse{}, fmt.Errorf("error"))

	_, err := ingester.GetProviderMarkets(ctx)
	require.Error(t, err)
}

// Test that the ingester only returns markets that are SPOT.
func TestIngesterIgnoresPerpMarkets(t *testing.T) {
	client := crypto_com_mocks.NewClient(t)
	ingester := crypto_com.NewWithClient(zap.NewNop(), client)

	ctx := context.Background()
	// Mock the products endpoint to return a mix of trading-enabled and disabled
	// tickers
	client.On("Instruments", ctx).Return(crypto_com.InstrumentsResponse{
		ID:     1,
		Method: "",
		Code:   0,
		Result: crypto_com.InstrumentsResult{
			Data: []crypto_com.InstrumentsData{
				{
					Symbol:      "BTC_USD",
					InstType:    crypto_com.InstrumentTypeCCYPair,
					DisplayName: "",
					BaseCcy:     "BTC",
					QuoteCcy:    "USD",
					Tradable:    true,
				},
				{
					Symbol:      "ETH_USD",
					InstType:    crypto_com.InstrumentTypeCCYPair,
					DisplayName: "",
					BaseCcy:     "ETH",
					QuoteCcy:    "USD",
					Tradable:    false,
				},
				{
					Symbol:      "BTC_USD-PERP",
					InstType:    crypto_com.InstrumentTypePerpetual,
					DisplayName: "",
					BaseCcy:     "BTC",
					QuoteCcy:    "USD",
					Tradable:    true,
				},
			},
		},
	}, nil)

	// mock an empty response from the volumes endpoint
	client.On("Tickers", ctx).Return(crypto_com.TickersResponse{
		ID:     1,
		Method: "",
		Code:   0,
		Result: crypto_com.TickersResult{
			Data: []crypto_com.TickerData{
				{
					H:           "4000",
					L:           "2000",
					I:           "BTC_USD",
					V:           "1000",
					Vv:          "10000",
					LatestPrice: "12.2352356",
				},
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
	require.Equal(t, float64(3000000), markets[0].Create.QuoteVolume)
	require.Equal(t, float64(10000), markets[0].Create.UsdVolume)
	require.Equal(t, 12.2352356, markets[0].Create.ReferencePrice)
}

func TestIngesterErrorsOnVolumesError(t *testing.T) {
	client := crypto_com_mocks.NewClient(t)
	ingester := crypto_com.NewWithClient(zap.NewNop(), client)

	ctx := context.Background()
	client.On("Instruments", ctx).Return(crypto_com.InstrumentsResponse{}, nil)

	client.On("Tickers", ctx).Return(crypto_com.TickersResponse{}, fmt.Errorf("error"))

	_, err := ingester.GetProviderMarkets(ctx)
	require.Error(t, err)
}
