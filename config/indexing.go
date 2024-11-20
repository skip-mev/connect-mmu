package config

import (
	"fmt"
)

type MarketConfig struct {
	Ingesters           []IngesterConfig    `json:"ingesters" mapstructure:"ingesters"`
	CoinMarketCapConfig CoinMarketCapConfig `json:"coinmarketcap" mapstructure:"coinmarketcap"`
	RaydiumNodes        []RaydiumNodeConfig `json:"raydium" mapstructure:"raydium"`

	// GeckoNetworkDexPairs is a configuration for the Gecko Terminal ingester. This configures the ingester to
	// ingest data from the specified pairs. Note: not all pairs are valid. Please see ingesters/gecko/utils.go for valid pairs.
	GeckoNetworkDexPairs []GeckoNetworkDexPair `json:"gecko_network_dex_pairs" mapstructure:"gecko_network_dex_pairs"`
}

var defaultIngesters = []IngesterConfig{
	{Name: "coinbase"},
	{Name: "gecko"},
}

var defaultGeckoNetworkDexPairs = []GeckoNetworkDexPair{
	{Network: "eth", Dex: "uniswap_v3"},
}

func DefaultMarketConfig() MarketConfig {
	return MarketConfig{
		Ingesters:            defaultIngesters,
		CoinMarketCapConfig:  CoinMarketCapConfig{APIKey: ""},
		RaydiumNodes:         []RaydiumNodeConfig{},
		GeckoNetworkDexPairs: defaultGeckoNetworkDexPairs,
	}
}

type GeckoNetworkDexPair struct {
	Network string `json:"network" mapstructure:"network"`
	Dex     string `json:"dex" mapstructure:"dex"`
}

type IngesterConfig struct {
	Name string `json:"name"`
}

func (pc *IngesterConfig) Validate() error {
	if pc.Name == "" {
		return fmt.Errorf("name cannot be invalid")
	}

	return nil
}

type CoinMarketCapConfig struct {
	APIKey string `json:"api_key" mapstructure:"api_key"`
}

func (cc *CoinMarketCapConfig) Validate() error {
	if cc.APIKey == "" {
		return fmt.Errorf("coinmarketcap_api_key is required")
	}

	return nil
}

type RaydiumNodeConfig struct {
	Endpoint string `json:"endpoint" mapstructure:"endpoint"`
	NodeKey  string `json:"node_key" mapstructure:"node_key"`
}

func (rc *RaydiumNodeConfig) Validate() error {
	if rc.Endpoint == "" {
		return fmt.Errorf("endpoint cannot be invalid")
	}

	if rc.NodeKey == "" {
		return fmt.Errorf("raydium_node_key is required")
	}

	return nil
}

func (c *MarketConfig) Validate() error {
	if err := c.CoinMarketCapConfig.Validate(); err != nil {
		return err
	}

	seen := make(map[string]struct{})

	for _, ingester := range c.Ingesters {
		if err := ingester.Validate(); err != nil {
			return fmt.Errorf("ingester %s invalid: %w", ingester.Name, err)
		}

		if _, found := seen[ingester.Name]; found {
			return fmt.Errorf("duplicate ingester %s found", ingester.Name)
		}

		seen[ingester.Name] = struct{}{}

		// extra validation for specific ingesters
		switch ingester.Name {
		case "raydium":
			for _, rc := range c.RaydiumNodes {
				if err := rc.Validate(); err != nil {
					return fmt.Errorf("raydium config invalid: %w", err)
				}
			}
		default:
			// do nothing
		}
	}

	return nil
}
