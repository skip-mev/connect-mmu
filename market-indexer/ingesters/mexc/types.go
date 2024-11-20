package mexc

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/skip-mev/connect-mmu/lib/symbols"
	"github.com/skip-mev/connect-mmu/store/provider"
)

// TickerData is the data payload returned from the mexc API on
// a Tickers request.
//
// Docs: https://mexcdevelop.github.io/apidocs/spot_v3_en/#24hr-ticker-price-change-statistics
//
// Ex.
// [
//
//	{
//	  "symbol": "BTCUSDT",
//	  "priceChange": "184.34",
//	  "priceChangePercent": "0.00400048",
//	  "prevClosePrice": "46079.37",
//	  "lastPrice": "46263.71",
//	  "bidPrice": "46260.38",
//	  "bidQty": "",
//	  "askPrice": "46260.41",
//	  "askQty": "",
//	  "openPrice": "46079.37",
//	  "highPrice": "47550.01",
//	  "lowPrice": "45555.5",
//	  "volume": "1732.461487",
//	  "quoteVolume": "923923.938",
//	  "openTime": 1641349500000,
//	  "closeTime": 1641349582808,
//	  "count": null
//	},
//
// ].
type TickerData struct {
	Symbol string `json:"symbol"`
	Volume string `json:"volume"`
	// QuoteVolume is the 24hr volume in terms of the quote.
	QuoteVolume string `json:"quoteVolume"`
	OpenPrice   string `json:"openPrice"`
}

func (td *TickerData) toProviderMarket() (provider.CreateProviderMarket, error) {
	base, quote, err := symbolToBaseQuote(td.Symbol)
	if err != nil {
		return provider.CreateProviderMarket{}, err
	}

	quoteVol, err := strconv.ParseFloat(td.QuoteVolume, 64)
	if err != nil {
		return provider.CreateProviderMarket{}, err
	}

	refPrice, err := strconv.ParseFloat(td.OpenPrice, 64)
	if err != nil {
		return provider.CreateProviderMarket{}, fmt.Errorf("failed to convert OpenPrice: %w", err)
	}

	targetBase, err := symbols.ToTickerString(base)
	if err != nil {
		return provider.CreateProviderMarket{}, err
	}
	targetQuote, err := symbols.ToTickerString(quote)
	if err != nil {
		return provider.CreateProviderMarket{}, err
	}

	pm := provider.CreateProviderMarket{
		Create: provider.CreateProviderMarketParams{
			TargetBase:     targetBase,
			TargetQuote:    targetQuote,
			OffChainTicker: td.Symbol,
			ProviderName:   ProviderName,
			QuoteVolume:    quoteVol,
			ReferencePrice: refPrice,
		},
	}

	return pm, pm.ValidateBasic()
}

// symbolToBaseQuote splits a ticker symbol into its base and quote components
// if the quote is a knownQuote.
func symbolToBaseQuote(symbol string) (string, string, error) {
	for _, knownQuote := range knownQuotes {
		base, cut := strings.CutSuffix(symbol, knownQuote)
		if cut {
			return base, knownQuote, nil
		}
	}

	return "", "", fmt.Errorf(`symbol "%s" does not have a known quote`, symbol)
}
