package types_test

import (
	"math/big"
	"testing"

	connecttypes "github.com/skip-mev/connect/v2/pkg/types"
	mmtypes "github.com/skip-mev/connect/v2/x/marketmap/types"
	"github.com/stretchr/testify/require"

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

	feedA := types.NewFeed(marketA.Ticker, marketA.ProviderConfigs[0], volumeA, volumeA, 0, liquidityInfoA, cmcInfo)
	feedC := types.NewFeed(marketA.Ticker, marketA.ProviderConfigs[0], volumeB, volumeB, 0, liquidityInfoB, cmcInfo)
	feedD := types.NewFeed(marketA.Ticker, marketA.ProviderConfigs[0], volumeA, volumeA, 0, liquidityInfo0, cmcInfo)
	feedE := types.NewFeed(marketA.Ticker, marketA.ProviderConfigs[0], volumeB, volumeB, 0, liquidityInfo0, cmcInfo)

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
			name: "quote rank is not considered in comparison",
			a: types.Feed{
				CMCInfo: mmutypes.NewCoinMarketCapInfo(0, 0, 1, 10),
				LiquidityInfo: mmutypes.LiquidityInfo{
					NegativeDepthTwo: 10,
					PositiveDepthTwo: 10,
				},
			},
			b: types.Feed{
				CMCInfo: mmutypes.NewCoinMarketCapInfo(0, 0, 1, 1),
				LiquidityInfo: mmutypes.LiquidityInfo{
					NegativeDepthTwo: 0,
					PositiveDepthTwo: 0,
				},
			},
			want:    false,
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
					DailyUsdVolume: big.NewFloat(10000),
				},
				{
					DailyUsdVolume: big.NewFloat(1000000),
				},
			},
			want: types.Feeds{
				{
					DailyUsdVolume: big.NewFloat(1000000),
				},
				{
					DailyUsdVolume: big.NewFloat(10000),
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
	tests := []struct {
		name    string
		f       types.Feeds
		want    mmtypes.MarketMap
		wantErr bool
	}{
		{
			name: "empty feeds",
			f:    types.Feeds{},
			want: mmtypes.MarketMap{
				Markets: make(map[string]mmtypes.Market),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.f.ToMarketMap()
			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			require.Equal(t, tt.want, got)
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
