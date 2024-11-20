package kraken

import (
	"fmt"
	"strconv"

	"github.com/skip-mev/connect-mmu/lib/symbols"
	"github.com/skip-mev/connect-mmu/store/provider"
)

const (
	StatusOnline = "online"
)

func (d *AssetData) toProviderMarket(offChainTicker string, data TickerData) (provider.CreateProviderMarket, error) {
	quoteVol, err := data.volumeInQuote()
	if err != nil {
		return provider.CreateProviderMarket{}, err
	}

	refPrice, err := data.referencePrice()
	if err != nil {
		return provider.CreateProviderMarket{}, err
	}

	targetBase, err := symbols.ToTickerString(decode(d.Base))
	if err != nil {
		return provider.CreateProviderMarket{}, err
	}
	targetQuote, err := symbols.ToTickerString(decode(d.Quote))
	if err != nil {
		return provider.CreateProviderMarket{}, err
	}

	pm := provider.CreateProviderMarket{
		Create: provider.CreateProviderMarketParams{
			TargetBase:     targetBase,
			TargetQuote:    targetQuote,
			OffChainTicker: offChainTicker,
			ProviderName:   ProviderName,
			QuoteVolume:    quoteVol,
			ReferencePrice: refPrice,
		},
	}

	return pm, pm.ValidateBasic()
}

type AssetPairsResponse struct {
	Errors []string             `json:"error" validate:"omitempty"`
	Result map[string]AssetData `json:"result"`
}

// AssetData is the struct representing an entry in the data payload map
// in the AssetPairs API response.
//
// Docs: https://docs.kraken.com/rest/#tag/Spot-Market-Data/operation/getTradableAssetPairs
//
// Ex.
//
//		{
//	 "error": [],
//	 "result": {
//	   "XETHXXBT": {
//	     "altname": "ETHXBT",
//	     "wsname": "ETH/XBT",
//	     "aclass_base": "currency",
//	     "base": "XETH",
//	     "aclass_quote": "currency",
//	     "quote": "XXBT",
//	     "lot": "unit",
//	     "cost_decimals": 6,
//	     "pair_decimals": 5,
//	     "lot_decimals": 8,
//	     "lot_multiplier": 1,
//	     "leverage_buy": [
//	       2,
//	       3,
//	       4,
//	       5
//	     ],
//	     "leverage_sell": [
//	       2,
//	       3,
//	       4,
//	       5
//	     ],
//	     "fee_volume_currency": "ZUSD",
//	     "margin_call": 80,
//	     "margin_stop": 40,
//	     "ordermin": "0.01",
//	     "costmin": "0.00002",
//	     "tick_size": "0.00001",
//	     "status": "online",
//	     "long_position_limit": 1100,
//	     "short_position_limit": 400
//	   },
//	 }
//	}
type AssetData struct {
	Altname     string `json:"altname"`
	Wsname      string `json:"wsname"`
	AclassBase  string `json:"aclass_base"`
	Base        string `json:"base"`
	AclassQuote string `json:"aclass_quote"`
	Quote       string `json:"quote"`
	Status      string `json:"status"`
}

// decode decodes an input ticker to the standardized Slinky ticker.
func decode(input string) string {
	switch input {
	case "ZEUR":
		return "EUR"
	case "ZUSD":
		return "USD"
	case "XXBT":
		return "BTC"
	case "XETH":
		return "ETH"
	case "XXRP":
		return "XRP"
	case "ZGBP":
		return "GBP"
	case "ZJPY":
		return "JPY"
	case "XLTC":
		return "LTC"
	case "XXDG":
		return "DOGE"
	case "ZCAD":
		return "CAD"
	case "XXMR":
		return "XMR"
	case "ZAUD":
		return "AUD"
	case "XREP":
		return "REP"
	case "XZEC":
		return "ZEC"
	case "XETC":
		return "ETC"
	case "XMLN":
		return "MLN"
	case "XXLM":
		return "XLM"

	default:
		return input
	}
}

type TickersResponse struct {
	Error  []interface{}         `json:"error"`
	Result map[string]TickerData `json:"result"`
}

// TickerData is the struct representing an entry in the data payload map
// in the Tickers API response.
//
// Docs: https://docs.kraken.com/rest/#tag/Spot-Market-Data/operation/getAssetInfo
//
// Ex.
//
//	{
//	 "error": [],
//	 "result": {
//	   "XXBTZUSD": {
//	     "a": [
//	       "30300.10000",
//	       "1",
//	       "1.000"
//	     ],
//	     "b": [
//	       "30300.00000",
//	       "1",
//	       "1.000"
//	     ],
//	     "c": [
//	       "30303.20000",
//	       "0.00067643"
//	     ],
//	     "v": [
//	       "4083.67001100",
//	       "4412.73601799"
//	     ],
//	     "p": [
//	       "30706.77771",
//	       "30689.13205"
//	     ],
//	     "t": [
//	       34619,
//	       38907
//	     ],
//	     "l": [
//	       "29868.30000",
//	       "29868.30000"
//	     ],
//	     "h": [
//	       "31631.00000",
//	       "31631.00000"
//	     ],
//	     "o": "30502.80000"
//	   }
//	 }
//	}
type TickerData struct {
	// V is volume.
	//	- market-index 0 is for today
	//	- market-index 1 is for the last 24hours.
	V []string `json:"v"`
	// L is the low price.
	//	- market-index 0 is for today
	//	- market-index 1 is for the last 24hours.
	L []string `json:"l"`
	// H is the high price.
	//	- market-index 0 is for today
	//	- market-index 1 is for the last 24hours.
	H []string `json:"h"`
	// C is the last trade price.
	// - market-index 0 is the price of the last trade.
	// - market-index 1 is for the volume of the last trade.
	C []string `json:"c"`
}

func (td *TickerData) referencePrice() (float64, error) {
	refPrice, err := strconv.ParseFloat(td.C[0], 64)
	if err != nil {
		return 0, fmt.Errorf("failed to convert TickerData.C[0]: %w", err)
	}
	return refPrice, nil
}

func (td *TickerData) volumeInQuote() (float64, error) {
	volume, err := strconv.ParseFloat(td.V[1], 64)
	if err != nil {
		return 0, err
	}

	low, err := strconv.ParseFloat(td.L[1], 64)
	if err != nil {
		return 0, err
	}

	high, err := strconv.ParseFloat(td.H[1], 64)
	if err != nil {
		return 0, err
	}

	avg := (low + high) / 2
	return volume * avg, nil
}
