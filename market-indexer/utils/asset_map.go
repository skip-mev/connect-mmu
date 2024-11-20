package utils

import (
	"github.com/skip-mev/connect-mmu/store/provider"
)

// AssetMap: asset symbol -> asset address ("" if not defi) -> asset info
type AssetMap map[string]AssetSubMap

// AssetSubMap: asset address ("" if not defi) -> asset info
type AssetSubMap map[string]provider.AssetInfo

// LookupAssetInfo wraps accessing an AssetMap by symbol and address.
func (m AssetMap) LookupAssetInfo(symbol, address string) (provider.AssetInfo, bool) {
	_, found := m[symbol]
	if !found {
		m[symbol] = make(map[string]provider.AssetInfo)
	}

	info, found := m[symbol][address]
	if !found {
		return provider.AssetInfo{}, false
	}

	return info, true
}

// LookupByCMCID wraps accessing an AssetInfo by a CoinMarketCap ID.
func (m AssetMap) LookupByCMCID(symbol string, cmcID int64) (provider.AssetInfo, bool) {
	for _, subMap := range m {
		for _, info := range subMap {
			if info.CMCID == cmcID && info.Symbol == symbol {
				return info, true
			}
		}
	}
	return provider.AssetInfo{}, false
}

// AddAssetFromInfo adds an asset to the underlying map from the given AssetInfo.
func (m AssetMap) AddAssetFromInfo(asset provider.AssetInfo) {
	multiAddresses := asset.MultiAddresses
	if len(multiAddresses) == 0 {
		multiAddresses = append(multiAddresses, []string{"unknown", ""})
	}

	for _, array := range multiAddresses {
		assetAddress := MustAssetAddressFromArray(array)
		got, found := m.LookupAssetInfo(asset.Symbol, assetAddress.Address)
		if !found {
			m[asset.Symbol][assetAddress.Address] = asset
		} else {
			// if we already had an entry, only replace if the new one is better rank
			m[asset.Symbol][assetAddress.Address] = compareAssetRank(got, asset)
		}
	}
}

// compareAssetRank compares two AssetInfos based on rank (lower rank is "better")
// if the field is null for one of the AssetInfos, AssetInfo a is returned.
// If both are null, AssetInfo a is returned.
func compareAssetRank(a, b provider.AssetInfo) provider.AssetInfo {
	if a.Rank <= b.Rank {
		return a
	}

	return b
}
