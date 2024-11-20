package gecko

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCommaSeparate(t *testing.T) {
	testCases := []struct {
		name     string
		input    []string
		expected string
	}{
		{
			name:     "Empty slice",
			input:    []string{},
			expected: "",
		},
		{
			name:     "Single token",
			input:    []string{"0x6B175474E89094C44Da98b954EedeAC495271d0F"},
			expected: "0x6B175474E89094C44Da98b954EedeAC495271d0F",
		},
		{
			name:     "Two tokens",
			input:    []string{"0xA0b86991c6218b36c1d19D4a2e9Eb0cE3606eB48", "0xdAC17F958D2ee523a2206206994597C13D831ec7"},
			expected: "0xA0b86991c6218b36c1d19D4a2e9Eb0cE3606eB48,0xdAC17F958D2ee523a2206206994597C13D831ec7",
		},
		{
			name:     "Multiple tokens",
			input:    []string{"0x6B175474E89094C44Da98b954EedeAC495271d0F", "0xA0b86991c6218b36c1d19D4a2e9Eb0cE3606eB48", "0xdAC17F958D2ee523a2206206994597C13D831ec7"},
			expected: "0x6B175474E89094C44Da98b954EedeAC495271d0F,0xA0b86991c6218b36c1d19D4a2e9Eb0cE3606eB48,0xdAC17F958D2ee523a2206206994597C13D831ec7",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := CommaSeparate(tc.input)
			require.Equal(t, tc.expected, result)
		})
	}
}
