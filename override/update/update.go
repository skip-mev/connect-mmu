package update

import (
	"encoding/json"
	"fmt"
	"regexp"
	"slices"
	"strings"

	types2 "github.com/skip-mev/connect/v2/pkg/types"
	mmtypes "github.com/skip-mev/connect/v2/x/marketmap/types"
	"github.com/skip-mev/connect/v2/x/marketmap/types/tickermetadata"
	"go.uber.org/zap"

	"github.com/skip-mev/connect-mmu/generator/types"
)

type Options struct {
	UpdateEnabled      bool
	OverwriteProviders bool
	ExistingOnly       bool
}

// CombineMarketMaps adds the given generated markets to the actual market.
// If the market in generated does not exist in actual, append the whole market.
// If the market in actual does not exist in generated, append the whole market.
// If the market exists in actual AND generated, only append to the provider configs.
func CombineMarketMaps(
	logger *zap.Logger,
	actual, generated mmtypes.MarketMap,
	options Options,
) (mmtypes.MarketMap, error) {
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

	// append add markets that are in actual, but NOT generated
	for ticker, market := range actual.Markets {
		if _, found := generated.Markets[ticker]; !found {
			logger.Debug("adding actual market that is not in the generated market map",
				zap.String("ticker", ticker),
			)
			combined.Markets[ticker] = market
		}
	}

	cmcIDToTickers, err := getCMCTickerMapping(combined)
	if err != nil {
		return mmtypes.MarketMap{}, err
	}
	merged, err := mergeCMCMIDMarkets(combined, cmcIDToTickers)
	if err != nil {
		return mmtypes.MarketMap{}, err
	}

	return merged, nil
}

var (
	defiTickerMatcher = regexp.MustCompile(`^[A-Z]+,[^/]+/USD$`)
)

// mergeCMCMIDMarkets merges markets that have the same CMC ID in their ticker metadata.
func mergeCMCMIDMarkets(mm mmtypes.MarketMap, cmcIDToTickers map[string][]string) (mmtypes.MarketMap, error) {
	for _, tickers := range cmcIDToTickers {
		// if there is only one ticker for this ID, we don't need to do anything.
		if len(tickers) <= 1 {
			continue
		}

		// keep track of the defi tickers we've seen, as well as the shortest ticker.
		// we do this so we can know if we:
		// 1. need to merge a defi set into a single set (i.e uniswap + raydium into one market)
		// 2. need to merge a mixed set into a single set (i.e. uniswap, coinbase, etc into one market).
		defiTickers := 0
		shortestTicker := tickers[0]
		for _, ticker := range tickers {
			if defiTickerMatcher.MatchString(ticker) {
				defiTickers++
			}
			if len(ticker) < len(shortestTicker) {
				shortestTicker = ticker
			}
		}

		// if all tickers are defi tickers, we need to
		// deconstruct the ticker and merge all markets into the deconstructed ticker.
		if defiTickers == len(tickers) {
			mergeMarketTicker := tickers[0] // we can just choose the first market to be the merger.
			deconstructedTicker, err := deconstructDeFiTicker(mergeMarketTicker)
			if err != nil {
				return mmtypes.MarketMap{}, fmt.Errorf("failed to deconstruct defi ticker: %w", err)
			}
			// check to make sure the deconstructed ticker doesn't already exist.
			if _, ok := mm.Markets[deconstructedTicker.String()]; ok {
				return mmtypes.MarketMap{}, fmt.Errorf("duplicate tickers found while attemtping to match CMC ID's: %q", deconstructedTicker.String())
			}
			newMarket := mm.Markets[mergeMarketTicker]
			newMarket.Ticker.CurrencyPair = deconstructedTicker
			// append the provider configs, and then remove the market from the map.
			for i := 1; i < len(tickers); i++ {
				newMarket.ProviderConfigs = appendIfNotExists(newMarket.ProviderConfigs, mm.Markets[tickers[i]].ProviderConfigs)
				delete(mm.Markets, tickers[i])
			}
			// set this new market into the map.
			mm.Markets[deconstructedTicker.String()] = newMarket
			delete(mm.Markets, tickers[0]) // remove the original defi market.
		} else {
			// otherwise, just take the shortest ticker we saw, and merge the others into it.
			consolidatedMarket := mm.Markets[shortestTicker]
			for _, ticker := range tickers {
				if ticker == shortestTicker {
					continue
				}
				otherMarket := mm.Markets[ticker]
				consolidatedMarket.ProviderConfigs = appendIfNotExists(consolidatedMarket.ProviderConfigs, otherMarket.ProviderConfigs)
				delete(mm.Markets, ticker)
			}
			mm.Markets[shortestTicker] = consolidatedMarket
		}
	}
	return mm, nil
}

// appendIfNotExists appends the config in newConfigs if it does not exist in src.
func appendIfNotExists(src []mmtypes.ProviderConfig, newConfigs []mmtypes.ProviderConfig) []mmtypes.ProviderConfig {
	appendedCfgs := make([]mmtypes.ProviderConfig, len(src))
	copy(appendedCfgs, src)
	for _, newConfig := range newConfigs {
		if !slices.ContainsFunc(src, func(config mmtypes.ProviderConfig) bool {
			return config.Name == newConfig.Name
		}) {
			appendedCfgs = append(appendedCfgs, newConfig)
		}
	}
	return appendedCfgs
}

func deconstructDeFiTicker(ticker string) (types2.CurrencyPair, error) {
	baseQuoteSplit := strings.Split(ticker, "/")
	if len(baseQuoteSplit) != 2 {
		return types2.CurrencyPair{}, fmt.Errorf("ticker %q is not valid defi ticker format (BASE,VENUE,ADDRESS/QUOTE)", ticker)
	}
	quote := baseQuoteSplit[1]

	baseVenueAddressSplit := strings.Split(baseQuoteSplit[0], ",")
	if len(baseVenueAddressSplit) != 3 {
		return types2.CurrencyPair{}, fmt.Errorf("base ticker %q is not valid defi ticker format (BASE,VENUE,ADDRESS)", ticker)
	}
	base := baseVenueAddressSplit[0]
	return types2.CurrencyPairFromString(base + "/" + quote)
}

func getCMCTickerMapping(mm mmtypes.MarketMap) (map[string][]string, error) {
	cmcIDToTickers := make(map[string][]string)
	for ticker, market := range mm.Markets {
		if market.Ticker.Metadata_JSON != "" {
			var md tickermetadata.CoreMetadata
			if err := json.Unmarshal([]byte(market.Ticker.Metadata_JSON), &md); err != nil {
				return nil, fmt.Errorf("failed to unmarshal market metadata for %q: %w", ticker, err)
			}
			for _, aggID := range md.AggregateIDs {
				if aggID.Venue == types.VenueCoinMarketcap {
					cmcIDToTickers[aggID.ID] = append(cmcIDToTickers[aggID.ID], ticker)
					break
				}
			}
		}
	}
	return cmcIDToTickers, nil
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
