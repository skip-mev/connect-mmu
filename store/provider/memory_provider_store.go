package provider

import (
	"context"
	"encoding/json"
	"errors"
	"sync"

	"github.com/skip-mev/connect-mmu/lib/file"
)

type MemoryStore struct {
	mu sync.Mutex

	providerMarketNextID int32
	assetInfoNextID      int32

	providerMarkets map[int32]*ProviderMarket
	assetInfos      map[int32]*AssetInfo

	providerMarketOffChainTickerProviderNameUniqueIndex map[string]map[string]int32
	assetInfoCMCIDUniqueIndex                           map[int64]int32
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		providerMarketNextID: 0,
		assetInfoNextID:      0,

		providerMarkets: make(map[int32]*ProviderMarket),
		assetInfos:      make(map[int32]*AssetInfo),

		providerMarketOffChainTickerProviderNameUniqueIndex: make(map[string]map[string]int32),
		assetInfoCMCIDUniqueIndex:                           make(map[int64]int32),
	}
}

func NewMemoryStoreFromFile(path string) (*MemoryStore, error) {
	jsonBz, err := file.ReadBytesFromFile(path)
	if err != nil {
		return nil, err
	}

	var document Document
	if err := json.Unmarshal(jsonBz, &document); err != nil {
		return nil, err
	}

	store := NewMemoryStore()

	maxAssetID := int32(-1)
	for _, assetInfo := range document.AssetInfos {
		if assetInfo.ID > maxAssetID {
			maxAssetID = assetInfo.ID
		}
		store.assetInfos[assetInfo.ID] = &AssetInfo{
			ID:             assetInfo.ID,
			Symbol:         assetInfo.Symbol,
			IsCrypto:       assetInfo.IsCrypto,
			CMCID:          assetInfo.CMCID,
			Rank:           assetInfo.Rank,
			MultiAddresses: assetInfo.MultiAddresses,
		}
	}
	store.assetInfoNextID = maxAssetID + 1

	maxProviderMarketID := int32(-1)
	for _, providerMarket := range document.ProviderMarkets {
		if providerMarket.ID > maxProviderMarketID {
			maxProviderMarketID = providerMarket.ID
		}
		store.providerMarkets[providerMarket.ID] = &ProviderMarket{
			ID:               providerMarket.ID,
			TargetBase:       providerMarket.TargetBase,
			TargetQuote:      providerMarket.TargetQuote,
			OffChainTicker:   providerMarket.OffChainTicker,
			ProviderName:     providerMarket.ProviderName,
			QuoteVolume:      providerMarket.QuoteVolume,
			BaseAssetInfoID:  providerMarket.BaseAssetInfoID,
			QuoteAssetInfoID: providerMarket.QuoteAssetInfoID,
			MetadataJSON:     providerMarket.MetadataJSON,
			ReferencePrice:   providerMarket.ReferencePrice,
			NegativeDepthTwo: providerMarket.NegativeDepthTwo,
			PositiveDepthTwo: providerMarket.PositiveDepthTwo,
		}
	}
	store.providerMarketNextID = maxProviderMarketID + 1

	return store, nil
}

func (w *MemoryStore) AddProviderMarket(_ context.Context, params CreateProviderMarketParams) (ProviderMarket, error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	if _, ok := w.providerMarketOffChainTickerProviderNameUniqueIndex[params.OffChainTicker]; ok {
		if id, ok := w.providerMarketOffChainTickerProviderNameUniqueIndex[params.OffChainTicker][params.ProviderName]; ok {
			return w.updateProviderMarket(params, id)
		}
	}

	providerMarket := ProviderMarket{
		ID:               w.providerMarketNextID,
		TargetBase:       params.TargetBase,
		TargetQuote:      params.TargetQuote,
		OffChainTicker:   params.OffChainTicker,
		ProviderName:     params.ProviderName,
		QuoteVolume:      params.QuoteVolume,
		BaseAssetInfoID:  params.BaseAssetInfoID,
		QuoteAssetInfoID: params.QuoteAssetInfoID,
		MetadataJSON:     string(params.MetadataJSON),
		ReferencePrice:   params.ReferencePrice,
		NegativeDepthTwo: params.NegativeDepthTwo,
		PositiveDepthTwo: params.PositiveDepthTwo,
	}

	w.providerMarketNextID++
	w.providerMarkets[providerMarket.ID] = &providerMarket
	if _, ok := w.providerMarketOffChainTickerProviderNameUniqueIndex[params.OffChainTicker]; !ok {
		w.providerMarketOffChainTickerProviderNameUniqueIndex[params.OffChainTicker] = make(map[string]int32)
	}
	w.providerMarketOffChainTickerProviderNameUniqueIndex[params.OffChainTicker][params.ProviderName] = providerMarket.ID

	return providerMarket, nil
}

func (w *MemoryStore) updateProviderMarket(params CreateProviderMarketParams, id int32) (ProviderMarket, error) {
	providerMarket, ok := w.providerMarkets[id]
	if !ok {
		return ProviderMarket{}, errors.New("attempted to update provider market at invalid id")
	}
	// Don't overwrite if the quote volume is lower.
	// e.g. we can have multiple provider markets for the same uniswap ticker because there can be multiple fee pools
	if providerMarket.QuoteVolume > params.QuoteVolume {
		return *providerMarket, nil
	}

	providerMarket.QuoteVolume = params.QuoteVolume
	providerMarket.BaseAssetInfoID = params.BaseAssetInfoID
	providerMarket.QuoteAssetInfoID = params.QuoteAssetInfoID
	providerMarket.ReferencePrice = params.ReferencePrice
	providerMarket.NegativeDepthTwo = params.NegativeDepthTwo
	providerMarket.PositiveDepthTwo = params.PositiveDepthTwo
	providerMarket.MetadataJSON = string(params.MetadataJSON)

	return *providerMarket, nil
}

func (w *MemoryStore) AddAssetInfo(_ context.Context, params CreateAssetInfoParams) (AssetInfo, error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	if id, ok := w.assetInfoCMCIDUniqueIndex[params.CmcID]; ok {
		return w.updateAssetInfo(params, id)
	}

	assetInfo := AssetInfo{
		ID: w.assetInfoNextID,

		Symbol:         params.Symbol,
		IsCrypto:       true, // it doesn't look like we actually use this
		CMCID:          params.CmcID,
		Rank:           params.Rank,
		MultiAddresses: params.MultiAddresses,
	}

	w.assetInfoNextID++
	w.assetInfos[assetInfo.ID] = &assetInfo
	w.assetInfoCMCIDUniqueIndex[params.CmcID] = assetInfo.ID

	return assetInfo, nil
}

func (w *MemoryStore) updateAssetInfo(params CreateAssetInfoParams, id int32) (AssetInfo, error) {
	assetInfo, ok := w.assetInfos[id]
	if !ok {
		return AssetInfo{}, errors.New("attempted to update asset info at invalid id")
	}

	assetInfo.Rank = params.Rank
	assetInfo.MultiAddresses = params.MultiAddresses

	return *assetInfo, nil
}

func (w *MemoryStore) GetProviderMarkets(_ context.Context, params GetFilteredProviderMarketsParams) ([]GetFilteredProviderMarketsRow, error) {
	targetProviderNames := make(map[string]struct{})
	for _, providerName := range params.ProviderNames {
		targetProviderNames[providerName] = struct{}{}
	}

	rows := make([]GetFilteredProviderMarketsRow, 0)
	for _, providerMarket := range w.providerMarkets {
		if _, ok := targetProviderNames[providerMarket.ProviderName]; !ok {
			continue
		}

		baseAssetInfo, ok := w.assetInfos[providerMarket.BaseAssetInfoID]
		if !ok {
			continue
		}

		quoteAssetInfo, ok := w.assetInfos[providerMarket.QuoteAssetInfoID]
		if !ok {
			continue
		}

		row := GetFilteredProviderMarketsRow{
			TargetBase:       providerMarket.TargetBase,
			TargetQuote:      providerMarket.TargetQuote,
			OffChainTicker:   providerMarket.OffChainTicker,
			ProviderName:     providerMarket.ProviderName,
			QuoteVolume:      providerMarket.QuoteVolume,
			MetadataJSON:     []byte(providerMarket.MetadataJSON),
			ReferencePrice:   providerMarket.ReferencePrice,
			NegativeDepthTwo: providerMarket.NegativeDepthTwo,
			PositiveDepthTwo: providerMarket.PositiveDepthTwo,
			BaseCmcID:        baseAssetInfo.CMCID,
			QuoteCmcID:       quoteAssetInfo.CMCID,
			BaseRank:         baseAssetInfo.Rank,
			QuoteRank:        quoteAssetInfo.Rank,
		}

		rows = append(rows, row)
	}

	return rows, nil
}

func (w *MemoryStore) CreateOutputDocument() Document {
	providerMarkets := make([]ProviderMarket, 0, len(w.providerMarkets))
	for _, providerMarket := range w.providerMarkets {
		providerMarkets = append(providerMarkets, *providerMarket)
	}

	assetInfos := make([]AssetInfo, 0, len(w.assetInfos))
	for _, assetInfo := range w.assetInfos {
		assetInfos = append(assetInfos, *assetInfo)
	}

	return Document{
		ProviderMarkets: providerMarkets,
		AssetInfos:      assetInfos,
	}
}

func (w *MemoryStore) WriteToPath(_ context.Context, path string) error {
	document := w.CreateOutputDocument()

	bz, err := json.Marshal(document)
	if err != nil {
		return err
	}

	return file.WriteBytesToFile(path, bz)
}
