package ingesters

import (
	"context"

	"github.com/skip-mev/connect-mmu/store/provider"
)

// Ingester is a general interface for a module that can ingest and parse market data for a provider
// (kraken, uniswap, etc.) to determine what markets that provider supports.
type Ingester interface {
	// GetProviderMarkets returns a list of CreateProviderMarket for the given Ingester.
	GetProviderMarkets(ctx context.Context) ([]provider.CreateProviderMarket, error)

	// Name returns the name of the Ingester.
	Name() string
}
