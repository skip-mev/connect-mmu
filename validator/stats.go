package validator

import (
	"math"

	"gonum.org/v1/gonum/stat"

	"github.com/skip-mev/connect-mmu/validator/types"
)

// successRates returns a mapping for each provider's success rates.
func successRates(counts types.ProviderCounts) map[string]float64 {
	rates := make(map[string]float64)
	for provider, count := range counts {
		successPercentage := (float64(count.Success) / float64(count.Success+count.Failure)) * 100
		rates[provider] = successPercentage
	}
	return rates
}

// ZScores returns a mapping for each provider's zScores.
func zScores(counts types.ProviderCounts) map[string]float64 {
	scores := make(map[string]float64)

	prices := pricesFromCounts(counts)
	valuesMean := stat.Mean(prices, nil)
	sd := stat.StdDev(prices, nil)

	for provider, count := range counts {
		if count.Success > 0 {
			z := stat.StdScore(valuesMean, sd, count.AveragePrice)
			// NaN occurs when it's a lone provider.
			if math.IsNaN(z) {
				z = 0.0
			}
			scores[provider] = z
		}
	}

	return scores
}

// percentDifference returns the percentage difference between two prices.
// The result is the absolute percentage difference, always positive
// Formula: |price1 - price2| / ((price1 + price2) / 2) * 100
func percentDifference(p1, p2 float64) float64 {
	if p1 == 0 && p2 == 0 {
		return 0
	}

	// Calculate absolute difference
	diff := p1 - p2
	if diff < 0 {
		diff = -diff
	}

	// Calculate average of the two prices
	avg := (p1 + p2) / 2

	// Calculate percentage difference
	percentDiff := (diff / avg) * 100

	return percentDiff
}

// pricesFromCounts returns a list of reported prices from providers.
// if a provider had no successful reports / no average price, it is ignored.
func pricesFromCounts(counts types.ProviderCounts) []float64 {
	prices := make([]float64, 0)
	for _, count := range counts {
		if count.Success > 0 && count.AveragePrice != 0.0 {
			prices = append(prices, count.AveragePrice)
		}
	}
	return prices
}
