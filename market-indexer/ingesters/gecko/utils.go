package gecko

import (
	"fmt"
	"math"
	"strings"

	"golang.org/x/exp/maps"

	"github.com/skip-mev/connect-mmu/config"
)

// getAfterUnderscore gets all characters that come after the first underscore.
// if no underscores are found, this function will return the original string.
func getAfterUnderscore(s string) string {
	str := strings.SplitAfterN(s, "_", 2)
	if len(str) < 2 {
		return s
	}
	return str[1]
}

func validatePairs(pairs []config.GeckoNetworkDexPair) error {
	if len(pairs) == 0 {
		return fmt.Errorf("no pairs specified")
	}
	for _, pair := range pairs {
		if _, exist := validPairs[pair]; !exist {
			return fmt.Errorf("invalid pair: %v: must be one of %v", pair, maps.Keys(validPairs))
		}
	}
	return nil
}

const (
	ProviderNameUniswapEth  = "uniswapv3_api-ethereum"
	ProviderNameUniswapBase = "uniswapv3_api-base"

	TickerVenueUniswapEth  = "UNISWAP_V3"
	TickerVenueUniswapBase = "UNISWAP_V3_BASE"

	GeckoVenueUniswapEth  = "uniswap_v3"
	GeckoVenueUniswapBase = "uniswap-v3-base"
)

func geckoDexToConnectDex(dex string) string {
	switch dex {
	case GeckoVenueUniswapEth:
		return ProviderNameUniswapEth
	case GeckoVenueUniswapBase:
		return ProviderNameUniswapBase
	default:
		return dex
	}
}

func geckoDexToConnectTickerVenue(dex string) string {
	switch dex {
	case GeckoVenueUniswapEth:
		return TickerVenueUniswapEth
	case GeckoVenueUniswapBase:
		return TickerVenueUniswapBase
	default:
		return dex
	}
}

func isValidFloat64(f float64) bool {
	f = math.Abs(f)
	if math.IsInf(f, 1) || f == 0.0 {
		return false
	}
	return true
}

var validPairs = map[config.GeckoNetworkDexPair]struct{}{
	{Network: "eth", Dex: GeckoVenueUniswapEth}:   {},
	{Network: "base", Dex: GeckoVenueUniswapBase}: {},
}
