package types_test

import (
	"fmt"
	"math/big"
	"slices"
	"strings"
	"testing"

	connecttypes "github.com/skip-mev/connect/v2/pkg/types"
	mmtypes "github.com/skip-mev/connect/v2/x/marketmap/types"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest"
	"gopkg.in/typ.v4/maps"

	"github.com/skip-mev/connect-mmu/generator/types"
	mmutypes "github.com/skip-mev/connect-mmu/types"
)

func TestCompareFeed(t *testing.T) {
	marketA := mmtypes.Market{
		Ticker: mmtypes.Ticker{
			CurrencyPair: connecttypes.NewCurrencyPair("BTC", "USD"),
		},
		ProviderConfigs: []mmtypes.ProviderConfig{
			{
				Name:            "test",
				OffChainTicker:  "btc-usd",
				NormalizeByPair: nil,
				Invert:          false,
				Metadata_JSON:   "",
			},
		},
	}

	volumeA := 1000.0
	volumeB := 2000.0
	liquidityA := 4000.0
	liquidityB := 93939.0

	liquidityInfoA := mmutypes.LiquidityInfo{
		NegativeDepthTwo: liquidityA,
		PositiveDepthTwo: liquidityA,
	}

	liquidityInfoB := mmutypes.LiquidityInfo{
		NegativeDepthTwo: liquidityB,
		PositiveDepthTwo: liquidityB,
	}

	liquidityInfo0 := mmutypes.LiquidityInfo{
		NegativeDepthTwo: 0,
		PositiveDepthTwo: 0,
	}

	cmcInfo := mmutypes.CoinMarketCapInfo{}

	feedA := types.NewFeed(marketA.Ticker, marketA.ProviderConfigs[0], volumeA, 0, liquidityInfoA, cmcInfo)
	feedC := types.NewFeed(marketA.Ticker, marketA.ProviderConfigs[0], volumeB, 0, liquidityInfoB, cmcInfo)
	feedD := types.NewFeed(marketA.Ticker, marketA.ProviderConfigs[0], volumeA, 0, liquidityInfo0, cmcInfo)
	feedE := types.NewFeed(marketA.Ticker, marketA.ProviderConfigs[0], volumeB, 0, liquidityInfo0, cmcInfo)

	tests := []struct {
		name        string
		feedA       types.Feed
		feedB       types.Feed
		expected    bool
		expectedErr bool
	}{
		{
			name:        "equal - choose A",
			feedA:       feedA,
			feedB:       feedA,
			expected:    false,
			expectedErr: false,
		},
		{
			name:     "different liquidity - choose C",
			feedA:    feedA,
			feedB:    feedC,
			expected: true,
		},
		{
			name:     "different liquidity - choose A",
			feedA:    feedC,
			feedB:    feedA,
			expected: false,
		},
		{
			name:     "different volume - no liquidity - choose E",
			feedA:    feedD,
			feedB:    feedE,
			expected: true,
		},
		{
			name:     "different volume - no liquidity  - choose D",
			feedA:    feedE,
			feedB:    feedD,
			expected: false,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := types.Compare(tc.feedA, tc.feedB)

			require.Equal(t, tc.expected, got)
		})
	}
}

func TestRemovalReasons(t *testing.T) {
	t.Run("test adding removal reason", func(t *testing.T) {
		reasons := types.NewRemovalReasons()
		btc := mmtypes.Ticker{
			CurrencyPair: connecttypes.NewCurrencyPair("BTC", "USD"),
		}

		reasons.AddRemovalReasonFromFeed(types.Feed{Ticker: btc}, "test", "test")

		reasons.AddRemovalReasonFromFeed(types.Feed{Ticker: btc}, "test", "test")

		// check that btc has 2 reasons
		require.Equal(t, 2, len(reasons[btc.String()]))
	})

	t.Run("test adding removal reason with different ticker", func(t *testing.T) {
		reasons := types.NewRemovalReasons()
		btc := mmtypes.Ticker{
			CurrencyPair: connecttypes.NewCurrencyPair("BTC", "USD"),
		}

		eth := mmtypes.Ticker{
			CurrencyPair: connecttypes.NewCurrencyPair("ETH", "USD"),
		}

		reasons.AddRemovalReasonFromFeed(types.Feed{Ticker: btc}, "test", "test")
		reasons.AddRemovalReasonFromFeed(types.Feed{Ticker: eth}, "test", "test")

		// check that btc has 1 reason
		require.Equal(t, 1, len(reasons[btc.String()]))
		// check that eth has 1 reason
		require.Equal(t, 1, len(reasons[eth.String()]))
	})

	t.Run("test merging removal reasons", func(t *testing.T) {
		reasons := types.NewRemovalReasons()
		btc := mmtypes.Ticker{
			CurrencyPair: connecttypes.NewCurrencyPair("BTC", "USD"),
		}

		eth := mmtypes.Ticker{
			CurrencyPair: connecttypes.NewCurrencyPair("ETH", "USD"),
		}

		reasons.AddRemovalReasonFromFeed(types.Feed{Ticker: btc}, "test", "test")

		// check that btc has 1 reason
		require.Equal(t, 1, len(reasons[btc.String()]))

		reasons2 := types.NewRemovalReasons()
		reasons.AddRemovalReasonFromFeed(types.Feed{Ticker: eth}, "test", "test")

		reasons.Merge(reasons2)

		// check that btc has 1 reason
		require.Equal(t, 1, len(reasons[btc.String()]))
		// check that eth has 1 reason
		require.Equal(t, 1, len(reasons[eth.String()]))
	})
}

func TestCompare(t *testing.T) {
	tests := []struct {
		name    string
		a       types.Feed
		b       types.Feed
		want    bool
		wantErr bool
	}{
		{
			name: "feed with lowest rank always takes precedence",
			a: types.Feed{
				CMCInfo: mmutypes.NewCoinMarketCapInfo(0, 0, 1, 1),
			},
			b: types.Feed{
				CMCInfo: mmutypes.NewCoinMarketCapInfo(0, 0, 10, 10),
			},
			want:    false,
			wantErr: false,
		},
		{
			name: "feed with lowest rank always takes precedence",
			a: types.Feed{
				CMCInfo: mmutypes.NewCoinMarketCapInfo(0, 0, 10, 10),
			},
			b: types.Feed{
				CMCInfo: mmutypes.NewCoinMarketCapInfo(0, 0, 1, 1),
			},
			want:    true,
			wantErr: false,
		},
		{
			name: "mismatched currency still returns",
			a: types.Feed{
				Ticker: mmtypes.Ticker{CurrencyPair: connecttypes.CurrencyPair{
					Base:  "TEST",
					Quote: "INVALID",
				}},
				CMCInfo: mmutypes.NewCoinMarketCapInfo(0, 0, 10, 10),
			},
			b: types.Feed{
				Ticker: mmtypes.Ticker{CurrencyPair: connecttypes.CurrencyPair{
					Base:  "INVALID",
					Quote: "TEST",
				}},
				CMCInfo: mmutypes.NewCoinMarketCapInfo(0, 0, 1, 1),
			},
			want:    true,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := types.Compare(tt.a, tt.b)

			require.Equal(t, tt.want, got)
		})
	}
}

func TestFeeds_Sort(t *testing.T) {
	tests := []struct {
		name string
		f    types.Feeds
		want types.Feeds
	}{
		{
			name: "empty feeds",
			f:    types.Feeds{},
			want: types.Feeds{},
		},
		{
			name: "sort based on CMC Rank",
			f: types.Feeds{
				{
					CMCInfo: mmutypes.NewCoinMarketCapInfo(0, 0, 10, 10),
				},
				{
					CMCInfo: mmutypes.NewCoinMarketCapInfo(0, 0, 1, 1),
				},
			},
			want: types.Feeds{
				{
					CMCInfo: mmutypes.NewCoinMarketCapInfo(0, 0, 1, 1),
				},
				{
					CMCInfo: mmutypes.NewCoinMarketCapInfo(0, 0, 10, 10),
				},
			},
		},
		{
			name: "sort based on volume",
			f: types.Feeds{
				{
					DailyQuoteVolume: big.NewFloat(10000),
				},
				{
					DailyQuoteVolume: big.NewFloat(1000000),
				},
			},
			want: types.Feeds{
				{
					DailyQuoteVolume: big.NewFloat(1000000),
				},
				{
					DailyQuoteVolume: big.NewFloat(10000),
				},
			},
		},
		{
			name: "sort based on liquidity",
			f: types.Feeds{
				{
					LiquidityInfo: mmutypes.LiquidityInfo{
						NegativeDepthTwo: 10000,
						PositiveDepthTwo: 10000,
					},
				},
				{
					LiquidityInfo: mmutypes.LiquidityInfo{
						NegativeDepthTwo: 1000000,
						PositiveDepthTwo: 1000000,
					},
				},
			},
			want: types.Feeds{
				{
					LiquidityInfo: mmutypes.LiquidityInfo{
						NegativeDepthTwo: 1000000,
						PositiveDepthTwo: 1000000,
					},
				},
				{
					LiquidityInfo: mmutypes.LiquidityInfo{
						NegativeDepthTwo: 10000,
						PositiveDepthTwo: 10000,
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.f.Sort()

			require.Equal(t, tt.want, tt.f)
		})
	}
}

func TestFeeds_ToMarketMap(t *testing.T) {
	logger := zaptest.NewLogger(t)
	type simpleMarket struct {
		ticker   string
		provider string
		id       int
	}
	tests := []struct {
		name                    string
		marketsByCMCID          []simpleMarket
		expectedMarketProviders map[string][]string
	}{
		{
			name: "uniswap should be combined",
			marketsByCMCID: []simpleMarket{
				{"FOO,UNISWAP,0XFOOBAR/USD", "uniswap", 40},
				{"FOO/USD", "binance", 40},
				{"FOO/USD", "coinbase", 40},

				{"BAZ/USD", "binance", 20},
				{"BAZ,UNISWAP,0XBAZBAR/USD", "uniswap", 20},
				{"BAZ/USD", "kraken", 20},

				{"BOOK/USD", "foobar", 30},
				{"BOOK/USD", "foobaz", -1},
				{"BOOK/USD", "bazbar", 30},

				{"DOG/USD", "binance", -1},
				{"DOG/USD", "kraken", -1},
			},
			expectedMarketProviders: map[string][]string{
				"FOO/USD":  {"binance", "uniswap", "coinbase"},
				"BAZ/USD":  {"binance", "uniswap", "kraken"},
				"BOOK/USD": {"foobar", "foobaz", "bazbar"},
				"DOG/USD":  {"binance", "kraken"},
			},
		},
		{
			name:                    "empty should work",
			marketsByCMCID:          []simpleMarket{},
			expectedMarketProviders: map[string][]string{},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			feeds := make(types.Feeds, 0)
			for _, market := range tc.marketsByCMCID {
				ticker := strings.Split(market.ticker, "/")
				metadata := ""
				if market.id > 0 {
					metadata = fmt.Sprintf("{\"reference_price\":1340000000,\"liquidity\":169321,\"aggregate_ids\":[{\"venue\":\"coinmarketcap\",\"ID\":\"%d\"}]}", market.id)
				}
				feeds = append(feeds, types.Feed{
					Ticker: mmtypes.Ticker{
						CurrencyPair:  connecttypes.CurrencyPair{Base: ticker[0], Quote: ticker[1]},
						Metadata_JSON: metadata,
					},
					ProviderConfig: mmtypes.ProviderConfig{Name: market.provider},
					ReferencePrice: new(big.Float).SetFloat64(40), // doesn't matter.
				})

			}
			logger.Debug("converting feeds", zap.Int("num_feeds", len(feeds)))
			mm, err := feeds.ToMarketMap()
			require.NoError(t, err)
			logger.Debug("resulting markets", zap.Any("markets", maps.Keys(mm.Markets)))
			for ticker, providers := range tc.expectedMarketProviders {
				market, ok := mm.Markets[ticker]
				require.True(t, ok, "expected market to exist: %s", ticker)
				logger.Debug("market providers", zap.String("market", market.Ticker.String()), zap.Int("providers", len(providers)))
				require.Equal(t, len(providers), len(market.ProviderConfigs))
				for _, provider := range providers {
					require.True(t, slices.ContainsFunc(market.ProviderConfigs, func(config mmtypes.ProviderConfig) bool {
						return config.Name == provider
					}))
				}
			}
		})
	}
}

func TestFeeds_ToProviderFeeds(t *testing.T) {
	tests := []struct {
		name string
		f    types.Feeds
		want types.ProviderFeeds
	}{
		{
			name: "empty feeds",
			f:    types.Feeds{},
			want: make(types.ProviderFeeds),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.f.ToProviderFeeds()
			require.Equal(t, tt.want, got)
		})
	}
}

func TestProviderFeeds_ToFeeds(t *testing.T) {
	tests := []struct {
		name string
		pf   types.ProviderFeeds
		want types.Feeds
	}{
		{
			name: "empty feeds",
			pf:   make(types.ProviderFeeds),
			want: types.Feeds{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.pf.ToFeeds()
			require.Equal(t, tt.want, got)
		})
	}
}

func TestRemovalReasons_Merge(t *testing.T) {
	tests := []struct {
		name  string
		r     types.RemovalReasons
		other types.RemovalReasons
		want  types.RemovalReasons
	}{
		{
			name:  "empty removal reasons",
			r:     types.RemovalReasons{},
			other: types.RemovalReasons{},
			want:  types.RemovalReasons{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.r.Merge(tt.other)
			require.Equal(t, tt.r, tt.want)
		})
	}
}
