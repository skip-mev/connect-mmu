package strategy

import (
	"fmt"

	mmtypes "github.com/skip-mev/connect/v2/x/marketmap/types"
	"go.uber.org/zap"
	"golang.org/x/exp/maps"

	"github.com/skip-mev/connect-mmu/errors"
)

// GetMarketMapUpserts returns the sequence of market-map updates required to translate actual (on chain) to generated.
// Specifically, for any markets for which actual.Markets[ticker] != generated.Markets[ticker], we'll return
// a MsgUpsertMarket setting actual.Markets[ticker] to generated.Markets[ticker].
func GetMarketMapUpserts(
	logger *zap.Logger,
	actual,
	generated mmtypes.MarketMap,
) ([]mmtypes.Market, error) {
	upserts := make([]mmtypes.Market, 0)

	var err error

	// we need to make a copy of actual, since we'll be modifying it
	actualCopy := mmtypes.MarketMap{
		Markets: maps.Clone(actual.Markets),
	}

	// short circuit if they're equal
	if actual.Equal(generated) {
		logger.Info("both markets are equal - returning")
		return upserts, nil
	}

	generated, err = PruneNormalizeByPairs(logger, generated)
	if err != nil {
		logger.Error("PruneNormalizeByPairs failed", zap.Error(err))
		return nil, err
	}

	// for each market in the generated market-map
	for ticker, market := range generated.Markets {
		// get the corresponding market in the actual market-map
		if actualMarket, ok := actualCopy.Markets[ticker]; ok {
			// if the market has not changed between the actual and generated market-maps, continue
			if market.Equal(actualMarket) {
				continue
			}
		}

		// otherwise, for all markets pointed to by a normalize-by-pair in the generated market, check if they are in the actual market
		// if they are not, add them to the upserts + to actual (so we don't add them again)
		for _, providerConfig := range market.ProviderConfigs {
			// if the market has a normalize-by-pair market, check if it exists in the actual market-map
			if normalizeByPair := providerConfig.NormalizeByPair; normalizeByPair != nil {
				if _, ok := actualCopy.Markets[normalizeByPair.String()]; !ok {
					// find the market in generated (if this does not exist, fail)
					if normalizeByPairMarket, ok := generated.Markets[normalizeByPair.String()]; ok {
						// adjust by market exists, add the adjust-by via an upsert + add to the
						// actual market-map
						upserts = append(upserts, normalizeByPairMarket)
						actualCopy.Markets[normalizeByPairMarket.Ticker.String()] = normalizeByPairMarket
					} else {
						logger.Error("market normalize-by pair not found in generated marketmap",
							zap.String("market", ticker), zap.String("normalize pair", normalizeByPair.String()))
						return nil, errors.NewMarketNotFoundError(
							fmt.Sprintf("market %s's normalize-by market %s not found in generated market-map", ticker, normalizeByPairMarket.String()),
						)
					}
				}
			}
		}

		// now that any of the necessary adjust-bys exist w/in the market-map, add the market to the upserts
		upserts = append(upserts, market)
		actualCopy.Markets[ticker] = market
	}

	// return all upserts + verify that the finalized market-map is valid
	if err := actualCopy.ValidateBasic(); err != nil {
		logger.Error("updated marketmap is invalid", zap.Error(err))
		return nil, errors.NewInvalidMarketMapError(fmt.Errorf("updated market-map is invalid: %w", err))
	}

	return upserts, nil
}

// PruneNormalizeByPairs removes any provider configs for enabled markets with providers with disabled normalized pairs from markets.
func PruneNormalizeByPairs(
	logger *zap.Logger,
	generated mmtypes.MarketMap,
) (mmtypes.MarketMap, error) {
	// make a copy of generated
	generatedCopy := mmtypes.MarketMap{
		Markets: maps.Clone(generated.Markets),
	}
	logger.Info("removing provider configs with disabled normalize by pairs")
	// remove any provider configs for enabled markets with providers with disabled adjust bys from markets
	for key, market := range generatedCopy.Markets {
		if market.Ticker.Enabled {
			var newProviderConfig []mmtypes.ProviderConfig
			for _, pc := range market.ProviderConfigs {
				if pc.NormalizeByPair != nil {
					norm, found := generatedCopy.Markets[pc.NormalizeByPair.String()]
					if !found {
						return mmtypes.MarketMap{}, fmt.Errorf("unable to find normalize for %s",
							pc.NormalizeByPair.String())
					}
					// only include enabled ticker that is a normalize by pair
					if norm.Ticker.Enabled {
						// include
						newProviderConfig = append(newProviderConfig, pc)
					} else {
						// exclude -> remove
						logger.Info("removing disabled provider configs",
							zap.String("market", market.Ticker.String()),
							zap.String("provider config", pc.NormalizeByPair.String()),
						)
					}
				} else {
					newProviderConfig = append(newProviderConfig, pc)
				}
			}
			// only add the market if it still has enough providers after pruning
			if uint64(len(newProviderConfig)) >= market.Ticker.MinProviderCount {
				market.ProviderConfigs = newProviderConfig
				generatedCopy.Markets[key] = market
			} else {
				delete(generatedCopy.Markets, key)
				logger.Debug("excluding market because it was pruned",
					zap.String("market", market.Ticker.String()),
					zap.Int("num providers", len(newProviderConfig)),
					zap.Uint64("required providers", market.Ticker.MinProviderCount),
				)
			}
		}
	}
	return generatedCopy, nil
}
