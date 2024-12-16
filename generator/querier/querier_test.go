package querier_test

import (
	"context"
	"math/big"
	"testing"

	connecttypes "github.com/skip-mev/connect/v2/pkg/types"
	mmtypes "github.com/skip-mev/connect/v2/x/marketmap/types"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/skip-mev/connect-mmu/config"
	"github.com/skip-mev/connect-mmu/generator/querier"
	"github.com/skip-mev/connect-mmu/generator/types"
	"github.com/skip-mev/connect-mmu/store/provider"
	mmutypes "github.com/skip-mev/connect-mmu/types"
)

var expectedFeeds = types.Feeds{
	types.Feed{
		Ticker: mmtypes.Ticker{
			CurrencyPair: connecttypes.CurrencyPair{
				Base:  "ETH",
				Quote: "USD",
			},
			Decimals:         8,
			MinProviderCount: 0,
			Enabled:          false,
			Metadata_JSON:    "",
		},
		ProviderConfig: mmtypes.ProviderConfig{
			Name:           "coinbase",
			OffChainTicker: "ETH-USD",
		},
		DailyQuoteVolume: big.NewFloat(1000.0),
		ReferencePrice:   big.NewFloat(100.0),
		CMCInfo: mmutypes.CoinMarketCapInfo{
			BaseID:    0,
			QuoteID:   1,
			BaseRank:  0,
			QuoteRank: 1,
		},
		LiquidityInfo: mmutypes.LiquidityInfo{
			NegativeDepthTwo: 100,
			PositiveDepthTwo: 100,
		},
	},
	types.Feed{
		Ticker: mmtypes.Ticker{
			CurrencyPair: connecttypes.CurrencyPair{
				Base:  "BTC",
				Quote: "USD",
			},
			Decimals:         8,
			MinProviderCount: 0,
			Enabled:          false,
			Metadata_JSON:    "",
		},
		ProviderConfig: mmtypes.ProviderConfig{
			Name:           "coinbase",
			OffChainTicker: "BTC-USD",
		},
		DailyQuoteVolume: big.NewFloat(1000.0),
		ReferencePrice:   big.NewFloat(100.0),
		CMCInfo: mmutypes.CoinMarketCapInfo{
			BaseID:    2,
			QuoteID:   1,
			BaseRank:  2,
			QuoteRank: 1,
		},
		LiquidityInfo: mmutypes.LiquidityInfo{
			NegativeDepthTwo: 100,
			PositiveDepthTwo: 100,
		},
	},
}

func TestFeeds(t *testing.T) {
	store := provider.NewMemoryStore()
	ctx := context.Background()
	ids := createAssets(ctx, t, store, []string{"ETH", "USD", "BTC"})

	_, err := store.AddProviderMarket(ctx, provider.CreateProviderMarketParams{
		TargetBase:       "ETH",
		TargetQuote:      "USD",
		OffChainTicker:   "ETH-USD",
		ProviderName:     "coinbase",
		BaseAssetInfoID:  ids[0],
		QuoteAssetInfoID: ids[1],
		QuoteVolume:      1000,
		NegativeDepthTwo: 100,
		PositiveDepthTwo: 100,
		ReferencePrice:   100,
	})
	require.NoError(t, err)
	_, err = store.AddProviderMarket(ctx, provider.CreateProviderMarketParams{
		TargetBase:       "BTC",
		TargetQuote:      "USD",
		OffChainTicker:   "BTC-USD",
		ProviderName:     "coinbase",
		BaseAssetInfoID:  ids[2],
		QuoteAssetInfoID: ids[1],
		QuoteVolume:      1000,
		NegativeDepthTwo: 100,
		PositiveDepthTwo: 100,
		ReferencePrice:   100,
	})
	require.NoError(t, err)
	log, err := zap.NewDevelopment()
	require.NoError(t, err)
	qr := querier.New(log, store)

	t.Run("get no feeds for empty query", func(t *testing.T) {
		feeds, err := qr.Feeds(ctx, config.GenerateConfig{})
		require.NoError(t, err)
		require.Len(t, feeds, 0)
	})

	t.Run("get no feeds for query with nonexistent Provider", func(t *testing.T) {
		feeds, err := qr.Feeds(ctx, config.GenerateConfig{Providers: map[string]config.ProviderConfig{"invalid": {}}})
		require.NoError(t, err)
		require.Len(t, feeds, 0)
	})

	t.Run("get all feeds for query with provider", func(t *testing.T) {
		feeds, err := qr.Feeds(ctx, config.GenerateConfig{Providers: map[string]config.ProviderConfig{"coinbase": {}}})
		require.NoError(t, err)
		require.Len(t, feeds, 2)
		expectedFeeds.Sort()
		feeds.Sort()
		require.True(t, expectedFeeds.Equal(feeds))
	})
}

func createAssets(ctx context.Context, t *testing.T, store provider.Store, assets []string) []int32 {
	t.Helper()
	ids := make([]int32, 0, len(assets))
	for i, asset := range assets {
		res, err := store.AddAssetInfo(ctx, provider.CreateAssetInfoParams{
			Symbol:         asset,
			MultiAddresses: [][]string{{"UNKNOWN", ""}},
			CmcID:          int64(i),
			Rank:           int64(i),
		})
		require.NoError(t, err)
		ids = append(ids, res.ID)
	}
	return ids
}
