package kucoin

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/skip-mev/connect-mmu/lib/symbols"
	"github.com/skip-mev/connect-mmu/store/provider"
)

const delimiter = "-"

// TickersResponse is thr response returned from the KuCoin
// API to a Tickers request.
type TickersResponse struct {
	Code string     `json:"code"`
	Data TickerData `json:"data"`
}

type TickerData struct {
	Tickers []Ticker `json:"ticker"`
}

// Ticker is the data payload included in a TickersResponse.
//
// Docs: https://www.kucoin.com/docs/rest/spot-trading/market-data/get-all-tickers#http-request
//
// Ex.
//
//	{
//	 "time": 1602832092060,
//	 "ticker": [
//	   {
//	     "symbol": "BTC-USDT", // symbol
//	     "symbolName": "BTC-USDT", // Name of trading pairs, it would change after renaming
//	     "buy": "11328.9", // bestAsk
//	     "sell": "11329", // bestBid
//		  "bestBidSize": "0.1",
//		  "bestAskSize": "1",
//	     "changeRate": "-0.0055", // 24h change rate
//	     "changePrice": "-63.6", // 24h change price
//	     "high": "11610", // 24h highest price
//	     "low": "11200", // 24h lowest price
//	     "vol": "2282.70993217", // 24h volume，the aggregated trading volume in BTC
//	     "volValue": "25984946.157790431", // 24h total, the trading volume in quote currency of last 24 hours
//	     "last": "11328.9", // last price
//	     "averagePrice": "11360.66065903", // 24h average transaction price yesterday
//	     "takerFeeRate": "0.001", // Basic Taker Fee
//	     "makerFeeRate": "0.001", // Basic Maker Fee
//	     "takerCoefficient": "1", // Taker Fee Coefficient
//	     "makerCoefficient": "1" // Maker Fee Coefficient
//	   }
//	 ]
//	}.
type Ticker struct {
	Symbol     string `json:"symbol"`
	SymbolName string `json:"symbolName"`
	// VolValue is the 24hr volume, quote denominated.
	VolValue     string `json:"volValue"`
	AveragePrice string `json:"averagePrice"`
}

func (td *Ticker) toProviderMarket() (provider.CreateProviderMarket, error) {
	quoteVol, err := strconv.ParseFloat(td.VolValue, 64)
	if err != nil {
		return provider.CreateProviderMarket{}, err
	}

	base, quote, err := symbolToBaseQuote(td.SymbolName)
	if err != nil {
		return provider.CreateProviderMarket{}, err
	}

	refPrice, err := strconv.ParseFloat(td.AveragePrice, 64)
	if err != nil {
		if td.AveragePrice == "" {
			refPrice = 0
		} else {
			return provider.CreateProviderMarket{}, fmt.Errorf("failed to convert AveragePrice: %w", err)
		}
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

func symbolToBaseQuote(symbol string) (string, string, error) {
	split := strings.Split(symbol, delimiter)
	if len(split) != 2 {
		return "", "", fmt.Errorf("invalid symbol: %s", symbol)
	}

	return split[0], split[1], nil
}
