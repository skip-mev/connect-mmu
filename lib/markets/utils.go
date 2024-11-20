package markets

import (
	mmtypes "github.com/skip-mev/connect/v2/x/marketmap/types"
)

// FindRemovedMarkets finds all markets that exist in actual but are not in generated.
func FindRemovedMarkets(actual, generated mmtypes.MarketMap) []mmtypes.Market {
	// identify + return any markets that are not in actual, but are in generated
	var removed []mmtypes.Market
	for _, market := range actual.Markets {
		if _, ok := generated.Markets[market.Ticker.CurrencyPair.String()]; !ok {
			removed = append(removed, market)
		}
	}

	return removed
}

// FindIntersectionAndExclusion separates the actual markets into two groups:
// 1. The intersection of actual and generated markets
// 2. The Markets in generated, that are not in actual.
func FindIntersectionAndExclusion(
	actual mmtypes.MarketMap,
	upserts []mmtypes.Market,
) (intersection []mmtypes.Market, exclusion []mmtypes.Market) {
	for _, market := range upserts {
		if _, ok := actual.Markets[market.Ticker.CurrencyPair.String()]; ok {
			intersection = append(intersection, market)
		} else {
			exclusion = append(exclusion, market)
		}
	}

	return intersection, exclusion
}
