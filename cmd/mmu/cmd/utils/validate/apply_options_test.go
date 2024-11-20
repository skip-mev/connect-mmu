package validate_test

import (
	"slices"
	"testing"

	"github.com/skip-mev/connect/v2/x/marketmap/types"
	"github.com/stretchr/testify/require"
	"golang.org/x/exp/maps"

	"github.com/skip-mev/connect-mmu/cmd/mmu/cmd/utils/validate"
)

func TestApplyOptionsToMarketMap(t *testing.T) {
	type simpleMarket struct {
		ticker  string
		enabled bool
	}

	tests := []struct {
		name            string
		markets         []simpleMarket
		enableAll       bool
		enableMarkets   []string
		enableOnly      []string
		expectedMarkets []string
		err             bool
	}{
		{
			name:            "normal path: all enabled",
			markets:         []simpleMarket{{"BTC/USD", true}, {"ETH/USD", false}, {"FOO/BAR", false}},
			enableAll:       true,
			enableMarkets:   nil,
			expectedMarkets: []string{"ETH/USD", "FOO/BAR", "BTC/USD"},
		},
		{
			name:            "only enable some, rest should be deleted",
			markets:         []simpleMarket{{"FOO/BAR", false}, {"ETH/USD", false}, {"ATOM/BTC", false}},
			enableMarkets:   []string{"FOO/BAR"},
			expectedMarkets: []string{"FOO/BAR"},
		},
		{
			name:            "only enable one, with another one already enabled",
			markets:         []simpleMarket{{"FOO/BAR", false}, {"ETH/USD", true}, {"ATOM/BTC", false}},
			enableMarkets:   []string{"FOO/BAR"},
			expectedMarkets: []string{"FOO/BAR", "ETH/USD"},
		},
		{
			name:            "all zeroed options should just give us back only enabled markets",
			markets:         []simpleMarket{{"ETH/USD", true}, {"ATOM/BTC", false}},
			expectedMarkets: []string{"ETH/USD"},
		},
		{
			name:            "enableOnly: should only give the markets specified",
			markets:         []simpleMarket{{"ETH/USD", false}, {"ATOM/BTC", true}, {"SOL/USD", true}},
			enableOnly:      []string{"ETH/USD", "SOL/USD"},
			expectedMarkets: []string{"ETH/USD", "SOL/USD"},
		},
		{
			name:          "cannot specify both enableOnly and enableMarkets",
			enableOnly:    []string{"FOO/BAR"},
			enableMarkets: []string{"ETH/USD"},
			err:           true,
		},
		{
			name:            "invalid ticker does nothing to marketmap, and returns error",
			markets:         []simpleMarket{{"ETH/USD", true}},
			enableMarkets:   []string{"blah.2024"},
			expectedMarkets: []string{"ETH/USD"},
			err:             true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// make the marketmap
			mm := types.MarketMap{Markets: map[string]types.Market{}}
			for _, market := range tc.markets {
				mm.Markets[market.ticker] = types.Market{Ticker: types.Ticker{Enabled: market.enabled}}
			}

			// apply the options
			err := validate.ApplyOptionsToMarketMap(mm, tc.enableAll, tc.enableOnly, tc.enableMarkets)
			if tc.err {
				require.Error(t, err)
			} else {
				require.NoError(t, err)

				// marketmap should only have enabled tickers at this point.
				for _, market := range mm.Markets {
					require.True(t, market.Ticker.Enabled)
				}

				// get list of all tickers and sort them so we can compare against expected.
				gotTickers := maps.Keys(mm.Markets)
				slices.Sort(gotTickers)
				slices.Sort(tc.expectedMarkets)

				// should be equal at this point.
				require.Equal(t, tc.expectedMarkets, gotTickers)
			}
		})
	}
}
