package transformer_test

import (
	"context"
	"math/big"
	"testing"

	"go.uber.org/zap/zaptest"

	connecttypes "github.com/skip-mev/connect/v2/pkg/types"
	mmtypes "github.com/skip-mev/connect/v2/x/marketmap/types"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/skip-mev/connect-mmu/config"
	"github.com/skip-mev/connect-mmu/generator/transformer"
	"github.com/skip-mev/connect-mmu/generator/types"
	mmutypes "github.com/skip-mev/connect-mmu/types"
)

var (
	krakenProvider  = "kraken_ws"
	binanceProvider = "binance_ws"
	bybitProvider   = "bybit_ws"

	btcusdt = mmtypes.Ticker{
		CurrencyPair: connecttypes.CurrencyPair{
			Base:  "BTC",
			Quote: "USDT",
		},
		Decimals:         8,
		MinProviderCount: 1,
		Enabled:          false,
		Metadata_JSON:    "",
	}

	btcusd = mmtypes.Ticker{
		CurrencyPair: connecttypes.CurrencyPair{
			Base:  "BTC",
			Quote: "USD",
		},
		Decimals:         8,
		MinProviderCount: 1,
		Enabled:          false,
		Metadata_JSON:    "",
	}

	usdtusd = mmtypes.Ticker{
		CurrencyPair: connecttypes.CurrencyPair{
			Base:  "USDT",
			Quote: "USD",
		},
		Decimals:         8,
		MinProviderCount: 1,
		Enabled:          false,
		Metadata_JSON:    "",
	}

	marketBtcUsdt = mmtypes.Market{
		Ticker: btcusdt,
		ProviderConfigs: []mmtypes.ProviderConfig{
			{
				Name:            krakenProvider,
				OffChainTicker:  "XXBTZUSDT",
				NormalizeByPair: nil,
				Invert:          false,
				Metadata_JSON:   "",
			},
		},
	}

	marketBtcUsd = mmtypes.Market{
		Ticker: btcusd,
		ProviderConfigs: []mmtypes.ProviderConfig{
			{
				Name:            krakenProvider,
				OffChainTicker:  "XXBTZUSD",
				NormalizeByPair: nil,
				Invert:          false,
				Metadata_JSON:   "",
			},
		},
	}

	marketUsdtUsd = mmtypes.Market{
		Ticker: usdtusd,
		ProviderConfigs: []mmtypes.ProviderConfig{
			{
				Name:            krakenProvider,
				OffChainTicker:  "USDTUSD",
				NormalizeByPair: nil,
				Invert:          false,
				Metadata_JSON:   "",
			},
		},
	}

	marketBtcUsdNormalized = mmtypes.Market{
		Ticker: btcusd,
		ProviderConfigs: []mmtypes.ProviderConfig{
			{
				Name:           krakenProvider,
				OffChainTicker: "XXBTZUSDT",
				NormalizeByPair: &connecttypes.CurrencyPair{
					Base:  "USDT",
					Quote: "USD",
				}, Invert: false,
				Metadata_JSON: "",
			},
		},
	}

	liquidityInfo1000 = mmutypes.LiquidityInfo{
		NegativeDepthTwo: 1000,
		PositiveDepthTwo: 1000,
	}

	liquidityInfo2000 = mmutypes.LiquidityInfo{
		NegativeDepthTwo: 2000,
		PositiveDepthTwo: 2000,
	}

	liquidityInfo4000 = mmutypes.LiquidityInfo{
		NegativeDepthTwo: 4000,
		PositiveDepthTwo: 4000,
	}

	cmcInfoNull = mmutypes.CoinMarketCapInfo{}
	cmcInfoA    = mmutypes.CoinMarketCapInfo{
		BaseID:    1,
		QuoteID:   2,
		BaseRank:  10,
		QuoteRank: 20,
	}

	cmcInfoAInverted = mmutypes.CoinMarketCapInfo{
		BaseID:    2,
		QuoteID:   1,
		BaseRank:  20,
		QuoteRank: 10,
	}

	cmcInfoB = mmutypes.CoinMarketCapInfo{
		BaseID:    3,
		QuoteID:   4,
		BaseRank:  30,
		QuoteRank: 40,
	}
)

func TestNormalizeBy(t *testing.T) {
	cfg := config.GenerateConfig{
		Quotes: map[string]config.QuoteConfig{
			"USDT": {
				MinProviderVolume: 100000,
				NormalizeByPair:   "USDT/USD",
			},
			"USD": {
				MinProviderVolume: 100000,
			},
		},
	}

	tests := []struct {
		name        string
		cfg         config.GenerateConfig
		feeds       types.Feeds
		transformed types.Feeds
		expectErr   bool
	}{
		{
			name:        "valid no markets",
			cfg:         cfg,
			feeds:       []types.Feed{},
			transformed: []types.Feed{},
			expectErr:   false,
		},
		{
			name: "valid 1 market to adjust",
			cfg:  cfg,
			feeds: []types.Feed{
				{
					Ticker:         marketBtcUsdt.Ticker,
					ProviderConfig: marketBtcUsdt.ProviderConfigs[0],
					ReferencePrice: big.NewFloat(10),
					CMCInfo:        cmcInfoA,
				},
				{
					Ticker:         marketUsdtUsd.Ticker,
					ProviderConfig: marketUsdtUsd.ProviderConfigs[0],
					ReferencePrice: big.NewFloat(1.1),
					CMCInfo:        usdtusdFeed.CMCInfo,
				},
			},
			transformed: []types.Feed{
				{
					Ticker:         marketBtcUsdNormalized.Ticker,
					ProviderConfig: marketBtcUsdNormalized.ProviderConfigs[0],
					ReferencePrice: big.NewFloat(11),
					CMCInfo:        cmcInfoA,
				},
				{
					Ticker:         marketUsdtUsd.Ticker,
					ProviderConfig: marketUsdtUsd.ProviderConfigs[0],
					ReferencePrice: big.NewFloat(1.1),
					CMCInfo:        usdtusdFeed.CMCInfo,
				},
			}, expectErr: false,
		},
		{
			name: "valid no markets to adjust",
			cfg:  cfg,
			feeds: []types.Feed{
				{
					Ticker:         marketBtcUsd.Ticker,
					ProviderConfig: marketBtcUsdt.ProviderConfigs[0],
					ReferencePrice: big.NewFloat(11),
					CMCInfo:        cmcInfoA,
				},
			},
			transformed: []types.Feed{
				{
					Ticker:         marketBtcUsd.Ticker,
					ProviderConfig: marketBtcUsdt.ProviderConfigs[0],
					ReferencePrice: big.NewFloat(11),
					CMCInfo:        cmcInfoA,
				},
			}, expectErr: false,
		},
		{
			name: "invalid quotes",
			cfg:  config.GenerateConfig{},
			feeds: []types.Feed{
				{
					Ticker:         marketBtcUsdt.Ticker,
					ProviderConfig: marketBtcUsdt.ProviderConfigs[0],
					ReferencePrice: big.NewFloat(0),
					CMCInfo:        cmcInfoA,
				},
			},
			expectErr: true,
		},
	}

	transform := transformer.NormalizeBy()
	ctx := context.Background()

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			transformed, _, err := transform(ctx, zaptest.NewLogger(t), tc.cfg, tc.feeds)
			if tc.expectErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			require.True(t, tc.transformed.Equal(transformed))
		})
	}
}

func TestResolveProviderConflicts(t *testing.T) {
	// run multiple times to check deterministic output
	numIters := 100

	tests := []struct {
		name        string
		cfg         config.GenerateConfig
		feeds       types.Feeds
		transformed types.Feeds
		expectErr   bool
	}{
		{
			name: "valid no markets",
			cfg: config.GenerateConfig{
				Quotes: map[string]config.QuoteConfig{
					"USDT": {
						MinProviderVolume: 100000,
						NormalizeByPair:   "USDT/USD",
					},
				},
			},
			feeds:       []types.Feed{},
			transformed: []types.Feed{},
			expectErr:   false,
		},
		{
			name: "valid 1 market no conflicts",
			cfg: config.GenerateConfig{
				Quotes: map[string]config.QuoteConfig{
					"USDT": {
						MinProviderVolume: 100000,
						NormalizeByPair:   "USDT/USD",
					},
				},
			},
			feeds: []types.Feed{
				{
					Ticker:         marketBtcUsdNormalized.Ticker,
					ProviderConfig: marketBtcUsdNormalized.ProviderConfigs[0],
					ReferencePrice: big.NewFloat(10),
					CMCInfo:        cmcInfoA,
				},
			},
			transformed: []types.Feed{
				{
					Ticker:         marketBtcUsdNormalized.Ticker,
					ProviderConfig: marketBtcUsdNormalized.ProviderConfigs[0],
					ReferencePrice: big.NewFloat(10),
					CMCInfo:        cmcInfoA,
				},
			}, expectErr: false,
		},
		{
			name: "choose higher volume and liquidity",
			cfg:  config.GenerateConfig{},
			feeds: []types.Feed{
				types.NewFeed(marketBtcUsdNormalized.Ticker, marketBtcUsdNormalized.ProviderConfigs[0], 10000.0, 10000.0,
					10000.0, liquidityInfo1000, cmcInfoA),
				types.NewFeed(marketBtcUsd.Ticker, marketBtcUsdt.ProviderConfigs[0], 20000.0, 20000.0, 20000.0,
					liquidityInfo2000, cmcInfoA),
			},
			transformed: []types.Feed{
				types.NewFeed(marketBtcUsd.Ticker, marketBtcUsdt.ProviderConfigs[0], 20000.0, 20000.0, 20000.0, liquidityInfo2000, cmcInfoA),
			},
			expectErr: false,
		},
		{
			name: "prioritize liquidity",
			cfg:  config.GenerateConfig{},
			feeds: []types.Feed{
				types.NewFeed(marketBtcUsdNormalized.Ticker, marketBtcUsdNormalized.ProviderConfigs[0], 10000.0, 10000.0,
					40000.0, liquidityInfo4000, cmcInfoA),
				types.NewFeed(marketBtcUsd.Ticker, marketBtcUsdt.ProviderConfigs[0], 20000.0, 20000.0, 20000.0, liquidityInfo2000, cmcInfoA),
			},
			transformed: []types.Feed{
				types.NewFeed(marketBtcUsdNormalized.Ticker, marketBtcUsdNormalized.ProviderConfigs[0], 10000.0, 10000.0,
					40000.0, liquidityInfo4000, cmcInfoA),
			},
			expectErr: false,
		},
		{
			name: "disjoint markets - retain both",
			cfg:  config.GenerateConfig{},
			feeds: []types.Feed{
				types.NewFeed(marketBtcUsdNormalized.Ticker, marketBtcUsdNormalized.ProviderConfigs[0], 10000.0, 10000.0,
					20000.0, liquidityInfo2000, cmcInfoA),
				types.NewFeed(marketBtcUsdt.Ticker, marketBtcUsdt.ProviderConfigs[0], 25000.0, 25000.0, 20000.0, liquidityInfo4000, cmcInfoA),
			},
			transformed: []types.Feed{
				types.NewFeed(marketBtcUsdt.Ticker, marketBtcUsdt.ProviderConfigs[0], 25000.0, 25000.0, 20000.0, liquidityInfo4000, cmcInfoA),
				types.NewFeed(marketBtcUsdNormalized.Ticker, marketBtcUsdNormalized.ProviderConfigs[0], 10000.0, 10000.0,
					20000.0, liquidityInfo2000, cmcInfoA),
			},
			expectErr: false,
		},
	}

	transform := transformer.ResolveConflictsForProvider()
	ctx := context.Background()

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			for range numIters {
				transformed, _, err := transform(ctx, zaptest.NewLogger(t), tc.cfg, tc.feeds)
				if tc.expectErr {
					require.Error(t, err)
					return
				}

				require.NoError(t, err)
				require.True(t, tc.transformed.Equal(transformed))
			}
		})
	}
}

func TestDropFeeds(t *testing.T) {
	tests := []struct {
		name        string
		cfg         config.GenerateConfig
		feeds       types.Feeds
		transformed types.Feeds
		dropped     []string
		expectErr   bool
	}{
		{
			name: "valid no markets",
			cfg: config.GenerateConfig{
				Providers: map[string]config.ProviderConfig{
					krakenProvider: {
						RequireAggregateIDs: true,
					},
				},
				Quotes: map[string]config.QuoteConfig{
					"USDT": {
						MinProviderVolume: 10000,
						NormalizeByPair:   "USDT/USD",
					},
				},
			},
			feeds:       []types.Feed{},
			transformed: []types.Feed{},
			expectErr:   false,
		},
		{
			name: "drop single market",
			cfg: config.GenerateConfig{
				Providers: map[string]config.ProviderConfig{
					krakenProvider: {
						RequireAggregateIDs: true,
					},
				},
			},
			feeds: []types.Feed{
				types.NewFeed(marketBtcUsdNormalized.Ticker, marketBtcUsdNormalized.ProviderConfigs[0], 10000.0, 10000.0,
					20000.0, liquidityInfo2000, cmcInfoNull),
				types.NewFeed(marketBtcUsd.Ticker, marketBtcUsdt.ProviderConfigs[0], 20000.0, 20000.0, 20000.0, liquidityInfo2000, cmcInfoA),
			},
			transformed: []types.Feed{
				types.NewFeed(marketBtcUsd.Ticker, marketBtcUsdt.ProviderConfigs[0], 20000.0, 20000.0, 20000.0, liquidityInfo2000, cmcInfoA),
			},
			dropped:   []string{marketBtcUsd.Ticker.String()},
			expectErr: false,
		},
		{
			name: "drop no markets",
			cfg: config.GenerateConfig{
				Providers: map[string]config.ProviderConfig{
					krakenProvider: {
						RequireAggregateIDs: true,
					},
				},
			},
			feeds: []types.Feed{
				types.NewFeed(marketBtcUsdNormalized.Ticker, marketBtcUsdNormalized.ProviderConfigs[0], 10000.0, 10000.0,
					20000.0, liquidityInfo2000, cmcInfoA),
				types.NewFeed(marketBtcUsdt.Ticker, marketBtcUsdt.ProviderConfigs[0], 20000.0, 20000.0, 20000.0, liquidityInfo2000, cmcInfoA),
			},
			transformed: []types.Feed{
				types.NewFeed(marketBtcUsdNormalized.Ticker, marketBtcUsdNormalized.ProviderConfigs[0], 10000.0, 10000.0,
					20000.0, liquidityInfo2000, cmcInfoA),
				types.NewFeed(marketBtcUsdt.Ticker, marketBtcUsdt.ProviderConfigs[0], 20000.0, 20000.0, 20000.0, liquidityInfo2000, cmcInfoA),
			},
			expectErr: false,
		},
	}

	transform := transformer.DropFeedsWithoutAggregatorIDs()
	ctx := context.Background()

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			transformed, dropped, err := transform(ctx, zap.NewNop(), tc.cfg, tc.feeds)
			if tc.expectErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			require.True(t, tc.transformed.Equal(transformed))
			var droppedKeys []string
			for k := range dropped {
				droppedKeys = append(droppedKeys, k)
			}
			require.Equal(t, tc.dropped, droppedKeys)
		})
	}
}

func TestInvert(t *testing.T) {
	tests := []struct {
		name        string
		cfg         config.GenerateConfig
		feeds       types.Feeds
		transformed types.Feeds
		dropped     []string
		expectErr   bool
	}{
		{
			name: "valid no markets",
			cfg: config.GenerateConfig{
				Providers: map[string]config.ProviderConfig{
					krakenProvider: {
						RequireAggregateIDs: true,
					},
				},
				Quotes: map[string]config.QuoteConfig{
					"USDT": {
						MinProviderVolume: 100000,
					},
				},
			},
			feeds:       []types.Feed{},
			transformed: []types.Feed{},
			expectErr:   false,
		},
		{
			name:        "valid no markets no config",
			cfg:         config.GenerateConfig{},
			feeds:       []types.Feed{},
			transformed: []types.Feed{},
			expectErr:   false,
		},
		{
			name: "valid base and quote not in config - drop",
			cfg: config.GenerateConfig{
				Providers: map[string]config.ProviderConfig{
					krakenProvider: {
						RequireAggregateIDs: true,
					},
				},
				Quotes: map[string]config.QuoteConfig{
					"USDT": {
						MinProviderVolume: 100000,
					},
				},
			},
			feeds: []types.Feed{
				{
					Ticker: mmtypes.Ticker{
						CurrencyPair: connecttypes.CurrencyPair{
							Base:  "BTC",
							Quote: "MOG",
						},
						Decimals:         10,
						MinProviderCount: 1,
						Enabled:          false,
						Metadata_JSON:    "",
					},
					ProviderConfig: mmtypes.ProviderConfig{
						Name:           "test",
						OffChainTicker: "test",
						Invert:         false,
					},
					ReferencePrice: big.NewFloat(2),
					CMCInfo:        cmcInfoA,
				},
			},
			transformed: []types.Feed{},
			dropped:     []string{"BTC/MOG"},
			expectErr:   false,
		},
		{
			name: "valid quote is in config - do nothing",
			cfg: config.GenerateConfig{
				Providers: map[string]config.ProviderConfig{
					krakenProvider: {
						RequireAggregateIDs: true,
					},
				},
				Quotes: map[string]config.QuoteConfig{
					"MOG": {
						MinProviderVolume: 100000,
					},
				},
			},
			feeds: []types.Feed{
				{
					Ticker: mmtypes.Ticker{
						CurrencyPair: connecttypes.CurrencyPair{
							Base:  "BTC",
							Quote: "MOG",
						},
						Decimals:         10,
						MinProviderCount: 1,
						Enabled:          false,
						Metadata_JSON:    "",
					},
					ProviderConfig: mmtypes.ProviderConfig{
						Name:           "test",
						OffChainTicker: "test",
						Invert:         false,
					},
					ReferencePrice: big.NewFloat(2),
					CMCInfo:        cmcInfoA,
				},
			},
			transformed: []types.Feed{
				{
					Ticker: mmtypes.Ticker{
						CurrencyPair: connecttypes.CurrencyPair{
							Base:  "BTC",
							Quote: "MOG",
						},
						Decimals:         10,
						MinProviderCount: 1,
						Enabled:          false,
						Metadata_JSON:    "",
					},
					ProviderConfig: mmtypes.ProviderConfig{
						Name:           "test",
						OffChainTicker: "test",
						Invert:         false,
					},
					ReferencePrice: big.NewFloat(2),
					CMCInfo:        cmcInfoA,
				},
			},
			expectErr: false,
		},
		{
			name: "valid base and quote is in config - do nothing",
			cfg: config.GenerateConfig{
				Providers: map[string]config.ProviderConfig{
					krakenProvider: {
						RequireAggregateIDs: true,
					},
				},
				Quotes: map[string]config.QuoteConfig{
					"MOG": {
						MinProviderVolume: 100000,
					},
					"BTC": {
						MinProviderVolume: 100000,
					},
				},
			},
			feeds: []types.Feed{
				{
					Ticker: mmtypes.Ticker{
						CurrencyPair: connecttypes.CurrencyPair{
							Base:  "BTC",
							Quote: "MOG",
						},
						Decimals:         10,
						MinProviderCount: 1,
						Enabled:          false,
						Metadata_JSON:    "",
					},
					ProviderConfig: mmtypes.ProviderConfig{
						Name:           "test",
						OffChainTicker: "test",
						Invert:         false,
					},
					ReferencePrice: big.NewFloat(2),
					CMCInfo:        cmcInfoA,
				},
			},
			transformed: []types.Feed{
				{
					Ticker: mmtypes.Ticker{
						CurrencyPair: connecttypes.CurrencyPair{
							Base:  "BTC",
							Quote: "MOG",
						},
						Decimals:         10,
						MinProviderCount: 1,
						Enabled:          false,
						Metadata_JSON:    "",
					},
					ProviderConfig: mmtypes.ProviderConfig{
						Name:           "test",
						OffChainTicker: "test",
						Invert:         false,
					},
					ReferencePrice: big.NewFloat(2),
					CMCInfo:        cmcInfoA,
				},
			},
			expectErr: false,
		},
		{
			name: "valid base is in config - invert",
			cfg: config.GenerateConfig{
				Providers: map[string]config.ProviderConfig{
					krakenProvider: {
						RequireAggregateIDs: true,
					},
				},
				MinCexProviderCount: 1,
				MinDexProviderCount: 1,
				Quotes: map[string]config.QuoteConfig{
					"BTC": {
						MinProviderVolume: 100000,
					},
				},
			},
			feeds: []types.Feed{
				{
					Ticker: mmtypes.Ticker{
						CurrencyPair: connecttypes.CurrencyPair{
							Base:  "BTC",
							Quote: "MOG",
						},
						Decimals:         10,
						MinProviderCount: 1,
						Enabled:          false,
						Metadata_JSON:    "",
					},
					ProviderConfig: mmtypes.ProviderConfig{
						Name:           krakenProvider,
						OffChainTicker: "test",
						Invert:         false,
					},
					ReferencePrice: big.NewFloat(2),
					CMCInfo:        cmcInfoA,
				},
			},
			transformed: []types.Feed{
				{
					Ticker: mmtypes.Ticker{
						CurrencyPair: connecttypes.CurrencyPair{
							Base:  "MOG",
							Quote: "BTC",
						},
						Decimals:         10,
						MinProviderCount: 1,
						Enabled:          false,
						Metadata_JSON:    "",
					},
					ProviderConfig: mmtypes.ProviderConfig{
						Name:           krakenProvider,
						OffChainTicker: "test",
						Invert:         true,
					},
					ReferencePrice: big.NewFloat(0.5),
					CMCInfo:        cmcInfoAInverted,
				},
			},
			expectErr: false,
		},
	}

	transform := transformer.InvertOrDrop()
	ctx := context.Background()

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			transformed, dropped, err := transform(ctx, zaptest.NewLogger(t), tc.cfg, tc.feeds)
			if tc.expectErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			require.True(t, tc.transformed.Equal(transformed))
			var droppedKeys []string
			for k := range dropped {
				droppedKeys = append(droppedKeys, k)
			}
			require.Equal(t, tc.dropped, droppedKeys)
		})
	}
}

func TestPruneByQuoteVolume(t *testing.T) {
	tests := []struct {
		name        string
		cfg         config.GenerateConfig
		feeds       types.Feeds
		transformed types.Feeds
		dropped     []string
		expectErr   bool
	}{
		{
			name: "valid no markets",
			cfg: config.GenerateConfig{
				Providers: map[string]config.ProviderConfig{
					krakenProvider: {
						RequireAggregateIDs: true,
					},
				},
				Quotes: map[string]config.QuoteConfig{
					"USD": {
						MinProviderVolume: 100000,
					},
				},
			},
			feeds:       []types.Feed{},
			transformed: []types.Feed{},
			expectErr:   false,
		},
		{
			name: "valid exclude a market with no associated quote config, keep others",
			cfg: config.GenerateConfig{
				Providers: map[string]config.ProviderConfig{
					krakenProvider: {
						RequireAggregateIDs: true,
					},
				},
				Quotes: map[string]config.QuoteConfig{
					"USD": {
						MinProviderVolume: 100000,
					},
				},
			},
			feeds: []types.Feed{
				types.NewFeed(marketBtcUsdNormalized.Ticker, marketBtcUsdNormalized.ProviderConfigs[0], 100000.0, 100000.0,
					20000.0, liquidityInfo2000, cmcInfoA),
				types.NewFeed(marketBtcUsdt.Ticker, marketBtcUsdt.ProviderConfigs[0], 200000.0, 200000.0, 20000.0, liquidityInfo2000, cmcInfoA),
			},
			transformed: []types.Feed{
				types.NewFeed(marketBtcUsdNormalized.Ticker, marketBtcUsdNormalized.ProviderConfigs[0], 100000.0, 100000.0,
					20000.0, liquidityInfo2000, cmcInfoA),
			},
			dropped:   []string{marketBtcUsdt.Ticker.String()},
			expectErr: false,
		},
		{
			name: "valid market with enough volume",
			cfg: config.GenerateConfig{
				Providers: map[string]config.ProviderConfig{
					krakenProvider: {
						RequireAggregateIDs: true,
					},
				},
				Quotes: map[string]config.QuoteConfig{
					"USD": {
						MinProviderVolume: 100000,
					},
				},
			},
			feeds: []types.Feed{
				types.NewFeed(marketBtcUsd.Ticker, marketBtcUsd.ProviderConfigs[0], 100000, 100000, 20000.0, liquidityInfo2000, cmcInfoA),
			},
			transformed: []types.Feed{
				types.NewFeed(marketBtcUsd.Ticker, marketBtcUsd.ProviderConfigs[0], 100000, 100000, 20000.0, liquidityInfo2000, cmcInfoA),
			},
			expectErr: false,
		},
		{
			name: "valid market with insufficient volume pruned",
			cfg: config.GenerateConfig{
				Providers: map[string]config.ProviderConfig{
					krakenProvider: {
						RequireAggregateIDs: true,
					},
				},
				Quotes: map[string]config.QuoteConfig{
					"USD": {
						MinProviderVolume: 100000,
					},
				},
			},
			feeds: []types.Feed{
				types.NewFeed(marketBtcUsd.Ticker, marketBtcUsd.ProviderConfigs[0], 0, 0, 20000.0, liquidityInfo2000, cmcInfoA),
			},
			transformed: []types.Feed{},
			dropped:     []string{marketBtcUsd.Ticker.String()},
			expectErr:   false,
		},
	}

	transform := transformer.PruneByQuoteVolume()
	ctx := context.Background()

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			transformed, dropped, err := transform(ctx, zaptest.NewLogger(t), tc.cfg, tc.feeds)
			if tc.expectErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			require.Equal(t, tc.transformed, transformed)
			var droppedKeys []string
			for k := range dropped {
				droppedKeys = append(droppedKeys, k)
			}
			require.Equal(t, tc.dropped, droppedKeys)
		})
	}
}

func TestPruneByLiquidity(t *testing.T) {
	tests := []struct {
		name        string
		cfg         config.GenerateConfig
		feeds       types.Feeds
		transformed types.Feeds
		dropped     []string
		expectErr   bool
	}{
		{
			name: "valid no markets",
			cfg: config.GenerateConfig{
				Providers: map[string]config.ProviderConfig{
					krakenProvider: {
						RequireAggregateIDs: true,
					},
				},
				MinCexProviderCount: 1,
				MinDexProviderCount: 1,
				Quotes: map[string]config.QuoteConfig{
					"USD": {
						MinProviderLiquidity: 1000,
					},
				},
			},
			feeds:       []types.Feed{},
			transformed: []types.Feed{},
			expectErr:   false,
		},
		{
			name: "valid exclude a market with no associated quote config, keep others",
			cfg: config.GenerateConfig{
				Providers: map[string]config.ProviderConfig{
					krakenProvider: {
						RequireAggregateIDs: true,
					},
				},
				MinCexProviderCount: 1,
				MinDexProviderCount: 1,
				Quotes: map[string]config.QuoteConfig{
					"USD": {
						MinProviderLiquidity: 1000,
					},
				},
			},
			feeds: []types.Feed{
				types.NewFeed(marketBtcUsdNormalized.Ticker, marketBtcUsdNormalized.ProviderConfigs[0], 100000.0, 100000.0,
					200000.0, liquidityInfo2000, cmcInfoNull),
				types.NewFeed(marketBtcUsdt.Ticker, marketBtcUsdt.ProviderConfigs[0], 200000.0, 200000.0, 200000.0, liquidityInfo2000, cmcInfoNull),
			},
			transformed: []types.Feed{
				types.NewFeed(marketBtcUsdNormalized.Ticker, marketBtcUsdNormalized.ProviderConfigs[0], 100000.0, 100000.0,
					200000.0, liquidityInfo2000, cmcInfoNull),
			},
			dropped:   []string{marketBtcUsdt.Ticker.String()},
			expectErr: false,
		},
		{
			name: "valid market with enough liquidity",
			cfg: config.GenerateConfig{
				Providers: map[string]config.ProviderConfig{
					krakenProvider: {
						RequireAggregateIDs: true,
					},
				},
				MinCexProviderCount: 1,
				MinDexProviderCount: 1,
				Quotes: map[string]config.QuoteConfig{
					"USD": {
						MinProviderLiquidity: 1000,
					},
				},
			},
			feeds: []types.Feed{
				types.NewFeed(marketBtcUsd.Ticker, marketBtcUsd.ProviderConfigs[0], 100000, 100000, 200000.0, liquidityInfo2000, cmcInfoNull),
			},
			transformed: []types.Feed{
				types.NewFeed(marketBtcUsd.Ticker, marketBtcUsd.ProviderConfigs[0], 100000, 100000, 200000.0, liquidityInfo2000, cmcInfoNull),
			},
			expectErr: false,
		},
		{
			name: "valid market with insufficient liquidity pruned",
			cfg: config.GenerateConfig{
				Providers: map[string]config.ProviderConfig{
					krakenProvider: {
						RequireAggregateIDs: true,
					},
				},
				MinCexProviderCount: 1,
				MinDexProviderCount: 1,
				Quotes: map[string]config.QuoteConfig{
					"USD": {
						MinProviderLiquidity: 100000,
					},
				},
			},
			feeds: []types.Feed{
				types.NewFeed(marketBtcUsd.Ticker, marketBtcUsd.ProviderConfigs[0], 0, 0, 20000.0, liquidityInfo2000, cmcInfoNull),
			},
			transformed: []types.Feed{},
			dropped:     []string{marketBtcUsd.Ticker.String()},
			expectErr:   false,
		},
	}

	transform := transformer.PruneByLiquidity()
	ctx := context.Background()

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			transformed, dropped, err := transform(ctx, zaptest.NewLogger(t), tc.cfg, tc.feeds)
			if tc.expectErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			require.True(t, tc.transformed.Equal(transformed))
			var droppedKeys []string
			for k := range dropped {
				droppedKeys = append(droppedKeys, k)
			}
			require.Equal(t, tc.dropped, droppedKeys)
		})
	}
}

func TestTopMarketsForProvider(t *testing.T) {
	cfg := config.GenerateConfig{
		Providers: map[string]config.ProviderConfig{
			krakenProvider: {
				IsSupplemental:      false,
				RequireAggregateIDs: false,
				Filters: config.Filters{
					TopMarkets: 2,
				},
			},
			binanceProvider: {
				IsSupplemental:      false,
				RequireAggregateIDs: false,
				Filters: config.Filters{
					TopMarkets: 2,
				},
			},
		},
	}

	tests := []struct {
		name           string
		feeds          types.Feeds
		cfg            config.GenerateConfig
		want           types.Feeds
		wantExclusions types.ExclusionReasons
		wantErr        bool
	}{
		{
			name: "return nothing for no invalid provider configs",
			feeds: types.Feeds{
				{
					Ticker: marketBtcUsd.Ticker,
					ProviderConfig: mmtypes.ProviderConfig{
						Name:            "invalid",
						OffChainTicker:  "",
						NormalizeByPair: nil,
						Invert:          false,
						Metadata_JSON:   "",
					},
					DailyUsdVolume: big.NewFloat(100),
					CMCInfo:        cmcInfoA,
					ReferencePrice: big.NewFloat(100),
				},
			},
			cfg:            cfg,
			want:           nil,
			wantExclusions: types.ExclusionReasons{},
			wantErr:        true,
		},
		{
			name:           "do nothing for no provider feeds",
			feeds:          nil,
			cfg:            cfg,
			want:           nil,
			wantExclusions: types.ExclusionReasons{},
			wantErr:        false,
		},
		{
			name: "retain all markets for a provider with no filter - should sort",
			feeds: types.Feeds{
				{
					Ticker:         marketBtcUsd.Ticker,
					ProviderConfig: marketBtcUsd.ProviderConfigs[0],
					DailyUsdVolume: big.NewFloat(100),
					ReferencePrice: big.NewFloat(100),
					CMCInfo:        cmcInfoA,
				},
				{
					Ticker:         marketBtcUsdt.Ticker,
					ProviderConfig: marketBtcUsdt.ProviderConfigs[0],
					DailyUsdVolume: big.NewFloat(1000),
					ReferencePrice: big.NewFloat(100),
					CMCInfo:        cmcInfoA,
				},
			},

			cfg: cfg,
			want: types.Feeds{
				{
					Ticker:         marketBtcUsdt.Ticker,
					ProviderConfig: marketBtcUsdt.ProviderConfigs[0],
					DailyUsdVolume: big.NewFloat(1000),
					ReferencePrice: big.NewFloat(100),
					CMCInfo:        cmcInfoA,
				},
				{
					Ticker:         marketBtcUsd.Ticker,
					ProviderConfig: marketBtcUsd.ProviderConfigs[0],
					DailyUsdVolume: big.NewFloat(100),
					ReferencePrice: big.NewFloat(100),
					CMCInfo:        cmcInfoA,
				},
			},
			wantExclusions: types.ExclusionReasons{},
			wantErr:        false,
		},
		{
			name: "exclude market with lower quote volume for provider with filter - will order feeds",
			feeds: types.Feeds{
				{
					Ticker:         marketBtcUsdt.Ticker,
					ProviderConfig: marketBtcUsdt.ProviderConfigs[0],
					DailyUsdVolume: big.NewFloat(1000),
					ReferencePrice: big.NewFloat(100),
					CMCInfo:        cmcInfoA,
					LiquidityInfo: mmutypes.LiquidityInfo{
						NegativeDepthTwo: 1000,
						PositiveDepthTwo: 1000,
					},
				},
				{
					Ticker:         marketBtcUsd.Ticker,
					ProviderConfig: marketBtcUsd.ProviderConfigs[0],
					DailyUsdVolume: big.NewFloat(100),
					ReferencePrice: big.NewFloat(100),
					CMCInfo:        cmcInfoA,
					LiquidityInfo: mmutypes.LiquidityInfo{
						NegativeDepthTwo: 100,
						PositiveDepthTwo: 100,
					},
				},
				{
					Ticker:         marketBtcUsdNormalized.Ticker,
					ProviderConfig: marketBtcUsdNormalized.ProviderConfigs[0],
					DailyUsdVolume: big.NewFloat(2000),
					ReferencePrice: big.NewFloat(100),
					CMCInfo:        cmcInfoA,
					LiquidityInfo: mmutypes.LiquidityInfo{
						NegativeDepthTwo: 2000,
						PositiveDepthTwo: 2000,
					},
				},
			},
			cfg: cfg,
			want: types.Feeds{
				{
					Ticker:         marketBtcUsdNormalized.Ticker,
					ProviderConfig: marketBtcUsdNormalized.ProviderConfigs[0],
					DailyUsdVolume: big.NewFloat(2000),
					ReferencePrice: big.NewFloat(100),
					CMCInfo:        cmcInfoA,
					LiquidityInfo: mmutypes.LiquidityInfo{
						NegativeDepthTwo: 2000,
						PositiveDepthTwo: 2000,
					},
				},
				{
					Ticker:         marketBtcUsdt.Ticker,
					ProviderConfig: marketBtcUsdt.ProviderConfigs[0],
					DailyUsdVolume: big.NewFloat(1000),
					ReferencePrice: big.NewFloat(100),
					CMCInfo:        cmcInfoA,
					LiquidityInfo: mmutypes.LiquidityInfo{
						NegativeDepthTwo: 1000,
						PositiveDepthTwo: 1000,
					},
				},
			},
			wantExclusions: types.ExclusionReasons{"BTC/USD": []types.ExclusionReason{{
				Reason:   "only selecting top 2 feeds for this provider",
				Provider: krakenProvider,
				Feed: types.Feed{
					Ticker:         marketBtcUsd.Ticker,
					ProviderConfig: marketBtcUsd.ProviderConfigs[0],
					DailyUsdVolume: big.NewFloat(100),
					ReferencePrice: big.NewFloat(100),
					CMCInfo:        cmcInfoA,
					LiquidityInfo: mmutypes.LiquidityInfo{
						NegativeDepthTwo: 100,
						PositiveDepthTwo: 100,
					},
				},
			}}},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		transform := transformer.TopFeedsForProvider()

		t.Run(tt.name, func(t *testing.T) {
			got, exclusions, err := transform(context.Background(), zap.NewNop(), tt.cfg, tt.feeds)
			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			require.True(t, tt.want.Equal(got))
			require.Equal(t, tt.wantExclusions, exclusions)
		})
	}
}

func TestResolveNamingAliases(t *testing.T) {
	tests := []struct {
		name        string
		cfg         config.GenerateConfig
		feeds       types.Feeds
		transformed types.Feeds
		dropped     []string
		expectErr   bool
	}{
		{
			name:        "valid no markets",
			cfg:         config.GenerateConfig{},
			feeds:       []types.Feed{},
			transformed: []types.Feed{},
			expectErr:   false,
		},
		{
			name: "valid two markets that do not intersect on TickerString",
			cfg:  config.GenerateConfig{},
			feeds: []types.Feed{
				types.NewFeed(
					marketBtcUsdNormalized.Ticker,
					marketBtcUsdNormalized.ProviderConfigs[0],
					100000.0,
					100000.0,
					20000.0,
					liquidityInfo2000,
					cmcInfoA,
				),
				types.NewFeed(
					marketBtcUsdt.Ticker,
					marketBtcUsdt.ProviderConfigs[0],
					200000.0,
					200000.0,
					20000.0,
					liquidityInfo2000,
					cmcInfoB,
				),
			},
			transformed: []types.Feed{
				types.NewFeed(
					marketBtcUsdNormalized.Ticker,
					marketBtcUsdNormalized.ProviderConfigs[0],
					100000.0,
					100000.0,
					20000.0,
					liquidityInfo2000,
					cmcInfoA,
				),
				types.NewFeed(
					marketBtcUsdt.Ticker,
					marketBtcUsdt.ProviderConfigs[0],
					200000.0,
					200000.0,
					20000.0,
					liquidityInfo2000,
					cmcInfoB,
				),
			},
			dropped:   nil,
			expectErr: false,
		},
		{
			name: "valid two markets that intersect on TickerString, but do not for CMC Info - prune 1 (choosing higher rank)",
			cfg:  config.GenerateConfig{},
			feeds: []types.Feed{
				types.NewFeed(
					marketBtcUsdt.Ticker,
					marketBtcUsdNormalized.ProviderConfigs[0],
					100000.0,
					100000.0,
					20000.0,
					liquidityInfo2000,
					cmcInfoA,
				),
				types.NewFeed(
					marketBtcUsdt.Ticker,
					marketBtcUsdt.ProviderConfigs[0],
					200000.0,
					200000.0,
					20000.0,
					liquidityInfo2000,
					cmcInfoB,
				),
			},
			transformed: []types.Feed{
				types.NewFeed(
					marketBtcUsdt.Ticker,
					marketBtcUsdNormalized.ProviderConfigs[0],
					100000.0,
					100000.0,
					20000.0,
					liquidityInfo2000,
					cmcInfoA,
				),
			},
			dropped:   []string{marketBtcUsdt.Ticker.String()},
			expectErr: false,
		},
		{
			name: "mix of intersecting and non-intersection Feeds",
			cfg:  config.GenerateConfig{},
			feeds: []types.Feed{
				types.NewFeed(
					marketBtcUsdNormalized.Ticker,
					marketBtcUsdNormalized.ProviderConfigs[0],
					100000.0,
					100000.0,
					20000.0,
					liquidityInfo2000,
					cmcInfoB,
				),
				types.NewFeed(
					marketBtcUsdt.Ticker,
					marketBtcUsdNormalized.ProviderConfigs[0],
					100000.0,
					100000.0,
					20000.0,
					liquidityInfo2000,
					cmcInfoA,
				),
				types.NewFeed(
					marketBtcUsdt.Ticker,
					marketBtcUsdt.ProviderConfigs[0],
					200000.0,
					200000.0,
					20000.0,
					liquidityInfo2000,
					cmcInfoB,
				),
			},
			transformed: []types.Feed{
				types.NewFeed(
					marketBtcUsdt.Ticker,
					marketBtcUsdNormalized.ProviderConfigs[0],
					100000.0,
					100000.0,
					20000.0,
					liquidityInfo2000,
					cmcInfoA,
				),
				types.NewFeed(
					marketBtcUsdNormalized.Ticker,
					marketBtcUsdNormalized.ProviderConfigs[0],
					100000.0,
					100000.0,
					20000.0,
					liquidityInfo2000,
					cmcInfoB),
			},
			dropped:   []string{marketBtcUsdt.Ticker.String()},
			expectErr: false,
		},
	}

	transform := transformer.ResolveNamingAliases()
	ctx := context.Background()

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			transformed, dropped, err := transform(ctx, zaptest.NewLogger(t), tc.cfg, tc.feeds)
			if tc.expectErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			require.Equal(t, len(tc.transformed), len(transformed))
			require.True(t, tc.transformed.Equal(transformed))
			var droppedKeys []string
			for k := range dropped {
				droppedKeys = append(droppedKeys, k)
			}
			require.Equal(t, tc.dropped, droppedKeys)
		})
	}
}

func TestResolveCMCConflictsForMarket(t *testing.T) {
	tests := []struct {
		name          string
		feeds         types.Feeds
		expectedFeeds types.Feeds
	}{
		{
			name: "no conflicts - single feed",
			feeds: types.Feeds{
				{
					Ticker: btcusdt,
					ProviderConfig: mmtypes.ProviderConfig{
						Name: krakenProvider,
					},
					CMCInfo: mmutypes.CoinMarketCapInfo{
						BaseID:   1,   // BTC
						QuoteID:  825, // USDT
						BaseRank: 1,
					},
					DailyUsdVolume: big.NewFloat(0),
				},
			},
			expectedFeeds: types.Feeds{
				{
					Ticker: btcusdt,
					ProviderConfig: mmtypes.ProviderConfig{
						Name: krakenProvider,
					},
					CMCInfo: mmutypes.CoinMarketCapInfo{
						BaseID:   1,   // BTC
						QuoteID:  825, // USDT
						BaseRank: 1,
					},
					DailyUsdVolume: big.NewFloat(0),
				},
			},
		},
		{
			name: "resolve conflicts - keep lowest CMC ID",
			feeds: types.Feeds{
				{
					Ticker: btcusdt,
					ProviderConfig: mmtypes.ProviderConfig{
						Name: krakenProvider,
					},
					CMCInfo: mmutypes.CoinMarketCapInfo{
						BaseID:   1,   // BTC
						QuoteID:  825, // USDT
						BaseRank: 1,
					},
					DailyUsdVolume: big.NewFloat(0),
				},
				{
					Ticker: btcusdt,
					ProviderConfig: mmtypes.ProviderConfig{
						Name: binanceProvider,
					},
					CMCInfo: mmutypes.CoinMarketCapInfo{
						BaseID:   1,   // BTC
						QuoteID:  825, // USDT
						BaseRank: 1,
					},
					DailyUsdVolume: big.NewFloat(10),
				},
				{
					Ticker: btcusdt,
					ProviderConfig: mmtypes.ProviderConfig{
						Name: bybitProvider,
					},
					CMCInfo: mmutypes.CoinMarketCapInfo{
						BaseID:   2,   // Different CMC ID for BTC (should be dropped)
						QuoteID:  825, // USDT
						BaseRank: 100,
					},
					DailyUsdVolume: big.NewFloat(20),
				},
			},
			expectedFeeds: types.Feeds{
				{
					Ticker: btcusdt,
					ProviderConfig: mmtypes.ProviderConfig{
						Name: binanceProvider,
					},
					CMCInfo: mmutypes.CoinMarketCapInfo{
						BaseID:   1,   // BTC
						QuoteID:  825, // USDT
						BaseRank: 1,
					},
					DailyUsdVolume: big.NewFloat(10),
				},
				{
					Ticker: btcusdt,
					ProviderConfig: mmtypes.ProviderConfig{
						Name: krakenProvider,
					},
					CMCInfo: mmutypes.CoinMarketCapInfo{
						BaseID:   1,   // BTC
						QuoteID:  825, // USDT
						BaseRank: 1,
					},
					DailyUsdVolume: big.NewFloat(0),
				},
			},
		},
		{
			name: "multiple tickers with conflicts",
			feeds: types.Feeds{
				{
					Ticker: btcusdt,
					ProviderConfig: mmtypes.ProviderConfig{
						Name: krakenProvider,
					},
					CMCInfo: mmutypes.CoinMarketCapInfo{
						BaseID:   1,   // BTC
						QuoteID:  825, // USDT
						BaseRank: 1,
					},
					DailyUsdVolume: big.NewFloat(0),
				},
				{
					Ticker: btcusd,
					ProviderConfig: mmtypes.ProviderConfig{
						Name: binanceProvider,
					},
					CMCInfo: mmutypes.CoinMarketCapInfo{
						BaseID:   1,    // BTC
						QuoteID:  2781, // USD
						BaseRank: 1,
					},
					DailyUsdVolume: big.NewFloat(0),
				},
				{
					Ticker: btcusd,
					ProviderConfig: mmtypes.ProviderConfig{
						Name: bybitProvider,
					},
					CMCInfo: mmutypes.CoinMarketCapInfo{
						BaseID:   2,    // Different CMC ID for BTC (should be dropped)
						QuoteID:  2781, // USD
						BaseRank: 100,
					},
					DailyUsdVolume: big.NewFloat(10),
				},
			},
			expectedFeeds: types.Feeds{
				{
					Ticker: btcusdt,
					ProviderConfig: mmtypes.ProviderConfig{
						Name: krakenProvider,
					},
					CMCInfo: mmutypes.CoinMarketCapInfo{
						BaseID:   1,   // BTC
						QuoteID:  825, // USDT
						BaseRank: 1,
					},
					DailyUsdVolume: big.NewFloat(0),
				},
				{
					Ticker: btcusd,
					ProviderConfig: mmtypes.ProviderConfig{
						Name: binanceProvider,
					},
					CMCInfo: mmutypes.CoinMarketCapInfo{
						BaseID:   1,    // BTC
						QuoteID:  2781, // USD
						BaseRank: 1,
					},
					DailyUsdVolume: big.NewFloat(0),
				},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			transform := transformer.ResolveCMCConflictsForMarket()
			result, _, err := transform(context.Background(), zaptest.NewLogger(t), config.GenerateConfig{}, tc.feeds)
			require.NoError(t, err)
			require.ElementsMatch(t, tc.expectedFeeds, result)
		})
	}
}
