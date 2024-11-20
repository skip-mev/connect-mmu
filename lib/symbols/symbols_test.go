package symbols_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/skip-mev/connect-mmu/lib/symbols"
)

func TestToTickerString(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
		expErr   bool
	}{
		{
			name:     "basic",
			input:    "BTC",
			expected: "BTC",
		},
		{
			name:     "dollar sign right",
			input:    "$BTC",
			expected: "BTC",
		},
		{
			name:     "dollar sign left",
			input:    "$BTC$",
			expected: "BTC",
		},
		{
			name:     "white space 1",
			input:    "    BTC   ",
			expected: "BTC",
		},
		{
			name:     "white space 2 ",
			input:    "    B         T C   ",
			expected: "B         T C",
		},
		{
			name:   "empty",
			input:  "",
			expErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := symbols.ToTickerString(tt.input)
			if tt.expErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tt.expected, got)
		})
	}
}
