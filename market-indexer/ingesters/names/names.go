package names

import (
	"github.com/skip-mev/connect-mmu/market-indexer/ingesters/binance"
	"github.com/skip-mev/connect-mmu/market-indexer/ingesters/bitfinex"
	"github.com/skip-mev/connect-mmu/market-indexer/ingesters/bitstamp"
	"github.com/skip-mev/connect-mmu/market-indexer/ingesters/bybit"
	"github.com/skip-mev/connect-mmu/market-indexer/ingesters/coinbase"
	crypto_com "github.com/skip-mev/connect-mmu/market-indexer/ingesters/crypto.com"
	"github.com/skip-mev/connect-mmu/market-indexer/ingesters/gate"
	"github.com/skip-mev/connect-mmu/market-indexer/ingesters/huobi"
	"github.com/skip-mev/connect-mmu/market-indexer/ingesters/kraken"
	"github.com/skip-mev/connect-mmu/market-indexer/ingesters/kucoin"
	"github.com/skip-mev/connect-mmu/market-indexer/ingesters/mexc"
	"github.com/skip-mev/connect-mmu/market-indexer/ingesters/okx"
	"github.com/skip-mev/connect-mmu/market-indexer/ingesters/raydium"
)

const (
	nameUnknown = "UNKNOWN"
)

// GetProviderName returns a provider name from the base name of an ingester.
func GetProviderName(name string) string {
	switch name {
	case binance.Name:
		return binance.ProviderName
	case bitfinex.Name:
		return bitfinex.ProviderName
	case bitstamp.Name:
		return bitstamp.ProviderName
	case bybit.Name:
		return bybit.ProviderName
	case coinbase.Name:
		return coinbase.ProviderName
	case crypto_com.Name:
		return crypto_com.ProviderName
	case gate.Name:
		return gate.ProviderName
	case huobi.Name:
		return huobi.ProviderName
	case kucoin.Name:
		return kucoin.ProviderName
	case kraken.Name:
		return kraken.ProviderName
	case mexc.Name:
		return mexc.ProviderName
	case okx.Name:
		return okx.ProviderName
	case raydium.Name:
		return raydium.ProviderName
	default:
		return nameUnknown
	}
}
