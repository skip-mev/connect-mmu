package binance

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/skip-mev/connect-mmu/lib/symbols"
	"github.com/skip-mev/connect-mmu/store/provider"
)

// TickerData is the data payload returned from the Binance API
// for the Tickers API request.
//
// Docs: https://binance-docs.github.io/apidocs/spot/en/#24hr-ticker-price-change-statistics
//
// Ex.
//
// [
//
//	{
//	  "symbol": "BNBBTC",
//	  "priceChange": "-94.99999800",
//	  "priceChangePercent": "-95.960",
//	  "weightedAvgPrice": "0.29628482",
//	  "prevClosePrice": "0.10002000",
//	  "lastPrice": "4.00000200",
//	  "lastQty": "200.00000000",
//	  "bidPrice": "4.00000000",
//	  "bidQty": "100.00000000",
//	  "askPrice": "4.00000200",
//	  "askQty": "100.00000000",
//	  "openPrice": "99.00000000",
//	  "highPrice": "100.00000000",
//	  "lowPrice": "0.10000000",
//	  "volume": "8913.30000000",
//	  "quoteVolume": "15.30000000",
//	  "openTime": 1499783499040,
//	  "closeTime": 1499869899040,
//	  "firstId": 28385,   // First tradeId
//	  "lastId": 28460,    // Last tradeId
//	  "count": 76         // Trade count
//	}
//
// ].
type TickerData struct {
	Symbol    string `json:"symbol"`
	HighPrice string `json:"highPrice"`
	LowPrice  string `json:"lowPrice"`
	// Volume is the 24hr volume for the ticker.
	Volume string `json:"volume"`
	// QuoteVolume is the 24hr volume for the ticker (price * ticker)
	QuoteVolume string `json:"quoteVolume"`
	// LastPrice is the last price of base/quote.
	LastPrice string `json:"lastPrice"`
	FirstID   int    `json:"firstId"`
	LastID    int    `json:"lastId"`
}

func (d *TickerData) toProviderMarket() (provider.CreateProviderMarket, error) {
	quoteVolume, err := strconv.ParseFloat(d.QuoteVolume, 64)
	if err != nil {
		return provider.CreateProviderMarket{}, fmt.Errorf("failed to convert quoteVolume: %w", err)
	}

	lastPrice, err := strconv.ParseFloat(d.LastPrice, 64)
	if err != nil {
		return provider.CreateProviderMarket{}, fmt.Errorf("failed to convert lastPrice: %w", err)
	}

	base, quote, err := symbolToBaseQuote(d.Symbol)
	if err != nil {
		return provider.CreateProviderMarket{}, err
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
			OffChainTicker: d.Symbol,
			ProviderName:   ProviderName,
			QuoteVolume:    quoteVolume,
			ReferencePrice: lastPrice,
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
