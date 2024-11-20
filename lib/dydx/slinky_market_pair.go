package dydx

import (
	"fmt"
	"strings"

	"github.com/skip-mev/connect/v2/pkg/types"
)

// MarketPairToCurrencyPair converts a base and quote pair from MarketPrice format (for example BTC-ETH)
// to a currency pair type. Returns an error if unable to convert.
func MarketPairToCurrencyPair(marketPair string) (types.CurrencyPair, error) {
	split := strings.Split(marketPair, "-")
	if len(split) != 2 {
		return types.CurrencyPair{}, fmt.Errorf("incorrectly formatted CurrencyPair: %s", marketPair)
	}
	cp := types.CurrencyPair{
		Base:  strings.ToUpper(split[0]),
		Quote: strings.ToUpper(split[1]),
	}

	return cp, cp.ValidateBasic()
}
