//nolint:revive
package crypto_com

import (
	"fmt"
	"strconv"

	"github.com/skip-mev/connect-mmu/lib/symbols"
	"github.com/skip-mev/connect-mmu/store/provider"
)

type InstrumentType string

type InstrumentsResponse struct {
	ID     int               `json:"id"`
	Method string            `json:"method"`
	Code   int               `json:"code"`
	Result InstrumentsResult `json:"result"`
}

type InstrumentsResult struct {
	Data []InstrumentsData `json:"data"`
}

// InstrumentsData is the data payload for an instruments request
// to the crypto.com API.
//
// https://api.crypto.com/exchange/v1/public/get-instruments
//
// Example:
//
//	{
//	 "id": 1,
//	 "method":"public/get-instruments",
//	 "code": 0,
//	 "result":{
//	   "data":[
//	     {
//	       "symbol":"BTCUSD-PERP",
//	       "inst_type":"PERPETUAL_SWAP",
//	       "display_name":"BTCUSD Perpetual",
//	       "base_ccy":"BTC",
//	       "quote_ccy":"USD",
//	       "quote_decimals":2,
//	       "quantity_decimals":4,
//	       "price_tick_size":"0.5",
//	       "qty_tick_size":"0.0001",
//	       "max_leverage":"50",
//	       "tradable":true,
//	       "expiry_timestamp_ms":1624012801123,
//	       "underlying_symbol": "BTCUSD-INDEX"
//	     }
//	   ]
//	 }
//	}
type InstrumentsData struct {
	Symbol      string `json:"symbol"`
	InstType    string `json:"inst_type"`
	DisplayName string `json:"display_name"`
	BaseCcy     string `json:"base_ccy"`
	QuoteCcy    string `json:"quote_ccy"`
	Tradable    bool   `json:"tradable"`
}

func (d *InstrumentsData) toProviderMarket(ticker TickerData) (provider.CreateProviderMarket, error) {
	baseVol, err := strconv.ParseFloat(ticker.V, 64)
	if err != nil {
		return provider.CreateProviderMarket{}, err
	}

	low, err := strconv.ParseFloat(ticker.L, 64)
	if err != nil {
		return provider.CreateProviderMarket{}, err
	}

	high, err := strconv.ParseFloat(ticker.H, 64)
	if err != nil {
		return provider.CreateProviderMarket{}, err
	}

	usdVol, err := strconv.ParseFloat(ticker.Vv, 64)
	if err != nil {
		return provider.CreateProviderMarket{}, err
	}

	refPrice, err := strconv.ParseFloat(ticker.LatestPrice, 64)
	if err != nil {
		return provider.CreateProviderMarket{}, fmt.Errorf("failed to convert latest price to float: %w", err)
	}

	avg := (high + low) / 2
	quoteVol := baseVol * avg

	targetBase, err := symbols.ToTickerString(d.BaseCcy)
	if err != nil {
		return provider.CreateProviderMarket{}, err
	}
	targetQuote, err := symbols.ToTickerString(d.QuoteCcy)
	if err != nil {
		return provider.CreateProviderMarket{}, err
	}

	pm := provider.CreateProviderMarket{
		Create: provider.CreateProviderMarketParams{
			TargetBase:     targetBase,
			TargetQuote:    targetQuote,
			OffChainTicker: d.Symbol,
			ProviderName:   ProviderName,
			QuoteVolume:    quoteVol,
			UsdVolume:      usdVol,
			ReferencePrice: refPrice,
		},
	}

	return pm, pm.ValidateBasic()
}

type TickersResponse struct {
	ID     int           `json:"id"`
	Method string        `json:"method"`
	Code   int           `json:"code"`
	Result TickersResult `json:"result"`
}

type TickersResult struct {
	Data []TickerData `json:"data"`
}

func (tr *TickersResponse) toMap() map[string]TickerData {
	m := make(map[string]TickerData, len(tr.Result.Data))

	for _, d := range tr.Result.Data {
		m[d.I] = d
	}

	return m
}

// TickerData is the data payload returned from the tickes request
// to the crypto.com API.
//
// https://exchange-docs.crypto.com/exchange/v1/rest-ws/index.html#public-get-tickers
//
// Example:
//
//	{
//	 "id": -1,
//	 "method": "public/get-tickers",
//	 "code": 0,
//	 "result": {
//	   "data": [{
//	     "h": "51790.00",        // Price of the 24h highest trade
//	     "l": "47895.50",        // Price of the 24h lowest trade, null if there weren't any trades
//	     "a": "51174.500000",    // The price of the latest trade, null if there weren't any trades
//	     "i": "BTCUSD-PERP",     // Instrument name
//	     "v": "879.5024",        // The total 24h traded volume
//	     "vv": "26370000.12",    // The total 24h traded volume value (in USD)
//	     "oi": "12345.12",       // Open interest
//	     "c": "0.03955106",      // 24-hour price change, null if there weren't any trades
//	     "b": "51170.000000",    // The current best bid price, null if there aren't any bids
//	     "k": "51180.000000",    // The current best ask price, null if there aren't any asks
//	     "t": 1613580710768
//	   }]
//	 }
//	}
type TickerData struct {
	// H is the price of the 24h highest trace (float).
	H string `json:"h"`
	// L is the price of the 24h lowest trace (float).
	L           string `json:"l"`
	LatestPrice string `json:"a"`
	I           string `json:"i"`
	// V is 24 hour volume (float).
	V string `json:"v"`
	// Vv is 24 hour volume in USD (float).
	Vv string `json:"vv"`
	Oi string `json:"oi"`
	C  string `json:"c"`
	B  string `json:"b"`
	K  string `json:"k"`
	T  int64  `json:"t"`
}
