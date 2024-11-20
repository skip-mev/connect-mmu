package validate

import (
	"fmt"
	"slices"

	connecttypes "github.com/skip-mev/connect/v2/pkg/types"
	mmtypes "github.com/skip-mev/connect/v2/x/marketmap/types"
)

// ApplyOptionsToMarketMap applies the options to the marketmap.
//
// if enable markets is specified, we enable those markets, and delete any that are disabled.
//
// if no enableMarkets are specified, we fall back to enableAll, which will enable all markets.
//
// if there are no markets specified, and enableMarkets is false, we just delete all disabled markets.
func ApplyOptionsToMarketMap(mm mmtypes.MarketMap, enableAll bool, enableOnly []string, enableMarkets []string) error {
	if len(enableOnly) > 0 && len(enableMarkets) > 0 {
		return fmt.Errorf("cannot specify both enableMarkets and enableAll at the same time")
	}

	if len(enableMarkets) > 0 {
		for _, enableMarket := range enableMarkets {
			if _, err := connecttypes.CurrencyPairFromString(enableMarket); err != nil {
				return fmt.Errorf("cannot enable market: invalid ticker %q: %w", enableMarket, err)
			}
		}
		for ticker, market := range mm.Markets {
			if slices.Contains(enableMarkets, ticker) {
				market.Ticker.Enabled = true
				mm.Markets[ticker] = market
			} else if !market.Ticker.Enabled {
				delete(mm.Markets, ticker)
			}
		}
		return nil
	}

	if len(enableOnly) > 0 {
		for _, enableMarket := range enableOnly {
			if _, err := connecttypes.CurrencyPairFromString(enableMarket); err != nil {
				return fmt.Errorf("cannot enable market: invalid ticker %q: %w", enableMarket, err)
			}
		}
		for ticker, market := range mm.Markets {
			if slices.Contains(enableOnly, ticker) {
				market.Ticker.Enabled = true
				mm.Markets[ticker] = market
			} else {
				delete(mm.Markets, ticker)
			}
		}
		return nil
	}

	if enableAll {
		for cp, market := range mm.Markets {
			market.Ticker.Enabled = true
			market.Ticker.MinProviderCount = 1
			mm.Markets[cp] = market
		}
	} else {
		for cp, market := range mm.Markets {
			if !market.Ticker.Enabled {
				delete(mm.Markets, cp)
			}
		}
	}

	return nil
}
