package okx_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/skip-mev/connect-mmu/market-indexer/ingesters/okx"
	"github.com/skip-mev/connect-mmu/market-indexer/ingesters/okx/mocks"
)

// Test that if okx ingester's products endpoint returns an error, the
// GetProviderMarkets method should return an error.
func TestIngesterReturnsErrorOnInstrumentsError(t *testing.T) {
	client := mocks.NewClient(t)
	ingester := okx.NewWithClient(zap.NewNop(), client)

	ctx := context.Background()
	client.On("Instruments", ctx).Return(okx.InstrumentsResponse{}, fmt.Errorf("error"))

	_, err := ingester.GetProviderMarkets(ctx)
	require.Error(t, err)
}

// Test that the ingester only returns markets that are SPOT.
func TestIngesterIgnoresPerpMarkets(t *testing.T) {
	client := mocks.NewClient(t)
	ingester := okx.NewWithClient(zap.NewNop(), client)

	ctx := context.Background()
	// Mock the products endpoint to return a mix of trading-enabled and disabled
	// tickers
	client.On("Instruments", ctx).Return(okx.InstrumentsResponse{
		Response: okx.Response{
			Code: "0",
			Msg:  "",
		},
		Data: []okx.InstrumentData{
			{
				BaseCcy:  "BTC",
				InstID:   "BTC-USD",
				InstType: "SPOT",
				QuoteCcy: "USD",
				State:    okx.StateLive,
			},
			{
				BaseCcy:  "ETH",
				InstID:   "ETH-USD",
				InstType: "SPOT",
				QuoteCcy: "USD",
				State:    okx.StateLive,
			},
			{
				BaseCcy:  "LTC",
				InstID:   "LTC-USD",
				InstType: "SPOT",
				QuoteCcy: "USD",
				State:    "",
			},
		},
	}, nil)

	// mock an empty response from the volumes endpoint
	client.On("Tickers", ctx).Return(okx.TickersResponse{
		Response: okx.Response{
			Code: "0",
			Msg:  "",
		},
		Data: []okx.TickerData{
			{
				InstType:  "SPOT",
				InstID:    "BTC-USD",
				VolCcy24H: "3000000",
				Open24h:   "67000.23",
			},
			{
				InstType:  "SPOT",
				InstID:    "ETH-USD",
				VolCcy24H: "6000000",
				Open24h:   "3200.32",
			},
			{
				InstType:  "SPOT",
				InstID:    "LTC-USD",
				VolCcy24H: "34235",
				Open24h:   "10.23",
			},
		},
	}, nil)

	markets, err := ingester.GetProviderMarkets(ctx)
	if err != nil {
		require.NoError(t, err)
	}

	require.Len(t, markets, 2)
	require.Equal(t, "BTC", markets[0].Create.TargetBase)
	require.Equal(t, "USD", markets[0].Create.TargetQuote)
	require.Equal(t, float64(3000000), markets[0].Create.QuoteVolume)
	require.Equal(t, 67000.23, markets[0].Create.ReferencePrice)

	require.Equal(t, "ETH", markets[1].Create.TargetBase)
	require.Equal(t, "USD", markets[1].Create.TargetQuote)
	require.Equal(t, float64(6000000), markets[1].Create.QuoteVolume)
	require.Equal(t, 3200.32, markets[1].Create.ReferencePrice)
}

func TestIngesterErrorsOnTickersError(t *testing.T) {
	client := mocks.NewClient(t)
	ingester := okx.NewWithClient(zap.NewNop(), client)

	ctx := context.Background()
	client.On("Instruments", ctx).Return(okx.InstrumentsResponse{}, nil)

	client.On("Tickers", ctx).Return(okx.TickersResponse{}, fmt.Errorf("error"))

	_, err := ingester.GetProviderMarkets(ctx)
	require.Error(t, err)
}
