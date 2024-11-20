package bybit

import (
	"fmt"
	"strconv"

	"github.com/skip-mev/connect-mmu/lib/symbols"
	"github.com/skip-mev/connect-mmu/store/provider"
)

// Response is a general shared struct for fields in all bybit API
// responses.
type Response struct {
	RetCode int    `json:"retCode"`
	RetMsg  string `json:"retMsg"`
}

type InstrumentsResponse struct {
	Response
	Result InstrumentsResult `json:"result"`
}

// InstrumentsResult is a list of InstrumentData.
type InstrumentsResult struct {
	List []InstrumentData `json:"list"`
}

// InstrumentData is the Data payload for a bybit API response to the
// get-instruments request.
//
// Docs: https://bybit-exchange.github.io/docs/v5/market/instrument
//
// Ex.
//
//	{
//		"symbol": "BTCUSDT",
//		"baseCoin": "BTC",
//		"quoteCoin": "USDT",
//		"innovation": "0",
//		"status": "Trading",
//		"marginTrading": "both",
//		"lotSizeFilter": {
//			"basePrecision": "0.000001",
//			"quotePrecision": "0.00000001",
//			"minOrderQty": "0.000048",
//			"maxOrderQty": "71.73956243",
//			"minOrderAmt": "1",
//			"maxOrderAmt": "2000000"
//		},
//		"priceFilter": {
//			"tickSize": "0.01"
//		}
//		"riskParameters": {
//			"limitParameter": "0.05",
//			"marketParameter": "0.05"
//		}
//	}
type InstrumentData struct {
	Symbol    string `json:"symbol"`
	Status    string `json:"status"`
	BaseCoin  string `json:"baseCoin"`
	QuoteCoin string `json:"quoteCoin"`
}

// toProviderMarket converts InstrumentData to a CreateProviderMarketParams object.
func (rd *InstrumentData) toProviderMarket(tickerData TickerData) (provider.CreateProviderMarket, error) {
	baseVol, err := strconv.ParseFloat(tickerData.Volume24H, 64)
	if err != nil {
		return provider.CreateProviderMarket{}, err
	}

	low, err := strconv.ParseFloat(tickerData.LowPrice24H, 64)
	if err != nil {
		return provider.CreateProviderMarket{}, err
	}

	high, err := strconv.ParseFloat(tickerData.HighPrice24H, 64)
	if err != nil {
		return provider.CreateProviderMarket{}, err
	}

	avg := (high + low) / 2
	quoteVol := baseVol * avg

	refPrice, err := strconv.ParseFloat(tickerData.LastPrice, 64)
	if err != nil {
		return provider.CreateProviderMarket{}, fmt.Errorf("failed to convert lastPrice: %w", err)
	}

	targetBase, err := symbols.ToTickerString(rd.BaseCoin)
	if err != nil {
		return provider.CreateProviderMarket{}, err
	}
	targetQuote, err := symbols.ToTickerString(rd.QuoteCoin)
	if err != nil {
		return provider.CreateProviderMarket{}, err
	}

	pm := provider.CreateProviderMarket{
		Create: provider.CreateProviderMarketParams{
			TargetBase:     targetBase,
			TargetQuote:    targetQuote,
			OffChainTicker: rd.Symbol,
			ProviderName:   ProviderName,
			QuoteVolume:    quoteVol,
			ReferencePrice: refPrice,
		},
	}

	return pm, pm.ValidateBasic()
}

type TickersResponse struct {
	Response
	Result TickersResult `json:"result"`
}

// TickersResult is a list of TickerData.
type TickersResult struct {
	List []TickerData `json:"list"`
}

func (tr *TickersResponse) toMap() map[string]TickerData {
	m := make(map[string]TickerData, len(tr.Result.List))

	for _, data := range tr.Result.List {
		m[data.Symbol] = data
	}

	return m
}

// TickerData is the Data payload for a bybit API response to the
// tickers request.
//
// Docs: https://bybit-exchange.github.io/docs/v5/market/tickers
//
// Ex.
//
//	{
//		"symbol": "BTCUSDT",
//		"bid1Price": "20517.96",
//		"bid1Size": "2",
//		"ask1Price": "20527.77",
//		"ask1Size": "1.862172",
//		"lastPrice": "20533.13",
//		"prevPrice24h": "20393.48",
//		"price24hPcnt": "0.0068",
//		"highPrice24h": "21128.12",
//		"lowPrice24h": "20318.89",
//		"turnover24h": "243765620.65899866",
//		"volume24h": "11801.27771",
//		"usdIndexPrice": "20784.12009279"
//	}
type TickerData struct {
	Symbol       string `json:"symbol"`
	LastPrice    string `json:"lastPrice"`
	IndexPrice   string `json:"indexPrice"`
	MarkPrice    string `json:"markPrice"`
	PrevPrice24H string `json:"prevPrice24h"`
	Price24HPcnt string `json:"price24hPcnt"`
	HighPrice24H string `json:"highPrice24h"`
	LowPrice24H  string `json:"lowPrice24h"`
	Volume24H    string `json:"volume24h"`
}
