package update

import (
	"testing"

	connecttypes "github.com/skip-mev/connect/v2/pkg/types"
	"github.com/skip-mev/connect/v2/x/marketmap/types"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
)

func TestCombineMarketMap(t *testing.T) {
	tests := []struct {
		name      string
		actual    types.MarketMap
		generated types.MarketMap
		options   Options
		want      types.MarketMap
		wantErr   bool
	}{
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
