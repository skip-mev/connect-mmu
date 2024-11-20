package bitfinex

import (
	"fmt"
	"strings"
)

// The following indices are used to parse the interface response
// returned from the Bitfinex API.
//
// Docs: https://docs.bitfinex.com/reference/rest-public-tickers
//
// Ex.
//
// [
//  [
//    "tBTCUSD",
//    62668,
//    7.66307317,
//    62677,
//    5.54677145,
//    -1190,
//    -0.01862868,
//    62690,
//    770.47964675,
//    64068,
//    62020
//  ],
//  [
//    "tLTCUSD",
//    83.614,
//    1417.08762132,
//    83.615,
//    1124.75543941,
//    -1.348,
//    -0.01588499,
//    83.512,
//    2211.76817607,
//    85.76,
//    81.914
//  ],

const (
	indexSymbol = iota
	indexBid
	indexBidSize
	indexAsk
	indexAskSize
	indexDailyChange
	indexDailyChangeRelative
	indexLastPrice
	indexVolume
	indexHigh
	indexLow

	symbolPrefixTrading = 't'
)

// checkAndTrimSymbol checks if the ticker is a trading pair
// and trims the trading symbol prefix.
func checkAndTrimSymbol(symbol string) (string, error) {
	trimmed := strings.TrimPrefix(symbol, string(symbolPrefixTrading))

	if trimmed == symbol {
		return "", fmt.Errorf("symbol is not for trading")
	}

	return trimmed, nil
}

// decodeSymbol splits a ticker symbol in half and returns
// base and quote.
func decodeSymbol(symbol string) (string, string, error) {
	split := strings.Split(symbol, ":")
	switch {
	case len(split) == 2:
		return split[0], split[1], nil
	case len(split) == 1:
		return checkKnownQuotes(symbol)
	default:
		return "", "", fmt.Errorf("invalid symbol %s", symbol)
	}
}

// checkKnownQuotes checks a ticker symbol against known quotes
// and returns a split base and quote.
func checkKnownQuotes(symbol string) (string, string, error) {
	for _, known := range knownQuotes {
		if strings.HasSuffix(symbol, known) {
			return strings.TrimSuffix(symbol, known), known, nil
		}
	}

	return "", "", fmt.Errorf("invalid symbol %s", symbol)
}

// replaceAliases replaces symbol aliases known from the Bitfinex API.
func replaceAliases(symbol string) string {
	switch symbol {
	case "UST":
		return "USDT"
	case "MNA":
		return "MANA"
	default:
		return symbol
	}
}

// getVolume parses volume from data interface response
// and returns the quote denominated volume.
func getVolume(data []interface{}) (float64, error) {
	baseVol, ok := data[indexVolume].(float64)
	if !ok {
		return 0, fmt.Errorf("received non-float type in response volume %v", data)
	}

	high, ok := data[indexHigh].(float64)
	if !ok {
		return 0, fmt.Errorf("received non-float type in response high %v", data)
	}

	low, ok := data[indexLow].(float64)
	if !ok {
		return 0, fmt.Errorf("received non-float type in response low %v", data)
	}

	avg := (low + high) / 2

	return baseVol * avg, nil
}
