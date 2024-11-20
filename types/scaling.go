package types

import (
	"math"
	"math/big"
)

// These values determine the range of decimals we are allowed to assign a market ticker.
//
// [ 4, 36 ].
const (
	minDecimals = 1
	maxDecimals = 36
)

// scalingFactorFromPrice returns a factor to scale a price value by given the price using
// the following formula:
//
// scale =  ceiling(9- log10(price))
//
//	This value is used so that a floating point price can be represented as a scaled integer.
func scalingFactorFromPrice(price float64) float64 {
	return math.Ceil(9 - math.Log10(price))
}

// ScalePriceToUint64 scales a price to a uint64 using the scaling factor
// obtained in ScalingFactorFromPrice.
func ScalePriceToUint64(price *big.Float) uint64 {
	decimalPlaces := DecimalPlacesFromPrice(price)
	decimalPlacesFloat := big.NewFloat(math.Pow10(int(decimalPlaces))) //nolint:gosec
	scaled := price.Mul(price, decimalPlacesFloat)

	// reduce precision to reduce jitter between runs when we calc ref prices
	// default precision is 53
	// reducing precision rounds the numbers, making them more close to work with
	reducePrecisionScaled := scaled.SetPrec(30)
	u, _ := reducePrecisionScaled.Uint64()

	return u
}

// DecimalPlacesFromPrice returns the decimal places a floating point price will be scaled
// by to be represented as an integer.  This value is obtained by scalingFactorFromPrice,
// but is also clamped between [minDecimals, maxDecimals].
func DecimalPlacesFromPrice(price *big.Float) uint64 {
	f, _ := price.Float64()

	scalingFactor := scalingFactorFromPrice(f)
	if f == 0.0 {
		return minDecimals
	}

	return min(max(uint64(scalingFactor), minDecimals), maxDecimals)
}
