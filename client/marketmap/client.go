package marketmap

import (
	"context"

	mmtypes "github.com/skip-mev/connect/v2/x/marketmap/types"
)

// Client is a client that provides a market-map from an external source.
//
//go:generate mockery --name=Client --output=mocks --case=underscore
type Client interface {
	// GetMarketMap retrieves a market-map from an external source.
	GetMarketMap(ctx context.Context) (mmtypes.MarketMap, error)
}
