package okx

import (
	"fmt"
	"strconv"

	"github.com/skip-mev/connect-mmu/lib/symbols"
	"github.com/skip-mev/connect-mmu/store/provider"
)

const StateLive = "live"

// Response is a common shared field for all responses from the
// okx API.
type Response struct {
	Code string `json:"code"`
	Msg  string `json:"msg"`
}

// InstrumentsResponse is a response to the GetInstruments
// API.
type InstrumentsResponse struct {
	Response
	Data []InstrumentData `json:"data"`
}

// Validate checks if the code is valid from the response.
func (ir *InstrumentsResponse) Validate() error {
	if ir.Code != "0" {
		return fmt.Errorf("invalid instruments response: %s", ir.Msg)
	}

	return nil
}

// InstrumentData is the data payload included in a
// InstrumentsResponse.
//
// Docs: https://www.okx.com/docs-v5/en/#public-data-rest-api-get-instruments
//
// Ex.
//
//	{
//	   "code":"0",
//	   "msg":"",
//	   "data":[
//	     {
//	           "alias": "",
//	           "baseCcy": "BTC",
//	           "category": "1",
//	           "ctMult": "",
//	           "ctType": "",
//	           "ctVal": "",
//	           "ctValCcy": "",
//	           "expTime": "",
//	           "instFamily": "",
//	           "instId": "BTC-USDT",
//	           "instType": "SPOT",
//	           "lever": "10",
//	           "listTime": "1606468572000",
//	           "lotSz": "0.00000001",
//	           "maxIcebergSz": "9999999999.0000000000000000",
//	           "maxLmtAmt": "1000000",
//	           "maxLmtSz": "9999999999",
//	           "maxMktAmt": "1000000",
//	           "maxMktSz": "",
//	           "maxStopSz": "",
//	           "maxTriggerSz": "9999999999.0000000000000000",
//	           "maxTwapSz": "9999999999.0000000000000000",
//	           "minSz": "0.00001",
//	           "optType": "",
//	           "quoteCcy": "USDT",
//	           "settleCcy": "",
//	           "state": "live",
//	           "stk": "",
//	           "tickSz": "0.1",
//	           "uly": ""
//	       }
//	   ]
//	}
type InstrumentData struct {
	BaseCcy  string `json:"baseCcy"`
	InstID   string `json:"instId"`
	InstType string `json:"instType"`
	QuoteCcy string `json:"quoteCcy"`
	State    string `json:"state"`
}

func (ir *InstrumentData) toProviderMarket(td TickerData) (provider.CreateProviderMarket, error) {
	volume, err := strconv.ParseFloat(td.VolCcy24H, 64)
	if err != nil {
		return provider.CreateProviderMarket{}, err
	}

	refPrice, err := strconv.ParseFloat(td.Open24h, 64)
	if err != nil {
		return provider.CreateProviderMarket{}, fmt.Errorf("failed to convert Open24h: %w", err)
	}

	targetBase, err := symbols.ToTickerString(ir.BaseCcy)
	if err != nil {
		return provider.CreateProviderMarket{}, err
	}
	targetQuote, err := symbols.ToTickerString(ir.QuoteCcy)
	if err != nil {
		return provider.CreateProviderMarket{}, err
	}

	pm := provider.CreateProviderMarket{
		Create: provider.CreateProviderMarketParams{
			TargetBase:     targetBase,
			TargetQuote:    targetQuote,
			OffChainTicker: ir.InstID,
			ProviderName:   ProviderName,
			QuoteVolume:    volume,
			ReferencePrice: refPrice,
		},
	}

	return pm, pm.ValidateBasic()
}

// TickersResponse is a response to the Tickers
// API.
type TickersResponse struct {
	Response
	Data []TickerData `json:"data"`
}

// Validate checks if the code is valid from the response.
func (tr *TickersResponse) Validate() error {
	if tr.Code != "0" {
		return fmt.Errorf("invalid tickers response: %s", tr.Msg)
	}

	return nil
}

// TickerData is the data payload included in a
// TickersResponse.
//
// Docs: https://www.okx.com/docs-v5/en/#order-book-trading-market-data-get-tickers
//
// Ex.
//
//		{
//	   "code":"0",
//	   "msg":"",
//	   "data":[
//	    {
//	       "instType":"SWAP",
//	       "instId":"LTC-USD-SWAP",
//	       "last":"9999.99",
//	       "lastSz":"1",
//	       "askPx":"9999.99",
//	       "askSz":"11",
//	       "bidPx":"8888.88",
//	       "bidSz":"5",
//	       "open24h":"9000",
//	       "high24h":"10000",
//	       "low24h":"8888.88",
//	       "volCcy24h":"2222",
//	       "vol24h":"2222",
//	       "sodUtc0":"0.1",
//	       "sodUtc8":"0.1",
//	       "ts":"1597026383085"
//	    },
//	    {
//	       "instType":"SWAP",
//	       "instId":"BTC-USD-SWAP",
//	       "last":"9999.99",
//	       "lastSz":"1",
//	       "askPx":"9999.99",
//	       "askSz":"11",
//	       "bidPx":"8888.88",
//	       "bidSz":"5",
//	       "open24h":"9000",
//	       "high24h":"10000",
//	       "low24h":"8888.88",
//	       "volCcy24h":"2222",
//	       "vol24h":"2222",
//	       "sodUtc0":"0.1",
//	       "sodUtc8":"0.1",
//	       "ts":"1597026383085"
//	   }
//	 ]
//	}
type TickerData struct {
	InstType  string `json:"instType"`
	InstID    string `json:"instId"`
	VolCcy24H string `json:"volCcy24h"`
	Open24h   string `json:"open24h"`
}

func (tr *TickersResponse) toMap() map[string]TickerData {
	m := make(map[string]TickerData, len(tr.Data))

	for _, d := range tr.Data {
		m[d.InstID] = d
	}

	return m
}
