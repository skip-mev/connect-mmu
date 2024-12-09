package config

import (
	"fmt"
	"math"
	"strings"

	connecttypes "github.com/skip-mev/connect/v2/pkg/types"
	"github.com/skip-mev/connect/v2/x/marketmap/types"
)

// ValidateProviderName checks if the name is valid.
func ValidateProviderName(n string) error {
	if n == "" {
		return fmt.Errorf("name cannot be empty")
	}

	return nil
}

// ProviderConfig contains all provider-specific configuration for generation.
// Ex. market filtering on a per-provider basis.
type ProviderConfig struct {
	// IsSupplemental indicates whether this provider is considered supplemental.
	// Supplemental providers are not counted towards the minimum number of
	// providers for a given data source due to potential reliability issues.
	IsSupplemental bool `json:"isSupplemental" mapstructure:"isSupplemental"`

	// RequireAggregateIDs ensures that the outputted market map only includes entries with aggregate IDs
	// for this provider.
	RequireAggregateIDs bool `json:"require_aggregate_ids" mapstructure:"require_aggregate_ids"`

	// Filters is a set of filters to apply to a market on a per-provider basis.
	Filters Filters `json:"filters" mapstructure:"filters"`

	// IgnoreLiquidity is a flag to ignore any filtering based on liquidity.
	IgnoreLiquidity bool `json:"ignore_liquidity" mapstructure:"ignore_liquidity"`

	// IgnoreVolume is a flag to ignore any filtering based on volume.
	IgnoreVolume bool `json:"ignore_volume" mapstructure:"ignore_volume"`

	// IsDefi is a flag that denotes that this provider is to be considered as a Defi venue.
	IsDefi bool `json:"is_defi" mapstructure:"is_defi"`

	// MinProviderVolume is the minimum volume threshold specific to this provider.
	// If set to 0, this threshold is ignored.
	MinProviderVolume float64 `json:"min_provider_volume" mapstructure:"min_provider_volume"`

	// MinProviderLiquidity is the minimum liquidity threshold specific to this provider.
	// If set to 0, this threshold is ignored.
	MinProviderLiquidity float64 `json:"min_provider_liquidity" mapstructure:"min_provider_liquidity"`
}

// Filters is a set of filters to apply to a market on a per-provider basis.
type Filters struct {
	// TopMarkets is the number of top markets to choose for this provider
	// based on a 24hr quote volume basis.  If the value is 0 (null value),
	// no filter will be applied.
	TopMarkets uint64 `json:"top_markets" mapstructure:"top_markets"`
}

// Validate checks if the ProviderConfig is valid.
func (pc *ProviderConfig) Validate() error {
	if pc.MinProviderVolume < 0 {
		return fmt.Errorf("min_provider_volume must be non-negative")
	}

	if pc.MinProviderLiquidity < 0 {
		return fmt.Errorf("min_provider_liquidity must be non-negative")
	}

	return nil
}

// QuoteConfig contains all quote-specific configuration for generation.
type QuoteConfig struct {
	// MinProviderVolume is the minimum volume per-provider for a market
	// with the given quote.
	MinProviderVolume float64 `json:"min_provider_volume" mapstructure:"min_provider_volume"`
	// MinProviderVolume is the minimum liquidity (on buy and sell side) per-provider for a market
	// denominated in USD.
	MinProviderLiquidity float64 `json:"min_provider_liquidity" mapstructure:"min_provider_liquidity"`
	// NormalizeByPair is the string representation of a connect currency pair
	// to be used to normalize this market. For example, Setting this to USDT/USD will convert USDT markets to be
	// in terms of USD.
	NormalizeByPair string `json:"normalize_by_pair" mapstructure:"normalize_by_pair"`
}

// Validate checks if the QuoteConfig is valid.
func (qc *QuoteConfig) Validate() error {
	if qc.MinProviderVolume < 0 {
		return fmt.Errorf("min_provider_volume must be non-negative")
	}

	if qc.MinProviderLiquidity < 0 {
		return fmt.Errorf("min_provider_liquidity must be non-negative")
	}

	if qc.NormalizeByPair != "" {
		if _, err := connecttypes.CurrencyPairFromString(qc.NormalizeByPair); err != nil {
			return fmt.Errorf("normalize_by_pair must be a valid currency pair: %w", err)
		}
	}

	return nil
}

// GenerateConfig contains all configuration for generating a market map from data input.
type GenerateConfig struct {
	// Providers is a map of provider name -> ProviderConfig
	Providers map[string]ProviderConfig `json:"providers" mapstructure:"providers"`
	Quotes    map[string]QuoteConfig    `json:"quotes" mapstructure:"quotes"`

	// MinCexProviderCount is the minimum number of cex providers needed to make a market valid.
	// For example, if this is set to 3, but only 2 providers reported data for the quote,
	// the data will be purged.
	// You must have MinProviderCountOverride <= MinCexProviderCount.
	MinCexProviderCount uint64 `json:"min_cex_provider_count" mapstructure:"min_cex_provider_count"`

	// MinDexProviderCount is the minimum number of dex providers needed to make a market valid.
	// For example, if this is set to 3, but only 2 providers reported data for the quote,
	// the data will be purged.
	// You must have MinProviderCountOverride <= MinDexProviderCount.
	MinDexProviderCount uint64 `json:"min_dex_provider_count" mapstructure:"min_dex_provider_count"`

	// DisableProviders specifies a list of providers that should not be allowed to provide for a market.
	// structure is map[market]providers.
	DisableProviders map[string][]string `json:"disable_providers" mapstructure:"disable_providers"`
	// ExcludeCurrencyPairs is a set of currency pairs to exclude when generating the market map.
	ExcludeCurrencyPairs map[string]struct{} `json:"exclude_pairs" mapstructure:"exclude_pairs"`
	// AllowedCurrencyPairs is a set of currency pairs to allow when generating the market map.
	AllowedCurrencyPairs map[string]struct{} `json:"allowed_currency_pairs" mapstructure:"allowed_currency_pairs"`
	// MarketMapOverride is a marketmap who's values will replace generated values. That is, if this marketmap specifies a USD/MOG configuration,
	// it will replace the generated marketmap's configuration of USD/MOG. This override will be applied in the FinalizeMarketMap function.
	MarketMapOverride types.MarketMap `json:"market_map_override" mapstructure:"market_map_override"`
	// EnableAll is a flag that designates whether all generated markets will have Enabled == true.
	// This can be used in Core markets where all markets will be enabled by default.
	EnableAll bool `json:"enable_all" mapstructure:"enable_all"`

	// MinProviderCountOverride is values which all markets will get for MinProviderCount.
	// This value will replace any configured MinProviderCount for all Markets.
	//
	// We require this value to be LTE the highest MinProviderCount for any Market in order to avoid producing markets
	// which would be unable to post prices ever.
	MinProviderCountOverride uint64 `json:"min_provider_count_override" mapstructure:"min_provider_count_override"`
}

var defaultProviders = map[string]ProviderConfig{
	"coinbase_ws":            {Filters: Filters{TopMarkets: 100}},
	"uniswapv3_api-ethereum": {Filters: Filters{TopMarkets: 50}, IgnoreLiquidity: true},
}

var defaultQuotes = map[string]QuoteConfig{
	"USD": {
		MinProviderVolume:    80000,
		MinProviderLiquidity: 1000,
	},
	"USDT": {
		MinProviderVolume:    80000,
		MinProviderLiquidity: 1000,
		NormalizeByPair:      "USDT/USD",
	},
	"BTC": {
		MinProviderVolume:    15,
		MinProviderLiquidity: 1000,
		NormalizeByPair:      "BTC/USD",
	},
	"ETH": {
		MinProviderVolume:    20,
		MinProviderLiquidity: 1000,
		NormalizeByPair:      "ETH/USD",
	},
	"WETH": {
		MinProviderVolume:    20,
		MinProviderLiquidity: 1000,
		NormalizeByPair:      "ETH/USD",
	},
	"SOL": {
		MinProviderVolume: 0,
		NormalizeByPair:   "SOL/USD",
	},
}

func DefaultGenerateConfig() GenerateConfig {
	return GenerateConfig{
		Providers:                defaultProviders,
		Quotes:                   defaultQuotes,
		MinCexProviderCount:      3,
		MinDexProviderCount:      1,
		DisableProviders:         map[string][]string{},
		ExcludeCurrencyPairs:     map[string]struct{}{},
		AllowedCurrencyPairs:     map[string]struct{}{},
		MarketMapOverride:        types.MarketMap{},
		EnableAll:                false,
		MinProviderCountOverride: 1,
	}
}

// Validate validates the GenerateConfig. Specifically, it checks that:
// - providers entries are valid and unique
// - target quote entries are valid
// - min provider volume for each quote is non-negative
// - min market volume for each quote is non-negative
// - min provider count is greater than zero
// - min volumes exist for each quote
// - min market-volume (per quote) >= min provider-volume * min-providers (per quote).
func (cfg *GenerateConfig) Validate() error {
	for name, providerCfg := range cfg.Providers {
		if err := ValidateProviderName(name); err != nil {
			return err
		}

		if err := providerCfg.Validate(); err != nil {
			return err
		}
	}

	// for each quote, min market volume must be greater than or equal to min provider volume * min providers
	for quote, quoteCfg := range cfg.Quotes {
		if quote == "" {
			return fmt.Errorf("quote cannot be empty")
		}

		if err := quoteCfg.Validate(); err != nil {
			return fmt.Errorf("invalid quote config for quote %q: %w", quote, err)
		}
	}

	if len(cfg.ExcludeCurrencyPairs) > 0 && len(cfg.AllowedCurrencyPairs) > 0 {
		return fmt.Errorf("invalid configuration: can only specify excluded currency pairs or allowed currency pairs")
	}

	for pair := range cfg.ExcludeCurrencyPairs {
		if _, err := connecttypes.CurrencyPairFromString(pair); err != nil {
			return fmt.Errorf("invalid currency pair %q in ExcludeCurrencyPairs: %w", pair, err)
		}
	}

	for pair := range cfg.AllowedCurrencyPairs {
		if _, err := connecttypes.CurrencyPairFromString(pair); err != nil {
			return fmt.Errorf("invalid currency pair %q in AllowedCurrencyPairs: %w", pair, err)
		}
	}

	if err := cfg.MarketMapOverride.ValidateBasic(); err != nil {
		if !strings.Contains(err.Error(), "pair for normalization") {
			return fmt.Errorf("invalid MarketMapOverride: %w", err)
		}
	}

	if cfg.MinProviderCountOverride < 1 {
		return fmt.Errorf(
			"invalid MinProviderCountOverride %d: must be GTE 1",
			cfg.MinProviderCountOverride,
		)
	}

	if cfg.MinCexProviderCount < 1 {
		return fmt.Errorf("min_cex_provider_count must be > 0, got %d", cfg.MinCexProviderCount)
	}

	if cfg.MinDexProviderCount < 1 {
		return fmt.Errorf("min_dex_provider_count must be > 0, got %d", cfg.MinDexProviderCount)
	}

	if cfg.MinProviderCountOverride > cfg.MinCexProviderCount || cfg.MinProviderCountOverride > cfg.MinDexProviderCount {
		return fmt.Errorf("invalid MinProviderCountOverride: must be less than %d, got %d", int(math.Min(float64(cfg.MinCexProviderCount), float64(cfg.MinDexProviderCount))), cfg.MinProviderCountOverride)
	}

	return nil
}

// IsCurrencyPairAllowed reports if a currency pair is allowed in the given configuration.
// Firstly, it checks if this pair is present in the "ExcludeCurrencyPairs" set.
// Then, if AllowedCurrencyPairs is not populated, the method will return true.
// Otherwise, it reports if the currency pair is present in the set of AllowedCurrencyPairs.
func (cfg *GenerateConfig) IsCurrencyPairAllowed(pair connecttypes.CurrencyPair) bool {
	// first check if this is an excluded pair.
	if _, excluded := cfg.ExcludeCurrencyPairs[pair.String()]; excluded {
		return false
	}
	// the pair wasn't excluded. check if there is an allowlist. if there is no allowlist, we can just return true.
	if len(cfg.AllowedCurrencyPairs) == 0 {
		return true
	}
	// there is an allowlist, return the result from the map accessor.
	_, exists := cfg.AllowedCurrencyPairs[pair.String()]
	return exists
}

// IsProviderDefi returns true iff
// - the provider exists
// - it is flagged as defi
func (cfg *GenerateConfig) IsProviderDefi(providerName string) bool {
	providerCfg, ok := cfg.Providers[providerName]
	if !ok {
		return false
	}

	return providerCfg.IsDefi
}
