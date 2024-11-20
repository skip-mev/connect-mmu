package indexer

import (
	"context"
	"errors"
	"fmt"

	"github.com/skip-mev/connect-mmu/lib/symbols"
	"github.com/skip-mev/connect-mmu/market-indexer/coinmarketcap"
)

func (idx *Indexer) CheckPair(
	ctx context.Context,
	marketPair coinmarketcap.ProviderMarketData,
) (int32, int32, error) {
	// closure to get the AssetInfo ID for a symbol.  If it does not exist, a new entry is created
	// in the db.
	findOrCreatesAssetInfo := func(symbol string, cmcID int64) (int32, error) {
		symbol, err := symbols.ToTickerString(symbol)
		if err != nil {
			return -1, errors.New("failed to convert symbol")
		}

		if info, found := idx.knownAssets.LookupByCMCID(symbol, cmcID); found {
			return info.ID, nil
		}

		// if not found, we should create the entry for this asset with market pair data
		quoteData, err := idx.cmcIndexer.Quote(ctx, cmcID)
		if err != nil {
			return -1, fmt.Errorf("failed to get quote: symbol: %s pair %v",
				symbol, marketPair)
		}

		createInfo, err := idx.providerStore.AddAssetInfo(ctx, CryptoAssetInfoFromQuoteData(quoteData))
		if err != nil {
			return -1, fmt.Errorf("error creating aggregator info for %q: %w", symbol, err)
		}
		idx.knownAssets.AddAssetFromInfo(createInfo)

		return createInfo.ID, nil
	}

	baseAssetInfoID, err := findOrCreatesAssetInfo(marketPair.BaseAsset, marketPair.CMCInfo.BaseID)
	if err != nil {
		return -1, -1, err
	}
	quoteAssetInfoID, err := findOrCreatesAssetInfo(marketPair.QuoteAsset, marketPair.CMCInfo.QuoteID)
	if err != nil {
		return -1, -1, err
	}

	return baseAssetInfoID, quoteAssetInfoID, nil
}
