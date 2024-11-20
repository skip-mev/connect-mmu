package gecko

import (
	"math"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/skip-mev/connect-mmu/config"
)

func TestGetAfterUnderscore(t *testing.T) {
	testCases := []struct {
		input    string
		expected string
	}{
		{"eth_0x88e6a0c2ddd26feeb64f039a2c41296fcb3f5640", "0x88e6a0c2ddd26feeb64f039a2c41296fcb3f5640"},
		{"bsc_0x55d398326f99059ff775485246999027b3197955", "0x55d398326f99059ff775485246999027b3197955"},
		{"polygon_0x7ceb23fd6bc0add59e62ac25578270cff1b9f619", "0x7ceb23fd6bc0add59e62ac25578270cff1b9f619"},
		{"arbitrum_0x82af49447d8a07e3bd95bd0d56f35241523fbab1", "0x82af49447d8a07e3bd95bd0d56f35241523fbab1"},
		{"optimism_0x4200000000000000000000000000000000000006", "0x4200000000000000000000000000000000000006"},
		{"nounderscore", "nounderscore"},
		{"_0x1234567890abcdef", "0x1234567890abcdef"},
		{"multiple_under_scores", "under_scores"},
		{"", ""},
	}

	for _, tc := range testCases {
		result := getAfterUnderscore(tc.input)
		require.Equal(t, tc.expected, result)
	}
}

func TestValidatePairs(t *testing.T) {
	tests := []struct {
		name   string
		pairs  []config.GeckoNetworkDexPair
		errMsg string
	}{
		{
			name:   "Empty pairs",
			pairs:  nil,
			errMsg: "no pairs specified",
		},
		{
			name: "Valid pairs",
			pairs: []config.GeckoNetworkDexPair{
				{Network: "eth", Dex: "uniswap_v3"},
			},
		},
		{
			name: "Invalid pair",
			pairs: []config.GeckoNetworkDexPair{
				{Network: "btc", Dex: "pancakeswap"},
			},
			errMsg: "invalid pair: {btc pancakeswap}",
		},
		{
			name: "Mixed valid and invalid pairs",
			pairs: []config.GeckoNetworkDexPair{
				{Network: "eth", Dex: "uniswap_v3"},
				{Network: "btc", Dex: "pancakeswap"},
			},
			errMsg: "invalid pair: {btc pancakeswap}",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validatePairs(tt.pairs)
			if tt.errMsg != "" {
				require.Error(t, err)
				require.ErrorContains(t, err, tt.errMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestIsValidFloat64(t *testing.T) {
	tests := []struct {
		name  string
		input float64
		want  bool
	}{
		{"Positive number", 1.23, true},
		{"Negative number", -4.56, true},
		{"Zero", 0.0, false},
		{"Positive infinity", math.Inf(1), false},
		{"Negative infinity", math.Inf(-1), false},
		{"Very large positive number", math.MaxFloat64, true},
		{"Very large negative number", -math.MaxFloat64, true},
		{"Very small positive number", math.SmallestNonzeroFloat64, true},
		{"Very small negative number", -math.SmallestNonzeroFloat64, true},
		{"NaN", math.NaN(), true}, // Note: NaN is considered valid in this function
		{"Epsilon", math.Nextafter(1, 2) - 1, true},
		{"Negative Epsilon", -(math.Nextafter(1, 2) - 1), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isValidFloat64(tt.input); got != tt.want {
				t.Errorf("isValidFloat64(%v) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}
