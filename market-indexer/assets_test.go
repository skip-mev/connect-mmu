package indexer_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	indexer "github.com/skip-mev/connect-mmu/market-indexer"
	"github.com/skip-mev/connect-mmu/market-indexer/coinmarketcap"
	"github.com/skip-mev/connect-mmu/market-indexer/utils"
	"github.com/skip-mev/connect-mmu/store/provider"
)

func TestAssetMap(t *testing.T) {
	tests := []struct {
		name  string
		asset provider.AssetInfo
	}{
		{
			name: "add single crypto asset",
			asset: provider.AssetInfo{
				Symbol:         "test",
				IsCrypto:       true,
				MultiAddresses: [][]string{{"test", "test"}},
				CMCID:          10,
			},
		},
		{
			name: "add single fiat asset",
			asset: provider.AssetInfo{
				Symbol:         "test",
				IsCrypto:       false,
				MultiAddresses: [][]string{{"test", ""}},
				CMCID:          10,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			kap := make(utils.AssetMap)
			kap.AddAssetFromInfo(tt.asset)

			got, found := kap.LookupAssetInfo(tt.asset.Symbol, tt.asset.MultiAddresses[0][1])
			require.True(t, found)
			require.Equal(t, tt.asset, got)

			foundMap, found := kap.LookupByCMCID(tt.asset.Symbol, tt.asset.CMCID)
			require.True(t, found)
			require.NotNil(t, foundMap)

			_, found = kap.LookupByCMCID("invalid", 0)
			require.False(t, found)

			_, found = kap.LookupAssetInfo("invalid", "invalid")
			require.False(t, found)
		})
	}
}

func TestCryptoAssetInfoFromData(t *testing.T) {
	tests := []struct {
		name   string
		data   coinmarketcap.WrappedCryptoIDMapData
		want   provider.CreateAssetInfoParams
		expErr bool
	}{
		{
			name: "add single crypto asset with address",
			data: coinmarketcap.WrappedCryptoIDMapData{
				IDMap: coinmarketcap.CryptoIDMapData{
					ID:                  10,
					Rank:                10,
					Name:                "test asset",
					Symbol:              "TEST",
					Slug:                "TEST",
					IsActive:            1,
					FirstHistoricalData: time.Time{},
					LastHistoricalData:  time.Time{},
					Platform: (*struct {
						ID           int    `json:"id"`
						Name         string `json:"name"`
						Symbol       string `json:"symbol"`
						Slug         string `json:"slug"`
						TokenAddress string `json:"token_address"`
					})(&struct {
						ID           int
						Name         string
						Symbol       string
						Slug         string
						TokenAddress string
					}{ID: 10, Name: "ETH", Symbol: "test", Slug: "test", TokenAddress: "address"}),
				},
				Info: coinmarketcap.InfoData{
					ContractAddress: []coinmarketcap.ContractAddress{
						{
							ContractAddress: "address",
							Platform: coinmarketcap.Platform{
								Name: "ETH",
							},
						},
					},
				},
			},
			want: provider.CreateAssetInfoParams{
				Symbol:         "TEST",
				MultiAddresses: [][]string{{"ETH", "address"}},
				CmcID:          10,
				Rank:           10,
			},
		},
		{
			name: "add single crypto asset with address - no platform",
			data: coinmarketcap.WrappedCryptoIDMapData{
				IDMap: coinmarketcap.CryptoIDMapData{
					ID:                  10,
					Rank:                10,
					Name:                "test asset",
					Symbol:              "TEST",
					Slug:                "TEST",
					IsActive:            1,
					FirstHistoricalData: time.Time{},
					LastHistoricalData:  time.Time{},
				},
				Info: coinmarketcap.InfoData{
					ContractAddress: []coinmarketcap.ContractAddress{},
				},
			},
			want: provider.CreateAssetInfoParams{
				Symbol:         "TEST",
				CmcID:          10,
				MultiAddresses: nil,
				Rank:           10,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := indexer.CryptoAssetInfoFromData(tt.data)
			if tt.expErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestFiatAssetInfoFromData(t *testing.T) {
	tests := []struct {
		name string
		data coinmarketcap.FiatData
		want provider.CreateAssetInfoParams
	}{
		{
			name: "create from fiat asset info",
			data: coinmarketcap.FiatData{
				ID:     10,
				Name:   "TEST",
				Sign:   "TEST",
				Symbol: "TEST",
			},
			want: provider.CreateAssetInfoParams{
				Symbol:         "TEST",
				MultiAddresses: [][]string{{indexer.VenueFiat, ""}},
				CmcID:          10,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := indexer.FiatAssetInfoFromData(tt.data)
			require.Equal(t, tt.want, got)
		})
	}
}
