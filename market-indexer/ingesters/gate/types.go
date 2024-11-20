package gate

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/skip-mev/connect-mmu/lib/symbols"
	"github.com/skip-mev/connect-mmu/store/provider"
)

const delimiter = "_"

// TickerData is the data payload returned from the Gate.io
// API in response to the Tickers request.
//
// Docs: https://www.gate.io/docs/developers/apiv4/#retrieve-ticker-information
//
// Ex.
// [
//
//	{
//	  "currency_pair": "BTC3L_USDT",
//	  "last": "2.46140352",
//	  "lowest_ask": "2.477",
//	  "highest_bid": "2.4606821",
//	  "change_percentage": "-8.91",
//	  "change_utc0": "-8.91",
//	  "change_utc8": "-8.91",
//	  "base_volume": "656614.0845820589",
//	  "quote_volume": "1602221.66468375534639404191",
//	  "high_24h": "2.7431",
//	  "low_24h": "1.9863",
//	  "etf_net_value": "2.46316141",
//	  "etf_pre_net_value": "2.43201848",
//	  "etf_pre_timestamp": 1611244800,
//	  "etf_leverage": "2.2803019447281203"
//	}
//
// ].
type TickerData struct {
	CurrencyPair string `json:"currency_pair"`
	// QuoteVolume is the 24hr volume in terms of the quote currency.
	QuoteVolume     string `json:"quote_volume"`
	EtfNetValue     string `json:"etf_net_value"`
	EtfPreNetValue  string `json:"etf_pre_net_value"`
	EtfPreTimestamp int    `json:"etf_pre_timestamp"`
	EtfLeverage     string `json:"etf_leverage"`
	LastPrice       string `json:"last"`
}

func (td *TickerData) isSpot() bool {
	return td.EtfNetValue == "" && td.EtfPreNetValue == "" && td.EtfPreTimestamp == 0 && td.EtfLeverage == ""
}

func (td *TickerData) toProviderMarket() (provider.CreateProviderMarket, error) {
	base, quote, err := symbolToBaseQuote(td.CurrencyPair)
	if err != nil {
		return provider.CreateProviderMarket{}, err
	}

	quoteVol, err := strconv.ParseFloat(td.QuoteVolume, 64)
	if err != nil {
		return provider.CreateProviderMarket{}, err
	}

	refPrice, err := strconv.ParseFloat(td.LastPrice, 64)
	if err != nil {
		return provider.CreateProviderMarket{}, fmt.Errorf("failed to convert last price: %w", err)
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
			OffChainTicker: td.CurrencyPair,
			ProviderName:   ProviderName,
			QuoteVolume:    quoteVol,
			ReferencePrice: refPrice,
		},
	}

	return pm, pm.ValidateBasic()
}

// symbolToBaseQuote splits a ticker to base and quote currencies.
func symbolToBaseQuote(symbol string) (string, string, error) {
	split := strings.Split(symbol, delimiter)
	if len(split) != 2 {
		return "", "", fmt.Errorf("symbol %s not supported", symbol)
	}

	return split[0], split[1], nil
}
