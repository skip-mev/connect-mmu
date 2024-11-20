package validator

import (
	"math"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/skip-mev/connect-mmu/validator/types"
)

// Helper function to create ProviderCounts
func createProviderCounts() types.ProviderCounts {
	return types.ProviderCounts{
		"provider1": &types.Counts{
			Success:      80,
			Failure:      20,
			AveragePrice: 100.0,
		},
		"provider2": &types.Counts{
			Success:      50,
			Failure:      50,
			AveragePrice: 200.0,
		},
		"provider3": &types.Counts{
			Success:      0,
			Failure:      0,
			AveragePrice: 0.0,
		},
	}
}

func TestSuccessRates(t *testing.T) {
	counts := createProviderCounts()
	expected := map[string]float64{
		"provider1": 80.0,
		"provider2": 50.0,
		"provider3": 0.0, // 0/0 handled as 0
	}

	result := successRates(counts)

	require.Len(t, result, len(expected))

	for provider, expRate := range expected {
		gotRate, exists := result[provider]
		require.True(t, exists)
		if math.Abs(gotRate-expRate) > 1e-6 {
			t.Errorf("provider %s: expected %f, got %f", provider, expRate, gotRate)
		}
	}
}

func TestPercentDifference(t *testing.T) {
	tests := []struct {
		p1, p2   float64
		expected float64
		testName string
	}{
		{100, 100, 0, "Same prices"},
		{100, 0, 200, "One price zero"},
		{0, 0, 0, "Both prices zero"},
		{150, 100, 40, "Positive difference"},
		{100, 150, 40, "Negative difference"},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			result := percentDifference(tt.p1, tt.p2)
			if math.Abs(result-tt.expected) > 1e-6 {
				t.Errorf("percentDifference(%f, %f) = %f; expected %f", tt.p1, tt.p2, result, tt.expected)
			}
		})
	}
}

func TestPricesFromCounts(t *testing.T) {
	counts := createProviderCounts()
	expected := []float64{100.0, 200.0} // provider3 is ignored

	result := pricesFromCounts(counts)

	require.Len(t, result, len(expected))

	priceMap := make(map[float64]bool)
	for _, price := range result {
		priceMap[price] = true
	}

	for _, expPrice := range expected {
		if !priceMap[expPrice] {
			t.Errorf("expected price %f in result", expPrice)
		}
	}
}
