package provider

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMemoryStoreWriteToPath(t *testing.T) {
	// Create initial store with test data
	store := NewMemoryStore()
	ctx := context.Background()

	// Add test asset infos
	btcAsset, err := store.AddAssetInfo(ctx, CreateAssetInfoParams{
		Symbol: "BTC",
		CmcID:  1,
		Rank:   1,
	})
	require.NoError(t, err)

	usdAsset, err := store.AddAssetInfo(ctx, CreateAssetInfoParams{
		Symbol: "USD",
		CmcID:  2,
		Rank:   2,
	})
	require.NoError(t, err)

	// Add test provider market
	original, err := store.AddProviderMarket(ctx, CreateProviderMarketParams{
		TargetBase:       "BTC",
		TargetQuote:      "USD",
		OffChainTicker:   "BTC-USD",
		ProviderName:     "test_provider",
		BaseAssetInfoID:  btcAsset.ID,
		QuoteAssetInfoID: usdAsset.ID,
		MetadataJSON:     []byte("test"),
		ReferencePrice:   100000.0,
		NegativeDepthTwo: 100.0,
		PositiveDepthTwo: 200.0,
		QuoteVolume:      1000.0,
	})
	require.NoError(t, err)

	// Create a temporary file for writing
	tmpfile, err := os.CreateTemp("", "test-store-*.json")
	require.NoError(t, err)
	defer os.Remove(tmpfile.Name())

	// Write store to file
	err = store.WriteToPath(ctx, tmpfile.Name())
	require.NoError(t, err)

	// Load the data into a new memory store
	newStore, err := NewMemoryStoreFromFile(tmpfile.Name())
	require.NoError(t, err)

	// Verify the data was preserved when querying the store read from file
	params := GetFilteredProviderMarketsParams{
		ProviderNames: []string{"test_provider"},
	}

	rows, err := newStore.GetProviderMarkets(ctx, params)
	require.NoError(t, err)
	require.Len(t, rows, 1)

	// Verify all fields are preserved correctly in the filtered rows.
	row := rows[0]
	require.Equal(t, original.TargetBase, row.TargetBase, "TargetBase should be preserved")
	require.Equal(t, original.TargetQuote, row.TargetQuote, "TargetQuote should be preserved")
	require.Equal(t, original.OffChainTicker, row.OffChainTicker, "OffChainTicker should be preserved")
	require.Equal(t, original.ProviderName, row.ProviderName, "ProviderName should be preserved")
	require.Equal(t, store.assetInfos[original.BaseAssetInfoID].CMCID, row.BaseCmcID, "BaseCmcID should be preserved")
	require.Equal(t, store.assetInfos[original.QuoteAssetInfoID].CMCID, row.QuoteCmcID, "QuoteCmcID should be preserved")
	require.Equal(t, original.MetadataJSON, string(row.MetadataJSON), "MetadataJSON should be preserved")
	require.Equal(t, original.ReferencePrice, row.ReferencePrice, "ReferencePrice should be preserved")
	require.Equal(t, original.NegativeDepthTwo, row.NegativeDepthTwo, "NegativeDepthTwo should be preserved")
	require.Equal(t, original.PositiveDepthTwo, row.PositiveDepthTwo, "PositiveDepthTwo should be preserved")
}
