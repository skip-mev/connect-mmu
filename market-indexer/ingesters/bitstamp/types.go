package bitstamp

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/skip-mev/connect-mmu/lib/symbols"
	"github.com/skip-mev/connect-mmu/store/provider"
)

const delimiter = "/"

// TickerData is the data payload returned from the Bitstamp API
// in response to a Tickers request.
//
// Docs: https://www.bitstamp.net/api/#tag/Tickers/operation/GetCurrencyPairTickers
//
// Ex.
// [
//
//	{
//	  "ask": "2211.00",
//	  "bid": "2188.97",
//	  "high": "2811.00",
//	  "last": "2211.00",
//	  "low": "2188.97",
//	  "open": "2211.00",
//	  "open_24": "2211.00",
//	  "pair": "BTC/USD",
//	  "percent_change_24": "13.57",
//	  "side": "0",
//	  "timestamp": "1643640186",
//	  "volume": "213.26801100",
//	  "vwap": "2189.80"
//	}
//
// ].
type TickerData struct {
	High      string `json:"high"`
	Low       string `json:"low"`
	Pair      string `json:"pair"`
	OpenPrice string `json:"open"`
	// Volume is the 24hr volume denominated in base asset
	Volume string `json:"volume"`
}

func symbolToBaseQuote(symbol string) (string, string, error) {
	splitSymbol := strings.Split(symbol, delimiter)
	if len(splitSymbol) != 2 {
		return "", "", fmt.Errorf("symbol %s is not valid", symbol)
	}

	return splitSymbol[0], splitSymbol[1], nil
}

func (td *TickerData) toProviderMarket() (provider.CreateProviderMarket, error) {
	baseVol, err := strconv.ParseFloat(td.Volume, 64)
	if err != nil {
		return provider.CreateProviderMarket{}, err
	}

	high, err := strconv.ParseFloat(td.High, 64)
	if err != nil {
		return provider.CreateProviderMarket{}, err
	}

	low, err := strconv.ParseFloat(td.Low, 64)
	if err != nil {
		return provider.CreateProviderMarket{}, err
	}

	avg := (high + low) / 2
	quoteVol := baseVol * avg

	base, quote, err := symbolToBaseQuote(td.Pair)
	if err != nil {
		return provider.CreateProviderMarket{}, err
	}

	refPrice, err := strconv.ParseFloat(td.OpenPrice, 64)
	if err != nil {
		return provider.CreateProviderMarket{}, fmt.Errorf("failed to convert open: %w", err)
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
			OffChainTicker: td.Pair,
			ProviderName:   ProviderName,
			QuoteVolume:    quoteVol,
			ReferencePrice: refPrice,
		},
	}

	return pm, pm.ValidateBasic()
}
