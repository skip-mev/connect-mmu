package coinbase

// Product is the representation of a single ticker returned
// from the coinbase products api (https://api.exchange.coinbase.com/products). The
// response contains an array of these objects
//
// Example response:
//
//	{
//	    "id": "BTC-USDT",
//	    "base_currency": "BTC",
//	    "quote_currency": "USDT",
//	    ...
//	    "status": "online",
//	    "trading_disabled": false,
//			...
//	}
type Product struct {
	ID              string `json:"id"`
	Base            string `json:"base_currency"`
	Quote           string `json:"quote_currency"`
	Status          string `json:"status"`
	TradingDisabled bool   `json:"trading_disabled"`
}

// Products is the response from the coinbase products endpoint.
type Products []Product

type Stats24Hour struct {
	High   string `json:"high"`   // highest price in the last 24 hours (denominated in quote)
	Low    string `json:"low"`    // lowest price in the last 24 hours (denominated in quote)
	Volume string `json:"volume"` // volume traded over this market (spot) in the last 24 hours (denominated in base)
	Last   string `json:"last"`
}

// StatsPerMarket is a struct that contains the stats for a single market (ticker-pair)
// according to the coinbase stats api (https://api.exchange.coinbase.com/products/stats)
//
// Example response:
//
//	"BTC-USD": {
//	    "stats_24hour": {
//	      "high": "64810",
//	      "low": "62389",
//	      "volume": "10560.32883125",
//	    }
//	}
type StatsPerMarket struct {
	Stats24Hour Stats24Hour `json:"stats_24hour"`
}

// Stats is a map of market id to the stats for that market.
type Stats map[string]StatsPerMarket
