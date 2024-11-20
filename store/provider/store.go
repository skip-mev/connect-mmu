package provider

import (
	"context"
)

//go:generate mockery --name Store  --filename mock_store.go
type Store interface {
	AddProviderMarket(ctx context.Context, params CreateProviderMarketParams) (ProviderMarket, error)
	AddAssetInfo(ctx context.Context, params CreateAssetInfoParams) (AssetInfo, error)

	// TODO(zrbecker): The params object here includes a MaxMarketAge field that is ignored by all implemented
	// proivder stores. This should be removed eventually. It should be considered an error to pass the field to this function.
	GetProviderMarkets(ctx context.Context, params GetFilteredProviderMarketsParams) ([]GetFilteredProviderMarketsRow, error)

	WriteToPath(ctx context.Context, path string) error
}
