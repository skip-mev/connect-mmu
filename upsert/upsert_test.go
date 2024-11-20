package upsert

import (
	"testing"

	connecttypes "github.com/skip-mev/connect/v2/pkg/types"
	mmtypes "github.com/skip-mev/connect/v2/x/marketmap/types"
	"github.com/stretchr/testify/require"
)

func TestOrderNormalizeMarketsFirst(t *testing.T) {
	testMarketA := mmtypes.Market{
		Ticker: mmtypes.Ticker{
			CurrencyPair: connecttypes.CurrencyPair{Base: "USDT", Quote: "USD"},
			Enabled:      false,
		},
		ProviderConfigs: []mmtypes.ProviderConfig{
			{
				Name:            "test 1",
				OffChainTicker:  "test 1",
				NormalizeByPair: nil,
			},
		},
	}

	// market normalized by market A
	testMarketB := mmtypes.Market{
		Ticker: mmtypes.Ticker{
			CurrencyPair: connecttypes.CurrencyPair{Base: "ETH", Quote: "USD"},
			Enabled:      false,
		},
		ProviderConfigs: []mmtypes.ProviderConfig{
			{
				Name:            "test 1",
				OffChainTicker:  "test 1",
				NormalizeByPair: &connecttypes.CurrencyPair{Base: "USDT", Quote: "USD"},
			},
			{
				Name:            "test 2",
				OffChainTicker:  "test 1",
				NormalizeByPair: &connecttypes.CurrencyPair{Base: "USDT", Quote: "USD"},
			},
			{
				Name:            "test 3",
				OffChainTicker:  "test 1",
				NormalizeByPair: &connecttypes.CurrencyPair{Base: "USDT", Quote: "USD"},
			},
			{
				Name:            "test 4",
				OffChainTicker:  "test 1",
				NormalizeByPair: &connecttypes.CurrencyPair{Base: "USDT", Quote: "USD"},
			},
		},
	}

	tests := []struct {
		name    string
		upserts []mmtypes.Market
		want    []mmtypes.Market
		wantErr bool
	}{
		{
			name:    "empty",
			upserts: []mmtypes.Market{},
			want:    []mmtypes.Market{},
			wantErr: false,
		},
		{
			name: "one market no reorder",
			upserts: []mmtypes.Market{
				testMarketA,
			},
			want: []mmtypes.Market{
				testMarketA,
			},
			wantErr: false,
		},
		{
			name: "two markets already in order",
			upserts: []mmtypes.Market{
				testMarketA,
				testMarketB,
			},
			want: []mmtypes.Market{
				testMarketA,
				testMarketB,
			},
			wantErr: false,
		},
		{
			name: "two markets - reordered",
			upserts: []mmtypes.Market{
				testMarketB,
				testMarketA,
			},
			want: []mmtypes.Market{
				testMarketA, // USDT/USD should be first because it is used by other markets
				testMarketB,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := orderNormalizeMarketsFirst(tt.upserts)
			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestRemoveFromUpserts(t *testing.T) {
	testCases := []struct {
		name     string
		markets  []string
		remove   []string
		expected []string
	}{
		{
			name: "empty",
		},
		{
			name:     "remove one",
			markets:  []string{"FOO/BAR", "BAZ/QUX"},
			remove:   []string{"FOO/BAR"},
			expected: []string{"BAZ/QUX"},
		},
		{
			name:     "remove all",
			markets:  []string{"FOO/BAR", "BAZ/QUX"},
			remove:   []string{"FOO/BAR", "BAZ/QUX"},
			expected: []string{},
		},
		{
			name:     "remove none",
			markets:  []string{"FOO/BAR", "BAZ/QUX"},
			expected: []string{"FOO/BAR", "BAZ/QUX"},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			markets := make([]mmtypes.Market, len(tc.markets))
			for i, m := range tc.markets {
				pair, err := connecttypes.CurrencyPairFromString(m)
				require.NoError(t, err)
				markets[i] = mmtypes.Market{Ticker: mmtypes.Ticker{CurrencyPair: pair}}
			}

			updated := removeFromUpserts(markets, tc.remove)
			require.Len(t, updated, len(tc.expected))
			for i, market := range updated {
				require.Equal(t, tc.expected[i], market.Ticker.String())
			}
		})
	}
}
