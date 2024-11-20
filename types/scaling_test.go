package types_test

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/skip-mev/connect-mmu/types"
)

func TestDecimalPlacesFromPrice(t *testing.T) {
	tests := []struct {
		name  string
		price float64
		want  uint64
	}{
		{
			name:  "zero will always return min decimal places",
			price: 0,
			want:  1,
		},
		{
			name:  "number with 2 digits should return 8",
			price: 10,
			want:  8,
		},
		{
			name:  "22 decimal places will return 31",
			price: 0.0000000000000000000001,
			want:  31,
		},
		{
			name:  "42 decimal places will return 36",
			price: 0.0000000000000000000000000000000000000000001,
			want:  36,
		},
		{
			name:  "2 decimal places will return 12",
			price: 0.001,
			want:  12,
		},
		{
			name:  "0 decimal places will return 9",
			price: 1,
			want:  9,
		},
		{
			name:  "bitcoin price",
			price: 66276.97,
			want:  5,
		},
		{
			name:  "eur usdt price",
			price: 1.1008,
			want:  9,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := types.DecimalPlacesFromPrice(big.NewFloat(tt.price))
			require.Equal(t, tt.want, got)
		})
	}
}

func TestScalePriceToUint64(t *testing.T) {
	tests := []struct {
		name  string
		price float64
		want  uint64
	}{
		{
			name:  "zero will always return 0",
			price: 0,
			want:  0,
		},
		{
			name:  "scale a decimal value",
			price: 0.001,
			want:  1000000000,
		},
		{
			name:  "scale a non-decimal value",
			price: 10,
			want:  1000000000,
		},
		{
			name:  "bitcoin price",
			price: 66276.97,
			want:  6627697000,
		},
		{
			name:  "eur usdt price",
			price: 1.1008,
			want:  1100800000,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := types.ScalePriceToUint64(big.NewFloat(tt.price))
			require.Equal(t, tt.want, got)
		})
	}
}
