package config_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	connecttypes "github.com/skip-mev/connect/v2/pkg/types"
	"github.com/skip-mev/connect/v2/x/marketmap/types"

	"github.com/skip-mev/connect-mmu/config"
)

func TestValidateGenerateConfig(t *testing.T) {
	tcs := []struct {
		name        string
		cfg         config.GenerateConfig
		expectedErr bool
	}{
		{
			"invalid provider config - invalid name",
			config.GenerateConfig{
				Providers: map[string]config.ProviderConfig{
					"": {},
				},
				MinCexProviderCount:      1,
				MinDexProviderCount:      1,
				MinProviderCountOverride: 1,
				Quotes: map[string]config.QuoteConfig{
					"BTC": {
						MinProviderVolume: 10,
						NormalizeByPair:   "",
					},
				},
			},
			true,
		},
		{
			"invalid target quote",
			config.GenerateConfig{
				Providers: map[string]config.ProviderConfig{
					"okx": {},
				},
				MinCexProviderCount: 1,
				MinDexProviderCount: 1,
				Quotes: map[string]config.QuoteConfig{
					"": {
						MinProviderVolume: 10,
						NormalizeByPair:   "",
					},
				},
				EnableAll: false,
			},
			true,
		},
		{
			"invalid currency pair in excluded pairs",
			config.GenerateConfig{
				Providers: map[string]config.ProviderConfig{
					"okx": {},
				},
				MinCexProviderCount: 1,
				MinDexProviderCount: 1,
				Quotes: map[string]config.QuoteConfig{
					"FOO": {
						MinProviderVolume: 10,
						NormalizeByPair:   "",
					},
				},
				ExcludeCurrencyPairs: map[string]struct{}{
					"FOOBAR": {},
				},
			},
			true,
		},
		{
			"min cex provider count is zero",
			config.GenerateConfig{
				Providers: map[string]config.ProviderConfig{
					"okx": {},
				},
				Quotes: map[string]config.QuoteConfig{
					"OK": {
						MinProviderVolume: 1,
					},
				},
				MinCexProviderCount: 0,
				MinDexProviderCount: 1,
			},
			true,
		},
		{
			"min dex provider count is zero",
			config.GenerateConfig{
				Providers: map[string]config.ProviderConfig{
					"okx": {},
				},
				Quotes: map[string]config.QuoteConfig{
					"OK": {
						MinProviderVolume: 1,
					},
				},
				MinCexProviderCount: 1,
				MinDexProviderCount: 0,
			},
			true,
		},
		{
			"min provider volume is negative",
			config.GenerateConfig{
				Providers: map[string]config.ProviderConfig{
					"okx": {},
				},
				MinCexProviderCount: 1,
				MinDexProviderCount: 1,
				Quotes: map[string]config.QuoteConfig{
					"BTC": {
						MinProviderVolume: -10,
						NormalizeByPair:   "",
					},
				},
			},
			true,
		},
		{
			"min provider liquidity is negative",
			config.GenerateConfig{
				Providers: map[string]config.ProviderConfig{
					"okx": {},
				},
				MinCexProviderCount: 1,
				MinDexProviderCount: 1,
				Quotes: map[string]config.QuoteConfig{
					"BTC": {
						MinProviderLiquidity: -10,
						NormalizeByPair:      "",
					},
				},
			},
			true,
		},
		{
			"min provider count > min provider count override",
			config.GenerateConfig{
				Providers: map[string]config.ProviderConfig{
					"okx": {},
				},
				MinCexProviderCount: 1,
				MinDexProviderCount: 1,
				Quotes: map[string]config.QuoteConfig{
					"BTC": {
						MinProviderVolume: 10,
						NormalizeByPair:   "",
					},
				},
				MinProviderCountOverride: 2,
			},

			true,
		},
		{
			"min provider count override < 1",
			config.GenerateConfig{
				Providers: map[string]config.ProviderConfig{
					"okx": {},
				},
				MinCexProviderCount:      1,
				MinDexProviderCount:      1,
				MinProviderCountOverride: 0,
				Quotes: map[string]config.QuoteConfig{
					"BTC": {
						MinProviderVolume: 10,
						NormalizeByPair:   "",
					},
				},
			},

			true,
		},
		{
			"invalid normalize by pair",
			config.GenerateConfig{
				Providers: map[string]config.ProviderConfig{
					"okx": {},
				},
				MinCexProviderCount: 1,
				MinDexProviderCount: 1,
				Quotes: map[string]config.QuoteConfig{
					"BTC": {
						MinProviderVolume: 10,
						NormalizeByPair:   "invalid",
					},
				},
			},
			true,
		},
		{
			name: "invalid AllowedCurrencyPair",
			cfg: config.GenerateConfig{
				Providers: map[string]config.ProviderConfig{
					"okx": {},
				},
				MinCexProviderCount: 1,
				MinDexProviderCount: 1,
				Quotes: map[string]config.QuoteConfig{
					"BTC": {
						MinProviderVolume: 10,
					},
				},
				AllowedCurrencyPairs: map[string]struct{}{
					"FOO BAR": {},
				},
			},
			expectedErr: true,
		},
		{
			"invalid market map override",
			config.GenerateConfig{
				Providers: map[string]config.ProviderConfig{
					"okx": {},
				},
				MinCexProviderCount: 1,
				MinDexProviderCount: 1,
				Quotes: map[string]config.QuoteConfig{
					"BTC": {
						MinProviderVolume: 10,
						NormalizeByPair:   "USDT/USD",
					},
				},
				MarketMapOverride: types.MarketMap{
					Markets: map[string]types.Market{
						"USD/FOO": {
							Ticker:          types.Ticker{},
							ProviderConfigs: nil,
						},
					},
				},
			},
			true,
		},
		{
			"invalid market map override with missing normalization - should pass",
			config.GenerateConfig{
				Providers: map[string]config.ProviderConfig{
					"okx": {},
				},
				MinCexProviderCount:      1,
				MinDexProviderCount:      1,
				MinProviderCountOverride: 1,
				Quotes: map[string]config.QuoteConfig{
					"BTC": {
						MinProviderVolume: 10,
						NormalizeByPair:   "USDT/USD",
					},
				},
				MarketMapOverride: types.MarketMap{
					Markets: map[string]types.Market{
						"USD/FOO": {
							Ticker: types.Ticker{
								CurrencyPair:     connecttypes.NewCurrencyPair("USD", "FOO"),
								Decimals:         20,
								MinProviderCount: 1,
							},
							ProviderConfigs: []types.ProviderConfig{
								{
									Name:            "test",
									OffChainTicker:  "test",
									NormalizeByPair: &connecttypes.CurrencyPair{Base: "NON", Quote: "EXISTENT"},
								},
							},
						},
					},
				},
			},
			false,
		},
		{
			"valid config",
			config.GenerateConfig{
				Providers: map[string]config.ProviderConfig{
					"okx": {},
				},
				MinCexProviderCount:      1,
				MinDexProviderCount:      1,
				MinProviderCountOverride: 1,
				Quotes: map[string]config.QuoteConfig{
					"BTC": {
						MinProviderVolume: 10,
						NormalizeByPair:   "",
					},
				},
			},
			false,
		},
		{
			"valid config with normalize by pair",
			config.GenerateConfig{
				Providers: map[string]config.ProviderConfig{
					"okx": {},
				},
				MinCexProviderCount:      1,
				MinDexProviderCount:      1,
				MinProviderCountOverride: 1,
				Quotes: map[string]config.QuoteConfig{
					"BTC": {
						MinProviderVolume: 10,
						NormalizeByPair:   "USDT/USD",
					},
				},
			},
			false,
		},
		{
			name: "invalid can't have both allowed and excluded pair configs set",
			cfg: config.GenerateConfig{
				Providers: map[string]config.ProviderConfig{
					"okx": {},
				},
				MinCexProviderCount: 1,
				MinDexProviderCount: 1,
				Quotes: map[string]config.QuoteConfig{
					"BTC": {
						MinProviderVolume: 10,
					},
				},
				AllowedCurrencyPairs: map[string]struct{}{"FOO/BAR": {}},
				ExcludeCurrencyPairs: map[string]struct{}{"FOO/BAR": {}},
			},
			expectedErr: true,
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.cfg.Validate()
			if tc.expectedErr && err == nil {
				t.Errorf("expected error, got nil")
			}
			if !tc.expectedErr && err != nil {
				t.Errorf("expected no error, got %v", err)
			}
		})
	}
}

func TestProviderConfig_Validate(t *testing.T) {
	tcs := []struct {
		name        string
		cfg         config.ProviderConfig
		expectedErr bool
	}{
		{
			name: "valid config - zero values",
			cfg: config.ProviderConfig{
				MinProviderVolume:    0,
				MinProviderLiquidity: 0,
			},
			expectedErr: false,
		},
		{
			name: "valid config - positive values",
			cfg: config.ProviderConfig{
				MinProviderVolume:    100,
				MinProviderLiquidity: 1000,
			},
			expectedErr: false,
		},
		{
			name: "invalid config - negative volume",
			cfg: config.ProviderConfig{
				MinProviderVolume:    -1,
				MinProviderLiquidity: 0,
			},
			expectedErr: true,
		},
		{
			name: "invalid config - negative liquidity",
			cfg: config.ProviderConfig{
				MinProviderVolume:    0,
				MinProviderLiquidity: -1,
			},
			expectedErr: true,
		},
		{
			name: "invalid config - both negative",
			cfg: config.ProviderConfig{
				MinProviderVolume:    -100,
				MinProviderLiquidity: -1000,
			},
			expectedErr: true,
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.cfg.Validate()
			if tc.expectedErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestGenerateConfig_IsCurrencyPairAllowed(t *testing.T) {
	tcs := []struct {
		name     string
		cfg      config.GenerateConfig
		pair     connecttypes.CurrencyPair
		expected bool
	}{
		{
			name:     "expect true when empty",
			cfg:      config.GenerateConfig{},
			pair:     connecttypes.NewCurrencyPair("FOO", "BAR"),
			expected: true,
		},
		{
			name: "expect true when present",
			cfg: config.GenerateConfig{
				AllowedCurrencyPairs: map[string]struct{}{
					"FOO/BAR": {},
				},
			},
			pair:     connecttypes.NewCurrencyPair("FOO", "BAR"),
			expected: true,
		},
		{
			name: "expect false when not present",
			cfg: config.GenerateConfig{
				AllowedCurrencyPairs: map[string]struct{}{
					"FOO/BAR": {},
				},
			},
			pair:     connecttypes.NewCurrencyPair("NOT", "PRESENT"),
			expected: false,
		},
	}
	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			allowed := tc.cfg.IsCurrencyPairAllowed(tc.pair)
			require.Equal(t, tc.expected, allowed, "expected %v, got %v", tc.expected, allowed)
		})
	}
}
