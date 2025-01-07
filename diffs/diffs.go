package diffs

import (
	"encoding/json"
	"fmt"

	mmtypes "github.com/skip-mev/connect/v2/x/marketmap/types"

	"github.com/skip-mev/connect-mmu/generator/types"
	"github.com/skip-mev/connect-mmu/lib/file"
)

func WriteExclusionReasonsToFile(filePath string, exclusionReasons types.ExclusionReasons) error {
	bz, err := json.MarshalIndent(exclusionReasons, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal exclusion reasons: %w", err)
	}

	return file.WriteBytesToFile(filePath, bz)
}

// FilterMarketUpdates identifies all fields in the updatedMarket.Ticker and updatedMarket.ProviderConfigs
// that are different from the corresponding fields in the currentMarket, and zeros (sets to default value)
// all fields in updatedMarket that are the same as currentMarket.
func FilterMarketUpdates(currentMarket, updatedMarket mmtypes.Market) mmtypes.Market {
	filteredMarket := mmtypes.Market{
		Ticker: mmtypes.Ticker{
			CurrencyPair: currentMarket.Ticker.CurrencyPair,
		},
		ProviderConfigs: make([]mmtypes.ProviderConfig, 0),
	}

	if currentMarket.Ticker.Decimals != updatedMarket.Ticker.Decimals {
		filteredMarket.Ticker.Decimals = updatedMarket.Ticker.Decimals
	}
	if currentMarket.Ticker.MinProviderCount != updatedMarket.Ticker.MinProviderCount {
		filteredMarket.Ticker.MinProviderCount = updatedMarket.Ticker.MinProviderCount
	}
	if currentMarket.Ticker.Metadata_JSON != updatedMarket.Ticker.Metadata_JSON {
		filteredMarket.Ticker.Metadata_JSON = updatedMarket.Ticker.Metadata_JSON
	}
	if currentMarket.Ticker.Enabled != updatedMarket.Ticker.Enabled {
		filteredMarket.Ticker.Enabled = updatedMarket.Ticker.Enabled
	}

	currentProviderConfigs := map[string]mmtypes.ProviderConfig{}
	for _, cfg := range currentMarket.ProviderConfigs {
		currentProviderConfigs[cfg.Name] = cfg
	}

	// Compare and filter ProviderConfigs
	for _, updatedConfig := range updatedMarket.ProviderConfigs {
		currentConfig, exists := currentProviderConfigs[updatedConfig.Name]
		if !exists {
			// If the provider doesn't exist in the current market, include it in the filtered market
			filteredMarket.ProviderConfigs = append(filteredMarket.ProviderConfigs, updatedConfig)
			continue
		}

		filteredConfig := mmtypes.ProviderConfig{}
		filteredConfig.Name = updatedConfig.Name

		if currentConfig.OffChainTicker != updatedConfig.OffChainTicker {
			filteredConfig.OffChainTicker = updatedConfig.OffChainTicker
		}

		// if the NormalizeByPair is non-nil for both, check equality, otherwise set to non-nil
		if currentConfig.NormalizeByPair != nil && updatedConfig.NormalizeByPair != nil {
			if !currentConfig.NormalizeByPair.Equal(*updatedConfig.NormalizeByPair) {
				filteredConfig.NormalizeByPair = updatedConfig.NormalizeByPair
			}
		} else if currentConfig.NormalizeByPair == nil && updatedConfig.NormalizeByPair != nil {
			filteredConfig.NormalizeByPair = updatedConfig.NormalizeByPair
		}

		if currentConfig.Invert != updatedConfig.Invert {
			filteredConfig.Invert = updatedConfig.Invert
		}

		if currentConfig.Metadata_JSON != updatedConfig.Metadata_JSON {
			filteredConfig.Metadata_JSON = updatedConfig.Metadata_JSON
		}

		if filteredConfig != (mmtypes.ProviderConfig{
			Name: currentConfig.Name,
		}) {
			filteredMarket.ProviderConfigs = append(filteredMarket.ProviderConfigs, filteredConfig)
		}
	}

	return filteredMarket
}
