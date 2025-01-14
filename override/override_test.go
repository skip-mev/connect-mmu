package override

import (
	"context"
	"testing"

	"github.com/skip-mev/connect-mmu/override/update"

	connecttypes "github.com/skip-mev/connect/v2/pkg/types"
	"github.com/skip-mev/connect/v2/x/marketmap/types"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"

	"github.com/skip-mev/connect-mmu/client/dydx"
	"github.com/skip-mev/connect-mmu/client/dydx/mocks"
)

func TestOverride(t *testing.T) {
	testCases := []struct {
		name             string
		actual           types.MarketMap
		generated        types.MarketMap
		expected         types.MarketMap
		expectedRemovals []string
		options          update.Options
	}{
		{
			name: "markets are consolidated",
			actual: types.MarketMap{Markets: map[string]types.Market{
				"FOO/USD": {
					Ticker: types.Ticker{
						CurrencyPair:  makeCurrencyPair(t, "FOO/USD"),
						Metadata_JSON: `{"aggregate_ids":[{"venue":"coinmarketcap","ID":"2"}]}`,
					},
					ProviderConfigs: []types.ProviderConfig{
						{Name: "coinbase"},
					},
				},
				"BAR/USD": {
					Ticker: types.Ticker{
						CurrencyPair:  makeCurrencyPair(t, "BAR/USD"),
						Metadata_JSON: `{"aggregate_ids":[{"venue":"coinmarketcap","ID":"3"}]}`,
					},
					ProviderConfigs: []types.ProviderConfig{
						{Name: "coinbase"},
					},
				},
			}},
			generated: types.MarketMap{Markets: map[string]types.Market{
				"FOO,UNISWAP,0XUNISWAP/USD": {
					Ticker: types.Ticker{
						CurrencyPair:  makeCurrencyPair(t, "FOO,UNISWAP,0XUNISWAP/USD"),
						Metadata_JSON: `{"aggregate_ids":[{"venue":"coinmarketcap","ID":"2"}]}`,
					},
					ProviderConfigs: []types.ProviderConfig{
						{Name: "uniswap"},
					},
				},
				"BAZ/USD": {
					Ticker: types.Ticker{
						CurrencyPair:  makeCurrencyPair(t, "BAZ/USD"),
						Metadata_JSON: `{"aggregate_ids":[{"venue":"coinmarketcap","ID":"4"}]}`,
					},
					ProviderConfigs: []types.ProviderConfig{
						{Name: "binance"},
					},
				},
			}},
			expected: types.MarketMap{Markets: map[string]types.Market{
				"FOO/USD": {
					Ticker: types.Ticker{
						CurrencyPair:  makeCurrencyPair(t, "FOO/USD"),
						Metadata_JSON: `{"aggregate_ids":[{"venue":"coinmarketcap","ID":"2"}]}`,
					},
					ProviderConfigs: []types.ProviderConfig{
						{Name: "coinbase"}, {Name: "uniswap"},
					},
				},
				"BAZ/USD": {
					Ticker: types.Ticker{
						CurrencyPair:  makeCurrencyPair(t, "BAZ/USD"),
						Metadata_JSON: `{"aggregate_ids":[{"venue":"coinmarketcap","ID":"4"}]}`,
					},
					ProviderConfigs: []types.ProviderConfig{
						{Name: "binance"},
					},
				},
			}},
			expectedRemovals: []string{"BAR/USD"},
			options:          update.Options{DisableDeFiMarketMerging: false},
		},
		{
			name: "markets are consolidated, actual market has DeFi ticker",
			actual: types.MarketMap{Markets: map[string]types.Market{
				"FOO,UNISWAP,0XUNISWAP/USD": {
					Ticker: types.Ticker{
						CurrencyPair:  makeCurrencyPair(t, "FOO,UNISWAP,0XUNISWAP/USD"),
						Metadata_JSON: `{"aggregate_ids":[{"venue":"coinmarketcap","ID":"2"}]}`,
					},
					ProviderConfigs: []types.ProviderConfig{
						{Name: "uniswap"},
					},
				},
				"BAR/USD": {
					Ticker: types.Ticker{
						CurrencyPair:  makeCurrencyPair(t, "BAR/USD"),
						Metadata_JSON: `{"aggregate_ids":[{"venue":"coinmarketcap","ID":"3"}]}`,
					},
					ProviderConfigs: []types.ProviderConfig{
						{Name: "coinbase"},
					},
				},
			}},
			generated: types.MarketMap{Markets: map[string]types.Market{
				"FOO/USD": {
					Ticker: types.Ticker{
						CurrencyPair:  makeCurrencyPair(t, "FOO/USD"),
						Metadata_JSON: `{"aggregate_ids":[{"venue":"coinmarketcap","ID":"2"}]}`,
					},
					ProviderConfigs: []types.ProviderConfig{
						{Name: "coinbase"}, {Name: "uniswap"},
					},
				},
				"BAZ/USD": {
					Ticker: types.Ticker{
						CurrencyPair:  makeCurrencyPair(t, "BAZ/USD"),
						Metadata_JSON: `{"aggregate_ids":[{"venue":"coinmarketcap","ID":"4"}]}`,
					},
					ProviderConfigs: []types.ProviderConfig{
						{Name: "binance"},
					},
				},
			}},
			expected: types.MarketMap{Markets: map[string]types.Market{
				"FOO,UNISWAP,0XUNISWAP/USD": {
					Ticker: types.Ticker{
						CurrencyPair:  makeCurrencyPair(t, "FOO,UNISWAP,0XUNISWAP/USD"),
						Metadata_JSON: `{"aggregate_ids":[{"venue":"coinmarketcap","ID":"2"}]}`,
					},
					ProviderConfigs: []types.ProviderConfig{
						{Name: "uniswap"}, {Name: "coinbase"},
					},
				},
				"BAZ/USD": {
					Ticker: types.Ticker{
						CurrencyPair:  makeCurrencyPair(t, "BAZ/USD"),
						Metadata_JSON: `{"aggregate_ids":[{"venue":"coinmarketcap","ID":"4"}]}`,
					},
					ProviderConfigs: []types.ProviderConfig{
						{Name: "binance"},
					},
				},
			}},
			expectedRemovals: []string{"BAR/USD"},
			options:          update.Options{DisableDeFiMarketMerging: false},
		},
	}
	mmo := NewCoreOverride()
	logger := zaptest.NewLogger(t)
	ctx := context.Background()
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			out, removals, err := Override(ctx, logger, mmo, tc.actual, tc.generated, tc.options)

			require.NoError(t, err)
			require.Equal(t, tc.expected, out, "unexpected output: %v", out)
			require.Equal(t, tc.expectedRemovals, removals, "unexpected removals: %v", removals)
		})
	}
}

func TestGetCMCIDMapping(t *testing.T) {
	tests := []struct {
		name        string
		in          types.MarketMap
		expected    map[string]string
		includeDeFi bool
	}{
		{
			name: "CMC IDs are extracted, no DeFi",
			in: types.MarketMap{
				Markets: map[string]types.Market{
					"FOO/USD": {
						Ticker:          types.Ticker{Metadata_JSON: "{\"reference_price\":1786788632,\"liquidity\":184445,\"aggregate_ids\":[{\"venue\":\"coinmarketcap\",\"ID\":\"32349\"}]}"},
						ProviderConfigs: nil,
					},
					"DOOM,UNISWAP,0XDOOM/USD": {
						Ticker:          types.Ticker{Metadata_JSON: "{\"reference_price\":1786788632,\"liquidity\":184445,\"aggregate_ids\":[{\"venue\":\"coinmarketcap\",\"ID\":\"33\"}]}"},
						ProviderConfigs: nil,
					},
					"BAR/USD": {
						Ticker: types.Ticker{Metadata_JSON: "{\"reference_price\":1786788632,\"liquidity\":184445,\"aggregate_ids\":[{\"venue\":\"coinmarketcap\",\"ID\":\"2\"}]}"},
					},
					"BAZ/USD": {},
				},
			},
			includeDeFi: false,
			expected: map[string]string{
				"32349": "FOO/USD",
				"2":     "BAR/USD",
			},
		},
		{
			name: "CMC IDs are extracted with DeFi",
			in: types.MarketMap{
				Markets: map[string]types.Market{
					"FOO,UNISWAP,0XFOO/USD": {
						Ticker:          types.Ticker{Metadata_JSON: "{\"reference_price\":1786788632,\"liquidity\":184445,\"aggregate_ids\":[{\"venue\":\"coinmarketcap\",\"ID\":\"32349\"}]}"},
						ProviderConfigs: nil,
					},
					"BAR/USD": {
						Ticker: types.Ticker{Metadata_JSON: "{\"reference_price\":1786788632,\"liquidity\":184445,\"aggregate_ids\":[{\"venue\":\"coinmarketcap\",\"ID\":\"2\"}]}"},
					},
					"BAZ/USD": {},
				},
			},
			includeDeFi: true,
			expected: map[string]string{
				"32349": "FOO,UNISWAP,0XFOO/USD",
				"2":     "BAR/USD",
			},
		},
		{
			name: "duplicates are ignored",
			in: types.MarketMap{
				Markets: map[string]types.Market{
					"FOO/USD": {
						Ticker:          types.Ticker{Metadata_JSON: "{\"reference_price\":1786788632,\"liquidity\":184445,\"aggregate_ids\":[{\"venue\":\"coinmarketcap\",\"ID\":\"2\"}]}"},
						ProviderConfigs: nil,
					},
					"BAR/USD": {
						Ticker: types.Ticker{Metadata_JSON: "{\"reference_price\":1786788632,\"liquidity\":184445,\"aggregate_ids\":[{\"venue\":\"coinmarketcap\",\"ID\":\"2\"}]}"},
					},
					"BAZ/USD": {},
				},
			},
			expected: map[string]string{},
		},
	}

	logger := zaptest.NewLogger(t)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out, err := getCMCTickerMapping(logger, tt.in, tt.includeDeFi)
			require.NoError(t, err)
			for id, ticker := range out {
				expected, ok := tt.expected[id]
				require.True(t, ok, "unexpected output id %s, ticker %s", id, ticker)
				require.Equal(t, expected, ticker)
			}
		})
	}
}

func TestConsolidateGeneratedMarkets(t *testing.T) {
	testCases := []struct {
		name        string
		generated   types.MarketMap
		actual      types.MarketMap
		expectedOut types.MarketMap
	}{
		{
			name:        "empty does nothing",
			generated:   types.MarketMap{},
			actual:      types.MarketMap{},
			expectedOut: types.MarketMap{},
		},
		{
			name: "market is consolidated",
			generated: types.MarketMap{
				Markets: map[string]types.Market{
					"BTC,UNISWAP,0XBITCOIN/USD": {
						Ticker: types.Ticker{
							CurrencyPair:  makeCurrencyPair(t, "BTC,UNISWAP,0XBITCOIN/USD"),
							Metadata_JSON: `{"aggregate_ids":[{"venue":"coinmarketcap","ID":"2"}]}`,
						},
						ProviderConfigs: []types.ProviderConfig{{Name: "uniswap"}},
					},
					"FOO/USD": {
						Ticker: types.Ticker{
							CurrencyPair:  makeCurrencyPair(t, "FOO/USD"),
							Metadata_JSON: `{"aggregate_ids":[{"venue":"coinmarketcap","ID":"3"}]}`,
						},
						ProviderConfigs: []types.ProviderConfig{{Name: "kucoin"}},
					},
				},
			},
			actual: types.MarketMap{
				Markets: map[string]types.Market{
					"BTC/USD": {
						Ticker: types.Ticker{
							CurrencyPair:  makeCurrencyPair(t, "BTC/USD"),
							Metadata_JSON: `{"aggregate_ids":[{"venue":"coinmarketcap","ID":"2"}]}`,
						},
						ProviderConfigs: []types.ProviderConfig{},
					},
				},
			},
			expectedOut: types.MarketMap{
				Markets: map[string]types.Market{
					"BTC/USD": {
						Ticker: types.Ticker{
							CurrencyPair:  makeCurrencyPair(t, "BTC/USD"),
							Metadata_JSON: `{"aggregate_ids":[{"venue":"coinmarketcap","ID":"2"}]}`,
						},
						ProviderConfigs: []types.ProviderConfig{{Name: "uniswap"}},
					},
					"FOO/USD": {
						Ticker: types.Ticker{
							CurrencyPair:  makeCurrencyPair(t, "FOO/USD"),
							Metadata_JSON: `{"aggregate_ids":[{"venue":"coinmarketcap","ID":"3"}]}`,
						},
						ProviderConfigs: []types.ProviderConfig{{Name: "kucoin"}},
					},
				},
			},
		},
		{
			name: "market is consolidated if actual is a defi ticker.",
			generated: types.MarketMap{
				Markets: map[string]types.Market{
					"BTC,UNISWAP,0XBITCOIN/USD": {
						Ticker: types.Ticker{
							CurrencyPair:  makeCurrencyPair(t, "BTC,UNISWAP,0XBITCOIN/USD"),
							Metadata_JSON: `{"aggregate_ids":[{"venue":"coinmarketcap","ID":"2"}]}`,
						},
						ProviderConfigs: []types.ProviderConfig{{Name: "uniswap"}},
					},
				},
			},
			actual: types.MarketMap{
				Markets: map[string]types.Market{
					"BTC,RAYDIUM,03231/USD": {
						Ticker: types.Ticker{
							CurrencyPair:  makeCurrencyPair(t, "BTC,RAYDIUM,03231/USD"),
							Metadata_JSON: `{"aggregate_ids":[{"venue":"coinmarketcap","ID":"2"}]}`,
						},
						ProviderConfigs: []types.ProviderConfig{{Name: "raydium"}},
					},
				},
			},
			expectedOut: types.MarketMap{
				Markets: map[string]types.Market{
					"BTC,RAYDIUM,03231/USD": {
						Ticker: types.Ticker{
							CurrencyPair:  makeCurrencyPair(t, "BTC,RAYDIUM,03231/USD"),
							Metadata_JSON: `{"aggregate_ids":[{"venue":"coinmarketcap","ID":"2"}]}`,
						},
						ProviderConfigs: []types.ProviderConfig{{Name: "uniswap"}},
					},
				},
			},
		},
	}

	logger := zaptest.NewLogger(t)
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			out, err := ConsolidateDeFiMarkets(logger, tc.generated, tc.actual)
			require.NoError(t, err)
			require.Equal(t, tc.expectedOut, out)
		})
	}
}

func makeCurrencyPair(t *testing.T, s string) connecttypes.CurrencyPair {
	t.Helper()
	cp, err := connecttypes.CurrencyPairFromString(s)
	require.NoError(t, err)
	return cp
}

func TestOverrideMarketMap(t *testing.T) {
	mockClient := mocks.NewClient(t)

	tests := []struct {
		name          string
		actual        types.MarketMap
		client        dydx.Client
		expect        func(client dydx.Client)
		generated     types.MarketMap
		want          types.MarketMap
		wantRemovals  []string
		updateEnabled bool
		wantInitErr   bool
		wantErr       bool
	}{
		{
			name:        "fail for a nil client",
			client:      nil,
			expect:      func(_ dydx.Client) {},
			wantInitErr: true,
		},
		{
			name:   "return error for nil response",
			client: mockClient,
			expect: func(_ dydx.Client) {
				mockClient.EXPECT().AllPerpetuals(mock.Anything).Return(nil, nil).Once()
			},
			actual: types.MarketMap{},
			generated: types.MarketMap{
				Markets: map[string]types.Market{
					"BTC/USD": {
						Ticker: types.Ticker{
							CurrencyPair:     connecttypes.NewCurrencyPair("BTC", "USD"),
							Decimals:         10,
							MinProviderCount: 1,
							Enabled:          false,
						},
						ProviderConfigs: []types.ProviderConfig{
							{
								Name:           "test",
								OffChainTicker: "test_offchain",
							},
						},
					},
				},
			},
			updateEnabled: false,
			wantErr:       true,
		},
		{
			name:   "override an empty market map",
			client: mockClient,
			expect: func(_ dydx.Client) {
				mockClient.EXPECT().AllPerpetuals(mock.Anything).Return(&dydx.AllPerpetualsResponse{}, nil).Once()
			},
			actual: types.MarketMap{},
			generated: types.MarketMap{
				Markets: map[string]types.Market{
					"BTC/USD": {
						Ticker: types.Ticker{
							CurrencyPair:     connecttypes.NewCurrencyPair("BTC", "USD"),
							Decimals:         10,
							MinProviderCount: 1,
							Enabled:          false,
						},
						ProviderConfigs: []types.ProviderConfig{
							{
								Name:           "test",
								OffChainTicker: "test_offchain",
							},
						},
					},
				},
			},
			want: types.MarketMap{
				Markets: map[string]types.Market{
					"BTC/USD": {
						Ticker: types.Ticker{
							CurrencyPair:     connecttypes.NewCurrencyPair("BTC", "USD"),
							Decimals:         10,
							MinProviderCount: 1,
							Enabled:          false,
						},
						ProviderConfigs: []types.ProviderConfig{
							{
								Name:           "test",
								OffChainTicker: "test_offchain",
							},
						},
					},
				},
			},
			wantRemovals:  []string{},
			updateEnabled: false,
			wantErr:       false,
		},
		{
			name:   "override an empty generated market map with enabled actual market",
			client: mockClient,
			expect: func(_ dydx.Client) {
				mockClient.EXPECT().AllPerpetuals(mock.Anything).Return(&dydx.AllPerpetualsResponse{}, nil).Once()
			},
			actual: types.MarketMap{
				Markets: map[string]types.Market{
					"BTC/USD": {
						Ticker: types.Ticker{
							CurrencyPair:     connecttypes.NewCurrencyPair("BTC", "USD"),
							Decimals:         10,
							MinProviderCount: 1,
							Enabled:          true,
						},
						ProviderConfigs: []types.ProviderConfig{
							{
								Name:           "test",
								OffChainTicker: "test_offchain",
							},
						},
					},
				},
			},
			generated: types.MarketMap{},
			want: types.MarketMap{
				Markets: map[string]types.Market{
					"BTC/USD": {
						Ticker: types.Ticker{
							CurrencyPair:     connecttypes.NewCurrencyPair("BTC", "USD"),
							Decimals:         10,
							MinProviderCount: 1,
							Enabled:          true,
						},
						ProviderConfigs: []types.ProviderConfig{
							{
								Name:           "test",
								OffChainTicker: "test_offchain",
							},
						},
					},
				},
			},
			wantRemovals:  []string{},
			updateEnabled: false,
			wantErr:       false,
		},
		{
			name:   "override an empty generated market map with disabled actual market",
			client: mockClient,
			expect: func(_ dydx.Client) {
				mockClient.EXPECT().AllPerpetuals(mock.Anything).Return(&dydx.AllPerpetualsResponse{}, nil).Once()
			},
			actual: types.MarketMap{
				Markets: map[string]types.Market{
					"BTC/USD": {
						Ticker: types.Ticker{
							CurrencyPair:     connecttypes.NewCurrencyPair("BTC", "USD"),
							Decimals:         10,
							MinProviderCount: 1,
							Enabled:          false,
						},
						ProviderConfigs: []types.ProviderConfig{
							{
								Name:           "test",
								OffChainTicker: "test_offchain",
							},
						},
					},
				},
			},
			generated:     types.MarketMap{},
			want:          types.MarketMap{Markets: map[string]types.Market{}},
			wantRemovals:  []string{"BTC/USD"},
			updateEnabled: false,
			wantErr:       false,
		},
		{
			name:   "disable a market that was enabled in the generated market map but does not exist in actual",
			client: mockClient,
			expect: func(_ dydx.Client) {
				mockClient.EXPECT().AllPerpetuals(mock.Anything).Return(&dydx.AllPerpetualsResponse{}, nil).Once()
			},
			actual: types.MarketMap{},
			generated: types.MarketMap{
				Markets: map[string]types.Market{
					"BTC/USD": {
						Ticker: types.Ticker{
							CurrencyPair:     connecttypes.NewCurrencyPair("BTC", "USD"),
							Decimals:         10,
							MinProviderCount: 1,
							Enabled:          true,
						},
						ProviderConfigs: []types.ProviderConfig{
							{
								Name:           "test",
								OffChainTicker: "test_offchain",
							},
						},
					},
				},
			},
			want: types.MarketMap{
				Markets: map[string]types.Market{
					"BTC/USD": {
						Ticker: types.Ticker{
							CurrencyPair:     connecttypes.NewCurrencyPair("BTC", "USD"),
							Decimals:         10,
							MinProviderCount: 1,
							Enabled:          false,
						},
						ProviderConfigs: []types.ProviderConfig{
							{
								Name:           "test",
								OffChainTicker: "test_offchain",
							},
						},
					},
				},
			},
			wantRemovals:  []string{},
			updateEnabled: false,
			wantErr:       false,
		},
		{
			name:   "do nothing if there is no diff between generated and generated",
			client: mockClient,
			expect: func(_ dydx.Client) {
				mockClient.EXPECT().AllPerpetuals(mock.Anything).Return(&dydx.AllPerpetualsResponse{}, nil).Once()
			},
			actual: types.MarketMap{
				Markets: map[string]types.Market{
					"BTC/USD": {
						Ticker: types.Ticker{
							CurrencyPair:     connecttypes.NewCurrencyPair("BTC", "USD"),
							Decimals:         10,
							MinProviderCount: 1,
							Enabled:          false,
						},
						ProviderConfigs: []types.ProviderConfig{
							{
								Name:           "test",
								OffChainTicker: "test_offchain",
							},
						},
					},
				},
			},
			generated: types.MarketMap{
				Markets: map[string]types.Market{
					"BTC/USD": {
						Ticker: types.Ticker{
							CurrencyPair:     connecttypes.NewCurrencyPair("BTC", "USD"),
							Decimals:         10,
							MinProviderCount: 1,
							Enabled:          false,
						},
						ProviderConfigs: []types.ProviderConfig{
							{
								Name:           "test",
								OffChainTicker: "test_offchain",
							},
						},
					},
				},
			},
			want: types.MarketMap{
				Markets: map[string]types.Market{
					"BTC/USD": {
						Ticker: types.Ticker{
							CurrencyPair:     connecttypes.NewCurrencyPair("BTC", "USD"),
							Decimals:         10,
							MinProviderCount: 1,
							Enabled:          false,
						},
						ProviderConfigs: []types.ProviderConfig{
							{
								Name:           "test",
								OffChainTicker: "test_offchain",
							},
						},
					},
				},
			},
			wantRemovals:  []string{},
			updateEnabled: false,
			wantErr:       false,
		},
		{
			name:   "enable a market that is enabled on chain, but disabled in generated",
			client: mockClient,
			expect: func(_ dydx.Client) {
				mockClient.EXPECT().AllPerpetuals(mock.Anything).Return(&dydx.AllPerpetualsResponse{}, nil).Once()
			},
			actual: types.MarketMap{
				Markets: map[string]types.Market{
					"BTC/USD": {
						Ticker: types.Ticker{
							CurrencyPair:     connecttypes.NewCurrencyPair("BTC", "USD"),
							Decimals:         10,
							MinProviderCount: 1,
							Enabled:          true,
						},
						ProviderConfigs: []types.ProviderConfig{
							{
								Name:           "test",
								OffChainTicker: "test_offchain",
							},
						},
					},
				},
			},
			generated: types.MarketMap{
				Markets: map[string]types.Market{
					"BTC/USD": {
						Ticker: types.Ticker{
							CurrencyPair:     connecttypes.NewCurrencyPair("BTC", "USD"),
							Decimals:         10,
							MinProviderCount: 1,
							Enabled:          false,
						},
						ProviderConfigs: []types.ProviderConfig{
							{
								Name:           "test",
								OffChainTicker: "test_offchain",
							},
						},
					},
				},
			},
			want: types.MarketMap{
				Markets: map[string]types.Market{
					"BTC/USD": {
						Ticker: types.Ticker{
							CurrencyPair:     connecttypes.NewCurrencyPair("BTC", "USD"),
							Decimals:         10,
							MinProviderCount: 1,
							Enabled:          true,
						},
						ProviderConfigs: []types.ProviderConfig{
							{
								Name:           "test",
								OffChainTicker: "test_offchain",
							},
						},
					},
				},
			},
			wantRemovals:  []string{},
			updateEnabled: false,
			wantErr:       false,
		},
		{
			name:   "override decimals and min provider count",
			client: mockClient,
			expect: func(_ dydx.Client) {
				mockClient.EXPECT().AllPerpetuals(mock.Anything).Return(&dydx.AllPerpetualsResponse{}, nil).Once()
			},
			actual: types.MarketMap{
				Markets: map[string]types.Market{
					"BTC/USD": {
						Ticker: types.Ticker{
							CurrencyPair:     connecttypes.NewCurrencyPair("BTC", "USD"),
							Decimals:         11,
							MinProviderCount: 4,
							Enabled:          true,
						},
						ProviderConfigs: []types.ProviderConfig{
							{
								Name:           "test",
								OffChainTicker: "test_offchain",
							},
						},
					},
				},
			},
			generated: types.MarketMap{
				Markets: map[string]types.Market{
					"BTC/USD": {
						Ticker: types.Ticker{
							CurrencyPair:     connecttypes.NewCurrencyPair("BTC", "USD"),
							Decimals:         10,
							MinProviderCount: 1,
							Enabled:          true,
						},
						ProviderConfigs: []types.ProviderConfig{
							{
								Name:           "test",
								OffChainTicker: "test_offchain",
							},
						},
					},
				},
			},
			want: types.MarketMap{
				Markets: map[string]types.Market{
					"BTC/USD": {
						Ticker: types.Ticker{
							CurrencyPair:     connecttypes.NewCurrencyPair("BTC", "USD"),
							Decimals:         11,
							MinProviderCount: 4,
							Enabled:          true,
						},
						ProviderConfigs: []types.ProviderConfig{
							{
								Name:           "test",
								OffChainTicker: "test_offchain",
							},
						},
					},
				},
			},
			wantRemovals:  []string{},
			updateEnabled: false,
			wantErr:       false,
		},
		{
			name:   "keep existing provider ticker for enabled market",
			client: mockClient,
			expect: func(_ dydx.Client) {
				mockClient.EXPECT().AllPerpetuals(mock.Anything).Return(&dydx.AllPerpetualsResponse{}, nil).Once()
			},
			actual: types.MarketMap{
				Markets: map[string]types.Market{
					"BTC/USD": {
						Ticker: types.Ticker{
							CurrencyPair:     connecttypes.NewCurrencyPair("BTC", "USD"),
							Decimals:         10,
							MinProviderCount: 1,
							Enabled:          true,
						},
						ProviderConfigs: []types.ProviderConfig{
							{
								Name:           "test",
								OffChainTicker: "test_offchain",
							},
						},
					},
				},
			},
			generated: types.MarketMap{
				Markets: map[string]types.Market{
					"BTC/USD": {
						Ticker: types.Ticker{
							CurrencyPair:     connecttypes.NewCurrencyPair("BTC", "USD"),
							Decimals:         10,
							MinProviderCount: 1,
							Enabled:          true,
						},
						ProviderConfigs: []types.ProviderConfig{
							{
								Name:           "test",
								OffChainTicker: "test_offchain_new",
							},
						},
					},
				},
			},
			want: types.MarketMap{
				Markets: map[string]types.Market{
					"BTC/USD": {
						Ticker: types.Ticker{
							CurrencyPair:     connecttypes.NewCurrencyPair("BTC", "USD"),
							Decimals:         10,
							MinProviderCount: 1,
							Enabled:          true,
						},
						ProviderConfigs: []types.ProviderConfig{
							{
								Name:           "test",
								OffChainTicker: "test_offchain",
							},
						},
					},
				},
			},
			wantRemovals:  []string{},
			updateEnabled: false,
			wantErr:       false,
		},
		{
			name:   "keep existing provider ticker for disabled market",
			client: mockClient,
			expect: func(_ dydx.Client) {
				mockClient.EXPECT().AllPerpetuals(mock.Anything).Return(&dydx.AllPerpetualsResponse{}, nil).Once()
			},
			actual: types.MarketMap{
				Markets: map[string]types.Market{
					"BTC/USD": {
						Ticker: types.Ticker{
							CurrencyPair:     connecttypes.NewCurrencyPair("BTC", "USD"),
							Decimals:         10,
							MinProviderCount: 1,
							Enabled:          false,
						},
						ProviderConfigs: []types.ProviderConfig{
							{
								Name:           "test",
								OffChainTicker: "test_offchain",
							},
						},
					},
				},
			},
			generated: types.MarketMap{
				Markets: map[string]types.Market{
					"BTC/USD": {
						Ticker: types.Ticker{
							CurrencyPair:     connecttypes.NewCurrencyPair("BTC", "USD"),
							Decimals:         10,
							MinProviderCount: 1,
							Enabled:          false,
						},
						ProviderConfigs: []types.ProviderConfig{
							{
								Name:           "test",
								OffChainTicker: "test_offchain_new",
							},
						},
					},
				},
			},
			want: types.MarketMap{
				Markets: map[string]types.Market{
					"BTC/USD": {
						Ticker: types.Ticker{
							CurrencyPair:     connecttypes.NewCurrencyPair("BTC", "USD"),
							Decimals:         10,
							MinProviderCount: 1,
							Enabled:          false,
						},
						ProviderConfigs: []types.ProviderConfig{
							{
								Name:           "test",
								OffChainTicker: "test_offchain",
							},
						},
					},
				},
			},
			wantRemovals:  []string{},
			updateEnabled: false,
			wantErr:       false,
		},
		{
			name:   "append market to existing one - disjoint provider configs",
			client: mockClient,
			expect: func(_ dydx.Client) {
				mockClient.EXPECT().AllPerpetuals(mock.Anything).Return(&dydx.AllPerpetualsResponse{}, nil).Once()
			},
			actual: types.MarketMap{
				Markets: map[string]types.Market{
					"BTC/USD": {
						Ticker: types.Ticker{
							CurrencyPair:     connecttypes.NewCurrencyPair("BTC", "USD"),
							Decimals:         10,
							MinProviderCount: 1,
							Enabled:          false,
						},
						ProviderConfigs: []types.ProviderConfig{
							{
								Name:           "test",
								OffChainTicker: "test_offchain",
							},
						},
					},
				},
			},
			generated: types.MarketMap{
				Markets: map[string]types.Market{
					"BTC/USD": {
						Ticker: types.Ticker{
							CurrencyPair:     connecttypes.NewCurrencyPair("BTC", "USD"),
							Decimals:         10,
							MinProviderCount: 1,
							Enabled:          false,
						},
						ProviderConfigs: []types.ProviderConfig{
							{
								Name:           "test_new",
								OffChainTicker: "test_offchain_new",
							},
						},
					},
				},
			},
			want: types.MarketMap{
				Markets: map[string]types.Market{
					"BTC/USD": {
						Ticker: types.Ticker{
							CurrencyPair:     connecttypes.NewCurrencyPair("BTC", "USD"),
							Decimals:         10,
							MinProviderCount: 1,
							Enabled:          false,
						},
						ProviderConfigs: []types.ProviderConfig{
							{
								Name:           "test",
								OffChainTicker: "test_offchain",
							},
							{
								Name:           "test_new",
								OffChainTicker: "test_offchain_new",
							},
						},
					},
				},
			},
			wantRemovals:  []string{},
			updateEnabled: false,
			wantErr:       false,
		},
		{
			name:   "append market to existing one - overlapping provider configs",
			client: mockClient,
			expect: func(_ dydx.Client) {
				mockClient.EXPECT().AllPerpetuals(mock.Anything).Return(&dydx.AllPerpetualsResponse{}, nil).Once()
			},
			actual: types.MarketMap{
				Markets: map[string]types.Market{
					"BTC/USD": {
						Ticker: types.Ticker{
							CurrencyPair:     connecttypes.NewCurrencyPair("BTC", "USD"),
							Decimals:         10,
							MinProviderCount: 1,
							Enabled:          false,
						},
						ProviderConfigs: []types.ProviderConfig{
							{
								Name:           "test",
								OffChainTicker: "test_offchain",
							},
						},
					},
				},
			},
			generated: types.MarketMap{
				Markets: map[string]types.Market{
					"BTC/USD": {
						Ticker: types.Ticker{
							CurrencyPair:     connecttypes.NewCurrencyPair("BTC", "USD"),
							Decimals:         10,
							MinProviderCount: 1,
							Enabled:          false,
						},
						ProviderConfigs: []types.ProviderConfig{
							{
								Name:           "test",
								OffChainTicker: "test_offchain",
							},
							{
								Name:           "test_new",
								OffChainTicker: "test_offchain_new",
							},
						},
					},
				},
			},
			want: types.MarketMap{
				Markets: map[string]types.Market{
					"BTC/USD": {
						Ticker: types.Ticker{
							CurrencyPair:     connecttypes.NewCurrencyPair("BTC", "USD"),
							Decimals:         10,
							MinProviderCount: 1,
							Enabled:          false,
						},
						ProviderConfigs: []types.ProviderConfig{
							{
								Name:           "test",
								OffChainTicker: "test_offchain",
							},
							{
								Name:           "test_new",
								OffChainTicker: "test_offchain_new",
							},
						},
					},
				},
			},
			wantRemovals:  []string{},
			updateEnabled: false,
			wantErr:       false,
		},
		{
			name:   "error if perpetual market is not in actual market map",
			client: mockClient,
			expect: func(_ dydx.Client) {
				mockClient.EXPECT().AllPerpetuals(mock.Anything).Return(&dydx.AllPerpetualsResponse{
					Perpetuals: []dydx.Perpetual{
						{
							Params: dydx.PerpetualParams{
								Ticker:     "BTC-USD",
								MarketType: dydx.PERPETUAL_MARKET_TYPE_CROSS,
							},
						},
					},
				}, nil).Once()
			},
			updateEnabled: false,
			wantErr:       true,
		},
		{
			name:   "set all fields equal to actual market map if it is CROSS MARGIN",
			client: mockClient,
			expect: func(_ dydx.Client) {
				mockClient.EXPECT().AllPerpetuals(mock.Anything).Return(&dydx.AllPerpetualsResponse{
					Perpetuals: []dydx.Perpetual{
						{
							Params: dydx.PerpetualParams{
								Ticker:     "BTC-USD",
								MarketType: dydx.PERPETUAL_MARKET_TYPE_CROSS,
							},
						},
					},
				}, nil).Once()
			},
			actual: types.MarketMap{
				Markets: map[string]types.Market{
					"BTC/USD": {
						Ticker: types.Ticker{
							CurrencyPair:     connecttypes.NewCurrencyPair("BTC", "USD"),
							Decimals:         10,
							MinProviderCount: 1,
							Enabled:          true,
						},
						ProviderConfigs: []types.ProviderConfig{
							{
								Name:           "test",
								OffChainTicker: "test_offchain",
							},
						},
					},
				},
			},
			generated: types.MarketMap{
				Markets: map[string]types.Market{
					"BTC/USD": {
						Ticker: types.Ticker{
							CurrencyPair:     connecttypes.NewCurrencyPair("BTC", "USD"),
							Decimals:         1,
							MinProviderCount: 13,
							Enabled:          false,
						},
						ProviderConfigs: []types.ProviderConfig{
							{
								Name:           "test",
								OffChainTicker: "test_offchain",
							},
						},
					},
				},
			},
			want: types.MarketMap{
				Markets: map[string]types.Market{
					"BTC/USD": {
						Ticker: types.Ticker{
							CurrencyPair:     connecttypes.NewCurrencyPair("BTC", "USD"),
							Decimals:         10,
							MinProviderCount: 1,
							Enabled:          true,
						},
						ProviderConfigs: []types.ProviderConfig{
							{
								Name:           "test",
								OffChainTicker: "test_offchain",
							},
						},
					},
				},
			},
			wantRemovals:  []string{},
			updateEnabled: false,
			wantErr:       false,
		},
		{
			name:   "set all fields equal to actual market map if it is CROSS MARGIN --update-enabled",
			client: mockClient,
			expect: func(_ dydx.Client) {
				mockClient.EXPECT().AllPerpetuals(mock.Anything).Return(&dydx.AllPerpetualsResponse{
					Perpetuals: []dydx.Perpetual{
						{
							Params: dydx.PerpetualParams{
								Ticker:     "BTC-USD",
								MarketType: dydx.PERPETUAL_MARKET_TYPE_CROSS,
							},
						},
					},
				}, nil).Once()
			},
			actual: types.MarketMap{
				Markets: map[string]types.Market{
					"BTC/USD": {
						Ticker: types.Ticker{
							CurrencyPair:     connecttypes.NewCurrencyPair("BTC", "USD"),
							Decimals:         10,
							MinProviderCount: 1,
							Enabled:          true,
						},
						ProviderConfigs: []types.ProviderConfig{
							{
								Name:           "test",
								OffChainTicker: "test_offchain",
							},
						},
					},
				},
			},
			generated: types.MarketMap{
				Markets: map[string]types.Market{
					"BTC/USD": {
						Ticker: types.Ticker{
							CurrencyPair:     connecttypes.NewCurrencyPair("BTC", "USD"),
							Decimals:         1,
							MinProviderCount: 13,
							Enabled:          false,
						},
						ProviderConfigs: []types.ProviderConfig{
							{
								Name:           "test",
								OffChainTicker: "test_offchain",
							},
						},
					},
				},
			},
			want: types.MarketMap{
				Markets: map[string]types.Market{
					"BTC/USD": {
						Ticker: types.Ticker{
							CurrencyPair:     connecttypes.NewCurrencyPair("BTC", "USD"),
							Decimals:         10,
							MinProviderCount: 1,
							Enabled:          true,
						},
						ProviderConfigs: []types.ProviderConfig{
							{
								Name:           "test",
								OffChainTicker: "test_offchain",
							},
						},
					},
				},
			},
			wantRemovals:  []string{},
			updateEnabled: true,
			wantErr:       false,
		},
		{
			name:   "set all fields equal to actual market map if it is CROSS MARGIN --update-enabled",
			client: mockClient,
			expect: func(_ dydx.Client) {
				mockClient.EXPECT().AllPerpetuals(mock.Anything).Return(&dydx.AllPerpetualsResponse{
					Perpetuals: []dydx.Perpetual{
						{
							Params: dydx.PerpetualParams{
								Ticker:     "BTC-USD",
								MarketType: dydx.PERPETUAL_MARKET_TYPE_CROSS,
							},
						},
					},
				}, nil).Once()
			},
			actual: types.MarketMap{
				Markets: map[string]types.Market{
					"BTC/USD": {
						Ticker: types.Ticker{
							CurrencyPair:     connecttypes.NewCurrencyPair("BTC", "USD"),
							Decimals:         10,
							MinProviderCount: 1,
							Enabled:          true,
						},
						ProviderConfigs: []types.ProviderConfig{
							{
								Name:           "test",
								OffChainTicker: "test_offchain",
							},
						},
					},
				},
			},
			generated: types.MarketMap{
				Markets: map[string]types.Market{
					"BTC/USD": {
						Ticker: types.Ticker{
							CurrencyPair:     connecttypes.NewCurrencyPair("BTC", "USD"),
							Decimals:         1,
							MinProviderCount: 13,
							Enabled:          false,
						},
						ProviderConfigs: []types.ProviderConfig{
							{
								Name:           "test",
								OffChainTicker: "test_offchain",
							},
						},
					},
				},
			},
			want: types.MarketMap{
				Markets: map[string]types.Market{
					"BTC/USD": {
						Ticker: types.Ticker{
							CurrencyPair:     connecttypes.NewCurrencyPair("BTC", "USD"),
							Decimals:         10,
							MinProviderCount: 1,
							Enabled:          true,
						},
						ProviderConfigs: []types.ProviderConfig{
							{
								Name:           "test",
								OffChainTicker: "test_offchain",
							},
						},
					},
				},
			},
			wantRemovals:  []string{},
			updateEnabled: true,
			wantErr:       false,
		},
		{
			name:   "do nothing to generated market map ticker if it is ISOLATED and disabled",
			client: mockClient,
			expect: func(_ dydx.Client) {
				mockClient.EXPECT().AllPerpetuals(mock.Anything).Return(&dydx.AllPerpetualsResponse{
					Perpetuals: []dydx.Perpetual{
						{
							Params: dydx.PerpetualParams{
								Ticker:     "BTC-USD",
								MarketType: dydx.PERPETUAL_MARKET_TYPE_ISOLATED,
							},
						},
					},
				}, nil).Once()
			},
			actual: types.MarketMap{
				Markets: map[string]types.Market{
					"BTC/USD": {
						Ticker: types.Ticker{
							CurrencyPair:     connecttypes.NewCurrencyPair("BTC", "USD"),
							Decimals:         10,
							MinProviderCount: 1,
							Enabled:          false,
						},
						ProviderConfigs: []types.ProviderConfig{
							{
								Name:           "test",
								OffChainTicker: "test_offchain",
							},
						},
					},
				},
			},
			generated: types.MarketMap{
				Markets: map[string]types.Market{
					"BTC/USD": {
						Ticker: types.Ticker{
							CurrencyPair:     connecttypes.NewCurrencyPair("BTC", "USD"),
							Decimals:         1,
							MinProviderCount: 13,
							Enabled:          false,
						},
						ProviderConfigs: []types.ProviderConfig{
							{
								Name:           "test",
								OffChainTicker: "test_offchain",
							},
						},
					},
				},
			},
			want: types.MarketMap{
				Markets: map[string]types.Market{
					"BTC/USD": {
						Ticker: types.Ticker{
							CurrencyPair:     connecttypes.NewCurrencyPair("BTC", "USD"),
							Decimals:         10,
							MinProviderCount: 1,
							Enabled:          false,
						},
						ProviderConfigs: []types.ProviderConfig{
							{
								Name:           "test",
								OffChainTicker: "test_offchain",
							},
						},
					},
				},
			},
			wantRemovals:  []string{},
			updateEnabled: false,
			wantErr:       false,
		},
		{
			name:   "do nothing to generated market map ticker if it is ISOLATED and disabled --update-enabled",
			client: mockClient,
			expect: func(_ dydx.Client) {
				mockClient.EXPECT().AllPerpetuals(mock.Anything).Return(&dydx.AllPerpetualsResponse{
					Perpetuals: []dydx.Perpetual{
						{
							Params: dydx.PerpetualParams{
								Ticker:     "BTC-USD",
								MarketType: dydx.PERPETUAL_MARKET_TYPE_ISOLATED,
							},
						},
					},
				}, nil).Once()
			},
			actual: types.MarketMap{
				Markets: map[string]types.Market{
					"BTC/USD": {
						Ticker: types.Ticker{
							CurrencyPair:     connecttypes.NewCurrencyPair("BTC", "USD"),
							Decimals:         10,
							MinProviderCount: 1,
							Enabled:          false,
						},
						ProviderConfigs: []types.ProviderConfig{
							{
								Name:           "test",
								OffChainTicker: "test_offchain",
							},
						},
					},
				},
			},
			generated: types.MarketMap{
				Markets: map[string]types.Market{
					"BTC/USD": {
						Ticker: types.Ticker{
							CurrencyPair:     connecttypes.NewCurrencyPair("BTC", "USD"),
							Decimals:         1,
							MinProviderCount: 13,
							Enabled:          false,
						},
						ProviderConfigs: []types.ProviderConfig{
							{
								Name:           "test",
								OffChainTicker: "test_offchain",
							},
						},
					},
				},
			},
			want: types.MarketMap{
				Markets: map[string]types.Market{
					"BTC/USD": {
						Ticker: types.Ticker{
							CurrencyPair:     connecttypes.NewCurrencyPair("BTC", "USD"),
							Decimals:         10,
							MinProviderCount: 1,
							Enabled:          false,
						},
						ProviderConfigs: []types.ProviderConfig{
							{
								Name:           "test",
								OffChainTicker: "test_offchain",
							},
						},
					},
				},
			},
			wantRemovals:  []string{},
			updateEnabled: true,
			wantErr:       false,
		},
		{
			name:   "do nothing to generated market map ticker if it is ISOLATED and enabled",
			client: mockClient,
			expect: func(_ dydx.Client) {
				mockClient.EXPECT().AllPerpetuals(mock.Anything).Return(&dydx.AllPerpetualsResponse{
					Perpetuals: []dydx.Perpetual{
						{
							Params: dydx.PerpetualParams{
								Ticker:     "BTC-USD",
								MarketType: dydx.PERPETUAL_MARKET_TYPE_ISOLATED,
							},
						},
					},
				}, nil).Once()
			},
			actual: types.MarketMap{
				Markets: map[string]types.Market{
					"BTC/USD": {
						Ticker: types.Ticker{
							CurrencyPair:     connecttypes.NewCurrencyPair("BTC", "USD"),
							Decimals:         10,
							MinProviderCount: 1,
							Enabled:          true,
						},
						ProviderConfigs: []types.ProviderConfig{
							{
								Name:           "test",
								OffChainTicker: "test_offchain",
							},
						},
					},
				},
			},
			generated: types.MarketMap{
				Markets: map[string]types.Market{
					"BTC/USD": {
						Ticker: types.Ticker{
							CurrencyPair:     connecttypes.NewCurrencyPair("BTC", "USD"),
							Decimals:         1,
							MinProviderCount: 13,
							Enabled:          true,
						},
						ProviderConfigs: []types.ProviderConfig{
							{
								Name:           "test",
								OffChainTicker: "test_offchain",
							},
						},
					},
				},
			},
			want: types.MarketMap{
				Markets: map[string]types.Market{
					"BTC/USD": {
						Ticker: types.Ticker{
							CurrencyPair:     connecttypes.NewCurrencyPair("BTC", "USD"),
							Decimals:         10,
							MinProviderCount: 1,
							Enabled:          true,
						},
						ProviderConfigs: []types.ProviderConfig{
							{
								Name:           "test",
								OffChainTicker: "test_offchain",
							},
						},
					},
				},
			},
			wantRemovals:  []string{},
			updateEnabled: false,
			wantErr:       false,
		},
		{
			name:   "do nothing to generated market map ticker if it is ISOLATED and enabled --update-enabled",
			client: mockClient,
			expect: func(_ dydx.Client) {
				mockClient.EXPECT().AllPerpetuals(mock.Anything).Return(&dydx.AllPerpetualsResponse{
					Perpetuals: []dydx.Perpetual{
						{
							Params: dydx.PerpetualParams{
								Ticker:     "BTC-USD",
								MarketType: dydx.PERPETUAL_MARKET_TYPE_ISOLATED,
							},
						},
					},
				}, nil).Once()
			},
			actual: types.MarketMap{
				Markets: map[string]types.Market{
					"BTC/USD": {
						Ticker: types.Ticker{
							CurrencyPair:     connecttypes.NewCurrencyPair("BTC", "USD"),
							Decimals:         10,
							MinProviderCount: 1,
							Enabled:          true,
						},
						ProviderConfigs: []types.ProviderConfig{
							{
								Name:           "test",
								OffChainTicker: "test_offchain",
							},
						},
					},
				},
			},
			generated: types.MarketMap{
				Markets: map[string]types.Market{
					"BTC/USD": {
						Ticker: types.Ticker{
							CurrencyPair:     connecttypes.NewCurrencyPair("BTC", "USD"),
							Decimals:         1,
							MinProviderCount: 13,
							Enabled:          true,
						},
						ProviderConfigs: []types.ProviderConfig{
							{
								Name:           "test",
								OffChainTicker: "test_offchain",
							},
						},
					},
				},
			},
			want: types.MarketMap{
				Markets: map[string]types.Market{
					"BTC/USD": {
						Ticker: types.Ticker{
							CurrencyPair:     connecttypes.NewCurrencyPair("BTC", "USD"),
							Decimals:         10,
							MinProviderCount: 1,
							Enabled:          true,
						},
						ProviderConfigs: []types.ProviderConfig{
							{
								Name:           "test",
								OffChainTicker: "test_offchain",
							},
						},
					},
				},
			},
			wantRemovals:  []string{},
			updateEnabled: true,
			wantErr:       false,
		},
		{
			name:   "append to existing market map provider configs for isolated, disabled market",
			client: mockClient,
			expect: func(_ dydx.Client) {
				mockClient.EXPECT().AllPerpetuals(mock.Anything).Return(&dydx.AllPerpetualsResponse{
					Perpetuals: []dydx.Perpetual{
						{
							Params: dydx.PerpetualParams{
								Ticker:     "BTC-USD",
								MarketType: dydx.PERPETUAL_MARKET_TYPE_ISOLATED,
							},
						},
					},
				}, nil).Once()
			},
			actual: types.MarketMap{
				Markets: map[string]types.Market{
					"BTC/USD": {
						Ticker: types.Ticker{
							CurrencyPair:     connecttypes.NewCurrencyPair("BTC", "USD"),
							Decimals:         10,
							MinProviderCount: 1,
							Enabled:          false,
						},
						ProviderConfigs: []types.ProviderConfig{
							{
								Name:           "test",
								OffChainTicker: "test_offchain",
							},
						},
					},
				},
			},
			generated: types.MarketMap{
				Markets: map[string]types.Market{
					"BTC/USD": {
						Ticker: types.Ticker{
							CurrencyPair:     connecttypes.NewCurrencyPair("BTC", "USD"),
							Decimals:         1,
							MinProviderCount: 13,
							Enabled:          false,
						},
						ProviderConfigs: []types.ProviderConfig{
							{
								Name:           "test",
								OffChainTicker: "test_offchain",
							},
							{
								Name:           "test_new",
								OffChainTicker: "test_offchain_new",
							},
						},
					},
				},
			},
			want: types.MarketMap{
				Markets: map[string]types.Market{
					"BTC/USD": {
						Ticker: types.Ticker{
							CurrencyPair:     connecttypes.NewCurrencyPair("BTC", "USD"),
							Decimals:         10,
							MinProviderCount: 1,
							Enabled:          false,
						},
						ProviderConfigs: []types.ProviderConfig{
							{
								Name:           "test",
								OffChainTicker: "test_offchain",
							},
							{
								Name:           "test_new",
								OffChainTicker: "test_offchain_new",
							},
						},
					},
				},
			},
			wantRemovals:  []string{},
			updateEnabled: false,
			wantErr:       false,
		},
		{
			name:   "append to existing market map provider configs for isolated, disabled market --update-enabled",
			client: mockClient,
			expect: func(_ dydx.Client) {
				mockClient.EXPECT().AllPerpetuals(mock.Anything).Return(&dydx.AllPerpetualsResponse{
					Perpetuals: []dydx.Perpetual{
						{
							Params: dydx.PerpetualParams{
								Ticker:     "BTC-USD",
								MarketType: dydx.PERPETUAL_MARKET_TYPE_ISOLATED,
							},
						},
					},
				}, nil).Once()
			},
			actual: types.MarketMap{
				Markets: map[string]types.Market{
					"BTC/USD": {
						Ticker: types.Ticker{
							CurrencyPair:     connecttypes.NewCurrencyPair("BTC", "USD"),
							Decimals:         10,
							MinProviderCount: 1,
							Enabled:          false,
						},
						ProviderConfigs: []types.ProviderConfig{
							{
								Name:           "test",
								OffChainTicker: "test_offchain",
							},
						},
					},
				},
			},
			generated: types.MarketMap{
				Markets: map[string]types.Market{
					"BTC/USD": {
						Ticker: types.Ticker{
							CurrencyPair:     connecttypes.NewCurrencyPair("BTC", "USD"),
							Decimals:         1,
							MinProviderCount: 13,
							Enabled:          false,
						},
						ProviderConfigs: []types.ProviderConfig{
							{
								Name:           "test",
								OffChainTicker: "test_offchain",
							},
							{
								Name:           "test_new",
								OffChainTicker: "test_offchain_new",
							},
						},
					},
				},
			},
			want: types.MarketMap{
				Markets: map[string]types.Market{
					"BTC/USD": {
						Ticker: types.Ticker{
							CurrencyPair:     connecttypes.NewCurrencyPair("BTC", "USD"),
							Decimals:         10,
							MinProviderCount: 1,
							Enabled:          false,
						},
						ProviderConfigs: []types.ProviderConfig{
							{
								Name:           "test",
								OffChainTicker: "test_offchain",
							},
							{
								Name:           "test_new",
								OffChainTicker: "test_offchain_new",
							},
						},
					},
				},
			},
			wantRemovals:  []string{},
			updateEnabled: true,
			wantErr:       false,
		},
		{
			name:   "append to existing market map provider configs for isolated, enabled market if update enabled is true",
			client: mockClient,
			expect: func(_ dydx.Client) {
				mockClient.EXPECT().AllPerpetuals(mock.Anything).Return(&dydx.AllPerpetualsResponse{
					Perpetuals: []dydx.Perpetual{
						{
							Params: dydx.PerpetualParams{
								Ticker:     "BTC-USD",
								MarketType: dydx.PERPETUAL_MARKET_TYPE_ISOLATED,
							},
						},
					},
				}, nil).Once()
			},
			actual: types.MarketMap{
				Markets: map[string]types.Market{
					"BTC/USD": {
						Ticker: types.Ticker{
							CurrencyPair:     connecttypes.NewCurrencyPair("BTC", "USD"),
							Decimals:         10,
							MinProviderCount: 1,
							Enabled:          true,
						},
						ProviderConfigs: []types.ProviderConfig{
							{
								Name:           "test",
								OffChainTicker: "test_offchain",
							},
						},
					},
				},
			},
			generated: types.MarketMap{
				Markets: map[string]types.Market{
					"BTC/USD": {
						Ticker: types.Ticker{
							CurrencyPair:     connecttypes.NewCurrencyPair("BTC", "USD"),
							Decimals:         1,
							MinProviderCount: 13,
							Enabled:          true,
						},
						ProviderConfigs: []types.ProviderConfig{
							{
								Name:           "test",
								OffChainTicker: "test_offchain",
							},
							{
								Name:           "test_new",
								OffChainTicker: "test_offchain_new",
							},
						},
					},
				},
			},
			want: types.MarketMap{
				Markets: map[string]types.Market{
					"BTC/USD": {
						Ticker: types.Ticker{
							CurrencyPair:     connecttypes.NewCurrencyPair("BTC", "USD"),
							Decimals:         10,
							MinProviderCount: 1,
							Enabled:          true,
						},
						ProviderConfigs: []types.ProviderConfig{
							{
								Name:           "test",
								OffChainTicker: "test_offchain",
							},
							{
								Name:           "test_new",
								OffChainTicker: "test_offchain_new",
							},
						},
					},
				},
			},
			wantRemovals:  []string{},
			updateEnabled: true,
			wantErr:       false,
		},
		{
			name:   "do nothing for enabled market when update enabled is false",
			client: mockClient,
			expect: func(_ dydx.Client) {
				mockClient.EXPECT().AllPerpetuals(mock.Anything).Return(&dydx.AllPerpetualsResponse{
					Perpetuals: []dydx.Perpetual{
						{
							Params: dydx.PerpetualParams{
								Ticker:     "BTC-USD",
								MarketType: dydx.PERPETUAL_MARKET_TYPE_ISOLATED,
							},
						},
					},
				}, nil).Once()
			},
			actual: types.MarketMap{
				Markets: map[string]types.Market{
					"BTC/USD": {
						Ticker: types.Ticker{
							CurrencyPair:     connecttypes.NewCurrencyPair("BTC", "USD"),
							Decimals:         10,
							MinProviderCount: 1,
							Enabled:          true,
						},
						ProviderConfigs: []types.ProviderConfig{
							{
								Name:           "test",
								OffChainTicker: "test_offchain",
							},
						},
					},
				},
			},
			generated: types.MarketMap{
				Markets: map[string]types.Market{
					"BTC/USD": {
						Ticker: types.Ticker{
							CurrencyPair:     connecttypes.NewCurrencyPair("BTC", "USD"),
							Decimals:         1,
							MinProviderCount: 13,
							Enabled:          true,
						},
						ProviderConfigs: []types.ProviderConfig{
							{
								Name:           "test",
								OffChainTicker: "test_offchain",
							},
							{
								Name:           "test_new",
								OffChainTicker: "test_offchain_new",
							},
						},
					},
				},
			},
			want: types.MarketMap{
				Markets: map[string]types.Market{
					"BTC/USD": {
						Ticker: types.Ticker{
							CurrencyPair:     connecttypes.NewCurrencyPair("BTC", "USD"),
							Decimals:         10,
							MinProviderCount: 1,
							Enabled:          true,
						},
						ProviderConfigs: []types.ProviderConfig{
							{
								Name:           "test",
								OffChainTicker: "test_offchain",
							},
						},
					},
				},
			},
			wantRemovals:  []string{},
			updateEnabled: false,
			wantErr:       false,
		},
		{
			name:   "no updates to existing market map provider configs for cross, enabled market",
			client: mockClient,
			expect: func(_ dydx.Client) {
				mockClient.EXPECT().AllPerpetuals(mock.Anything).Return(&dydx.AllPerpetualsResponse{
					Perpetuals: []dydx.Perpetual{
						{
							Params: dydx.PerpetualParams{
								Ticker:     "BTC-USD",
								MarketType: dydx.PERPETUAL_MARKET_TYPE_CROSS,
							},
						},
					},
				}, nil).Once()
			},
			actual: types.MarketMap{
				Markets: map[string]types.Market{
					"BTC/USD": {
						Ticker: types.Ticker{
							CurrencyPair:     connecttypes.NewCurrencyPair("BTC", "USD"),
							Decimals:         10,
							MinProviderCount: 1,
							Enabled:          true,
						},
						ProviderConfigs: []types.ProviderConfig{
							{
								Name:           "test",
								OffChainTicker: "test_offchain",
							},
						},
					},
				},
			},
			generated: types.MarketMap{
				Markets: map[string]types.Market{
					"BTC/USD": {
						Ticker: types.Ticker{
							CurrencyPair:     connecttypes.NewCurrencyPair("BTC", "USD"),
							Decimals:         1,
							MinProviderCount: 13,
							Enabled:          true,
						},
						ProviderConfigs: []types.ProviderConfig{
							{
								Name:           "test",
								OffChainTicker: "test_offchain",
							},
							{
								Name:           "test_new",
								OffChainTicker: "test_offchain_new",
							},
						},
					},
				},
			},
			want: types.MarketMap{
				Markets: map[string]types.Market{
					"BTC/USD": {
						Ticker: types.Ticker{
							CurrencyPair:     connecttypes.NewCurrencyPair("BTC", "USD"),
							Decimals:         10,
							MinProviderCount: 1,
							Enabled:          true,
						},
						ProviderConfigs: []types.ProviderConfig{
							{
								Name:           "test",
								OffChainTicker: "test_offchain",
							},
						},
					},
				},
			},
			wantRemovals:  []string{},
			updateEnabled: false,
			wantErr:       false,
		},
		{
			name:   "no updates to existing market map provider configs for cross, enabled market --update-enabled",
			client: mockClient,
			expect: func(_ dydx.Client) {
				mockClient.EXPECT().AllPerpetuals(mock.Anything).Return(&dydx.AllPerpetualsResponse{
					Perpetuals: []dydx.Perpetual{
						{
							Params: dydx.PerpetualParams{
								Ticker:     "BTC-USD",
								MarketType: dydx.PERPETUAL_MARKET_TYPE_CROSS,
							},
						},
					},
				}, nil).Once()
			},
			actual: types.MarketMap{
				Markets: map[string]types.Market{
					"BTC/USD": {
						Ticker: types.Ticker{
							CurrencyPair:     connecttypes.NewCurrencyPair("BTC", "USD"),
							Decimals:         10,
							MinProviderCount: 1,
							Enabled:          true,
						},
						ProviderConfigs: []types.ProviderConfig{
							{
								Name:           "test",
								OffChainTicker: "test_offchain",
							},
						},
					},
				},
			},
			generated: types.MarketMap{
				Markets: map[string]types.Market{
					"BTC/USD": {
						Ticker: types.Ticker{
							CurrencyPair:     connecttypes.NewCurrencyPair("BTC", "USD"),
							Decimals:         1,
							MinProviderCount: 13,
							Enabled:          true,
						},
						ProviderConfigs: []types.ProviderConfig{
							{
								Name:           "test",
								OffChainTicker: "test_offchain",
							},
							{
								Name:           "test_new",
								OffChainTicker: "test_offchain_new",
							},
						},
					},
				},
			},
			want: types.MarketMap{
				Markets: map[string]types.Market{
					"BTC/USD": {
						Ticker: types.Ticker{
							CurrencyPair:     connecttypes.NewCurrencyPair("BTC", "USD"),
							Decimals:         10,
							MinProviderCount: 1,
							Enabled:          true,
						},
						ProviderConfigs: []types.ProviderConfig{
							{
								Name:           "test",
								OffChainTicker: "test_offchain",
							},
						},
					},
				},
			},
			wantRemovals:  []string{},
			updateEnabled: true,
			wantErr:       false,
		},
		{
			name:   "no updates to existing market map provider configs for cross, disabled market",
			client: mockClient,
			expect: func(_ dydx.Client) {
				mockClient.EXPECT().AllPerpetuals(mock.Anything).Return(&dydx.AllPerpetualsResponse{
					Perpetuals: []dydx.Perpetual{
						{
							Params: dydx.PerpetualParams{
								Ticker:     "BTC-USD",
								MarketType: dydx.PERPETUAL_MARKET_TYPE_CROSS,
							},
						},
					},
				}, nil).Once()
			},
			actual: types.MarketMap{
				Markets: map[string]types.Market{
					"BTC/USD": {
						Ticker: types.Ticker{
							CurrencyPair:     connecttypes.NewCurrencyPair("BTC", "USD"),
							Decimals:         10,
							MinProviderCount: 1,
							Enabled:          false,
						},
						ProviderConfigs: []types.ProviderConfig{
							{
								Name:           "test",
								OffChainTicker: "test_offchain",
							},
						},
					},
				},
			},
			generated: types.MarketMap{
				Markets: map[string]types.Market{
					"BTC/USD": {
						Ticker: types.Ticker{
							CurrencyPair:     connecttypes.NewCurrencyPair("BTC", "USD"),
							Decimals:         1,
							MinProviderCount: 13,
							Enabled:          true,
						},
						ProviderConfigs: []types.ProviderConfig{
							{
								Name:           "test",
								OffChainTicker: "test_offchain",
							},
							{
								Name:           "test_new",
								OffChainTicker: "test_offchain_new",
							},
						},
					},
				},
			},
			want: types.MarketMap{
				Markets: map[string]types.Market{
					"BTC/USD": {
						Ticker: types.Ticker{
							CurrencyPair:     connecttypes.NewCurrencyPair("BTC", "USD"),
							Decimals:         10,
							MinProviderCount: 1,
							Enabled:          false,
						},
						ProviderConfigs: []types.ProviderConfig{
							{
								Name:           "test",
								OffChainTicker: "test_offchain",
							},
						},
					},
				},
			},
			wantRemovals:  []string{},
			updateEnabled: false,
			wantErr:       false,
		},
		{
			name:   "no updates to existing market map provider configs for cross, disabled market --update-enabled",
			client: mockClient,
			expect: func(_ dydx.Client) {
				mockClient.EXPECT().AllPerpetuals(mock.Anything).Return(&dydx.AllPerpetualsResponse{
					Perpetuals: []dydx.Perpetual{
						{
							Params: dydx.PerpetualParams{
								Ticker:     "BTC-USD",
								MarketType: dydx.PERPETUAL_MARKET_TYPE_CROSS,
							},
						},
					},
				}, nil).Once()
			},
			actual: types.MarketMap{
				Markets: map[string]types.Market{
					"BTC/USD": {
						Ticker: types.Ticker{
							CurrencyPair:     connecttypes.NewCurrencyPair("BTC", "USD"),
							Decimals:         10,
							MinProviderCount: 1,
							Enabled:          false,
						},
						ProviderConfigs: []types.ProviderConfig{
							{
								Name:           "test",
								OffChainTicker: "test_offchain",
							},
						},
					},
				},
			},
			generated: types.MarketMap{
				Markets: map[string]types.Market{
					"BTC/USD": {
						Ticker: types.Ticker{
							CurrencyPair:     connecttypes.NewCurrencyPair("BTC", "USD"),
							Decimals:         1,
							MinProviderCount: 13,
							Enabled:          true,
						},
						ProviderConfigs: []types.ProviderConfig{
							{
								Name:           "test",
								OffChainTicker: "test_offchain",
							},
							{
								Name:           "test_new",
								OffChainTicker: "test_offchain_new",
							},
						},
					},
				},
			},
			want: types.MarketMap{
				Markets: map[string]types.Market{
					"BTC/USD": {
						Ticker: types.Ticker{
							CurrencyPair:     connecttypes.NewCurrencyPair("BTC", "USD"),
							Decimals:         10,
							MinProviderCount: 1,
							Enabled:          false,
						},
						ProviderConfigs: []types.ProviderConfig{
							{
								Name:           "test",
								OffChainTicker: "test_offchain",
							},
						},
					},
				},
			},
			wantRemovals:  []string{},
			updateEnabled: false,
			wantErr:       false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// set up mocks
			tt.expect(tt.client)
			marketOverride, err := NewDyDxOverride(tt.client)
			if tt.wantInitErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)

			got, removals, err := marketOverride.OverrideGeneratedMarkets(
				context.Background(),
				zaptest.NewLogger(t),
				tt.actual,
				tt.generated,
				update.Options{
					UpdateEnabled:      tt.updateEnabled,
					OverwriteProviders: false,
					ExistingOnly:       false,
				},
			)
			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			require.Equal(t, tt.want, got)
			require.Equal(t, tt.wantRemovals, removals)
		})
	}
}

func TestOverrideMarketMapOverwriteProviders(t *testing.T) {
	mockClient := mocks.NewClient(t)

	tests := []struct {
		name          string
		actual        types.MarketMap
		client        dydx.Client
		expect        func(client dydx.Client)
		generated     types.MarketMap
		want          types.MarketMap
		wantRemovals  []string
		updateEnabled bool
		wantInitErr   bool
		wantErr       bool
	}{
		{
			name:        "fail for a nil client",
			client:      nil,
			expect:      func(_ dydx.Client) {},
			wantInitErr: true,
		},
		{
			name:   "return error for nil response",
			client: mockClient,
			expect: func(_ dydx.Client) {
				mockClient.EXPECT().AllPerpetuals(mock.Anything).Return(nil, nil).Once()
			},
			actual: types.MarketMap{},
			generated: types.MarketMap{
				Markets: map[string]types.Market{
					"BTC/USD": {
						Ticker: types.Ticker{
							CurrencyPair:     connecttypes.NewCurrencyPair("BTC", "USD"),
							Decimals:         10,
							MinProviderCount: 1,
							Enabled:          false,
						},
						ProviderConfigs: []types.ProviderConfig{
							{
								Name:           "test",
								OffChainTicker: "test_offchain",
							},
						},
					},
				},
			},
			updateEnabled: false,
			wantErr:       true,
		},
		{
			name:   "override an empty market map",
			client: mockClient,
			expect: func(_ dydx.Client) {
				mockClient.EXPECT().AllPerpetuals(mock.Anything).Return(&dydx.AllPerpetualsResponse{}, nil).Once()
			},
			actual: types.MarketMap{},
			generated: types.MarketMap{
				Markets: map[string]types.Market{
					"BTC/USD": {
						Ticker: types.Ticker{
							CurrencyPair:     connecttypes.NewCurrencyPair("BTC", "USD"),
							Decimals:         10,
							MinProviderCount: 1,
							Enabled:          false,
						},
						ProviderConfigs: []types.ProviderConfig{
							{
								Name:           "test",
								OffChainTicker: "test_offchain",
							},
						},
					},
				},
			},
			want: types.MarketMap{
				Markets: map[string]types.Market{
					"BTC/USD": {
						Ticker: types.Ticker{
							CurrencyPair:     connecttypes.NewCurrencyPair("BTC", "USD"),
							Decimals:         10,
							MinProviderCount: 1,
							Enabled:          false,
						},
						ProviderConfigs: []types.ProviderConfig{
							{
								Name:           "test",
								OffChainTicker: "test_offchain",
							},
						},
					},
				},
			},
			wantRemovals:  []string{},
			updateEnabled: false,
			wantErr:       false,
		},
		{
			name:   "override an empty generated market map with enabled actual market",
			client: mockClient,
			expect: func(_ dydx.Client) {
				mockClient.EXPECT().AllPerpetuals(mock.Anything).Return(&dydx.AllPerpetualsResponse{}, nil).Once()
			},
			actual: types.MarketMap{
				Markets: map[string]types.Market{
					"BTC/USD": {
						Ticker: types.Ticker{
							CurrencyPair:     connecttypes.NewCurrencyPair("BTC", "USD"),
							Decimals:         10,
							MinProviderCount: 1,
							Enabled:          true,
						},
						ProviderConfigs: []types.ProviderConfig{
							{
								Name:           "test",
								OffChainTicker: "test_offchain",
							},
						},
					},
				},
			},
			generated: types.MarketMap{},
			want: types.MarketMap{
				Markets: map[string]types.Market{
					"BTC/USD": {
						Ticker: types.Ticker{
							CurrencyPair:     connecttypes.NewCurrencyPair("BTC", "USD"),
							Decimals:         10,
							MinProviderCount: 1,
							Enabled:          true,
						},
						ProviderConfigs: []types.ProviderConfig{
							{
								Name:           "test",
								OffChainTicker: "test_offchain",
							},
						},
					},
				},
			},
			wantRemovals:  []string{},
			updateEnabled: false,
			wantErr:       false,
		},
		{
			name:   "override an empty generated market map with disabled actual market",
			client: mockClient,
			expect: func(_ dydx.Client) {
				mockClient.EXPECT().AllPerpetuals(mock.Anything).Return(&dydx.AllPerpetualsResponse{}, nil).Once()
			},
			actual: types.MarketMap{
				Markets: map[string]types.Market{
					"BTC/USD": {
						Ticker: types.Ticker{
							CurrencyPair:     connecttypes.NewCurrencyPair("BTC", "USD"),
							Decimals:         10,
							MinProviderCount: 1,
							Enabled:          false,
						},
						ProviderConfigs: []types.ProviderConfig{
							{
								Name:           "test",
								OffChainTicker: "test_offchain",
							},
						},
					},
				},
			},
			generated:     types.MarketMap{},
			want:          types.MarketMap{Markets: map[string]types.Market{}},
			wantRemovals:  []string{"BTC/USD"},
			updateEnabled: false,
			wantErr:       false,
		},
		{
			name:   "disable a market that was enabled in the generated market map but does not exist in actual",
			client: mockClient,
			expect: func(_ dydx.Client) {
				mockClient.EXPECT().AllPerpetuals(mock.Anything).Return(&dydx.AllPerpetualsResponse{}, nil).Once()
			},
			actual: types.MarketMap{},
			generated: types.MarketMap{
				Markets: map[string]types.Market{
					"BTC/USD": {
						Ticker: types.Ticker{
							CurrencyPair:     connecttypes.NewCurrencyPair("BTC", "USD"),
							Decimals:         10,
							MinProviderCount: 1,
							Enabled:          true,
						},
						ProviderConfigs: []types.ProviderConfig{
							{
								Name:           "test",
								OffChainTicker: "test_offchain",
							},
						},
					},
				},
			},
			want: types.MarketMap{
				Markets: map[string]types.Market{
					"BTC/USD": {
						Ticker: types.Ticker{
							CurrencyPair:     connecttypes.NewCurrencyPair("BTC", "USD"),
							Decimals:         10,
							MinProviderCount: 1,
							Enabled:          false,
						},
						ProviderConfigs: []types.ProviderConfig{
							{
								Name:           "test",
								OffChainTicker: "test_offchain",
							},
						},
					},
				},
			},
			wantRemovals:  []string{},
			updateEnabled: false,
			wantErr:       false,
		},
		{
			name:   "do nothing if there is no diff between generated and generated",
			client: mockClient,
			expect: func(_ dydx.Client) {
				mockClient.EXPECT().AllPerpetuals(mock.Anything).Return(&dydx.AllPerpetualsResponse{}, nil).Once()
			},
			actual: types.MarketMap{
				Markets: map[string]types.Market{
					"BTC/USD": {
						Ticker: types.Ticker{
							CurrencyPair:     connecttypes.NewCurrencyPair("BTC", "USD"),
							Decimals:         10,
							MinProviderCount: 1,
							Enabled:          false,
						},
						ProviderConfigs: []types.ProviderConfig{
							{
								Name:           "test",
								OffChainTicker: "test_offchain",
							},
						},
					},
				},
			},
			generated: types.MarketMap{
				Markets: map[string]types.Market{
					"BTC/USD": {
						Ticker: types.Ticker{
							CurrencyPair:     connecttypes.NewCurrencyPair("BTC", "USD"),
							Decimals:         10,
							MinProviderCount: 1,
							Enabled:          false,
						},
						ProviderConfigs: []types.ProviderConfig{
							{
								Name:           "test",
								OffChainTicker: "test_offchain",
							},
						},
					},
				},
			},
			want: types.MarketMap{
				Markets: map[string]types.Market{
					"BTC/USD": {
						Ticker: types.Ticker{
							CurrencyPair:     connecttypes.NewCurrencyPair("BTC", "USD"),
							Decimals:         10,
							MinProviderCount: 1,
							Enabled:          false,
						},
						ProviderConfigs: []types.ProviderConfig{
							{
								Name:           "test",
								OffChainTicker: "test_offchain",
							},
						},
					},
				},
			},
			wantRemovals:  []string{},
			updateEnabled: false,
			wantErr:       false,
		},
		{
			name:   "enable a market that is enabled on chain, but disabled in generated",
			client: mockClient,
			expect: func(_ dydx.Client) {
				mockClient.EXPECT().AllPerpetuals(mock.Anything).Return(&dydx.AllPerpetualsResponse{}, nil).Once()
			},
			actual: types.MarketMap{
				Markets: map[string]types.Market{
					"BTC/USD": {
						Ticker: types.Ticker{
							CurrencyPair:     connecttypes.NewCurrencyPair("BTC", "USD"),
							Decimals:         10,
							MinProviderCount: 1,
							Enabled:          true,
						},
						ProviderConfigs: []types.ProviderConfig{
							{
								Name:           "test",
								OffChainTicker: "test_offchain",
							},
						},
					},
				},
			},
			generated: types.MarketMap{
				Markets: map[string]types.Market{
					"BTC/USD": {
						Ticker: types.Ticker{
							CurrencyPair:     connecttypes.NewCurrencyPair("BTC", "USD"),
							Decimals:         10,
							MinProviderCount: 1,
							Enabled:          false,
						},
						ProviderConfigs: []types.ProviderConfig{
							{
								Name:           "test",
								OffChainTicker: "test_offchain",
							},
						},
					},
				},
			},
			want: types.MarketMap{
				Markets: map[string]types.Market{
					"BTC/USD": {
						Ticker: types.Ticker{
							CurrencyPair:     connecttypes.NewCurrencyPair("BTC", "USD"),
							Decimals:         10,
							MinProviderCount: 1,
							Enabled:          true,
						},
						ProviderConfigs: []types.ProviderConfig{
							{
								Name:           "test",
								OffChainTicker: "test_offchain",
							},
						},
					},
				},
			},
			wantRemovals:  []string{},
			updateEnabled: false,
			wantErr:       false,
		},
		{
			name:   "override decimals and min provider count",
			client: mockClient,
			expect: func(_ dydx.Client) {
				mockClient.EXPECT().AllPerpetuals(mock.Anything).Return(&dydx.AllPerpetualsResponse{}, nil).Once()
			},
			actual: types.MarketMap{
				Markets: map[string]types.Market{
					"BTC/USD": {
						Ticker: types.Ticker{
							CurrencyPair:     connecttypes.NewCurrencyPair("BTC", "USD"),
							Decimals:         11,
							MinProviderCount: 4,
							Enabled:          true,
						},
						ProviderConfigs: []types.ProviderConfig{
							{
								Name:           "test",
								OffChainTicker: "test_offchain",
							},
						},
					},
				},
			},
			generated: types.MarketMap{
				Markets: map[string]types.Market{
					"BTC/USD": {
						Ticker: types.Ticker{
							CurrencyPair:     connecttypes.NewCurrencyPair("BTC", "USD"),
							Decimals:         10,
							MinProviderCount: 1,
							Enabled:          true,
						},
						ProviderConfigs: []types.ProviderConfig{
							{
								Name:           "test",
								OffChainTicker: "test_offchain",
							},
						},
					},
				},
			},
			want: types.MarketMap{
				Markets: map[string]types.Market{
					"BTC/USD": {
						Ticker: types.Ticker{
							CurrencyPair:     connecttypes.NewCurrencyPair("BTC", "USD"),
							Decimals:         11,
							MinProviderCount: 4,
							Enabled:          true,
						},
						ProviderConfigs: []types.ProviderConfig{
							{
								Name:           "test",
								OffChainTicker: "test_offchain",
							},
						},
					},
				},
			},
			wantRemovals:  []string{},
			updateEnabled: false,
			wantErr:       false,
		},
		{
			name:   "keep existing provider ticker for enabled market",
			client: mockClient,
			expect: func(_ dydx.Client) {
				mockClient.EXPECT().AllPerpetuals(mock.Anything).Return(&dydx.AllPerpetualsResponse{}, nil).Once()
			},
			actual: types.MarketMap{
				Markets: map[string]types.Market{
					"BTC/USD": {
						Ticker: types.Ticker{
							CurrencyPair:     connecttypes.NewCurrencyPair("BTC", "USD"),
							Decimals:         10,
							MinProviderCount: 1,
							Enabled:          true,
						},
						ProviderConfigs: []types.ProviderConfig{
							{
								Name:           "test",
								OffChainTicker: "test_offchain",
							},
						},
					},
				},
			},
			generated: types.MarketMap{
				Markets: map[string]types.Market{
					"BTC/USD": {
						Ticker: types.Ticker{
							CurrencyPair:     connecttypes.NewCurrencyPair("BTC", "USD"),
							Decimals:         10,
							MinProviderCount: 1,
							Enabled:          true,
						},
						ProviderConfigs: []types.ProviderConfig{
							{
								Name:           "test",
								OffChainTicker: "test_offchain_new",
							},
						},
					},
				},
			},
			want: types.MarketMap{
				Markets: map[string]types.Market{
					"BTC/USD": {
						Ticker: types.Ticker{
							CurrencyPair:     connecttypes.NewCurrencyPair("BTC", "USD"),
							Decimals:         10,
							MinProviderCount: 1,
							Enabled:          true,
						},
						ProviderConfigs: []types.ProviderConfig{
							{
								Name:           "test",
								OffChainTicker: "test_offchain",
							},
						},
					},
				},
			},
			wantRemovals:  []string{},
			updateEnabled: false,
			wantErr:       false,
		},
		{
			name:   "overwrite provider ticker for disabled market",
			client: mockClient,
			expect: func(_ dydx.Client) {
				mockClient.EXPECT().AllPerpetuals(mock.Anything).Return(&dydx.AllPerpetualsResponse{}, nil).Once()
			},
			actual: types.MarketMap{
				Markets: map[string]types.Market{
					"BTC/USD": {
						Ticker: types.Ticker{
							CurrencyPair:     connecttypes.NewCurrencyPair("BTC", "USD"),
							Decimals:         10,
							MinProviderCount: 1,
							Enabled:          false,
						},
						ProviderConfigs: []types.ProviderConfig{
							{
								Name:           "test",
								OffChainTicker: "test_offchain",
							},
						},
					},
				},
			},
			generated: types.MarketMap{
				Markets: map[string]types.Market{
					"BTC/USD": {
						Ticker: types.Ticker{
							CurrencyPair:     connecttypes.NewCurrencyPair("BTC", "USD"),
							Decimals:         10,
							MinProviderCount: 1,
							Enabled:          false,
						},
						ProviderConfigs: []types.ProviderConfig{
							{
								Name:           "test",
								OffChainTicker: "test_offchain_new",
							},
						},
					},
				},
			},
			want: types.MarketMap{
				Markets: map[string]types.Market{
					"BTC/USD": {
						Ticker: types.Ticker{
							CurrencyPair:     connecttypes.NewCurrencyPair("BTC", "USD"),
							Decimals:         10,
							MinProviderCount: 1,
							Enabled:          false,
						},
						ProviderConfigs: []types.ProviderConfig{
							{
								Name:           "test",
								OffChainTicker: "test_offchain_new",
							},
						},
					},
				},
			},
			wantRemovals:  []string{},
			updateEnabled: false,
			wantErr:       false,
		},
		{
			name:   "overwrite provider configs - disjoint providers",
			client: mockClient,
			expect: func(_ dydx.Client) {
				mockClient.EXPECT().AllPerpetuals(mock.Anything).Return(&dydx.AllPerpetualsResponse{}, nil).Once()
			},
			actual: types.MarketMap{
				Markets: map[string]types.Market{
					"BTC/USD": {
						Ticker: types.Ticker{
							CurrencyPair:     connecttypes.NewCurrencyPair("BTC", "USD"),
							Decimals:         10,
							MinProviderCount: 1,
							Enabled:          false,
						},
						ProviderConfigs: []types.ProviderConfig{
							{
								Name:           "test",
								OffChainTicker: "test_offchain",
							},
						},
					},
				},
			},
			generated: types.MarketMap{
				Markets: map[string]types.Market{
					"BTC/USD": {
						Ticker: types.Ticker{
							CurrencyPair:     connecttypes.NewCurrencyPair("BTC", "USD"),
							Decimals:         10,
							MinProviderCount: 1,
							Enabled:          false,
						},
						ProviderConfigs: []types.ProviderConfig{
							{
								Name:           "test_new",
								OffChainTicker: "test_offchain_new",
							},
						},
					},
				},
			},
			want: types.MarketMap{
				Markets: map[string]types.Market{
					"BTC/USD": {
						Ticker: types.Ticker{
							CurrencyPair:     connecttypes.NewCurrencyPair("BTC", "USD"),
							Decimals:         10,
							MinProviderCount: 1,
							Enabled:          false,
						},
						ProviderConfigs: []types.ProviderConfig{
							{
								Name:           "test_new",
								OffChainTicker: "test_offchain_new",
							},
						},
					},
				},
			},
			wantRemovals:  []string{},
			updateEnabled: false,
			wantErr:       false,
		},
		{
			name:   "overwrite market to existing one - overlapping provider configs",
			client: mockClient,
			expect: func(_ dydx.Client) {
				mockClient.EXPECT().AllPerpetuals(mock.Anything).Return(&dydx.AllPerpetualsResponse{}, nil).Once()
			},
			actual: types.MarketMap{
				Markets: map[string]types.Market{
					"BTC/USD": {
						Ticker: types.Ticker{
							CurrencyPair:     connecttypes.NewCurrencyPair("BTC", "USD"),
							Decimals:         10,
							MinProviderCount: 1,
							Enabled:          false,
						},
						ProviderConfigs: []types.ProviderConfig{
							{
								Name:           "test",
								OffChainTicker: "test_offchain",
							},
						},
					},
				},
			},
			generated: types.MarketMap{
				Markets: map[string]types.Market{
					"BTC/USD": {
						Ticker: types.Ticker{
							CurrencyPair:     connecttypes.NewCurrencyPair("BTC", "USD"),
							Decimals:         10,
							MinProviderCount: 1,
							Enabled:          false,
						},
						ProviderConfigs: []types.ProviderConfig{
							{
								Name:           "test",
								OffChainTicker: "test_offchain",
							},
							{
								Name:           "test_new",
								OffChainTicker: "test_offchain_new",
							},
						},
					},
				},
			},
			want: types.MarketMap{
				Markets: map[string]types.Market{
					"BTC/USD": {
						Ticker: types.Ticker{
							CurrencyPair:     connecttypes.NewCurrencyPair("BTC", "USD"),
							Decimals:         10,
							MinProviderCount: 1,
							Enabled:          false,
						},
						ProviderConfigs: []types.ProviderConfig{
							{
								Name:           "test",
								OffChainTicker: "test_offchain",
							},
							{
								Name:           "test_new",
								OffChainTicker: "test_offchain_new",
							},
						},
					},
				},
			},
			wantRemovals:  []string{},
			updateEnabled: false,
			wantErr:       false,
		},
		{
			name:   "error if perpetual market is not in actual market map",
			client: mockClient,
			expect: func(_ dydx.Client) {
				mockClient.EXPECT().AllPerpetuals(mock.Anything).Return(&dydx.AllPerpetualsResponse{
					Perpetuals: []dydx.Perpetual{
						{
							Params: dydx.PerpetualParams{
								Ticker:     "BTC-USD",
								MarketType: dydx.PERPETUAL_MARKET_TYPE_CROSS,
							},
						},
					},
				}, nil).Once()
			},
			wantRemovals:  []string{},
			updateEnabled: false,
			wantErr:       true,
		},
		{
			name:   "set all fields equal to actual market map if it is CROSS MARGIN",
			client: mockClient,
			expect: func(_ dydx.Client) {
				mockClient.EXPECT().AllPerpetuals(mock.Anything).Return(&dydx.AllPerpetualsResponse{
					Perpetuals: []dydx.Perpetual{
						{
							Params: dydx.PerpetualParams{
								Ticker:     "BTC-USD",
								MarketType: dydx.PERPETUAL_MARKET_TYPE_CROSS,
							},
						},
					},
				}, nil).Once()
			},
			actual: types.MarketMap{
				Markets: map[string]types.Market{
					"BTC/USD": {
						Ticker: types.Ticker{
							CurrencyPair:     connecttypes.NewCurrencyPair("BTC", "USD"),
							Decimals:         10,
							MinProviderCount: 1,
							Enabled:          true,
						},
						ProviderConfigs: []types.ProviderConfig{
							{
								Name:           "test",
								OffChainTicker: "test_offchain",
							},
						},
					},
				},
			},
			generated: types.MarketMap{
				Markets: map[string]types.Market{
					"BTC/USD": {
						Ticker: types.Ticker{
							CurrencyPair:     connecttypes.NewCurrencyPair("BTC", "USD"),
							Decimals:         1,
							MinProviderCount: 13,
							Enabled:          false,
						},
						ProviderConfigs: []types.ProviderConfig{
							{
								Name:           "test",
								OffChainTicker: "test_offchain",
							},
						},
					},
				},
			},
			want: types.MarketMap{
				Markets: map[string]types.Market{
					"BTC/USD": {
						Ticker: types.Ticker{
							CurrencyPair:     connecttypes.NewCurrencyPair("BTC", "USD"),
							Decimals:         10,
							MinProviderCount: 1,
							Enabled:          true,
						},
						ProviderConfigs: []types.ProviderConfig{
							{
								Name:           "test",
								OffChainTicker: "test_offchain",
							},
						},
					},
				},
			},
			wantRemovals:  []string{},
			updateEnabled: false,
			wantErr:       false,
		},
		{
			name:   "set all fields equal to actual market map if it is CROSS MARGIN --update-enabled",
			client: mockClient,
			expect: func(_ dydx.Client) {
				mockClient.EXPECT().AllPerpetuals(mock.Anything).Return(&dydx.AllPerpetualsResponse{
					Perpetuals: []dydx.Perpetual{
						{
							Params: dydx.PerpetualParams{
								Ticker:     "BTC-USD",
								MarketType: dydx.PERPETUAL_MARKET_TYPE_CROSS,
							},
						},
					},
				}, nil).Once()
			},
			actual: types.MarketMap{
				Markets: map[string]types.Market{
					"BTC/USD": {
						Ticker: types.Ticker{
							CurrencyPair:     connecttypes.NewCurrencyPair("BTC", "USD"),
							Decimals:         10,
							MinProviderCount: 1,
							Enabled:          true,
						},
						ProviderConfigs: []types.ProviderConfig{
							{
								Name:           "test",
								OffChainTicker: "test_offchain",
							},
						},
					},
				},
			},
			generated: types.MarketMap{
				Markets: map[string]types.Market{
					"BTC/USD": {
						Ticker: types.Ticker{
							CurrencyPair:     connecttypes.NewCurrencyPair("BTC", "USD"),
							Decimals:         1,
							MinProviderCount: 13,
							Enabled:          false,
						},
						ProviderConfigs: []types.ProviderConfig{
							{
								Name:           "test",
								OffChainTicker: "test_offchain",
							},
						},
					},
				},
			},
			want: types.MarketMap{
				Markets: map[string]types.Market{
					"BTC/USD": {
						Ticker: types.Ticker{
							CurrencyPair:     connecttypes.NewCurrencyPair("BTC", "USD"),
							Decimals:         10,
							MinProviderCount: 1,
							Enabled:          true,
						},
						ProviderConfigs: []types.ProviderConfig{
							{
								Name:           "test",
								OffChainTicker: "test_offchain",
							},
						},
					},
				},
			},
			wantRemovals:  []string{},
			updateEnabled: true,
			wantErr:       false,
		},
		{
			name:   "do nothing to generated market map ticker if it is ISOLATED and disabled",
			client: mockClient,
			expect: func(_ dydx.Client) {
				mockClient.EXPECT().AllPerpetuals(mock.Anything).Return(&dydx.AllPerpetualsResponse{
					Perpetuals: []dydx.Perpetual{
						{
							Params: dydx.PerpetualParams{
								Ticker:     "BTC-USD",
								MarketType: dydx.PERPETUAL_MARKET_TYPE_ISOLATED,
							},
						},
					},
				}, nil).Once()
			},
			actual: types.MarketMap{
				Markets: map[string]types.Market{
					"BTC/USD": {
						Ticker: types.Ticker{
							CurrencyPair:     connecttypes.NewCurrencyPair("BTC", "USD"),
							Decimals:         10,
							MinProviderCount: 1,
							Enabled:          false,
						},
						ProviderConfigs: []types.ProviderConfig{
							{
								Name:           "test",
								OffChainTicker: "test_offchain",
							},
						},
					},
				},
			},
			generated: types.MarketMap{
				Markets: map[string]types.Market{
					"BTC/USD": {
						Ticker: types.Ticker{
							CurrencyPair:     connecttypes.NewCurrencyPair("BTC", "USD"),
							Decimals:         1,
							MinProviderCount: 13,
							Enabled:          false,
						},
						ProviderConfigs: []types.ProviderConfig{
							{
								Name:           "test",
								OffChainTicker: "test_offchain",
							},
						},
					},
				},
			},
			want: types.MarketMap{
				Markets: map[string]types.Market{
					"BTC/USD": {
						Ticker: types.Ticker{
							CurrencyPair:     connecttypes.NewCurrencyPair("BTC", "USD"),
							Decimals:         10,
							MinProviderCount: 1,
							Enabled:          false,
						},
						ProviderConfigs: []types.ProviderConfig{
							{
								Name:           "test",
								OffChainTicker: "test_offchain",
							},
						},
					},
				},
			},
			wantRemovals:  []string{},
			updateEnabled: false,
			wantErr:       false,
		},
		{
			name:   "do nothing to generated market map ticker if it is ISOLATED and disabled --update-enabled",
			client: mockClient,
			expect: func(_ dydx.Client) {
				mockClient.EXPECT().AllPerpetuals(mock.Anything).Return(&dydx.AllPerpetualsResponse{
					Perpetuals: []dydx.Perpetual{
						{
							Params: dydx.PerpetualParams{
								Ticker:     "BTC-USD",
								MarketType: dydx.PERPETUAL_MARKET_TYPE_ISOLATED,
							},
						},
					},
				}, nil).Once()
			},
			actual: types.MarketMap{
				Markets: map[string]types.Market{
					"BTC/USD": {
						Ticker: types.Ticker{
							CurrencyPair:     connecttypes.NewCurrencyPair("BTC", "USD"),
							Decimals:         10,
							MinProviderCount: 1,
							Enabled:          false,
						},
						ProviderConfigs: []types.ProviderConfig{
							{
								Name:           "test",
								OffChainTicker: "test_offchain",
							},
						},
					},
				},
			},
			generated: types.MarketMap{
				Markets: map[string]types.Market{
					"BTC/USD": {
						Ticker: types.Ticker{
							CurrencyPair:     connecttypes.NewCurrencyPair("BTC", "USD"),
							Decimals:         1,
							MinProviderCount: 13,
							Enabled:          false,
						},
						ProviderConfigs: []types.ProviderConfig{
							{
								Name:           "test",
								OffChainTicker: "test_offchain",
							},
						},
					},
				},
			},
			want: types.MarketMap{
				Markets: map[string]types.Market{
					"BTC/USD": {
						Ticker: types.Ticker{
							CurrencyPair:     connecttypes.NewCurrencyPair("BTC", "USD"),
							Decimals:         10,
							MinProviderCount: 1,
							Enabled:          false,
						},
						ProviderConfigs: []types.ProviderConfig{
							{
								Name:           "test",
								OffChainTicker: "test_offchain",
							},
						},
					},
				},
			},
			wantRemovals:  []string{},
			updateEnabled: true,
			wantErr:       false,
		},
		{
			name:   "do nothing to generated market map ticker if it is ISOLATED and enabled",
			client: mockClient,
			expect: func(_ dydx.Client) {
				mockClient.EXPECT().AllPerpetuals(mock.Anything).Return(&dydx.AllPerpetualsResponse{
					Perpetuals: []dydx.Perpetual{
						{
							Params: dydx.PerpetualParams{
								Ticker:     "BTC-USD",
								MarketType: dydx.PERPETUAL_MARKET_TYPE_ISOLATED,
							},
						},
					},
				}, nil).Once()
			},
			actual: types.MarketMap{
				Markets: map[string]types.Market{
					"BTC/USD": {
						Ticker: types.Ticker{
							CurrencyPair:     connecttypes.NewCurrencyPair("BTC", "USD"),
							Decimals:         10,
							MinProviderCount: 1,
							Enabled:          true,
						},
						ProviderConfigs: []types.ProviderConfig{
							{
								Name:           "test",
								OffChainTicker: "test_offchain",
							},
						},
					},
				},
			},
			generated: types.MarketMap{
				Markets: map[string]types.Market{
					"BTC/USD": {
						Ticker: types.Ticker{
							CurrencyPair:     connecttypes.NewCurrencyPair("BTC", "USD"),
							Decimals:         1,
							MinProviderCount: 13,
							Enabled:          true,
						},
						ProviderConfigs: []types.ProviderConfig{
							{
								Name:           "test",
								OffChainTicker: "test_offchain",
							},
						},
					},
				},
			},
			want: types.MarketMap{
				Markets: map[string]types.Market{
					"BTC/USD": {
						Ticker: types.Ticker{
							CurrencyPair:     connecttypes.NewCurrencyPair("BTC", "USD"),
							Decimals:         10,
							MinProviderCount: 1,
							Enabled:          true,
						},
						ProviderConfigs: []types.ProviderConfig{
							{
								Name:           "test",
								OffChainTicker: "test_offchain",
							},
						},
					},
				},
			},
			wantRemovals:  []string{},
			updateEnabled: false,
			wantErr:       false,
		},
		{
			name:   "do nothing to generated market map ticker if it is ISOLATED and enabled --update-enabled",
			client: mockClient,
			expect: func(_ dydx.Client) {
				mockClient.EXPECT().AllPerpetuals(mock.Anything).Return(&dydx.AllPerpetualsResponse{
					Perpetuals: []dydx.Perpetual{
						{
							Params: dydx.PerpetualParams{
								Ticker:     "BTC-USD",
								MarketType: dydx.PERPETUAL_MARKET_TYPE_ISOLATED,
							},
						},
					},
				}, nil).Once()
			},
			actual: types.MarketMap{
				Markets: map[string]types.Market{
					"BTC/USD": {
						Ticker: types.Ticker{
							CurrencyPair:     connecttypes.NewCurrencyPair("BTC", "USD"),
							Decimals:         10,
							MinProviderCount: 1,
							Enabled:          true,
						},
						ProviderConfigs: []types.ProviderConfig{
							{
								Name:           "test",
								OffChainTicker: "test_offchain",
							},
						},
					},
				},
			},
			generated: types.MarketMap{
				Markets: map[string]types.Market{
					"BTC/USD": {
						Ticker: types.Ticker{
							CurrencyPair:     connecttypes.NewCurrencyPair("BTC", "USD"),
							Decimals:         1,
							MinProviderCount: 13,
							Enabled:          true,
						},
						ProviderConfigs: []types.ProviderConfig{
							{
								Name:           "test",
								OffChainTicker: "test_offchain",
							},
						},
					},
				},
			},
			want: types.MarketMap{
				Markets: map[string]types.Market{
					"BTC/USD": {
						Ticker: types.Ticker{
							CurrencyPair:     connecttypes.NewCurrencyPair("BTC", "USD"),
							Decimals:         10,
							MinProviderCount: 1,
							Enabled:          true,
						},
						ProviderConfigs: []types.ProviderConfig{
							{
								Name:           "test",
								OffChainTicker: "test_offchain",
							},
						},
					},
				},
			},
			wantRemovals:  []string{},
			updateEnabled: true,
			wantErr:       false,
		},
		{
			name:   "append to existing market map provider configs for isolated, disabled market",
			client: mockClient,
			expect: func(_ dydx.Client) {
				mockClient.EXPECT().AllPerpetuals(mock.Anything).Return(&dydx.AllPerpetualsResponse{
					Perpetuals: []dydx.Perpetual{
						{
							Params: dydx.PerpetualParams{
								Ticker:     "BTC-USD",
								MarketType: dydx.PERPETUAL_MARKET_TYPE_ISOLATED,
							},
						},
					},
				}, nil).Once()
			},
			actual: types.MarketMap{
				Markets: map[string]types.Market{
					"BTC/USD": {
						Ticker: types.Ticker{
							CurrencyPair:     connecttypes.NewCurrencyPair("BTC", "USD"),
							Decimals:         10,
							MinProviderCount: 1,
							Enabled:          false,
						},
						ProviderConfigs: []types.ProviderConfig{
							{
								Name:           "test",
								OffChainTicker: "test_offchain",
							},
						},
					},
				},
			},
			generated: types.MarketMap{
				Markets: map[string]types.Market{
					"BTC/USD": {
						Ticker: types.Ticker{
							CurrencyPair:     connecttypes.NewCurrencyPair("BTC", "USD"),
							Decimals:         1,
							MinProviderCount: 13,
							Enabled:          false,
						},
						ProviderConfigs: []types.ProviderConfig{
							{
								Name:           "test",
								OffChainTicker: "test_offchain",
							},
							{
								Name:           "test_new",
								OffChainTicker: "test_offchain_new",
							},
						},
					},
				},
			},
			want: types.MarketMap{
				Markets: map[string]types.Market{
					"BTC/USD": {
						Ticker: types.Ticker{
							CurrencyPair:     connecttypes.NewCurrencyPair("BTC", "USD"),
							Decimals:         10,
							MinProviderCount: 1,
							Enabled:          false,
						},
						ProviderConfigs: []types.ProviderConfig{
							{
								Name:           "test",
								OffChainTicker: "test_offchain",
							},
							{
								Name:           "test_new",
								OffChainTicker: "test_offchain_new",
							},
						},
					},
				},
			},
			wantRemovals:  []string{},
			updateEnabled: false,
			wantErr:       false,
		},
		{
			name:   "append to existing market map provider configs for isolated, disabled market --update-enabled",
			client: mockClient,
			expect: func(_ dydx.Client) {
				mockClient.EXPECT().AllPerpetuals(mock.Anything).Return(&dydx.AllPerpetualsResponse{
					Perpetuals: []dydx.Perpetual{
						{
							Params: dydx.PerpetualParams{
								Ticker:     "BTC-USD",
								MarketType: dydx.PERPETUAL_MARKET_TYPE_ISOLATED,
							},
						},
					},
				}, nil).Once()
			},
			actual: types.MarketMap{
				Markets: map[string]types.Market{
					"BTC/USD": {
						Ticker: types.Ticker{
							CurrencyPair:     connecttypes.NewCurrencyPair("BTC", "USD"),
							Decimals:         10,
							MinProviderCount: 1,
							Enabled:          false,
						},
						ProviderConfigs: []types.ProviderConfig{
							{
								Name:           "test",
								OffChainTicker: "test_offchain",
							},
						},
					},
				},
			},
			generated: types.MarketMap{
				Markets: map[string]types.Market{
					"BTC/USD": {
						Ticker: types.Ticker{
							CurrencyPair:     connecttypes.NewCurrencyPair("BTC", "USD"),
							Decimals:         1,
							MinProviderCount: 13,
							Enabled:          false,
						},
						ProviderConfigs: []types.ProviderConfig{
							{
								Name:           "test",
								OffChainTicker: "test_offchain",
							},
							{
								Name:           "test_new",
								OffChainTicker: "test_offchain_new",
							},
						},
					},
				},
			},
			want: types.MarketMap{
				Markets: map[string]types.Market{
					"BTC/USD": {
						Ticker: types.Ticker{
							CurrencyPair:     connecttypes.NewCurrencyPair("BTC", "USD"),
							Decimals:         10,
							MinProviderCount: 1,
							Enabled:          false,
						},
						ProviderConfigs: []types.ProviderConfig{
							{
								Name:           "test",
								OffChainTicker: "test_offchain",
							},
							{
								Name:           "test_new",
								OffChainTicker: "test_offchain_new",
							},
						},
					},
				},
			},
			wantRemovals:  []string{},
			updateEnabled: true,
			wantErr:       false,
		},
		{
			name:   "append to existing market map provider configs for isolated, enabled market if update enabled is true",
			client: mockClient,
			expect: func(_ dydx.Client) {
				mockClient.EXPECT().AllPerpetuals(mock.Anything).Return(&dydx.AllPerpetualsResponse{
					Perpetuals: []dydx.Perpetual{
						{
							Params: dydx.PerpetualParams{
								Ticker:     "BTC-USD",
								MarketType: dydx.PERPETUAL_MARKET_TYPE_ISOLATED,
							},
						},
					},
				}, nil).Once()
			},
			actual: types.MarketMap{
				Markets: map[string]types.Market{
					"BTC/USD": {
						Ticker: types.Ticker{
							CurrencyPair:     connecttypes.NewCurrencyPair("BTC", "USD"),
							Decimals:         10,
							MinProviderCount: 1,
							Enabled:          true,
						},
						ProviderConfigs: []types.ProviderConfig{
							{
								Name:           "test",
								OffChainTicker: "test_offchain",
							},
						},
					},
				},
			},
			generated: types.MarketMap{
				Markets: map[string]types.Market{
					"BTC/USD": {
						Ticker: types.Ticker{
							CurrencyPair:     connecttypes.NewCurrencyPair("BTC", "USD"),
							Decimals:         1,
							MinProviderCount: 13,
							Enabled:          true,
						},
						ProviderConfigs: []types.ProviderConfig{
							{
								Name:           "test",
								OffChainTicker: "test_offchain",
							},
							{
								Name:           "test_new",
								OffChainTicker: "test_offchain_new",
							},
						},
					},
				},
			},
			want: types.MarketMap{
				Markets: map[string]types.Market{
					"BTC/USD": {
						Ticker: types.Ticker{
							CurrencyPair:     connecttypes.NewCurrencyPair("BTC", "USD"),
							Decimals:         10,
							MinProviderCount: 1,
							Enabled:          true,
						},
						ProviderConfigs: []types.ProviderConfig{
							{
								Name:           "test",
								OffChainTicker: "test_offchain",
							},
							{
								Name:           "test_new",
								OffChainTicker: "test_offchain_new",
							},
						},
					},
				},
			},
			wantRemovals:  []string{},
			updateEnabled: true,
			wantErr:       false,
		},
		{
			name:   "do nothing for enabled market when update enabled is false",
			client: mockClient,
			expect: func(_ dydx.Client) {
				mockClient.EXPECT().AllPerpetuals(mock.Anything).Return(&dydx.AllPerpetualsResponse{
					Perpetuals: []dydx.Perpetual{
						{
							Params: dydx.PerpetualParams{
								Ticker:     "BTC-USD",
								MarketType: dydx.PERPETUAL_MARKET_TYPE_ISOLATED,
							},
						},
					},
				}, nil).Once()
			},
			actual: types.MarketMap{
				Markets: map[string]types.Market{
					"BTC/USD": {
						Ticker: types.Ticker{
							CurrencyPair:     connecttypes.NewCurrencyPair("BTC", "USD"),
							Decimals:         10,
							MinProviderCount: 1,
							Enabled:          true,
						},
						ProviderConfigs: []types.ProviderConfig{
							{
								Name:           "test",
								OffChainTicker: "test_offchain",
							},
						},
					},
				},
			},
			generated: types.MarketMap{
				Markets: map[string]types.Market{
					"BTC/USD": {
						Ticker: types.Ticker{
							CurrencyPair:     connecttypes.NewCurrencyPair("BTC", "USD"),
							Decimals:         1,
							MinProviderCount: 13,
							Enabled:          true,
						},
						ProviderConfigs: []types.ProviderConfig{
							{
								Name:           "test",
								OffChainTicker: "test_offchain",
							},
							{
								Name:           "test_new",
								OffChainTicker: "test_offchain_new",
							},
						},
					},
				},
			},
			want: types.MarketMap{
				Markets: map[string]types.Market{
					"BTC/USD": {
						Ticker: types.Ticker{
							CurrencyPair:     connecttypes.NewCurrencyPair("BTC", "USD"),
							Decimals:         10,
							MinProviderCount: 1,
							Enabled:          true,
						},
						ProviderConfigs: []types.ProviderConfig{
							{
								Name:           "test",
								OffChainTicker: "test_offchain",
							},
						},
					},
				},
			},
			wantRemovals:  []string{},
			updateEnabled: false,
			wantErr:       false,
		},
		{
			name:   "no updates to existing market map provider configs for cross, enabled market",
			client: mockClient,
			expect: func(_ dydx.Client) {
				mockClient.EXPECT().AllPerpetuals(mock.Anything).Return(&dydx.AllPerpetualsResponse{
					Perpetuals: []dydx.Perpetual{
						{
							Params: dydx.PerpetualParams{
								Ticker:     "BTC-USD",
								MarketType: dydx.PERPETUAL_MARKET_TYPE_CROSS,
							},
						},
					},
				}, nil).Once()
			},
			actual: types.MarketMap{
				Markets: map[string]types.Market{
					"BTC/USD": {
						Ticker: types.Ticker{
							CurrencyPair:     connecttypes.NewCurrencyPair("BTC", "USD"),
							Decimals:         10,
							MinProviderCount: 1,
							Enabled:          true,
						},
						ProviderConfigs: []types.ProviderConfig{
							{
								Name:           "test",
								OffChainTicker: "test_offchain",
							},
						},
					},
				},
			},
			generated: types.MarketMap{
				Markets: map[string]types.Market{
					"BTC/USD": {
						Ticker: types.Ticker{
							CurrencyPair:     connecttypes.NewCurrencyPair("BTC", "USD"),
							Decimals:         1,
							MinProviderCount: 13,
							Enabled:          true,
						},
						ProviderConfigs: []types.ProviderConfig{
							{
								Name:           "test",
								OffChainTicker: "test_offchain",
							},
							{
								Name:           "test_new",
								OffChainTicker: "test_offchain_new",
							},
						},
					},
				},
			},
			want: types.MarketMap{
				Markets: map[string]types.Market{
					"BTC/USD": {
						Ticker: types.Ticker{
							CurrencyPair:     connecttypes.NewCurrencyPair("BTC", "USD"),
							Decimals:         10,
							MinProviderCount: 1,
							Enabled:          true,
						},
						ProviderConfigs: []types.ProviderConfig{
							{
								Name:           "test",
								OffChainTicker: "test_offchain",
							},
						},
					},
				},
			},
			wantRemovals:  []string{},
			updateEnabled: false,
			wantErr:       false,
		},
		{
			name:   "no updates to existing market map provider configs for cross, enabled market --update-enabled",
			client: mockClient,
			expect: func(_ dydx.Client) {
				mockClient.EXPECT().AllPerpetuals(mock.Anything).Return(&dydx.AllPerpetualsResponse{
					Perpetuals: []dydx.Perpetual{
						{
							Params: dydx.PerpetualParams{
								Ticker:     "BTC-USD",
								MarketType: dydx.PERPETUAL_MARKET_TYPE_CROSS,
							},
						},
					},
				}, nil).Once()
			},
			actual: types.MarketMap{
				Markets: map[string]types.Market{
					"BTC/USD": {
						Ticker: types.Ticker{
							CurrencyPair:     connecttypes.NewCurrencyPair("BTC", "USD"),
							Decimals:         10,
							MinProviderCount: 1,
							Enabled:          true,
						},
						ProviderConfigs: []types.ProviderConfig{
							{
								Name:           "test",
								OffChainTicker: "test_offchain",
							},
						},
					},
				},
			},
			generated: types.MarketMap{
				Markets: map[string]types.Market{
					"BTC/USD": {
						Ticker: types.Ticker{
							CurrencyPair:     connecttypes.NewCurrencyPair("BTC", "USD"),
							Decimals:         1,
							MinProviderCount: 13,
							Enabled:          true,
						},
						ProviderConfigs: []types.ProviderConfig{
							{
								Name:           "test",
								OffChainTicker: "test_offchain",
							},
							{
								Name:           "test_new",
								OffChainTicker: "test_offchain_new",
							},
						},
					},
				},
			},
			want: types.MarketMap{
				Markets: map[string]types.Market{
					"BTC/USD": {
						Ticker: types.Ticker{
							CurrencyPair:     connecttypes.NewCurrencyPair("BTC", "USD"),
							Decimals:         10,
							MinProviderCount: 1,
							Enabled:          true,
						},
						ProviderConfigs: []types.ProviderConfig{
							{
								Name:           "test",
								OffChainTicker: "test_offchain",
							},
						},
					},
				},
			},
			wantRemovals:  []string{},
			updateEnabled: true,
			wantErr:       false,
		},
		{
			name:   "no updates to existing market map provider configs for cross, disabled market",
			client: mockClient,
			expect: func(_ dydx.Client) {
				mockClient.EXPECT().AllPerpetuals(mock.Anything).Return(&dydx.AllPerpetualsResponse{
					Perpetuals: []dydx.Perpetual{
						{
							Params: dydx.PerpetualParams{
								Ticker:     "BTC-USD",
								MarketType: dydx.PERPETUAL_MARKET_TYPE_CROSS,
							},
						},
					},
				}, nil).Once()
			},
			actual: types.MarketMap{
				Markets: map[string]types.Market{
					"BTC/USD": {
						Ticker: types.Ticker{
							CurrencyPair:     connecttypes.NewCurrencyPair("BTC", "USD"),
							Decimals:         10,
							MinProviderCount: 1,
							Enabled:          false,
						},
						ProviderConfigs: []types.ProviderConfig{
							{
								Name:           "test",
								OffChainTicker: "test_offchain",
							},
						},
					},
				},
			},
			generated: types.MarketMap{
				Markets: map[string]types.Market{
					"BTC/USD": {
						Ticker: types.Ticker{
							CurrencyPair:     connecttypes.NewCurrencyPair("BTC", "USD"),
							Decimals:         1,
							MinProviderCount: 13,
							Enabled:          true,
						},
						ProviderConfigs: []types.ProviderConfig{
							{
								Name:           "test",
								OffChainTicker: "test_offchain",
							},
							{
								Name:           "test_new",
								OffChainTicker: "test_offchain_new",
							},
						},
					},
				},
			},
			want: types.MarketMap{
				Markets: map[string]types.Market{
					"BTC/USD": {
						Ticker: types.Ticker{
							CurrencyPair:     connecttypes.NewCurrencyPair("BTC", "USD"),
							Decimals:         10,
							MinProviderCount: 1,
							Enabled:          false,
						},
						ProviderConfigs: []types.ProviderConfig{
							{
								Name:           "test",
								OffChainTicker: "test_offchain",
							},
						},
					},
				},
			},
			wantRemovals:  []string{},
			updateEnabled: false,
			wantErr:       false,
		},
		{
			name:   "no updates to existing market map provider configs for cross, disabled market --update-enabled",
			client: mockClient,
			expect: func(_ dydx.Client) {
				mockClient.EXPECT().AllPerpetuals(mock.Anything).Return(&dydx.AllPerpetualsResponse{
					Perpetuals: []dydx.Perpetual{
						{
							Params: dydx.PerpetualParams{
								Ticker:     "BTC-USD",
								MarketType: dydx.PERPETUAL_MARKET_TYPE_CROSS,
							},
						},
					},
				}, nil).Once()
			},
			actual: types.MarketMap{
				Markets: map[string]types.Market{
					"BTC/USD": {
						Ticker: types.Ticker{
							CurrencyPair:     connecttypes.NewCurrencyPair("BTC", "USD"),
							Decimals:         10,
							MinProviderCount: 1,
							Enabled:          false,
						},
						ProviderConfigs: []types.ProviderConfig{
							{
								Name:           "test",
								OffChainTicker: "test_offchain",
							},
						},
					},
				},
			},
			generated: types.MarketMap{
				Markets: map[string]types.Market{
					"BTC/USD": {
						Ticker: types.Ticker{
							CurrencyPair:     connecttypes.NewCurrencyPair("BTC", "USD"),
							Decimals:         1,
							MinProviderCount: 13,
							Enabled:          true,
						},
						ProviderConfigs: []types.ProviderConfig{
							{
								Name:           "test",
								OffChainTicker: "test_offchain",
							},
							{
								Name:           "test_new",
								OffChainTicker: "test_offchain_new",
							},
						},
					},
				},
			},
			want: types.MarketMap{
				Markets: map[string]types.Market{
					"BTC/USD": {
						Ticker: types.Ticker{
							CurrencyPair:     connecttypes.NewCurrencyPair("BTC", "USD"),
							Decimals:         10,
							MinProviderCount: 1,
							Enabled:          false,
						},
						ProviderConfigs: []types.ProviderConfig{
							{
								Name:           "test",
								OffChainTicker: "test_offchain",
							},
						},
					},
				},
			},
			wantRemovals:  []string{},
			updateEnabled: false,
			wantErr:       false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// set up mocks
			tt.expect(tt.client)
			marketOverride, err := NewDyDxOverride(tt.client)
			if tt.wantInitErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)

			got, removals, err := marketOverride.OverrideGeneratedMarkets(
				context.Background(),
				zaptest.NewLogger(t),
				tt.actual,
				tt.generated,
				update.Options{
					UpdateEnabled:      tt.updateEnabled,
					OverwriteProviders: true,
					ExistingOnly:       false,
				},
			)
			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			require.Equal(t, tt.want, got)
			require.Equal(t, tt.wantRemovals, removals)
		})
	}
}
