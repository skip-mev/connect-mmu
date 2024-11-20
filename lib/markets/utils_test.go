package markets_test

import (
	"testing"

	connecttypes "github.com/skip-mev/connect/v2/pkg/types"
	mmtypes "github.com/skip-mev/connect/v2/x/marketmap/types"
	"github.com/stretchr/testify/require"

	"github.com/skip-mev/connect-mmu/lib/markets"
)

func TestFindRemovedMarkets(t *testing.T) {
	var (
		btcUSD = mmtypes.Market{
			Ticker: mmtypes.Ticker{
				CurrencyPair: connecttypes.NewCurrencyPair("BTC", "USD"),
				Decimals:     8,
			},
		}

		ethUSD = mmtypes.Market{
			Ticker: mmtypes.Ticker{
				CurrencyPair: connecttypes.NewCurrencyPair("ETH", "USD"),
			},
		}

		solUSD = mmtypes.Market{
			Ticker: mmtypes.Ticker{
				CurrencyPair: connecttypes.NewCurrencyPair("SOL", "USD"),
			},
		}
	)

	tt := []struct {
		name              string
		actual, generated mmtypes.MarketMap
		markets           []mmtypes.Market
	}{
		{
			name: "no overlap",
			actual: mmtypes.MarketMap{
				Markets: map[string]mmtypes.Market{
					"BTC/USD": btcUSD,
				},
			},
			generated: mmtypes.MarketMap{
				Markets: map[string]mmtypes.Market{
					"ETH/USD": ethUSD,
				},
			},
			markets: []mmtypes.Market{btcUSD},
		},
		{
			name: "ignore overlap",
			actual: mmtypes.MarketMap{
				Markets: map[string]mmtypes.Market{
					"BTC/USD": btcUSD,
					"ETH/USD": ethUSD,
				},
			},
			generated: mmtypes.MarketMap{
				Markets: map[string]mmtypes.Market{
					"ETH/USD": ethUSD,
					"SOL/USD": solUSD,
				},
			},
			markets: []mmtypes.Market{btcUSD},
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			removedMarkets := markets.FindRemovedMarkets(tc.actual, tc.generated)
			require.Equal(t, tc.markets, removedMarkets)
		})
	}
}

func TestFindIntersectionAndExclusion(t *testing.T) {
	var (
		btcUSD = mmtypes.Market{
			Ticker: mmtypes.Ticker{
				CurrencyPair: connecttypes.NewCurrencyPair("BTC", "USD"),
				Decimals:     8,
			},
		}

		ethUSD = mmtypes.Market{
			Ticker: mmtypes.Ticker{
				CurrencyPair: connecttypes.NewCurrencyPair("ETH", "USD"),
			},
		}

		solUSD = mmtypes.Market{
			Ticker: mmtypes.Ticker{
				CurrencyPair: connecttypes.NewCurrencyPair("SOL", "USD"),
			},
		}
	)

	tt := []struct {
		name      string
		actual    mmtypes.MarketMap
		generated []mmtypes.Market

		added, intersection []mmtypes.Market
	}{
		{
			name: "only exclusion",
			actual: mmtypes.MarketMap{
				Markets: map[string]mmtypes.Market{},
			},
			generated:    []mmtypes.Market{btcUSD, ethUSD, solUSD},
			added:        []mmtypes.Market{btcUSD, ethUSD, solUSD},
			intersection: []mmtypes.Market{},
		},
		{
			name: "only intersection",
			actual: mmtypes.MarketMap{
				Markets: map[string]mmtypes.Market{
					"BTC/USD": btcUSD,
					"ETH/USD": ethUSD,
				},
			},
			generated: []mmtypes.Market{btcUSD, ethUSD},
			added:     []mmtypes.Market{},
			intersection: []mmtypes.Market{
				btcUSD, ethUSD,
			},
		},
		{
			name: "intersection and exclusion",
			actual: mmtypes.MarketMap{
				Markets: map[string]mmtypes.Market{
					"BTC/USD": btcUSD,
					"ETH/USD": ethUSD,
				},
			},
			generated: []mmtypes.Market{btcUSD, ethUSD, solUSD},
			added:     []mmtypes.Market{solUSD},
			intersection: []mmtypes.Market{
				btcUSD, ethUSD,
			},
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			intersection, exclusion := markets.FindIntersectionAndExclusion(tc.actual,
				tc.generated)

			// check that intersection / tc.intersection have same elements (order doesn't matter)
			require.ElementsMatch(t, tc.intersection, intersection)
			require.ElementsMatch(t, tc.added, exclusion)
		})
	}
}
