package types

type LiquidityInfo struct {
	// NegativeDepthTwo is negative 2% depth in USD.
	NegativeDepthTwo float64 `json:"negative_two_depth"`
	// PositiveDepthTwo is positive 2% depth in USD.
	PositiveDepthTwo float64 `json:"positive_two_depth"`
}

func (l *LiquidityInfo) TotalLiquidity() float64 {
	return l.NegativeDepthTwo + l.PositiveDepthTwo
}

func (l *LiquidityInfo) IsZero() bool {
	return l.TotalLiquidity() == 0
}

// IsSufficient returns true if both sides of liquidity are greater than or equal to the required amount.
func (l *LiquidityInfo) IsSufficient(required float64) bool {
	return l.NegativeDepthTwo >= required && l.PositiveDepthTwo >= required
}
