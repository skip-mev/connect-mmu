package transformer_test

import (
	"context"
	"math/big"
	"testing"

	connecttypes "github.com/skip-mev/connect/v2/pkg/types"
	mmtypes "github.com/skip-mev/connect/v2/x/marketmap/types"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"

	"github.com/skip-mev/connect-mmu/config"
	"github.com/skip-mev/connect-mmu/generator/transformer"
	"github.com/skip-mev/connect-mmu/generator/types"
	mmutypes "github.com/skip-mev/connect-mmu/types"
)

var usdtusdFeed = types.Feed{
	Ticker: mmtypes.Ticker{
		CurrencyPair: connecttypes.CurrencyPair{
			Base:  "USDT",
			Quote: "USD",
		},
		Decimals:         8,
		MinProviderCount: 1,
		Enabled:          false,
		Metadata_JSON:    "",
	},
	ProviderConfig: mmtypes.ProviderConfig{
		Name:            krakenProvider,
		OffChainTicker:  "usdt-usd",
		NormalizeByPair: nil,
		Invert:          false,
		Metadata_JSON:   "",
	},
	ReferencePrice:   big.NewFloat(1.1),
	DailyQuoteVolume: big.NewFloat(400),
	DailyUsdVolume:   big.NewFloat(400),
	CMCInfo: mmutypes.CoinMarketCapInfo{
		BaseID:    825,
		QuoteID:   2781,
		BaseRank:  3,
		QuoteRank: 0,
	},
}

func TestDefaultTransformer_TransformFeeds(t *testing.T) {
	cfg := config.GenerateConfig{
		Providers: map[string]config.ProviderConfig{
			krakenProvider: {},
		},
		MinCexProviderCount: 1,
		MinDexProviderCount: 1,
		Quotes: map[string]config.QuoteConfig{
			"USD": {
				MinProviderVolume: 100,
				NormalizeByPair:   "",
			},
			"USDT": {
				MinProviderVolume: 100,
				NormalizeByPair:   "USDT/USD",
			},
			"ETH": {
				MinProviderVolume: 100,
				NormalizeByPair:   "",
			},
		},
	}

	tests := []struct {
		name    string
		cfg     config.GenerateConfig
		feeds   types.Feeds
		want    types.Feeds
		wantErr bool
	}{
		{
			name: "no transforms for one valid feed",
			cfg:  cfg,
			feeds: types.Feeds{
				{
					Ticker: mmtypes.Ticker{
						CurrencyPair: connecttypes.CurrencyPair{
							Base:  "BTC",
							Quote: "USD",
						},
						Decimals:         8,
						MinProviderCount: 1,
						Enabled:          false,
						Metadata_JSON:    "",
					},
					ProviderConfig: mmtypes.ProviderConfig{
						Name:            krakenProvider,
						OffChainTicker:  "btc-usd",
						NormalizeByPair: nil,
						Invert:          false,
						Metadata_JSON:   "",
					},
					DailyQuoteVolume: big.NewFloat(200),
					DailyUsdVolume:   big.NewFloat(200),
					ReferencePrice:   big.NewFloat(1.1),
					CMCInfo:          cmcInfoA,
				},
			},
			want: types.Feeds{
				{
					Ticker: mmtypes.Ticker{
						CurrencyPair: connecttypes.CurrencyPair{
							Base:  "BTC",
							Quote: "USD",
						},
						Decimals:         8,
						MinProviderCount: 1,
						Enabled:          false,
						Metadata_JSON:    "",
					},
					ProviderConfig: mmtypes.ProviderConfig{
						Name:            krakenProvider,
						OffChainTicker:  "btc-usd",
						NormalizeByPair: nil,
						Invert:          false,
						Metadata_JSON:   "",
					},
					DailyQuoteVolume: big.NewFloat(200),
					DailyUsdVolume:   big.NewFloat(200),
					ReferencePrice:   big.NewFloat(1.1),
					CMCInfo:          cmcInfoA,
				},
			},
			wantErr: false,
		},
		{
			name: "invert a feed",
			cfg:  cfg,
			feeds: types.Feeds{
				{
					Ticker: mmtypes.Ticker{
						CurrencyPair: connecttypes.CurrencyPair{
							Base:  "USD",
							Quote: "BTC",
						},
						Decimals:         8,
						MinProviderCount: 1,
						Enabled:          false,
						Metadata_JSON:    "",
					},
					ProviderConfig: mmtypes.ProviderConfig{
						Name:            krakenProvider,
						OffChainTicker:  "usd-btc",
						NormalizeByPair: nil,
						Invert:          false,
						Metadata_JSON:   "",
					},
					DailyQuoteVolume: big.NewFloat(200),
					DailyUsdVolume:   big.NewFloat(200),
					ReferencePrice:   big.NewFloat(100),
					CMCInfo:          cmcInfoA,
				},
			},
			want: types.Feeds{
				{
					Ticker: mmtypes.Ticker{
						CurrencyPair: connecttypes.CurrencyPair{
							Base:  "BTC",
							Quote: "USD",
						},
						Decimals:         8,
						MinProviderCount: 1,
						Enabled:          false,
						Metadata_JSON:    "",
					},
					ProviderConfig: mmtypes.ProviderConfig{
						Name:            krakenProvider,
						OffChainTicker:  "usd-btc",
						NormalizeByPair: nil,
						Invert:          true,
						Metadata_JSON:   "",
					},
					DailyQuoteVolume: big.NewFloat(200),
					DailyUsdVolume:   big.NewFloat(200),
					ReferencePrice:   big.NewFloat(0.01),
					CMCInfo:          cmcInfoAInverted,
				},
			},
			wantErr: false,
		},
		{
			name: "normalize a feed",
			cfg:  cfg,
			feeds: types.Feeds{
				{
					Ticker: mmtypes.Ticker{
						CurrencyPair: connecttypes.CurrencyPair{
							Base:  "BTC",
							Quote: "USDT",
						},
						Decimals:         8,
						MinProviderCount: 1,
						Enabled:          false,
						Metadata_JSON:    "",
					},
					ProviderConfig: mmtypes.ProviderConfig{
						Name:            krakenProvider,
						OffChainTicker:  "btc-usdt",
						NormalizeByPair: nil,
						Invert:          false,
						Metadata_JSON:   "",
					},
					ReferencePrice:   big.NewFloat(100),
					DailyQuoteVolume: big.NewFloat(200),
					DailyUsdVolume:   big.NewFloat(200),
					CMCInfo:          cmcInfoA,
				},
				usdtusdFeed,
			},
			want: types.Feeds{
				usdtusdFeed,
				{
					Ticker: mmtypes.Ticker{
						CurrencyPair: connecttypes.CurrencyPair{
							Base:  "BTC",
							Quote: "USD",
						},
						Decimals:         8,
						MinProviderCount: 1,
						Enabled:          false,
						Metadata_JSON:    "",
					},
					ProviderConfig: mmtypes.ProviderConfig{
						Name:           krakenProvider,
						OffChainTicker: "btc-usdt",
						NormalizeByPair: &connecttypes.CurrencyPair{
							Base:  "USDT",
							Quote: "USD",
						},
						Invert:        false,
						Metadata_JSON: "",
					},
					ReferencePrice:   big.NewFloat(110.00000000000001),
					DailyQuoteVolume: big.NewFloat(200),
					DailyUsdVolume:   big.NewFloat(200),
					CMCInfo:          cmcInfoA,
				},
			},
			wantErr: false,
		},
		{
			name: "invert and normalize a feed",
			cfg:  cfg,
			feeds: types.Feeds{
				{
					Ticker: mmtypes.Ticker{
						CurrencyPair: connecttypes.CurrencyPair{
							Base:  "USDT",
							Quote: "BTC",
						},
						Decimals:         8,
						MinProviderCount: 1,
						Enabled:          false,
						Metadata_JSON:    "",
					},
					ProviderConfig: mmtypes.ProviderConfig{
						Name:            krakenProvider,
						OffChainTicker:  "usdt-btc",
						NormalizeByPair: nil,
						Invert:          false,
						Metadata_JSON:   "",
					},
					DailyQuoteVolume: big.NewFloat(200),
					DailyUsdVolume:   big.NewFloat(200),
					ReferencePrice:   big.NewFloat(1),
					CMCInfo:          cmcInfoA,
				},
				usdtusdFeed,
			},
			want: types.Feeds{
				usdtusdFeed,
				{
					Ticker: mmtypes.Ticker{
						CurrencyPair: connecttypes.CurrencyPair{
							Base:  "BTC",
							Quote: "USD",
						},
						Decimals:         8,
						MinProviderCount: 1,
						Enabled:          false,
						Metadata_JSON:    "",
					},
					ProviderConfig: mmtypes.ProviderConfig{
						Name:           krakenProvider,
						OffChainTicker: "usdt-btc",
						NormalizeByPair: &connecttypes.CurrencyPair{
							Base:  "USDT",
							Quote: "USD",
						},
						Invert:        true,
						Metadata_JSON: "",
					},
					DailyQuoteVolume: big.NewFloat(200),
					DailyUsdVolume:   big.NewFloat(200),
					ReferencePrice:   big.NewFloat(1.1),
					CMCInfo:          cmcInfoAInverted,
				},
			},
			wantErr: false,
		},
		{
			name: "drop a feed because it does not have enough volume",
			cfg:  cfg,
			feeds: types.Feeds{
				{
					Ticker: mmtypes.Ticker{
						CurrencyPair: connecttypes.CurrencyPair{
							Base:  "USDT",
							Quote: "BTC",
						},
						Decimals:         8,
						MinProviderCount: 1,
						Enabled:          false,
						Metadata_JSON:    "",
					},
					ProviderConfig: mmtypes.ProviderConfig{
						Name:            krakenProvider,
						OffChainTicker:  "usdt-btc",
						NormalizeByPair: nil,
						Invert:          false,
						Metadata_JSON:   "",
					},
					DailyQuoteVolume: big.NewFloat(1),
					DailyUsdVolume:   big.NewFloat(100000),
					ReferencePrice:   big.NewFloat(100),
					CMCInfo:          cmcInfoA,
				},
				usdtusdFeed,
			},
			want: types.Feeds{
				usdtusdFeed,
			},
			wantErr: false,
		},
		{
			name: "remove because quote is not in config",
			cfg:  cfg,
			feeds: types.Feeds{
				{
					Ticker: mmtypes.Ticker{
						CurrencyPair: connecttypes.CurrencyPair{
							Base:  "BTC",
							Quote: "INCORRECT",
						},
						Decimals:         8,
						MinProviderCount: 1,
						Enabled:          false,
						Metadata_JSON:    "",
					},
					ProviderConfig: mmtypes.ProviderConfig{
						Name:            krakenProvider,
						OffChainTicker:  "btc-incorrect",
						NormalizeByPair: nil,
						Invert:          false,
						Metadata_JSON:   "",
					},
					ReferencePrice:   big.NewFloat(0.011000000000000001),
					DailyQuoteVolume: big.NewFloat(1),
					DailyUsdVolume:   big.NewFloat(100000),
					CMCInfo:          cmcInfoA,
				},
			},
			want:    types.Feeds{},
			wantErr: false,
		},
		{
			name: "resolve conflicts between two feeds that will be transformed",
			cfg:  cfg,
			feeds: types.Feeds{
				{
					Ticker: mmtypes.Ticker{
						CurrencyPair: connecttypes.CurrencyPair{
							Base:  "USDT",
							Quote: "BTC",
						},
						Decimals:         8,
						MinProviderCount: 1,
						Enabled:          false,
						Metadata_JSON:    "",
					},
					ProviderConfig: mmtypes.ProviderConfig{
						Name:            krakenProvider,
						OffChainTicker:  "usdt-btc",
						NormalizeByPair: nil,
						Invert:          false,
						Metadata_JSON:   "",
					},
					ReferencePrice:   big.NewFloat(100),
					DailyQuoteVolume: big.NewFloat(2200),
					DailyUsdVolume:   big.NewFloat(220000000),
					CMCInfo:          cmcInfoAInverted,
				},
				{
					Ticker: mmtypes.Ticker{
						CurrencyPair: connecttypes.CurrencyPair{
							Base:  "BTC",
							Quote: "USD",
						},
						Decimals:         8,
						MinProviderCount: 1,
						Enabled:          false,
						Metadata_JSON:    "",
					},
					ProviderConfig: mmtypes.ProviderConfig{
						Name:            krakenProvider,
						OffChainTicker:  "btc-usd",
						NormalizeByPair: nil,
						Invert:          false,
						Metadata_JSON:   "",
					},
					ReferencePrice:   big.NewFloat(100),
					DailyQuoteVolume: big.NewFloat(0),
					DailyUsdVolume:   big.NewFloat(0),
					CMCInfo:          cmcInfoA,
				},
				usdtusdFeed,
			},
			want: types.Feeds{
				usdtusdFeed,
				{
					Ticker: mmtypes.Ticker{
						CurrencyPair: connecttypes.CurrencyPair{
							Base:  "BTC",
							Quote: "USD",
						},
						Decimals:         8,
						MinProviderCount: 1,
						Enabled:          false,
						Metadata_JSON:    "",
					},
					ProviderConfig: mmtypes.ProviderConfig{
						Name:           krakenProvider,
						OffChainTicker: "usdt-btc",
						NormalizeByPair: &connecttypes.CurrencyPair{
							Base:  "USDT",
							Quote: "USD",
						},
						Invert:        true,
						Metadata_JSON: "",
					},
					ReferencePrice:   big.NewFloat(0.011000000000000001),
					DailyQuoteVolume: big.NewFloat(2200),
					DailyUsdVolume:   big.NewFloat(220000000),
					CMCInfo:          cmcInfoA,
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := transformer.New(zaptest.NewLogger(t))
			got, _, err := d.TransformFeeds(context.Background(), tt.cfg, tt.feeds)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.Equal(t, len(tt.want), len(got))
			require.True(t, tt.want.Equal(got))
		})
	}
}

func TestDefaultTransformer_TransformMarketMap(t *testing.T) {
	cfg := config.GenerateConfig{
		Providers: map[string]config.ProviderConfig{
			krakenProvider:  {},
			binanceProvider: {},
			bybitProvider:   {},
		},
		MinCexProviderCount: 1,
		MinDexProviderCount: 1,
		Quotes: map[string]config.QuoteConfig{
			"USD": {
				MinProviderVolume: 100,
				NormalizeByPair:   "",
			},
			"USDT": {
				MinProviderVolume: 100,
				NormalizeByPair:   "USDT/USD",
			},
		},
		ExcludeCurrencyPairs: map[string]struct{}{
			"BTC/ETH": {},
		},
	}

	cfgWithOverride := cfg
	cfgWithOverride.MarketMapOverride = mmtypes.MarketMap{
		Markets: map[string]mmtypes.Market{
			"TEST/USD": {
				Ticker: mmtypes.Ticker{
					CurrencyPair: connecttypes.CurrencyPair{
						Base:  "TEST",
						Quote: "USD",
					},
					Decimals:         36,
					MinProviderCount: 1,
					Enabled:          true,
					Metadata_JSON:    "",
				},
				ProviderConfigs: []mmtypes.ProviderConfig{
					{
						Name:            binanceProvider,
						OffChainTicker:  "testUSD",
						NormalizeByPair: nil,
						Invert:          false,
						Metadata_JSON:   "",
					},
				},
			},
		},
	}

	tests := []struct {
		name      string
		cfg       config.GenerateConfig
		marketMap mmtypes.MarketMap
		want      mmtypes.MarketMap
		wantErr   bool
	}{
		{
			name:      "error for nil markets",
			cfg:       cfg,
			marketMap: mmtypes.MarketMap{},
			want:      mmtypes.MarketMap{},
			wantErr:   true,
		},
		{
			name: "do nothing for no markets",
			cfg:  cfg,
			marketMap: mmtypes.MarketMap{
				Markets: make(map[string]mmtypes.Market),
			},
			want: mmtypes.MarketMap{
				Markets: make(map[string]mmtypes.Market),
			},
			wantErr: false,
		},
		{
			name: "do nothing for a valid market",
			cfg:  cfg,
			marketMap: mmtypes.MarketMap{
				Markets: map[string]mmtypes.Market{
					"BTC/USD": {
						Ticker: mmtypes.Ticker{
							CurrencyPair: connecttypes.CurrencyPair{
								Base:  "BTC",
								Quote: "USD",
							},
							Decimals:         8,
							MinProviderCount: 3,
							Enabled:          false,
							Metadata_JSON:    "",
						},
						ProviderConfigs: []mmtypes.ProviderConfig{
							{
								Name:            krakenProvider,
								OffChainTicker:  "btc-usd",
								NormalizeByPair: nil,
								Invert:          false,
								Metadata_JSON:   "",
							},
							{
								Name:            binanceProvider,
								OffChainTicker:  "btc-usd",
								NormalizeByPair: nil,
								Invert:          false,
								Metadata_JSON:   "",
							},
							{
								Name:            bybitProvider,
								OffChainTicker:  "btc-usd",
								NormalizeByPair: nil,
								Invert:          false,
								Metadata_JSON:   "",
							},
						},
					},
				},
			},
			want: mmtypes.MarketMap{
				Markets: map[string]mmtypes.Market{
					"BTC/USD": {
						Ticker: mmtypes.Ticker{
							CurrencyPair: connecttypes.CurrencyPair{
								Base:  "BTC",
								Quote: "USD",
							},
							Decimals:         8,
							MinProviderCount: 3,
							Enabled:          false,
							Metadata_JSON:    "",
						},
						ProviderConfigs: []mmtypes.ProviderConfig{
							{
								Name:            krakenProvider,
								OffChainTicker:  "btc-usd",
								NormalizeByPair: nil,
								Invert:          false,
								Metadata_JSON:   "",
							},
							{
								Name:            binanceProvider,
								OffChainTicker:  "btc-usd",
								NormalizeByPair: nil,
								Invert:          false,
								Metadata_JSON:   "",
							},
							{
								Name:            bybitProvider,
								OffChainTicker:  "btc-usd",
								NormalizeByPair: nil,
								Invert:          false,
								Metadata_JSON:   "",
							},
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "prune an excluded market",
			cfg:  cfg,
			marketMap: mmtypes.MarketMap{
				Markets: map[string]mmtypes.Market{
					"BTC/ETH": {
						Ticker: mmtypes.Ticker{
							CurrencyPair: connecttypes.CurrencyPair{
								Base:  "BTC",
								Quote: "ETH",
							},
							Decimals:         8,
							MinProviderCount: 3,
							Enabled:          false,
							Metadata_JSON:    "",
						},
						ProviderConfigs: []mmtypes.ProviderConfig{
							{
								Name:            krakenProvider,
								OffChainTicker:  "btc-eth",
								NormalizeByPair: nil,
								Invert:          false,
								Metadata_JSON:   "",
							},
						},
					},
				},
			},
			want: mmtypes.MarketMap{
				Markets: map[string]mmtypes.Market{},
			},
			wantErr: false,
		},
		{
			name: "remove insufficiently provided markets",
			cfg:  cfg,
			marketMap: mmtypes.MarketMap{
				Markets: map[string]mmtypes.Market{
					"BTC/USD": {
						Ticker: mmtypes.Ticker{
							CurrencyPair: connecttypes.CurrencyPair{
								Base:  "BTC",
								Quote: "USD",
							},
							Decimals:         8,
							MinProviderCount: 3,
							Enabled:          false,
							Metadata_JSON:    "",
						},
						ProviderConfigs: []mmtypes.ProviderConfig{
							{
								Name:            krakenProvider,
								OffChainTicker:  "btc-usd",
								NormalizeByPair: nil,
								Invert:          false,
								Metadata_JSON:   "",
							},
						},
					},
				},
			},
			want: mmtypes.MarketMap{
				Markets: map[string]mmtypes.Market{},
			},
			wantErr: false,
		},
		{
			name: "override existing market",
			cfg:  cfgWithOverride,
			marketMap: mmtypes.MarketMap{
				Markets: map[string]mmtypes.Market{
					"TEST/USD": {
						Ticker: mmtypes.Ticker{
							CurrencyPair: connecttypes.CurrencyPair{
								Base:  "TEST",
								Quote: "USD",
							},
							Decimals:         8,
							MinProviderCount: 3,
							Enabled:          false,
							Metadata_JSON:    "",
						},
						ProviderConfigs: []mmtypes.ProviderConfig{
							{
								Name:            krakenProvider,
								OffChainTicker:  "test-usd",
								NormalizeByPair: nil,
								Invert:          false,
								Metadata_JSON:   "",
							},
						},
					},
				},
			},
			want: mmtypes.MarketMap{
				Markets: map[string]mmtypes.Market{
					"TEST/USD": {
						Ticker: mmtypes.Ticker{
							CurrencyPair: connecttypes.CurrencyPair{
								Base:  "TEST",
								Quote: "USD",
							},
							Decimals:         36,
							MinProviderCount: 1,
							Enabled:          true,
							Metadata_JSON:    "",
						},
						ProviderConfigs: []mmtypes.ProviderConfig{
							{
								Name:            binanceProvider,
								OffChainTicker:  "testUSD",
								NormalizeByPair: nil,
								Invert:          false,
								Metadata_JSON:   "",
							},
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "override non-existing market",
			cfg:  cfgWithOverride,
			marketMap: mmtypes.MarketMap{
				Markets: make(map[string]mmtypes.Market),
			},
			want: mmtypes.MarketMap{
				Markets: map[string]mmtypes.Market{
					"TEST/USD": {
						Ticker: mmtypes.Ticker{
							CurrencyPair: connecttypes.CurrencyPair{
								Base:  "TEST",
								Quote: "USD",
							},
							Decimals:         36,
							MinProviderCount: 1,
							Enabled:          true,
							Metadata_JSON:    "",
						},
						ProviderConfigs: []mmtypes.ProviderConfig{
							{
								Name:            binanceProvider,
								OffChainTicker:  "testUSD",
								NormalizeByPair: nil,
								Invert:          false,
								Metadata_JSON:   "",
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
			d := transformer.New(zaptest.NewLogger(t))
			got, _, err := d.TransformMarketMap(context.Background(), tt.cfg, tt.marketMap)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestPruneByProviderLiquidity(t *testing.T) {
	tests := []struct {
		name            string
		feeds           types.Feeds
		config          config.GenerateConfig
		expectedFeeds   types.Feeds
		expectedRemoved int
	}{
		{
			name: "provider ignores liquidity filter",
			feeds: types.Feeds{
				{
					Ticker: btcusd,
					ProviderConfig: mmtypes.ProviderConfig{
						Name: krakenProvider,
					},
					LiquidityInfo: mmutypes.LiquidityInfo{
						NegativeDepthTwo: 100,
						PositiveDepthTwo: 100,
					},
				},
			},
			config: config.GenerateConfig{
				Providers: map[string]config.ProviderConfig{
					krakenProvider: {
						IgnoreLiquidity:      true,
						MinProviderLiquidity: 1000,
					},
				},
			},
			expectedFeeds: types.Feeds{
				{
					Ticker: btcusd,
					ProviderConfig: mmtypes.ProviderConfig{
						Name: krakenProvider,
					},
					LiquidityInfo: mmutypes.LiquidityInfo{
						NegativeDepthTwo: 100,
						PositiveDepthTwo: 100,
					},
				},
			},
			expectedRemoved: 0,
		},
		{
			name: "insufficient liquidity",
			feeds: types.Feeds{
				{
					Ticker: btcusd,
					ProviderConfig: mmtypes.ProviderConfig{
						Name: krakenProvider,
					},
					LiquidityInfo: mmutypes.LiquidityInfo{
						NegativeDepthTwo: 100,
						PositiveDepthTwo: 100,
					},
				},
			},
			config: config.GenerateConfig{
				Providers: map[string]config.ProviderConfig{
					krakenProvider: {
						MinProviderLiquidity: 1000,
					},
				},
			},
			expectedFeeds:   types.Feeds{},
			expectedRemoved: 1,
		},
		{
			name: "sufficient liquidity",
			feeds: types.Feeds{
				{
					Ticker: btcusd,
					ProviderConfig: mmtypes.ProviderConfig{
						Name: krakenProvider,
					},
					LiquidityInfo: mmutypes.LiquidityInfo{
						NegativeDepthTwo: 2000,
						PositiveDepthTwo: 2000,
					},
				},
			},
			config: config.GenerateConfig{
				Providers: map[string]config.ProviderConfig{
					krakenProvider: {
						MinProviderLiquidity: 1000,
					},
				},
			},
			expectedFeeds: types.Feeds{
				{
					Ticker: btcusd,
					ProviderConfig: mmtypes.ProviderConfig{
						Name: krakenProvider,
					},
					LiquidityInfo: mmutypes.LiquidityInfo{
						NegativeDepthTwo: 2000,
						PositiveDepthTwo: 2000,
					},
				},
			},
			expectedRemoved: 0,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			logger := zaptest.NewLogger(t)
			transform := transformer.PruneByProviderLiquidity()
			feeds, removals, err := transform(context.Background(), logger, tc.config, tc.feeds)
			require.NoError(t, err)
			require.Equal(t, tc.expectedFeeds, feeds)
			require.Equal(t, tc.expectedRemoved, len(removals))
		})
	}
}

func TestPruneByProviderUsdVolume(t *testing.T) {
	tests := []struct {
		name            string
		feeds           types.Feeds
		config          config.GenerateConfig
		expectedFeeds   types.Feeds
		expectedRemoved int
	}{
		{
			name: "provider ignores volume filter",
			feeds: types.Feeds{
				{
					Ticker: btcusd,
					ProviderConfig: mmtypes.ProviderConfig{
						Name: krakenProvider,
					},
					DailyQuoteVolume: big.NewFloat(100),
					DailyUsdVolume:   big.NewFloat(100),
				},
			},
			config: config.GenerateConfig{
				Providers: map[string]config.ProviderConfig{
					krakenProvider: {
						IgnoreVolume:      true,
						MinProviderVolume: 1000,
					},
				},
			},
			expectedFeeds: types.Feeds{
				{
					Ticker: btcusd,
					ProviderConfig: mmtypes.ProviderConfig{
						Name: krakenProvider,
					},
					DailyQuoteVolume: big.NewFloat(100),
					DailyUsdVolume:   big.NewFloat(100),
				},
			},
			expectedRemoved: 0,
		},
		{
			name: "insufficient volume",
			feeds: types.Feeds{
				{
					Ticker: btcusd,
					ProviderConfig: mmtypes.ProviderConfig{
						Name: krakenProvider,
					},
					DailyQuoteVolume: big.NewFloat(100),
					DailyUsdVolume:   big.NewFloat(100),
				},
			},
			config: config.GenerateConfig{
				Providers: map[string]config.ProviderConfig{
					krakenProvider: {
						MinProviderVolume: 1000,
					},
				},
			},
			expectedFeeds:   types.Feeds{},
			expectedRemoved: 1,
		},
		{
			name: "sufficient volume",
			feeds: types.Feeds{
				{
					Ticker: btcusd,
					ProviderConfig: mmtypes.ProviderConfig{
						Name: krakenProvider,
					},
					DailyQuoteVolume: big.NewFloat(1),
					DailyUsdVolume:   big.NewFloat(2000),
				},
			},
			config: config.GenerateConfig{
				Providers: map[string]config.ProviderConfig{
					krakenProvider: {
						MinProviderVolume: 1000,
					},
				},
			},
			expectedFeeds: types.Feeds{
				{
					Ticker: btcusd,
					ProviderConfig: mmtypes.ProviderConfig{
						Name: krakenProvider,
					},
					DailyQuoteVolume: big.NewFloat(1),
					DailyUsdVolume:   big.NewFloat(2000),
				},
			},
			expectedRemoved: 0,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			logger := zaptest.NewLogger(t)
			transform := transformer.PruneByProviderUsdVolume()
			feeds, removals, err := transform(context.Background(), logger, tc.config, tc.feeds)
			require.NoError(t, err)
			require.Equal(t, tc.expectedFeeds, feeds)
			require.Equal(t, tc.expectedRemoved, len(removals))
		})
	}
}
