package diffs_test

import (
	"testing"

	"github.com/skip-mev/connect-mmu/diffs"

	connecttypes "github.com/skip-mev/connect/v2/pkg/types"
	mmtypes "github.com/skip-mev/connect/v2/x/marketmap/types"
	"github.com/stretchr/testify/require"
)

func TestFilterMarketUpdates(t *testing.T) {
	tt := []struct {
		name          string
		currentMarket mmtypes.Market
		update        mmtypes.Market
		expected      mmtypes.Market
	}{
		{
			name: "no updates",
			currentMarket: mmtypes.Market{
				Ticker: mmtypes.Ticker{
					CurrencyPair: connecttypes.NewCurrencyPair("BTC", "USD"),
				},
				ProviderConfigs: []mmtypes.ProviderConfig{
					{
						Name: "provider1",
					},
				},
			},
			update: mmtypes.Market{
				Ticker: mmtypes.Ticker{
					CurrencyPair: connecttypes.NewCurrencyPair("BTC", "USD"),
				},
				ProviderConfigs: []mmtypes.ProviderConfig{
					{
						Name: "provider1",
					},
				},
			},
			expected: mmtypes.Market{
				Ticker: mmtypes.Ticker{
					CurrencyPair: connecttypes.NewCurrencyPair("BTC", "USD"),
				},
			},
		},
		{
			name: "decimals changed",
			currentMarket: mmtypes.Market{
				Ticker: mmtypes.Ticker{
					CurrencyPair: connecttypes.NewCurrencyPair("BTC", "USD"),
					Decimals:     8,
				},
				ProviderConfigs: []mmtypes.ProviderConfig{
					{
						Name: "provider1",
					},
				},
			},
			update: mmtypes.Market{
				Ticker: mmtypes.Ticker{
					CurrencyPair: connecttypes.NewCurrencyPair("BTC", "USD"),
					Decimals:     9,
				},
				ProviderConfigs: []mmtypes.ProviderConfig{
					{
						Name: "provider1",
					},
				},
			},
			expected: mmtypes.Market{
				Ticker: mmtypes.Ticker{
					CurrencyPair: connecttypes.NewCurrencyPair("BTC", "USD"),
					Decimals:     9,
				},
			},
		},
		{
			name: "min provider count changed",
			currentMarket: mmtypes.Market{
				Ticker: mmtypes.Ticker{
					CurrencyPair:     connecttypes.NewCurrencyPair("BTC", "USD"),
					Decimals:         8,
					MinProviderCount: 2,
				},
				ProviderConfigs: []mmtypes.ProviderConfig{
					{
						Name: "provider1",
					},
				},
			},
			update: mmtypes.Market{
				Ticker: mmtypes.Ticker{
					CurrencyPair:     connecttypes.NewCurrencyPair("BTC", "USD"),
					Decimals:         8,
					MinProviderCount: 1,
				},
				ProviderConfigs: []mmtypes.ProviderConfig{
					{
						Name: "provider1",
					},
				},
			},
			expected: mmtypes.Market{
				Ticker: mmtypes.Ticker{
					CurrencyPair:     connecttypes.NewCurrencyPair("BTC", "USD"),
					MinProviderCount: 1,
				},
			},
		},
		{
			name: "metadata-json changed",
			currentMarket: mmtypes.Market{
				Ticker: mmtypes.Ticker{
					CurrencyPair:     connecttypes.NewCurrencyPair("BTC", "USD"),
					Decimals:         8,
					MinProviderCount: 2,
					Metadata_JSON:    "{}",
				},
				ProviderConfigs: []mmtypes.ProviderConfig{
					{
						Name: "provider1",
					},
				},
			},
			update: mmtypes.Market{
				Ticker: mmtypes.Ticker{
					CurrencyPair:     connecttypes.NewCurrencyPair("BTC", "USD"),
					Decimals:         8,
					MinProviderCount: 2,
					Metadata_JSON:    "{\"a\": \"b\"}",
				},
				ProviderConfigs: []mmtypes.ProviderConfig{
					{
						Name: "provider1",
					},
				},
			},
			expected: mmtypes.Market{
				Ticker: mmtypes.Ticker{
					CurrencyPair:  connecttypes.NewCurrencyPair("BTC", "USD"),
					Metadata_JSON: "{\"a\": \"b\"}",
				},
			},
		},
		{
			name: "enabled changed",
			currentMarket: mmtypes.Market{
				Ticker: mmtypes.Ticker{
					CurrencyPair:     connecttypes.NewCurrencyPair("BTC", "USD"),
					Decimals:         8,
					MinProviderCount: 2,
				},
				ProviderConfigs: []mmtypes.ProviderConfig{
					{
						Name: "provider1",
					},
				},
			},
			update: mmtypes.Market{
				Ticker: mmtypes.Ticker{
					CurrencyPair:     connecttypes.NewCurrencyPair("BTC", "USD"),
					Decimals:         8,
					MinProviderCount: 1,
					Enabled:          true,
				},
				ProviderConfigs: []mmtypes.ProviderConfig{
					{
						Name: "provider1",
					},
				},
			},
			expected: mmtypes.Market{
				Ticker: mmtypes.Ticker{
					CurrencyPair:     connecttypes.NewCurrencyPair("BTC", "USD"),
					MinProviderCount: 1,
					Enabled:          true,
				},
			},
		},
		{
			name: "updated provider config",
			currentMarket: mmtypes.Market{
				Ticker: mmtypes.Ticker{
					CurrencyPair: connecttypes.NewCurrencyPair("BTC", "USD"),
				},
				ProviderConfigs: []mmtypes.ProviderConfig{
					{
						Name:           "provider1",
						OffChainTicker: "abc",
					},
				},
			},
			update: mmtypes.Market{
				Ticker: mmtypes.Ticker{
					CurrencyPair: connecttypes.NewCurrencyPair("BTC", "USD"),
				},
				ProviderConfigs: []mmtypes.ProviderConfig{
					{
						Name:           "provider1",
						OffChainTicker: "def",
					},
				},
			},
			expected: mmtypes.Market{
				Ticker: mmtypes.Ticker{
					CurrencyPair: connecttypes.NewCurrencyPair("BTC", "USD"),
				},
				ProviderConfigs: []mmtypes.ProviderConfig{
					{
						Name:           "provider1",
						OffChainTicker: "def",
					},
				},
			},
		},
		{
			name: "net new provider config",
			currentMarket: mmtypes.Market{
				Ticker: mmtypes.Ticker{
					CurrencyPair: connecttypes.NewCurrencyPair("BTC", "USD"),
				},
				ProviderConfigs: []mmtypes.ProviderConfig{
					{
						Name:           "provider1",
						OffChainTicker: "abc",
					},
				},
			},
			update: mmtypes.Market{
				Ticker: mmtypes.Ticker{
					CurrencyPair: connecttypes.NewCurrencyPair("BTC", "USD"),
				},
				ProviderConfigs: []mmtypes.ProviderConfig{
					{
						Name:           "provider2",
						OffChainTicker: "def",
					},
				},
			},
			expected: mmtypes.Market{
				Ticker: mmtypes.Ticker{
					CurrencyPair: connecttypes.NewCurrencyPair("BTC", "USD"),
				},
				ProviderConfigs: []mmtypes.ProviderConfig{
					{
						Name:           "provider2",
						OffChainTicker: "def",
					},
				},
			},
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			require.True(t, tc.expected.Equal(diffs.FilterMarketUpdates(tc.currentMarket, tc.update)))
		})
	}
}
