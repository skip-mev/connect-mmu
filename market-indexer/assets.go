package indexer

import (
	"context"
	"fmt"
	"strings"

	"go.uber.org/zap"
	"golang.org/x/exp/maps"

	"github.com/skip-mev/connect-mmu/lib/file"
	"github.com/skip-mev/connect-mmu/lib/symbols"
	"github.com/skip-mev/connect-mmu/market-indexer/coinmarketcap"
	"github.com/skip-mev/connect-mmu/market-indexer/utils"
	"github.com/skip-mev/connect-mmu/store/provider"
)

const (
	ValueUnknown = "UNKNOWN"
	VenueFiat    = "fiat"
)

// SetupAssets setup up the Indexer database with AssetInfo for all assets it can scrape.
// These assets are later referenced when ProviderMarkets are indexed as they consist of a pair of two assets.
func (idx *Indexer) SetupAssets(ctx context.Context) (coinmarketcap.ProviderMarketPairs, error) {
	// index everything since we have no assets in the db
	return idx.IndexKnownAssetInfo(ctx)
}

func (idx *Indexer) IndexKnownAssetInfo(ctx context.Context) (coinmarketcap.ProviderMarketPairs, error) {
	// Get CMC crypto map data
	cmcCryptoData, err := idx.cmcIndexer.CryptoIDMap(ctx)
	if err != nil {
		return coinmarketcap.ProviderMarketPairs{}, err
	}

	if err := idx.archiveIntermediateFile(cmcCryptoData, "cmc_crypto_data.json"); err != nil {
		return coinmarketcap.ProviderMarketPairs{}, err
	}

	// Get CMC fiat map data
	cmcFiatData, err := idx.cmcIndexer.FiatIDMap(ctx)
	if err != nil {
		return coinmarketcap.ProviderMarketPairs{}, err
	}

	if err := idx.archiveIntermediateFile(cmcFiatData, "cmc_fiat_data.json"); err != nil {
		return coinmarketcap.ProviderMarketPairs{}, err
	}

	// TODO: get other aggregator sources
	idx.knownAssets = make(utils.AssetMap, len(cmcCryptoData)+len(cmcFiatData))

	// create entries for known non-crypto assets
	for _, data := range cmcFiatData {
		create := FiatAssetInfoFromData(data)
		info, err := idx.providerStore.AddAssetInfo(ctx, create)
		if err != nil {
			idx.logger.Error("error creating aggregator info", zap.Error(err), zap.Any("data", data), zap.Any("create", create))
			return coinmarketcap.ProviderMarketPairs{}, err
		}

		idx.knownAssets.AddAssetFromInfo(info)
	}

	// set crypto Asset data from aggregators in db
	for _, data := range cmcCryptoData {
		create, err := CryptoAssetInfoFromData(data)
		if err != nil {
			idx.logger.Debug("unable to create asset info from data", zap.Error(err), zap.Any("data", data), zap.Any("create", create))
			continue
		}
		info, err := idx.providerStore.AddAssetInfo(ctx, create)
		if err != nil {
			idx.logger.Error("error creating aggregator info", zap.Error(err), zap.Any("data", data), zap.Any("create", create))
			return coinmarketcap.ProviderMarketPairs{}, err
		}

		idx.knownAssets.AddAssetFromInfo(info)
	}

	// iterate through market pairs we care about and add any extra info to the DB:
	cmcMarketPairs, err := idx.cmcIndexer.GetProviderMarketsPairs(ctx, idx.config)
	if err != nil {
		idx.logger.Error("error setting up provider market pairs", zap.Error(err))
		return coinmarketcap.ProviderMarketPairs{}, err
	}

	idSet := make(map[int64]struct{})
	for _, pair := range cmcMarketPairs.Data {
		idSet[pair.CMCInfo.BaseID] = struct{}{}
		idSet[pair.CMCInfo.QuoteID] = struct{}{}
	}
	ids := maps.Keys(idSet)
	if err := idx.cmcIndexer.CacheQuotes(ctx, ids); err != nil {
		return coinmarketcap.ProviderMarketPairs{}, err
	}

	for _, pair := range cmcMarketPairs.Data {
		_, _, err := idx.CheckPair(ctx, pair)
		if err != nil {
			idx.logger.Error("error checking pair", zap.Error(err), zap.Any("pair", pair))
			return coinmarketcap.ProviderMarketPairs{}, err
		}
	}

	idx.logger.Info("committing aggregate info tx to db...")

	if err := idx.archiveIntermediateFile(cmcMarketPairs, "cmc_market_pairs.json"); err != nil {
		return coinmarketcap.ProviderMarketPairs{}, err
	}

	return cmcMarketPairs, nil
}

// achiveIntermediateFile writes data to a JSON file in the tmp directory if the --archive-intermediate-steps flag is true
func (idx *Indexer) archiveIntermediateFile(data interface{}, filename string) error {
	if !idx.archiveIntermediateSteps {
		return nil
	}

	filepath := fmt.Sprintf("tmp/%s", filename)
	return file.CreateAndWriteJSONToFile(filepath, data)
}

// FiatAssetInfoFromData creates a fiat asset from coinmarketcap data.
func FiatAssetInfoFromData(data coinmarketcap.FiatData) provider.CreateAssetInfoParams {
	assetAddress := utils.AssetAddress{
		Venue:   VenueFiat,
		Address: "",
	}

	create := provider.CreateAssetInfoParams{
		Symbol:         data.Symbol,
		MultiAddresses: [][]string{assetAddress.ToArray()},
		CmcID:          int64(data.ID),
	}

	return create
}

// CryptoAssetInfoFromQuoteData creates a crypto asset from coinmarketcap quote data.
func CryptoAssetInfoFromQuoteData(data coinmarketcap.QuoteData) provider.CreateAssetInfoParams {
	// use zero values
	assetAddress := utils.AssetAddress{}

	if data.Platform.Name != "" {
		assetAddress.Venue = data.Platform.Name
		assetAddress.Address = data.Platform.TokenAddress
	}

	create := provider.CreateAssetInfoParams{
		Symbol:         data.Symbol,
		MultiAddresses: [][]string{assetAddress.ToArray()},
		CmcID:          int64(data.ID),
		Rank:           int64(data.CmcRank),
	}

	return create
}

// CryptoAssetInfoFromData creates a crypto asset from coinmarketcap data.
func CryptoAssetInfoFromData(data coinmarketcap.WrappedCryptoIDMapData) (provider.CreateAssetInfoParams, error) {
	// use zero values
	var multiAddresses [][]string //nolint:prealloc

	for _, contractAddress := range data.Info.ContractAddress {
		assetSlug := contractAddress.Platform.Coin.Slug
		contractAddressValue := contractAddress.ContractAddress

		// CoinGecko lowercases their EVM addresses. EVM addresses are case-insensitive.
		// Solana addresses are case-sensitive
		if assetSlug == "base" || assetSlug == "ethereum" {
			contractAddressValue = strings.ToLower(contractAddressValue)
		}

		assetAddress := utils.AssetAddress{
			Venue:   contractAddress.Platform.Name,
			Address: contractAddressValue,
		}

		multiAddresses = append(multiAddresses, assetAddress.ToArray())
	}

	symbol, err := symbols.ToTickerString(data.IDMap.Symbol)
	if err != nil {
		return provider.CreateAssetInfoParams{}, fmt.Errorf("error converting symbol to ticker string: %w", err)
	}

	return provider.CreateAssetInfoParams{
		Symbol:         symbol,
		MultiAddresses: multiAddresses,
		CmcID:          int64(data.IDMap.ID),
		Rank:           int64(data.IDMap.Rank),
	}, nil
}

func (idx *Indexer) AssociateCoinMarketCap(
	ctx context.Context,
	inputs []provider.CreateProviderMarket,
	providerMarketPairs coinmarketcap.ProviderMarketPairs,
) ([]provider.CreateProviderMarket, error) {
	var err error
	associatedInputs := make([]provider.CreateProviderMarket, 0, len(inputs))
	for _, input := range inputs {
		// check pairs
		data, found := providerMarketPairs.Data[coinmarketcap.ProviderMarketPairKey(
			input.Create.ProviderName,
			input.Create.TargetBase,
			input.Create.TargetQuote,
		)]
		if found && input.BaseAddress == "" { // pair data does use addresses for matching, so do not use for defi
			idx.logger.Debug("using exchange pair info for CMC info",
				zap.String("base", input.Create.TargetBase),
				zap.String("quote", input.Create.TargetQuote),
				zap.String("provider name", input.Create.ProviderName),
			)

			input.Create.BaseAssetInfoID, input.Create.QuoteAssetInfoID, err = idx.CheckPair(ctx, data)
			if err != nil {
				idx.logger.Debug("failed to check pair info for CMC info")
				continue
			}

		} else {
			idx.logger.Debug("using asset info for CMC info",
				zap.String("base", input.Create.TargetBase),
				zap.String("quote", input.Create.TargetQuote),
				zap.String("provider name", input.Create.ProviderName),
			)

			// check individual assets if we cannot match a pair
			info, ok := idx.knownAssets.LookupAssetInfo(input.Create.TargetBase, input.BaseAddress)
			if !ok {
				idx.logger.Debug("failed to check known base asset info for CMC info", zap.Any("input", input))
				continue
			}
			input.Create.BaseAssetInfoID = info.ID

			info, ok = idx.knownAssets.LookupAssetInfo(input.Create.TargetQuote, input.QuoteAddress)
			if !ok {
				idx.logger.Debug("failed to check known quote asset info for CMC info", zap.Any("input", input))
				continue
			}
			input.Create.QuoteAssetInfoID = info.ID
		}

		// add pair data to supplement
		if found {
			input = addPairDataToCreateProviderMarket(input, data)
		}

		associatedInputs = append(associatedInputs, input)
	}

	return associatedInputs, nil
}

func addPairDataToCreateProviderMarket(
	create provider.CreateProviderMarket,
	data coinmarketcap.ProviderMarketData,
) provider.CreateProviderMarket {
	// use coinmarketcap values if there are no primary source values
	if create.Create.QuoteVolume == 0 {
		create.Create.QuoteVolume = data.QuoteVolume
	}

	if create.Create.ReferencePrice == 0 {
		create.Create.ReferencePrice = data.ReferencePrice
	}

	if create.Create.NegativeDepthTwo == 0 {
		create.Create.NegativeDepthTwo = data.LiquidityInfo.NegativeDepthTwo
	}

	if create.Create.PositiveDepthTwo == 0 {
		create.Create.PositiveDepthTwo = data.LiquidityInfo.PositiveDepthTwo
	}

	return create
}
