package update

import (
	mmtypes "github.com/skip-mev/connect/v2/x/marketmap/types"
	"go.uber.org/zap"
)

type Options struct {
	UpdateEnabled            bool
	OverwriteProviders       bool
	ExistingOnly             bool
	DisableDeFiMarketMerging bool
}

// CombineMarketMaps adds the given generated markets to the actual market.
// If the market in generated does not exist in actual, append the whole market.
// If the market in actual does not exist in generated, append the whole market.
// If the market exists in actual AND generated, only append to the provider configs.
func CombineMarketMaps(
	logger *zap.Logger,
	actual, generated mmtypes.MarketMap,
	options Options,
) (mmtypes.MarketMap, []string, error) {
	// allow for the input of fully empty market maps.  It is a valid case if the on-chain or generated market map is empty.
	if actual.Markets == nil {
		actual.Markets = make(map[string]mmtypes.Market)
	}

	if generated.Markets == nil {
		generated.Markets = make(map[string]mmtypes.Market)
	}

	combined := mmtypes.MarketMap{
		Markets: make(map[string]mmtypes.Market),
	}

	// update the enabled field of each market in the generated market-map
	for ticker, market := range generated.Markets {
		// generated market exists in the actual on chain map
		actualMarket, found := actual.Markets[ticker]
		if !found && options.ExistingOnly {
			// do not use markets that are not on chain if we only want to modify existing markets
			logger.Debug("not adding market because it is not in the actual market map",
				zap.String("ticker", ticker),
				zap.Bool("existing-only", options.ExistingOnly),
			)
			continue
		}

		if found {
			if actualMarket.Ticker.Enabled && !options.UpdateEnabled {
				// if the market is enabled, but we are NOT updating enabled, keep it the set to actual
				logger.Debug("not updating market because it is already in the actual market map",
					zap.String("ticker", ticker),
					zap.Bool("update-enabled", options.UpdateEnabled),
				)
				market = actualMarket
			} else {
				logger.Debug("updating market that is is already in the actual market map",
					zap.String("ticker", ticker),
					zap.Bool("update-enabled", options.UpdateEnabled),
				)

				market.Ticker.Enabled = actualMarket.Ticker.Enabled
				market.Ticker.MinProviderCount = actualMarket.Ticker.MinProviderCount
				market.Ticker.Decimals = actualMarket.Ticker.Decimals

				updatedProviderConfigs := market.ProviderConfigs
				if !options.OverwriteProviders {
					updatedProviderConfigs = appendToProviders(actualMarket, market)
				}
				market.ProviderConfigs = updatedProviderConfigs
			}
		} else {
			logger.Debug("adding generated market that is not in the actual market map",
				zap.String("ticker", ticker),
			)

			// if not found in the on chain marketmap, add, but disable
			market.Ticker.Enabled = false
		}
		combined.Markets[ticker] = market
	}

	// append remove markets that are in generated, but NOT actual, unless it is enabled
	removals := make([]string, 0)
	for ticker, market := range actual.Markets {
		if _, found := generated.Markets[ticker]; !found {
			if market.Ticker.Enabled {
				logger.Warn("Adding actual market that is not in the generated market map because it is enabled",
					zap.String("ticker", ticker),
				)
				combined.Markets[ticker] = market
			} else {
				removals = append(removals, ticker)
				logger.Debug("removing actual market that is not in the generated market map",
					zap.String("ticker", ticker),
				)
			}
		}
	}

	return combined, removals, nil
}

func appendToProviders(actual, generated mmtypes.Market) []mmtypes.ProviderConfig {
	// create map of configs by their provider name
	actualProviderConfigsMap := make(map[string]mmtypes.ProviderConfig)
	for _, config := range actual.ProviderConfigs {
		actualProviderConfigsMap[config.Name] = config
	}

	// only update to the ProviderConfigs when they are new
	appendedProviderConfigs := actual.ProviderConfigs
	for _, generatedProviderConfig := range generated.ProviderConfigs {
		if _, found := actualProviderConfigsMap[generatedProviderConfig.Name]; !found {
			// if the provider config is not in the actual set, add it
			appendedProviderConfigs = append(appendedProviderConfigs, generatedProviderConfig)
		}
	}

	return appendedProviderConfigs
}
