package validator

import (
	"context"
	"fmt"

	mmtypes "github.com/skip-mev/connect/v2/x/marketmap/types"

	"github.com/skip-mev/connect-mmu/validator/types"
)

type Validator struct {
	mm        mmtypes.MarketMap
	cmcAPIKey string
}

// New returns a new Validator. A CMC API key may be optionally passed in to generate reference price checks for
// markets that have a CMC ID in their metadata.
func New(mm mmtypes.MarketMap, opts ...Option) *Validator {
	v := &Validator{mm: mm}
	for _, opt := range opts {
		opt(v)
	}
	return v
}

// Report generates a report for the provider health. Specifically, for each market it will generate:
// - each provider's success rate
// - each provider's relative Z-Score
// - each provider's reference price difference, if reference prices were fetched (see options).
func (v *Validator) Report(ctx context.Context, health types.MarketHealth) ([]types.Report, error) {
	reports := make([]types.Report, 0, len(health))
	refPrices, err := v.referencePrices(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get reference prices: %w", err)
	}

	for ticker, counts := range health {
		report := types.Report{Ticker: ticker}
		refPrice, hasRefPrice := refPrices[ticker]
		if hasRefPrice {
			report.ReferencePrice = &refPrice
		}
		zScores := zScores(counts)
		successRates := successRates(counts)

		for provider, count := range counts {
			pReport := types.ProviderReport{
				Name:         provider,
				SuccessRate:  successRates[provider],
				ZScore:       zScores[provider],
				AveragePrice: count.AveragePrice,
			}
			if hasRefPrice {
				jitter := percentDifference(count.AveragePrice, refPrice)
				pReport.ReferencePriceDiff = &jitter
			}
			report.ProviderReports = append(report.ProviderReports, pReport)
		}
		reports = append(reports, report)
	}
	return reports, nil
}

// MissingReports will return any markets/providers that exist in the marketmap, but were not reported in the market health.
func (v *Validator) MissingReports(health types.MarketHealth) map[string][]string {
	missing := make(map[string][]string)
	for ticker, market := range v.mm.Markets {
		for _, provider := range market.ProviderConfigs {
			if _, ok := health[ticker][provider.Name]; !ok {
				missing[ticker] = append(missing[ticker], provider.Name)
			}
		}
	}
	return missing
}

// referencePrices returns reference prices.
func (v *Validator) referencePrices(ctx context.Context) (map[string]float64, error) {
	if v.cmcAPIKey == "" {
		return nil, nil
	}
	cmcIDs, err := getCMCIDMapping(v.mm)
	if err != nil {
		return nil, fmt.Errorf("failed to get cmc ID mapping")
	}
	prices, err := getReferencePrices(ctx, v.cmcAPIKey, cmcIDs)
	if err != nil {
		return nil, fmt.Errorf("error getting cmc prices: %w", err)
	}
	return prices, nil
}
