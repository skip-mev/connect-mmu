package update

import (
	"slices"
	"testing"

	connecttypes "github.com/skip-mev/connect/v2/pkg/types"
	"github.com/skip-mev/connect/v2/x/marketmap/types"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
)

func TestMergeCMCIDMarkets(t *testing.T) {
	tests := []struct {
		name           string
		mm             types.MarketMap
		cmcIDToTickers map[string][]string
		expected       types.MarketMap
		expErr         bool
	}{
		{
			name: "simple merge",
			mm: types.MarketMap{Markets: map[string]types.Market{
				"FOO/USD": {
					Ticker:          types.Ticker{CurrencyPair: connecttypes.CurrencyPair{Base: "FOO", Quote: "USD"}},
					ProviderConfigs: []types.ProviderConfig{{Name: "coinbase"}},
				},
				"FOO,UNISWAP,0XFOO/USD": {ProviderConfigs: []types.ProviderConfig{{Name: "uniswap"}}},
			}},
			cmcIDToTickers: map[string][]string{
				"2": {"FOO/USD", "FOO,UNISWAP,0XFOO/USD"},
			},
			expected: types.MarketMap{Markets: map[string]types.Market{
				"FOO/USD": {
					Ticker: types.Ticker{CurrencyPair: connecttypes.CurrencyPair{Base: "FOO", Quote: "USD"}},
					ProviderConfigs: []types.ProviderConfig{
						{Name: "coinbase"},
						{Name: "uniswap"},
					},
				},
			}},
		},
		{
			name: "defi merge",
			mm: types.MarketMap{Markets: map[string]types.Market{
				"FOO,UNISWAP,0XFOO/USD":   {ProviderConfigs: []types.ProviderConfig{{Name: "uniswap"}}},
				"FOO,RAYDIUM,ABCDEFG/USD": {ProviderConfigs: []types.ProviderConfig{{Name: "raydium"}}},
			}},
			cmcIDToTickers: map[string][]string{
				"500": {"FOO,UNISWAP,0XFOO/USD", "FOO,RAYDIUM,ABCDEFG/USD"},
			},
			expected: types.MarketMap{Markets: map[string]types.Market{
				"FOO/USD": {
					Ticker: types.Ticker{CurrencyPair: connecttypes.CurrencyPair{Base: "FOO", Quote: "USD"}},
					ProviderConfigs: []types.ProviderConfig{
						{Name: "raydium"},
						{Name: "uniswap"},
					},
				},
			}},
		},
		{
			name: "no clobbered markets",
			mm: types.MarketMap{Markets: map[string]types.Market{
				"FOO,RAYDIUM,SLDKJFLKSDJF/USD": {ProviderConfigs: []types.ProviderConfig{{Name: "raydium"}}},
				"FOO/USD": {
					Ticker: types.Ticker{CurrencyPair: connecttypes.CurrencyPair{Base: "FOO", Quote: "USD"}},
					ProviderConfigs: []types.ProviderConfig{
						{Name: "coinbase"},
						{Name: "binance"},
					},
				},
				"BAR/USD": {
					Ticker: types.Ticker{CurrencyPair: connecttypes.CurrencyPair{Base: "BAR", Quote: "USD"}},
					ProviderConfigs: []types.ProviderConfig{
						{Name: "coinbase"},
					},
				},
			}},
			cmcIDToTickers: map[string][]string{
				"10": {"FOO,RAYDIUM,SLDKJFLKSDJF/USD", "FOO/USD"},
			},
			expected: types.MarketMap{Markets: map[string]types.Market{
				"FOO/USD": {
					Ticker:          types.Ticker{CurrencyPair: connecttypes.CurrencyPair{Base: "FOO", Quote: "USD"}},
					ProviderConfigs: []types.ProviderConfig{{Name: "binance"}, {Name: "coinbase"}, {Name: "raydium"}},
				},
				"BAR/USD": {
					Ticker:          types.Ticker{CurrencyPair: connecttypes.CurrencyPair{Base: "BAR", Quote: "USD"}},
					ProviderConfigs: []types.ProviderConfig{{Name: "coinbase"}},
				},
			}},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			out, err := mergeCMCMIDMarkets(tc.mm, tc.cmcIDToTickers)
			if tc.expErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.True(t, out.Equal(tc.expected), out)
			}
		})
	}
}

func TestDeconstructDefiTicker(t *testing.T) {
	tests := []struct {
		name     string
		ticker   string
		expected string
		expErr   bool
	}{
		{
			name:     "valid ticker",
			ticker:   "BLUE,RAYDIUM,CWQVQTKUH1IU8ZSFFFVAUXAVZLZQU1E8GYU5D6ECGBNE/USD",
			expected: "BLUE/USD",
		},
		{
			name:   "invalid ticker - no separator",
			ticker: "BLUE,RAYDIUM,CWQVQTKUH1IU8ZSFFFVAUXAVZLZQU1E8GYU5D6ECGBNEUSD",
			expErr: true,
		},
		{
			name:   "invalid ticker - invalid parts",
			ticker: "BLUE,RAY,DIUM,CWQVQTKUH1IU8ZSFFFVAUXAVZLZQU1E8GYU5D6ECGBNE/USD",
			expErr: true,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			pair, err := deconstructDeFiTicker(tc.ticker)
			if tc.expErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.expected, pair.String())
			}
		})
	}
}

func TestGetCMCIDMapping(t *testing.T) {
	tests := []struct {
		name     string
		in       types.MarketMap
		expected map[string][]string
	}{
		{
			name: "CMC IDs are extracted",
			in: types.MarketMap{
				Markets: map[string]types.Market{
					"FOO/USD": {
						Ticker:          types.Ticker{Metadata_JSON: "{\"reference_price\":1786788632,\"liquidity\":184445,\"aggregate_ids\":[{\"venue\":\"coinmarketcap\",\"ID\":\"32349\"}]}"},
						ProviderConfigs: nil,
					},
					"FOO,UNISWAP,0XFOO/USD": {
						Ticker: types.Ticker{Metadata_JSON: "{\"reference_price\":1786788632,\"liquidity\":184445,\"aggregate_ids\":[{\"venue\":\"coinmarketcap\",\"ID\":\"32349\"}]}"},
					},
					"BAR/USD": {
						Ticker: types.Ticker{Metadata_JSON: "{\"reference_price\":1786788632,\"liquidity\":184445,\"aggregate_ids\":[{\"venue\":\"coinmarketcap\",\"ID\":\"2\"}]}"},
					},
					"BAZ/USD": {},
				},
			},
			expected: map[string][]string{
				"32349": {"FOO/USD", "FOO,UNISWAP,0XFOO/USD"},
				"2":     {"BAR/USD"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out, err := getCMCTickerMapping(tt.in)
			require.NoError(t, err)
			for id, tickers := range out {
				expected, ok := tt.expected[id]
				require.True(t, ok)
				slices.Sort(expected)
				slices.Sort(tickers)
				require.Equal(t, expected, tickers)
			}
		})
	}
}

func TestAppendIfNotExists(t *testing.T) {
	tests := []struct {
		name       string
		src        []types.ProviderConfig
		newConfigs []types.ProviderConfig
		expected   []types.ProviderConfig
	}{
		{
			name:       "provider configs appended",
			src:        []types.ProviderConfig{{Name: "foo"}},
			newConfigs: []types.ProviderConfig{{Name: "bar"}},
			expected:   []types.ProviderConfig{{Name: "foo"}, {Name: "bar"}},
		},
		{
			name:       "not appended if exists",
			src:        []types.ProviderConfig{{Name: "foo"}, {Name: "bar"}},
			newConfigs: []types.ProviderConfig{{Name: "bar"}},
			expected:   []types.ProviderConfig{{Name: "foo"}, {Name: "bar"}},
		},
		{
			name:       "empty appends all",
			src:        []types.ProviderConfig{},
			newConfigs: []types.ProviderConfig{{Name: "foo"}, {Name: "bar"}},
			expected:   []types.ProviderConfig{{Name: "foo"}, {Name: "bar"}},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			out := appendIfNotExists(tc.src, tc.newConfigs)
			require.Equal(t, tc.expected, out)
		})
	}
}

func TestCombineMarketMap(t *testing.T) {
	cmcID1 := "{\"aggregate_ids\":[{\"venue\":\"coinmarketcap\",\"ID\":\"1\"}]}"
	cmcID2 := "{\"aggregate_ids\":[{\"venue\":\"coinmarketcap\",\"ID\":\"2\"}]}"
	cmcID3 := "{\"aggregate_ids\":[{\"venue\":\"coinmarketcap\",\"ID\":\"3\"}]}"
	tests := []struct {
		name      string
		actual    types.MarketMap
		generated types.MarketMap
		options   Options
		want      types.MarketMap
		wantErr   bool
	}{
		{
			name: "CMC ID match markets consolidate",
			actual: types.MarketMap{
				Markets: map[string]types.Market{
					"FOO/USD": {
						Ticker:          types.Ticker{CurrencyPair: connecttypes.CurrencyPair{Base: "FOO", Quote: "USD"}, Metadata_JSON: cmcID1},
						ProviderConfigs: []types.ProviderConfig{{Name: "foo"}, {Name: "bar"}},
					},
					"BUX/USD": {
						Ticker:          types.Ticker{CurrencyPair: connecttypes.CurrencyPair{Base: "BUX", Quote: "USD"}, Metadata_JSON: cmcID3},
						ProviderConfigs: []types.ProviderConfig{{Name: "bux"}},
					},
					"BAR,UNISWAP,0XFOOBAR/USD": {
						Ticker:          types.Ticker{CurrencyPair: connecttypes.CurrencyPair{Base: "BAR,UNISWAP,0XFOOBAR", Quote: "USD"}, Metadata_JSON: cmcID2},
						ProviderConfigs: []types.ProviderConfig{{Name: "uniswap"}},
					},
					"BAZ/USD": { // non cmc id should just match on ticker..
						Ticker:          types.Ticker{CurrencyPair: connecttypes.CurrencyPair{Base: "BAZ", Quote: "USD"}},
						ProviderConfigs: []types.ProviderConfig{{Name: "foo"}},
					},
				},
			},
			generated: types.MarketMap{
				Markets: map[string]types.Market{
					"BUX/USD": { // should merge with non-defi market above
						Ticker:          types.Ticker{CurrencyPair: connecttypes.CurrencyPair{Base: "BUX", Quote: "USD"}, Metadata_JSON: cmcID3},
						ProviderConfigs: []types.ProviderConfig{{Name: "baz"}},
					},
					"FOO/USD": { // should merge into non defi market above
						Ticker:          types.Ticker{CurrencyPair: connecttypes.CurrencyPair{Base: "FOO", Quote: "USD"}, Metadata_JSON: cmcID1},
						ProviderConfigs: []types.ProviderConfig{{Name: "baz"}},
					},
					"FOO,UNISWAP,0XFOOBAR/USD": { // should consolidate with non-defi market above
						Ticker:          types.Ticker{CurrencyPair: connecttypes.CurrencyPair{Base: "FOO,UNISWAP,0XFOOBAR", Quote: "USD"}, Metadata_JSON: cmcID1},
						ProviderConfigs: []types.ProviderConfig{{Name: "uniswap"}},
					},
					"BAR,RAYDIUM,0XFOOBAR/USD": { // should merge with other defi market above.
						Ticker:          types.Ticker{CurrencyPair: connecttypes.CurrencyPair{Base: "BAR,RAYDIUM,0XFOOBAR", Quote: "USD"}, Metadata_JSON: cmcID2},
						ProviderConfigs: []types.ProviderConfig{{Name: "raydium"}},
					},
					"BAZ/USD": { // this non-cmc ID market should just merge into the above one based on ticker.
						Ticker:          types.Ticker{CurrencyPair: connecttypes.CurrencyPair{Base: "BAZ", Quote: "USD"}},
						ProviderConfigs: []types.ProviderConfig{{Name: "bar"}},
					},
				},
			},
			want: types.MarketMap{
				Markets: map[string]types.Market{
					"FOO/USD": {
						Ticker:          types.Ticker{CurrencyPair: connecttypes.CurrencyPair{Base: "FOO", Quote: "USD"}, Metadata_JSON: cmcID1},
						ProviderConfigs: []types.ProviderConfig{{Name: "bar"}, {Name: "baz"}, {Name: "foo"}, {Name: "uniswap"}},
					},
					"BAZ/USD": {
						Ticker:          types.Ticker{CurrencyPair: connecttypes.CurrencyPair{Base: "BAZ", Quote: "USD"}},
						ProviderConfigs: []types.ProviderConfig{{Name: "bar"}, {Name: "foo"}},
					},
					"BAR/USD": {
						Ticker:          types.Ticker{CurrencyPair: connecttypes.CurrencyPair{Base: "BAR", Quote: "USD"}, Metadata_JSON: cmcID2},
						ProviderConfigs: []types.ProviderConfig{{Name: "raydium"}, {Name: "uniswap"}},
					},
					"BUX/USD": {
						Ticker:          types.Ticker{CurrencyPair: connecttypes.CurrencyPair{Base: "BUX", Quote: "USD"}, Metadata_JSON: cmcID3},
						ProviderConfigs: []types.ProviderConfig{{Name: "baz"}, {Name: "bux"}},
					},
				},
			},
		},
		{
			name: "do nothing for empty - nil",
			want: types.MarketMap{
				Markets: make(map[string]types.Market),
			},
		},
		{
			name: "do nothing for empty",
			actual: types.MarketMap{
				Markets: make(map[string]types.Market),
			},
			generated: types.MarketMap{
				Markets: make(map[string]types.Market),
			},
			want: types.MarketMap{
				Markets: make(map[string]types.Market),
			},
		},
		{
			name:   "override an empty market map",
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
			wantErr: false,
		},
		{
			name:   "disable a market that was enabled in the generated market map but does not exist in actual",
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
			wantErr: false,
		},
		{
			name: "do nothing if there is no diff between generated and generated",
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
			wantErr: false,
		},
		{
			name: "enable a market that is enabled on chain, but disabled in generated",
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
			wantErr: false,
		},
		{
			name: "override decimals and min provider count",
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
			wantErr: false,
		},
		{
			name: "keep existing provider ticker for enabled market",
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
			wantErr: false,
		},
		{
			name: "keep existing provider ticker for disabled market",
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
			wantErr: false,
		},
		{
			name: "append market to existing one - disjoint provider configs",
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
			wantErr: false,
		},
		{
			name: "append market to existing one - overlapping provider configs",
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
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := CombineMarketMaps(zaptest.NewLogger(t), tt.actual, tt.generated, tt.options)
			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			require.Equal(t, tt.want, got)
		})
	}
}
