package huobi

import (
	"fmt"
	"strings"

	"github.com/skip-mev/connect-mmu/lib/symbols"
	"github.com/skip-mev/connect-mmu/store/provider"
)

const StatusOK = "ok"

// TickersResponse is the API response from Huobi for the
// Tickers request.
type TickersResponse struct {
	Status string       `json:"status"`
	Data   []TickerData `json:"data"`
}

// Validate checks if the TickerResponse got an ok status.
func (tr *TickersResponse) Validate() error {
	if tr.Status != StatusOK {
		return fmt.Errorf("status is not ok %s", tr.Status)
	}

	return nil
}

// TickerData is the data payload returned from the Huobi API
// from a Tickers request.
//
// Docs: https://huobiapi.github.io/docs/spot/v1/en/#get-latest-tickers-for-all-pairs
//
// Ex.
//
//	{
//	   "status":"ok",
//	   "ts":1629789355531,
//	   "data":[
//	       {
//	           "symbol":"smtusdt",
//	           "open":0.004659,
//	           "high":0.004696,
//	           "low":0.0046,
//	           "close":0.00468,
//	           "amount":36551302.17544405,
//	           "vol":170526.0643855023,
//	           "count":1709,
//	           "bid":0.004651,
//	           "bidSize":54300.341,
//	           "ask":0.004679,
//	           "askSize":1923.4879
//	       },
//	       {
//	           "symbol":"ltcht",
//	           "open":12.795626,
//	           "high":12.918053,
//	           "low":12.568926,
//	           "close":12.918053,
//	           "amount":1131.801675005825,
//	           "vol":14506.9381937385,
//	           "count":923,
//	           "bid":12.912687,
//	           "bidSize":0.1068,
//	           "ask":12.927032,
//	           "askSize":5.3228
//	       }
//	   ]
//	}
type TickerData struct {
	Symbol string  `json:"symbol"`
	High   float64 `json:"high"`
	Low    float64 `json:"low"`
	// Vol is quote-denominated volume over the last 24 hours.
	Vol  float64 `json:"vol"`
	Open float64 `json:"open"`
}

func (td *TickerData) toProviderMarket() (provider.CreateProviderMarket, error) {
	base, quote, err := symbolToBaseQuote(td.Symbol)
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
			OffChainTicker: td.Symbol,
			ProviderName:   ProviderName,
			QuoteVolume:    td.Vol,
			ReferencePrice: td.Open,
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
			return strings.ToUpper(base), strings.ToUpper(knownQuote), nil
		}
	}

	return "", "", fmt.Errorf(`symbol "%s" does not have a known quote`, symbol)
}

func ignore(symbol string) bool {
	_, found := ignoreSymbols[symbol]

	return found
}
