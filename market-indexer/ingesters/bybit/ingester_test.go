package bybit_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/skip-mev/connect-mmu/market-indexer/ingesters/bybit"
	"github.com/skip-mev/connect-mmu/market-indexer/ingesters/bybit/mocks"
)

// Test that if bybit ingester's products endpoint returns an error, the
// GetProviderMarkets method should return an error.
func TestIngesterReturnsErrorOnInstrumentsError(t *testing.T) {
	client := mocks.NewClient(t)
	ingester := bybit.NewWithClient(zap.NewNop(), client)

	ctx := context.Background()
	client.On("Instruments", ctx).Return(bybit.InstrumentsResponse{}, fmt.Errorf("error"))

	_, err := ingester.GetProviderMarkets(ctx)
	require.Error(t, err)
}

// Test that the ingester only returns markets that are Trading.
func TestIngesterGetsValidMarkets(t *testing.T) {
	client := mocks.NewClient(t)
	ingester := bybit.NewWithClient(zap.NewNop(), client)

	ctx := context.Background()
	// Mock the products endpoint to return a mix of trading-enabled and disabled
	// tickers
	client.On("Instruments", ctx).Return(bybit.InstrumentsResponse{
		Response: bybit.Response{
			RetCode: 0,
			RetMsg:  "",
		},
		Result: bybit.InstrumentsResult{
			List: []bybit.InstrumentData{
				{
					Symbol:    "BTCUSDT",
					Status:    bybit.StatusTrading,
					BaseCoin:  "BTC",
					QuoteCoin: "USDT",
				},
				{
					Symbol:    "ETHUSDT",
					Status:    "",
					BaseCoin:  "ETH",
					QuoteCoin: "USDT",
				},
			},
		},
	}, nil)

	// mock an empty response from the volumes endpoint
	client.On("Tickers", ctx).Return(bybit.TickersResponse{
		Response: bybit.Response{
			RetCode: 0,
			RetMsg:  "",
		},
		Result: bybit.TickersResult{
			List: []bybit.TickerData{
				{
					Symbol:       "BTCUSDT",
					HighPrice24H: "1000",
					LowPrice24H:  "500",
					Volume24H:    "1000",
					LastPrice:    "124.32003",
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
	require.Equal(t, "USDT", markets[0].Create.TargetQuote)
	require.Equal(t, float64(750000), markets[0].Create.QuoteVolume)
	require.Equal(t, 124.32003, markets[0].Create.ReferencePrice)
}

func TestIngesterErrorsOnTickersError(t *testing.T) {
	client := mocks.NewClient(t)
	ingester := bybit.NewWithClient(zap.NewNop(), client)

	ctx := context.Background()
	client.On("Instruments", ctx).Return(bybit.InstrumentsResponse{}, nil)

	client.On("Tickers", ctx).Return(bybit.TickersResponse{}, fmt.Errorf("error"))

	_, err := ingester.GetProviderMarkets(ctx)
	require.Error(t, err)
}
