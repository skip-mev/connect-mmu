package raydium

// Pairs is an alias for an array of PairData objects.
type Pairs []PairData

// PairData is the type returned from the /pairs API on Raydium v2.
//
// https://api.raydium.io/v2/main/pairs
//
// Ex:
//
//	 {
//	  "name": "SRM-USDC",
//	  "ammId": "8tzS7SkUZyHPQY7gLqsMCXZ5EDCgjESUHcB17tiR1h3Z",
//	  "lpMint": "9XnZd82j34KxNLgQfz29jGbYdxsYznTWRpvZE3SRE7JG",
//	  "baseMint": "SRMuApVNdxXokk5GT7XD5cUUgXMBCoAz2LHeuAoKWRt",
//	  "quoteMint": "EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v",
//	  "market": "ByRys5tuUWDgL73G8JBAEfkdFf8JWBzPBDHsBVQ5vbQA",
//	  "liquidity": 101760.526754,
//	  "volume24h": 1875.1074019999996,
//	  "volume24hQuote": 1875.1074019999996,
//	  "fee24h": 4.1252362844,
//	  "fee24hQuote": 4.1252362844,
//	  "volume7d": 9657.792483999996,
//	  "volume7dQuote": 9657.792483999996,
//	  "fee7d": 21.247143464799993,
//	  "fee7dQuote": 21.247143464799993,
//	  "volume30d": 30305.02346200001,
//	  "volume30dQuote": 30305.02346200001,
//	  "fee30d": 66.67105161640002,
//	  "fee30dQuote": 66.67105161640002,
//	  "price": 0.035400681679734834,
//	  "lpPrice": 3.8314572218218066,
//	  "tokenAmountCoin": 1437267.898887,
//	  "tokenAmountPc": 50880.263377,
//	  "tokenAmountLp": 26559.222996,
//	  "apr24h": 1.48,
//	  "apr7d": 1.09,
//	  "apr30d": 0.8
//	}
type PairData struct {
	Name            string  `json:"name"`
	AmmID           string  `json:"ammId"`
	LpMint          string  `json:"lpMint"`
	BaseMint        string  `json:"baseMint"`
	QuoteMint       string  `json:"quoteMint"`
	Market          string  `json:"market"`
	Liquidity       float64 `json:"liquidity"`
	Volume24H       float64 `json:"volume24h"`
	Volume24HQuote  float64 `json:"volume24hQuote"`
	Fee24H          float64 `json:"fee24h"`
	Fee24HQuote     float64 `json:"fee24hQuote"`
	Volume7D        float64 `json:"volume7d"`
	Volume7DQuote   float64 `json:"volume7dQuote"`
	Fee7D           float64 `json:"fee7d"`
	Fee7DQuote      float64 `json:"fee7dQuote"`
	Volume30D       float64 `json:"volume30d"`
	Volume30DQuote  float64 `json:"volume30dQuote"`
	Fee30D          float64 `json:"fee30d"`
	Fee30DQuote     float64 `json:"fee30dQuote"`
	Price           float64 `json:"price"`
	LpPrice         float64 `json:"lpPrice"`
	TokenAmountCoin float64 `json:"tokenAmountCoin"`
	TokenAmountPc   float64 `json:"tokenAmountPc"`
	TokenAmountLp   float64 `json:"tokenAmountLp"`
	Apr24H          float64 `json:"apr24h"`
	Apr7D           float64 `json:"apr7d"`
	Apr30D          float64 `json:"apr30d"`
}

type TokenMetadataResponse struct {
	Content []Content `json:"content"`
}

// Content is the data payload returned from the Solflare API
//
// https://token-list-api.solana.cloud/v1/list
//
// Ex.
//
//	 {
//	  "address": "ETPz31G7uXGCAv8o2bDhWmx9ejZvNdmirg9x62N3AAga",
//	  "chainId": 101,
//	  "name": "RAID TOKEN",
//	  "symbol": "RAID",
//	  "verified": true,
//	  "decimals": 2,
//	  "holders": 142,
//	  "logoURI": "https://raw.githubusercontent.com/DefiTokens/assets/main/RAID%20TOKEN.png",
//	  "tags": []
//	}.
type Content struct {
	Address    string   `json:"address"`
	ChainID    int      `json:"chainId"`
	Name       string   `json:"name"`
	Symbol     string   `json:"symbol"`
	Verified   bool     `json:"verified"`
	Decimals   int      `json:"decimals"`
	Holders    *int     `json:"holders"`
	LogoURI    *string  `json:"logoURI"`
	Tags       []string `json:"tags"`
	Extensions struct {
		CoingeckoID string `json:"coingeckoId"`
	} `json:"extensions,omitempty"`
}
