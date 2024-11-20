package coinmarketcap

import (
	"fmt"
	"net/http"
	"time"
)

const (
	Name = "coinmarketcap"
)

// CryptoIDMapResponse is the payload returned form coinmarketcap using the
//
// https://pro-api.coinmarketcap.com/v1/cryptocurrency/map
//
// Query. More documentation can be found here: https://coinmarketcap.com/api/documentation/v1/#operation/getV1CryptocurrencyMap
//
// Example:
// {
// "data": [
// {
// "id": 1,
// "rank": 1,
// "name": "Bitcoin",
// "symbol": "BTC",
// "slug": "bitcoin",
// "is_active": 1,
// "first_historical_data": "2013-04-28T18:47:21.000Z",
// "last_historical_data": "2020-05-05T20:44:01.000Z",
// "platform": null
// },
// {
// "id": 1839,
// "rank": 3,
// "name": "Binance Coin",
// "symbol": "BNB",
// "slug": "binance-coin",
// "is_active": 1,
// "first_historical_data": "2017-07-25T04:30:05.000Z",
// "last_historical_data": "2020-05-05T20:44:02.000Z",
// "platform": {
// "id": 1027,
// "name": "Ethereum",
// "symbol": "ETH",
// "slug": "ethereum",
// "token_address": "0xB8c77482e45F1F44dE1745F52C74426C631bDD52"
// }
// },
// {
// "id": 825,
// "rank": 5,
// "name": "Tether",
// "symbol": "USDT",
// "slug": "tether",
// "is_active": 1,
// "first_historical_data": "2015-02-25T13:34:26.000Z",
// "last_historical_data": "2020-05-05T20:44:01.000Z",
// "platform": {
// "id": 1027,
// "name": "Ethereum",
// "symbol": "ETH",
// "slug": "ethereum",
// "token_address": "0xdac17f958d2ee523a2206206994597c13d831ec7"
// }
// }
// ],
// "status": {
// "timestamp": "2018-06-02T22:51:28.209Z",
// "error_code": 0,
// "error_message": "",
// "elapsed": 10,
// "credit_count": 1
// }
// }.
type CryptoIDMapResponse struct {
	Data   []CryptoIDMapData `json:"data"`
	Status Status            `json:"status"`
}

type CryptoIDMapData struct {
	ID                  int       `json:"id"`
	Rank                int       `json:"rank"`
	Name                string    `json:"name"`
	Symbol              string    `json:"symbol"`
	Slug                string    `json:"slug"`
	IsActive            int       `json:"is_active"`
	FirstHistoricalData time.Time `json:"first_historical_data"`
	LastHistoricalData  time.Time `json:"last_historical_data"`
	Platform            *struct {
		ID           int    `json:"id"`
		Name         string `json:"name"`
		Symbol       string `json:"symbol"`
		Slug         string `json:"slug"`
		TokenAddress string `json:"token_address"`
	} `json:"platform"`
}

// ExchangeIDMapResponse is the payload returned form coinmarketcap using the
//
// https://pro-api.coinmarketcap.com/v1/exchange/map
//
// Query. More documentation can be found here: https://coinmarketcap.com/api/documentation/v1/#operation/getV1ExchangeMap
//
// Ex:
//
// {
// "data": [
//
//	{
//	"id": 270,
//	"name": "Binance",
//	"slug": "binance",
//	"is_active": 1,
//	"status": "active",
//	"first_historical_data": "2018-04-26T00:45:00.000Z",
//	"last_historical_data": "2019-06-02T21:25:00.000Z"
//	}
//
// ],
//
//	"status": {
//		"timestamp": "2024-06-12T12:30:30.867Z",
//		"error_code": 0,
//		"error_message": "",
//		"elapsed": 10,
//		"credit_count": 1,
//		"notice": ""
//		}
//	}.
type ExchangeIDMapResponse struct {
	Data   []ExchangeIDMapData `json:"data"`
	Status Status              `json:"status"`
}

type ExchangeIDMapData struct {
	ID                  int       `json:"id"`
	Name                string    `json:"name"`
	Slug                string    `json:"slug"`
	IsActive            int       `json:"is_active"`
	Status              string    `json:"status"`
	FirstHistoricalData time.Time `json:"first_historical_data"`
	LastHistoricalData  time.Time `json:"last_historical_data"`
}

// ExchangeAssetsResponse is the payload returned form coinmarketcap using the
//
// https://pro-api.coinmarketcap.com/v1/exchange/assets
//
// Query. More documentation can be found here: https://coinmarketcap.com/api/documentation/v1/#operation/getV1ExchangeAssets
//
// {
//
//	"status": {
//		"timestamp": "2022-11-24T08:23:22.028Z",
//		"error_code": 0,
//		"error_message": null,
//		"elapsed": 1828,
//		"credit_count": 0,
//		"notice": null
//	},
//
// "data": [
//
//	{
//		"wallet_address": "0x5a52e96bacdabb82fd05763e25335261b270efcb",
//		"balance": 45000000,
//		"platform": {
//		"crypto_id": 1027,
//		"symbol": "ETH",
//		"name": "Ethereum"
//	},
//
//	"currency": {
//		"crypto_id": 5117,
//		"price_usd": 0.10241799413549,
//		"symbol": "OGN",
//		"name": "Origin Protocol"
//		}
//	},
//
// ]
// }.
type ExchangeAssetsResponse struct {
	Data   []ExchangeAssetData `json:"data"`
	Status Status              `json:"status"`
}

type ExchangeAssetData struct {
	WalletAddress string  `json:"wallet_address"`
	Balance       float64 `json:"balance"`
	Platform      struct {
		CryptoID int    `json:"crypto_id"`
		Symbol   string `json:"symbol"`
		Name     string `json:"name"`
	} `json:"platform"`
	Currency struct {
		CryptoID int     `json:"crypto_id"`
		PriceUSD float64 `json:"price_usd"`
		Symbol   string  `json:"symbol"`
		Name     string  `json:"name"`
	} `json:"currency"`
}

// ExchangeMarketsResponse is the expected response for the following query to CMC.
//
// https://pro-api.coinmarketcap.com/v1/exchange/market-pairs/latest
//
// The data payload is as follows:
//
// "data": {
// "id": 270,
// "name": "Binance",
// "slug": "binance",
// "num_market_pairs": 473,
// "volume_24h": 769291636.239632,
// "market_pairs": [
// {
// "market_id": 9933,
// "market_pair": "BTC/USDT",
// "category": "spot",
// "fee_type": "percentage",
// "outlier_detected": 0,
// "exclusions": null,
// "market_pair_base": {
// "currency_id": 1,
// "currency_symbol": "BTC",
// "exchange_symbol": "BTC",
// "currency_type": "cryptocurrency"
// },
// "market_pair_quote": {
// "currency_id": 825,
// "currency_symbol": "USDT",
// "exchange_symbol": "USDT",
// "currency_type": "cryptocurrency"
// },
// "quote": {
// "exchange_reported": {
// "price": 7901.83,
// "volume_24h_base": 47251.3345550653,
// "volume_24h_quote": 373372012.927251,
// "volume_percentage": 19.4346563602467,
// "last_updated": "2019-05-24T01:40:10.000Z"
// },
// "USD": {
// "price": 7933.66233493434,
// "volume_24h": 374876133.234903,
// "depth_negative_two": 40654.68019906,
// "depth_positive_two": 17352.9964811,
// "last_updated": "2019-05-24T01:40:10.000Z"
// }
// }
// },
// {
// "market_id": 36329,
// "market_pair": "MATIC/BTC",
// "category": "spot",
// "fee_type": "percentage",
// "outlier_detected": 0,
// "exclusions": null,
// "market_pair_base": {
// "currency_id": 3890,
// "currency_symbol": "MATIC",
// "exchange_symbol": "MATIC",
// "currency_type": "cryptocurrency"
// },
// "market_pair_quote": {
// "currency_id": 1,
// "currency_symbol": "BTC",
// "exchange_symbol": "BTC",
// "currency_type": "cryptocurrency"
// },
// "quote": {
// "exchange_reported": {
// "price": 0.0000034,
// "volume_24h_base": 8773968381.05,
// "volume_24h_quote": 29831.49249557,
// "volume_percentage": 19.4346563602467,
// "last_updated": "2019-05-24T01:41:16.000Z"
// },
// "USD": {
// "price": 0.0269295015799739,
// "volume_24h": 236278595.380127,
// "depth_negative_two": 40654.68019906,
// "depth_positive_two": 17352.9964811,
// "last_updated": "2019-05-24T01:41:16.000Z"
// }
// }
// }
// ]
// }.
//
// More information can be found here: https://coinmarketcap.com/api/documentation/v1/#operation/getV1ExchangeListingsLatest.
type ExchangeMarketsResponse struct {
	Data   ExchangeMarketsData `json:"data"`
	Status Status              `json:"status"`
}

type ExchangeMarketsData struct {
	ID             int     `json:"id"`
	Name           string  `json:"name"`
	Slug           string  `json:"slug"`
	NumMarketPairs int     `json:"num_market_pairs"`
	Volume24H      float64 `json:"volume_24h"`
	MarketPairs    []struct {
		MarketID        int         `json:"market_id"`
		MarketPair      string      `json:"market_pair"`
		Category        string      `json:"category"`
		FeeType         string      `json:"fee_type"`
		OutlierDetected int         `json:"outlier_detected"`
		Exclusions      interface{} `json:"exclusions"`
		MarketPairBase  struct {
			CurrencyID     int    `json:"currency_id"`
			CurrencySymbol string `json:"currency_symbol"`
			ExchangeSymbol string `json:"exchange_symbol"`
			CurrencyType   string `json:"currency_type"`
		} `json:"market_pair_base"`
		MarketPairQuote struct {
			CurrencyID     int    `json:"currency_id"`
			CurrencySymbol string `json:"currency_symbol"`
			ExchangeSymbol string `json:"exchange_symbol"`
			CurrencyType   string `json:"currency_type"`
		} `json:"market_pair_quote"`
		Quote struct {
			ExchangeReported struct {
				Price            float64   `json:"price"`
				Volume24HBase    float64   `json:"volume_24h_base"`
				Volume24HQuote   float64   `json:"volume_24h_quote"`
				VolumePercentage float64   `json:"volume_percentage"`
				LastUpdated      time.Time `json:"last_updated"`
			} `json:"exchange_reported"`
			USD struct {
				Price            float64   `json:"price"`
				Volume24H        float64   `json:"volume_24h"`
				DepthNegativeTwo float64   `json:"depth_negative_two"`
				DepthPositiveTwo float64   `json:"depth_positive_two"`
				LastUpdated      time.Time `json:"last_updated"`
			} `json:"USD"`
		} `json:"quote"`
	} `json:"market_pairs"`
}

// FiatResponse is the response returned by the CoinMarketCap API for the
//
// https://pro-api.coinmarketcap.com/v1/fiat/map
//
// request.
//
// Example:
//
// {
// "data": [
// {
// "id": 2781,
// "name": "United States Dollar",
// "sign": "$",
// "symbol": "USD"
// },
// {
// "id": 2787,
// "name": "Chinese Yuan",
// "sign": "¥",
// "symbol": "CNY"
// },
// {
// "id": 2781,
// "name": "South Korean Won",
// "sign": "₩",
// "symbol": "KRW"
// }
// ],
// "status": {
// "timestamp": "2020-01-07T22:51:28.209Z",
// "error_code": 0,
// "error_message": "",
// "elapsed": 3,
// "credit_count": 1
// }
// }
//
// More information can be found here:https://coinmarketcap.com/api/documentation/v1/#tag/fiat.
type FiatResponse struct {
	Data   []FiatData `json:"data"`
	Status Status     `json:"status"`
}

type FiatData struct {
	ID     int    `json:"id"`
	Name   string `json:"name"`
	Sign   string `json:"sign"`
	Symbol string `json:"symbol"`
}

// QuoteResponse is the response returned by the CoinMarketCap API for the
//
// https://pro-api.coinmarketcap.com/v2/cryptocurrency/quotes/latest
//
// request.
//
// Example:
//
// {
// "data": {
// "1": {
// "id": 1,
// "name": "Bitcoin",
// "symbol": "BTC",
// "slug": "bitcoin",
// "is_active": 1,
// "is_fiat": 0,
// "circulating_supply": 17199862,
// "total_supply": 17199862,
// "max_supply": 21000000,
// "date_added": "2013-04-28T00:00:00.000Z",
// "num_market_pairs": 331,
// "cmc_rank": 1,
// "last_updated": "2018-08-09T21:56:28.000Z",
// "tags": [
// "mineable"
// ],
// "platform": null,
// "self_reported_circulating_supply": null,
// "self_reported_market_cap": null,
// "quote": {
// "USD": {
// "price": 6602.60701122,
// "volume_24h": 4314444687.5194,
// "volume_change_24h": -0.152774,
// "percent_change_1h": 0.988615,
// "percent_change_24h": 4.37185,
// "percent_change_7d": -12.1352,
// "percent_change_30d": -12.1352,
// "market_cap": 852164659250.2758,
// "market_cap_dominance": 51,
// "fully_diluted_market_cap": 952835089431.14,
// "last_updated": "2018-08-09T21:56:28.000Z"
// }
// }
// }
// },
// "status": {
// "timestamp": "2024-07-19T10:27:30.125Z",
// "error_code": 0,
// "error_message": "",
// "elapsed": 10,
// "credit_count": 1,
// "notice": ""
// }
// }
//
// More information can be found here:https://coinmarketcap.com/api/documentation/v1/#operation/getV2CryptocurrencyQuotesLatest.
type QuoteResponse struct {
	Status Status               `json:"status"`
	Data   map[string]QuoteData `json:"data"`
}

type QuoteData struct {
	ID             int       `json:"id"`
	Name           string    `json:"name"`
	Symbol         string    `json:"symbol"`
	Slug           string    `json:"slug"`
	NumMarketPairs int       `json:"num_market_pairs"`
	DateAdded      time.Time `json:"date_added"`
	Platform       struct {
		ID           int    `json:"id"`
		Name         string `json:"name"`
		Symbol       string `json:"symbol"`
		Slug         string `json:"slug"`
		TokenAddress string `json:"token_address"`
	} `json:"platform"`
	IsActive       int       `json:"is_active"`
	InfiniteSupply bool      `json:"infinite_supply"`
	CmcRank        int       `json:"cmc_rank"`
	IsFiat         int       `json:"is_fiat"`
	LastUpdated    time.Time `json:"last_updated"`
	// Quote is a map of price to
	Quote map[string]struct {
		Price     float64 `json:"price"`
		Volume24H float64 `json:"volume_24h"`
	} `json:"quote"`
}

// InfoResponse is the payload returned from the info query to CoinMarketCap
//
// Ex:
//
//	{
//	 "status": {
//	   "timestamp": "2024-08-28T17:04:12.022Z",
//	   "error_code": 0,
//	   "error_message": null,
//	   "elapsed": 25,
//	   "credit_count": 1,
//	   "notice": null
//	 },
//	 "data": {
//	   "2396": {
//	     "id": 2396,
//	     "name": "WETH",
//	     "symbol": "WETH",
//	     "category": "token",
//	     "description": "WETH (WETH) is a cryptocurrency and operates on the Ethereum platform. WETH has a current supply of 3,375,317.5926469. The last known price of WETH is 2,485.55761368 USD and is down -3.66 over the last 24 hours. It is currently trading on 19073 active market(s) with $997,029,114.23 traded over the last 24 hours. More information can be found at https://weth.io/.",
//	     "slug": "weth",
//	     "logo": "https://s2.coinmarketcap.com/static/img/coins/64x64/2396.png",
//	     "subreddit": "",
//	     "notice": "",
//	     "tags": [
//	       "wrapped-tokens",
//	       "arbitrum-ecosytem",
//	       "optimism-ecosystem",
//	       "linea-ecosystem",
//	       "rehypothecated-crypto"
//	     ],
//	     "tag-names": [
//	       "Wrapped Tokens",
//	       "Arbitrum Ecosystem",
//	       "Optimism Ecosystem",
//	       "Linea Ecosystem",
//	       "Rehypothecated Crypto"
//	     ],
//	     "tag-groups": [
//	       "PLATFORM",
//	       "PLATFORM",
//	       "PLATFORM",
//	       "PLATFORM",
//	       "CATEGORY"
//	     ],
//	     "urls": {
//	       "website": [
//	         "https://weth.io/"
//	       ],
//	       "twitter": [],
//	       "message_board": [],
//	       "chat": [],
//	       "facebook": [],
//	       "explorer": [
//	         "https://solscan.io/token/7vfCXTUXx5WJV5JADk17DUJ4ksgau7utNKj4b963voxs",
//	         "https://etherscan.io/token/0xc02aaa39b223fe8d0a0e5c4f27ead9083c756cc2",
//	         "https://polygonscan.com/token/0x7ceb23fd6bc0add59e62ac25578270cff1b9f619",
//	         "https://ftmscan.com/token/0x74b23882a30290451A17c44f4F05243b6b58C76d",
//	         "https://nearblocks.io/token/c02aaa39b223fe8d0a0e5c4f27ead9083c756cc2.factory.bridge.near"
//	       ],
//	       "reddit": [],
//	       "technical_doc": [],
//	       "source_code": [],
//	       "announcement": []
//	     },
//	     "platform": {
//	       "id": "1027",
//	       "name": "Ethereum",
//	       "slug": "ethereum",
//	       "symbol": "ETH",
//	       "token_address": "0xc02aaa39b223fe8d0a0e5c4f27ead9083c756cc2"
//	     },
//	     "date_added": "2018-01-14T00:00:00.000Z",
//	     "twitter_username": "",
//	     "is_hidden": 0,
//	     "date_launched": null,
//	     "contract_address": [
//	       {
//	         "contract_address": "0xc02aaa39b223fe8d0a0e5c4f27ead9083c756cc2",
//	         "platform": {
//	           "name": "Ethereum",
//	           "coin": {
//	             "id": "1027",
//	             "name": "Ethereum",
//	             "symbol": "ETH",
//	             "slug": "ethereum"
//	           }
//	         }
//	       },
//	       {
//	         "contract_address": "0x74b23882a30290451A17c44f4F05243b6b58C76d",
//	         "platform": {
//	           "name": "Fantom",
//	           "coin": {
//	             "id": "3513",
//	             "name": "Fantom",
//	             "symbol": "FTM",
//	             "slug": "fantom"
//	           }
//	         }
//	       },
//	       {
//	         "contract_address": "0x6a023ccd1ff6f2045c3309768ead9e68f978f6e1",
//	         "platform": {
//	           "name": "Gnosis Chain",
//	           "coin": {
//	             "id": "1659",
//	             "name": "Gnosis",
//	             "symbol": "GNO",
//	             "slug": "gnosis-gno"
//	           }
//	         }
//	       },
//	       {
//	         "contract_address": "0x7ceb23fd6bc0add59e62ac25578270cff1b9f619",
//	         "platform": {
//	           "name": "Polygon",
//	           "coin": {
//	             "id": "3890",
//	             "name": "Polygon",
//	             "symbol": "MATIC",
//	             "slug": "polygon"
//	           }
//	         }
//	       },
//	       {
//	         "contract_address": "0x82af49447d8a07e3bd95bd0d56f35241523fbab1",
//	         "platform": {
//	           "name": "Arbitrum",
//	           "coin": {
//	             "id": "11841",
//	             "name": "Arbitrum",
//	             "symbol": "ARB",
//	             "slug": "arbitrum"
//	           }
//	         }
//	       },
//	       {
//	         "contract_address": "0x49D5c2BdFfac6CE2BFdB6640F4F80f226bc10bAB",
//	         "platform": {
//	           "name": "Avalanche C-Chain",
//	           "coin": {
//	             "id": "5805",
//	             "name": "Avalanche",
//	             "symbol": "AVAX",
//	             "slug": "avalanche"
//	           }
//	         }
//	       },
//	       {
//	         "contract_address": "zil19j33tapjje2xzng7svslnsjjjgge930jx0w09v",
//	         "platform": {
//	           "name": "Zilliqa",
//	           "coin": {
//	             "id": "2469",
//	             "name": "Zilliqa",
//	             "symbol": "ZIL",
//	             "slug": "zilliqa"
//	           }
//	         }
//	       },
//	       {
//	         "contract_address": "0xe44Fd7fCb2b1581822D0c862B68222998a0c299a",
//	         "platform": {
//	           "name": "Cronos",
//	           "coin": {
//	             "id": "3635",
//	             "name": "Cronos",
//	             "symbol": "CRO",
//	             "slug": "cronos"
//	           }
//	         }
//	       },
//	       {
//	         "contract_address": "0x6983d1e6def3690c4d616b13597a09e6193ea013",
//	         "platform": {
//	           "name": "Harmony",
//	           "coin": {
//	             "id": "3945",
//	             "name": "Harmony",
//	             "symbol": "ONE",
//	             "slug": "harmony"
//	           }
//	         }
//	       },
//	       {
//	         "contract_address": "0xdeaddeaddeaddeaddeaddeaddeaddeaddead0000",
//	         "platform": {
//	           "name": "Boba Network",
//	           "coin": {
//	             "id": "14556",
//	             "name": "Boba Network",
//	             "symbol": "BOBA",
//	             "slug": "boba-network"
//	           }
//	         }
//	       },
//	       {
//	         "contract_address": "7vfCXTUXx5WJV5JADk17DUJ4ksgau7utNKj4b963voxs",
//	         "platform": {
//	           "name": "Solana",
//	           "coin": {
//	             "id": "5426",
//	             "name": "Solana",
//	             "symbol": "SOL",
//	             "slug": "solana"
//	           }
//	         }
//	       },
//	       {
//	         "contract_address": "0x4DB5a66E937A9F4473fA95b1cAF1d1E1D62E29EA",
//	         "platform": {
//	           "name": "BNB Smart Chain (BEP20)",
//	           "coin": {
//	             "id": "1839",
//	             "name": "BNB",
//	             "symbol": "BNB",
//	             "slug": "bnb"
//	           }
//	         }
//	       },
//	       {
//	         "contract_address": "terra14tl83xcwqjy0ken9peu4pjjuu755lrry2uy25r",
//	         "platform": {
//	           "name": "Terra Classic",
//	           "coin": {
//	             "id": "4172",
//	             "name": "Terra Classic",
//	             "symbol": "LUNC",
//	             "slug": "terra-luna"
//	           }
//	         }
//	       },
//	       {
//	         "contract_address": "0xc99a6a985ed2cac1ef41640596c5a5f9f4e19ef5",
//	         "platform": {
//	           "name": "Ronin",
//	           "coin": {
//	             "id": "14101",
//	             "name": "Ronin",
//	             "symbol": "RON",
//	             "slug": "ronin"
//	           }
//	         }
//	       },
//	       {
//	         "contract_address": "0x122013fd7dF1C6F636a5bb8f03108E876548b455",
//	         "platform": {
//	           "name": "Celo",
//	           "coin": {
//	             "id": "5567",
//	             "name": "Celo",
//	             "symbol": "CELO",
//	             "slug": "celo"
//	           }
//	         }
//	       },
//	       {
//	         "contract_address": "0xfa9343c3897324496a05fc75abed6bac29f8a40f",
//	         "platform": {
//	           "name": "Moonbeam",
//	           "coin": {
//	             "id": "6836",
//	             "name": "Moonbeam",
//	             "symbol": "GLMR",
//	             "slug": "moonbeam"
//	           }
//	         }
//	       },
//	       {
//	         "contract_address": "0xc9bdeed33cd01541e1eed10f90519d2c06fe3feb",
//	         "platform": {
//	           "name": "Aurora",
//	           "coin": {
//	             "id": "14803",
//	             "name": "Aurora",
//	             "symbol": "AURORA",
//	             "slug": "aurora-near"
//	           }
//	         }
//	       },
//	       {
//	         "contract_address": "0xA0fB8cd450c8Fd3a11901876cD5f17eB47C6bc50",
//	         "platform": {
//	           "name": "Telos",
//	           "coin": {
//	             "id": "4660",
//	             "name": "Telos",
//	             "symbol": "TLOS",
//	             "slug": "telos"
//	           }
//	         }
//	       },
//	       {
//	         "contract_address": "0x420000000000000000000000000000000000000a",
//	         "platform": {
//	           "name": "Metis Andromeda",
//	           "coin": {
//	             "id": "9640",
//	             "name": "Metis",
//	             "symbol": "METIS",
//	             "slug": "metisdao"
//	           }
//	         }
//	       },
//	       {
//	         "contract_address": "0xA1588dC914e236bB5AE4208Ce3081246f7A00193",
//	         "platform": {
//	           "name": "Hoo Smart Chain",
//	           "coin": {
//	             "id": "15165",
//	             "name": "Hoo Smart Chain",
//	             "symbol": "HSC",
//	             "slug": "hoo-smart-chain"
//	           }
//	         }
//	       },
//	       {
//	         "contract_address": "0x3223f17957Ba502cbe71401D55A0DB26E5F7c68F",
//	         "platform": {
//	           "name": "Oasis Network",
//	           "coin": {
//	             "id": "7653",
//	             "name": "Oasis",
//	             "symbol": "ROSE",
//	             "slug": "oasis-network"
//	           }
//	         }
//	       },
//	       {
//	         "contract_address": "0xa722c13135930332eb3d749b2f0906559d2c5b99",
//	         "platform": {
//	           "name": "Fuse",
//	           "coin": {
//	             "id": "5634",
//	             "name": "Fuse",
//	             "symbol": "FUSE",
//	             "slug": "fuse-network"
//	           }
//	         }
//	       },
//	       {
//	         "contract_address": "0xf55aF137A98607F7ED2eFEfA4cd2DfE70E4253b1",
//	         "platform": {
//	           "name": "KCC",
//	           "coin": {
//	             "id": "2087",
//	             "name": "KuCoin Token",
//	             "symbol": "KCS",
//	             "slug": "kucoin-token"
//	           }
//	         }
//	       },
//	       {
//	         "contract_address": "0x802c3e839E4fDb10aF583E3E759239ec7703501e",
//	         "platform": {
//	           "name": "Elastos",
//	           "coin": {
//	             "id": "2492",
//	             "name": "Elastos",
//	             "symbol": "ELA",
//	             "slug": "elastos"
//	           }
//	         }
//	       },
//	       {
//	         "contract_address": "0x0258866edaf84d6081df17660357ab20a07d0c80",
//	         "platform": {
//	           "name": "IoTex",
//	           "coin": {
//	             "id": "2777",
//	             "name": "IoTeX",
//	             "symbol": "IOTX",
//	             "slug": "iotex"
//	           }
//	         }
//	       },
//	       {
//	         "contract_address": "0x1540020a94aa8bc189aa97639da213a4ca49d9a7",
//	         "platform": {
//	           "name": "KardiaChain",
//	           "coin": {
//	             "id": "5453",
//	             "name": "KardiaChain",
//	             "symbol": "KAI",
//	             "slug": "kardiachain"
//	           }
//	         }
//	       },
//	       {
//	         "contract_address": "0x81ecac0d6be0550a00ff064a4f9dd2400585fe9c",
//	         "platform": {
//	           "name": "Milkomeda",
//	           "coin": {
//	             "id": "2010",
//	             "name": "Cardano",
//	             "symbol": "ADA",
//	             "slug": "cardano"
//	           }
//	         }
//	       },
//	       {
//	         "contract_address": "0x79a61d3a28f8c8537a3df63092927cfa1150fb3c",
//	         "platform": {
//	           "name": "Meter",
//	           "coin": {
//	             "id": "5919",
//	             "name": "Meter Governance",
//	             "symbol": "MTRG",
//	             "slug": "meter-governance"
//	           }
//	         }
//	       },
//	       {
//	         "contract_address": "c02aaa39b223fe8d0a0e5c4f27ead9083c756cc2.factory.bridge.near",
//	         "platform": {
//	           "name": "Near",
//	           "coin": {
//	             "id": "6535",
//	             "name": "NEAR Protocol",
//	             "symbol": "NEAR",
//	             "slug": "near-protocol"
//	           }
//	         }
//	       },
//	       {
//	         "contract_address": "0x4200000000000000000000000000000000000006",
//	         "platform": {
//	           "name": "Optimism",
//	           "coin": {
//	             "id": "11840",
//	             "name": "Optimism",
//	             "symbol": "OP",
//	             "slug": "optimism-ethereum"
//	           }
//	         }
//	       },
//	       {
//	         "contract_address": "0:59b6b64ac6798aacf385ae9910008a525a84fc6dcf9f942ae81f8e8485fe160d",
//	         "platform": {
//	           "name": "Everscale",
//	           "coin": {
//	             "id": "7505",
//	             "name": "Everscale",
//	             "symbol": "EVER",
//	             "slug": "everscale"
//	           }
//	         }
//	       },
//	       {
//	         "contract_address": "0xa47f43de2f9623acb395ca4905746496d2014d57",
//	         "platform": {
//	           "name": "Conflux",
//	           "coin": {
//	             "id": "7334",
//	             "name": "Conflux",
//	             "symbol": "CFX",
//	             "slug": "conflux-network"
//	           }
//	         }
//	       },
//	       {
//	         "contract_address": "474jTeYx2r2Va35794tCScAXWJG9hU2HcgxzMowaZUnu",
//	         "platform": {
//	           "name": "Waves",
//	           "coin": {
//	             "id": "1274",
//	             "name": "Waves",
//	             "symbol": "WAVES",
//	             "slug": "waves"
//	           }
//	         }
//	       },
//	       {
//	         "contract_address": "0x576fDe3f61B7c97e381c94e7A03DBc2e08Af1111",
//	         "platform": {
//	           "name": "Moonriver",
//	           "coin": {
//	             "id": "9285",
//	             "name": "Moonriver",
//	             "symbol": "MOVR",
//	             "slug": "moonriver"
//	           }
//	         }
//	       },
//	       {
//	         "contract_address": "0x5FD55A1B9FC24967C4dB09C513C3BA0DFa7FF687",
//	         "platform": {
//	           "name": "Canto",
//	           "coin": {
//	             "id": "21516",
//	             "name": "CANTO",
//	             "symbol": "CANTO",
//	             "slug": "canto"
//	           }
//	         }
//	       },
//	       {
//	         "contract_address": "0xc02aaa39b223fe8d0a0e5c4f27ead9083c756cc2",
//	         "platform": {
//	           "name": "EthereumPoW",
//	           "coin": {
//	             "id": "21296",
//	             "name": "EthereumPoW",
//	             "symbol": "ETHW",
//	             "slug": "ethereum-pow"
//	           }
//	         }
//	       },
//	       {
//	         "contract_address": "0xf22bede237a07e121b56d91a491eb7bcdfd1f5907926a9e58338f964a01b17fa::asset::WETH",
//	         "platform": {
//	           "name": "Aptos",
//	           "coin": {
//	             "id": "21794",
//	             "name": "Aptos",
//	             "symbol": "APT",
//	             "slug": "aptos"
//	           }
//	         }
//	       },
//	       {
//	         "contract_address": "0x5aea5775959fbc2557cc8789bc1bf90a239d9a91",
//	         "platform": {
//	           "name": "zkSync Era",
//	           "coin": {
//	             "id": "24091",
//	             "name": "zkSync",
//	             "symbol": "ZK",
//	             "slug": "zksync"
//	           }
//	         }
//	       },
//	       {
//	         "contract_address": "0x4f9a0e7fd2bf6067db6994cf12e4495df938e6e9",
//	         "platform": {
//	           "name": "Polygon zkEVM",
//	           "coin": {
//	             "id": "3890",
//	             "name": "Polygon",
//	             "symbol": "MATIC",
//	             "slug": "polygon"
//	           }
//	         }
//	       },
//	       {
//	         "contract_address": "0x765277EebeCA2e31912C9946eAe1021199B39C61",
//	         "platform": {
//	           "name": "Wemix",
//	           "coin": {
//	             "id": "7548",
//	             "name": "WEMIX",
//	             "symbol": "WEMIX",
//	             "slug": "wemix"
//	           }
//	         }
//	       },
//	       {
//	         "contract_address": "0x4200000000000000000000000000000000000006",
//	         "platform": {
//	           "name": "Base",
//	           "coin": {
//	             "id": "27716",
//	             "name": "Base",
//	             "symbol": "TBA",
//	             "slug": "base"
//	           }
//	         }
//	       },
//	       {
//	         "contract_address": "0xe5d7c2a44ffddf6b295a15c148167daaaf5cf34f",
//	         "platform": {
//	           "name": "Linea",
//	           "coin": {
//	             "id": "27657",
//	             "name": "Linea",
//	             "symbol": "TBA",
//	             "slug": "linea"
//	           }
//	         }
//	       },
//	       {
//	         "contract_address": "0x5300000000000000000000000000000000000004",
//	         "platform": {
//	           "name": "Scroll",
//	           "coin": {
//	             "id": "26998",
//	             "name": "Scroll",
//	             "symbol": "SCROLL",
//	             "slug": "scroll"
//	           }
//	         }
//	       },
//	       {
//	         "contract_address": "0xdeaddeaddeaddeaddeaddeaddeaddeaddead1111",
//	         "platform": {
//	           "name": "Mantle",
//	           "coin": {
//	             "id": "27075",
//	             "name": "Mantle",
//	             "symbol": "MNT",
//	             "slug": "mantle"
//	           }
//	         }
//	       },
//	       {
//	         "contract_address": "0x4300000000000000000000000000000000000004",
//	         "platform": {
//	           "name": "Blast",
//	           "coin": {
//	             "id": "28480",
//	             "name": "Blast",
//	             "symbol": "BLAST",
//	             "slug": "blast"
//	           }
//	         }
//	       },
//	       {
//	         "contract_address": "0x4200000000000000000000000000000000000006",
//	         "platform": {
//	           "name": "Mode",
//	           "coin": {
//	             "id": "31016",
//	             "name": "Mode",
//	             "symbol": "MODE",
//	             "slug": "mode"
//	           }
//	         }
//	       },
//	       {
//	         "contract_address": "0x5a77f1443d16ee5761d310e38b62f77f726bc71c",
//	         "platform": {
//	           "name": "X Layer",
//	           "coin": {
//	             "id": "3897",
//	             "name": "OKB",
//	             "symbol": "OKB",
//	             "slug": "okb"
//	           }
//	         }
//	       },
//	       {
//	         "contract_address": "0xa51894664a773981c6c112c43ce576f315d5b1b6",
//	         "platform": {
//	           "name": "Taiko",
//	           "coin": {
//	             "id": "31525",
//	             "name": "Taiko",
//	             "symbol": "TAIKO",
//	             "slug": "taiko"
//	           }
//	         }
//	       },
//	       {
//	         "contract_address": "0x160345fc359604fc6e70e3c5facbde5f7a9342d8",
//	         "platform": {
//	           "name": "Sei V2",
//	           "coin": {
//	             "id": "23149",
//	             "name": "Sei",
//	             "symbol": "SEI",
//	             "slug": "sei"
//	           }
//	         }
//	       }
//	     ],
//	     "self_reported_circulating_supply": null,
//	     "self_reported_tags": null,
//	     "self_reported_market_cap": null,
//	     "infinite_supply": false
//	   }
//	 }
//	}
type InfoResponse struct {
	Status Status      `json:"status"`
	Data   InfoDataMap `json:"data"`
}

type InfoDataMap map[string]InfoData

type InfoData struct {
	ID              int               `json:"id"`
	Name            string            `json:"name"`
	Symbol          string            `json:"symbol"`
	Category        string            `json:"category"`
	Description     string            `json:"description"`
	Slug            string            `json:"slug"`
	ContractAddress []ContractAddress `json:"contract_address"`
}

type ContractAddress struct {
	ContractAddress string   `json:"contract_address"`
	Platform        Platform `json:"platform"`
}

type Platform struct {
	Name string `json:"name"`
	Coin struct {
		ID     string `json:"id"`
		Name   string `json:"name"`
		Symbol string `json:"symbol"`
		Slug   string `json:"slug"`
	} `json:"coin"`
}

const (
	exchangeStatusActive   = 1
	exchangeStatusInactive = 0
)

type Status struct {
	Timestamp    time.Time `json:"timestamp"`
	ErrorCode    int       `json:"error_code"`
	ErrorMessage string    `json:"error_message"`
	Elapsed      int       `json:"elapsed"`
	CreditCount  int       `json:"credit_count"`
}

func (s *Status) Validate() error {
	switch s.ErrorCode {
	case http.StatusOK:
		return nil
	case 0:
		return nil
	default:
		return fmt.Errorf("invalid http response: %s: %s", http.StatusText(s.ErrorCode), s.ErrorMessage)
	}
}
