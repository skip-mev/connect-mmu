package transformer

import (
	"context"
	"fmt"
	"math"
	"math/big"
	"strings"

	connecttypes "github.com/skip-mev/connect/v2/pkg/types"
	"go.uber.org/zap"
	"golang.org/x/exp/maps"

	"github.com/skip-mev/connect-mmu/config"
	"github.com/skip-mev/connect-mmu/generator/types"
)

// TransformFeed is a function that performs some transformation on the given input markets.
type TransformFeed func(ctx context.Context, logger *zap.Logger, cfg config.GenerateConfig, feeds types.Feeds) (types.Feeds, types.RemovalReasons, error)

// NormalizeBy returns a TransformFeed that adds NormalizeBy feeds to all configured markets based on an input config.
//
// For example, if we have a feed for BTC/USDT with a quote config for USDT indicating to adjustby USDT/USD:
// - add a NormalizeByPair to the ProviderConfig of USDT/USD.
// - change the ticker to be BTC/USD.
func NormalizeBy() TransformFeed {
	return func(_ context.Context, logger *zap.Logger, cfg config.GenerateConfig, feeds types.Feeds) (types.Feeds, types.RemovalReasons, error) {
		logger.Info("adding normalize by pairs", zap.Int("feeds", len(feeds)))

		avgRefPrices, err := types.CalculateAverageReferencePrices(feeds)
		if err != nil {
			logger.Error("failed to calculate average reference prices", zap.Error(err))
			return nil, types.RemovalReasons{}, err
		}

		logger.Info("using quotes", zap.Any("configs", cfg.Quotes))

		transformedFeeds := make([]types.Feed, 0, len(feeds))
		for _, feed := range feeds {
			ticker := feed.Ticker
			quoteConfig, ok := cfg.Quotes[ticker.CurrencyPair.Quote]
			if !ok {
				return nil, nil, fmt.Errorf("quote %s not found in config for normalizing pair",
					ticker.CurrencyPair.Quote)
			}

			// normalize the pair if NormalizeByPair is specified.
			if quoteConfig.NormalizeByPair != "" {
				logger.Debug("normalizing by pair", zap.Any("feed", feed))

				normPair, err := connecttypes.CurrencyPairFromString(quoteConfig.NormalizeByPair)
				if err != nil {
					return nil, nil, err
				}
				newQuote := normPair.Quote
				feed.ProviderConfig.NormalizeByPair = &normPair
				feed.Ticker.CurrencyPair.Quote = newQuote

				adjustPrice, ok := avgRefPrices[normPair.String()]
				if !ok {
					return nil, nil, fmt.Errorf("adjust price for %s not found", normPair.String())
				}

				// example:
				// feed = BTC/USD provided by BTC/USDT adjusted by USDT/USD
				// reference price ( BTC in terms of USD)
				// is equal to (BTC in terms of USDT) times (USDT in terms of USD)
				feed.ReferencePrice = new(big.Float).Mul(feed.ReferencePrice, adjustPrice)

				logger.Debug("normalized by pair", zap.Any("feed", feed))
			}

			transformedFeeds = append(transformedFeeds, feed)
		}

		logger.Info("added normalize by pairs", zap.Int("remaining feeds", len(feeds)))
		return transformedFeeds, nil, nil
	}
}

// ResolveCMCConflictsForMarket resolves issues where the feeds for a market may be referring to different
// base assets.
//
// An example conflict is if we have three feeds for GOAT/USD, from binance, kraken, and uniswap base.
// - GOAT/USD from binance and kraken referring to Goatseus Maximus (CMC ID 33440)
// - GOAT/USD from uniswap base referring to GOAT on Base (CMC ID 34935)
//
// For each market, we sort the feeds and select the base asset's CMC ID of the first sorted feed. This
// will have the best CMC rank. We then filter out all feeds for this market that do not match this CMC ID.
func ResolveCMCConflictsForMarket() TransformFeed {
	return func(_ context.Context, logger *zap.Logger, _ config.GenerateConfig, feeds types.Feeds) (types.Feeds,
		types.RemovalReasons, error,
	) {
		logger.Info("resolving CMC conflicts", zap.Int("feeds", len(feeds)))

		tickerToFeeds := make(map[string]types.Feeds, len(feeds))
		for _, feed := range feeds {
			ticker := feed.TickerString()
			tickerToFeeds[ticker] = append(tickerToFeeds[ticker], feed)
		}

		out := make([]types.Feed, 0, len(feeds))
		removals := types.NewRemovalReasons()

		for ticker, feeds := range tickerToFeeds {
			feeds.Sort()
			bestCMCId := feeds[0].CMCInfo.BaseID
			bestCMCRank := feeds[0].CMCInfo.BaseRank
			for _, feed := range feeds {
				if feed.CMCInfo.BaseRank < bestCMCRank {
					panic(fmt.Sprintf("found feed for %s with lower CMC rank than the best one for ticker %s. best CMC rank %d, feed CMC rank %d", feed.ProviderConfig.Name, ticker, bestCMCRank, feed.CMCInfo.BaseRank))
				}
				if feed.CMCInfo.BaseID == bestCMCId {
					out = append(out, feed)
				} else {
					removals.AddRemovalReasonFromFeed(feed, feed.ProviderConfig.Name,
						fmt.Sprintf("Transform ResolveCMCConflictsForMarket: BestCMCID: %d, FeedCMCID: %d, BestCMCRank: %d, FeedCMCRank: %d", bestCMCId,
							feed.CMCInfo.BaseID, bestCMCRank, feed.CMCInfo.BaseRank))
					logger.Debug("dropping feed with worse CMC ID", zap.Any("ticker", feed.Ticker.String()), zap.Any("provider", feed.ProviderConfig.Name))

				}
			}
		}

		logger.Info("resolved CMC conflicts", zap.Int("remaining feeds", len(out)))

		return out, nil, nil
	}
}

// ResolveConflictsForProvider resolves all conflicts between feeds.  Conflicts arise when the feeds have overlapping CurrencyPairs.
//
// An example conflict could arise if we desire markets quoted in USD and have two feeds:
// - BTC/USD from kraken using the btc/usd ticker
// - BTC/USD from kraken using the btc/usdt ticker adjusted by BTC/USD
//
// This conflict would have been created in the NormalizeBy transform, and we must choose one of the feeds for this
// given provider. We choose based on comparing the Liquidity and 24HR Volume for each feed.
func ResolveConflictsForProvider() TransformFeed {
	return func(_ context.Context, logger *zap.Logger, _ config.GenerateConfig, feeds types.Feeds) (types.Feeds,
		types.RemovalReasons, error,
	) {
		logger.Info("resolving conflicts", zap.Int("feeds", len(feeds)))

		cpToProvider := make(map[string]types.Feed, len(feeds))
		for _, feed := range feeds {
			key := keyCurrencyPairProviderName(feed.TickerString(), feed.ProviderConfig.Name)

			got, found := cpToProvider[key]
			if !found {
				cpToProvider[key] = feed
				continue
			}

			replace := types.Compare(got, feed)
			if replace {
				logger.Debug("replacing on conflict", zap.Any("old", got), zap.Any("new", feed))
				cpToProvider[key] = feed
			} else {
				logger.Debug("conflict found but not replacing", zap.Any("old", got), zap.Any("new", feed))
			}
		}

		out := maps.Values(cpToProvider)
		logger.Info("resolved conflicts", zap.Int("remaining feeds", len(out)))

		// sort for stable output
		types.Feeds(out).Sort()

		return out, nil, nil
	}
}

// DropFeedsWithoutAggregatorIDs drops feeds based on the given config.
//
// Feeds can be dropped if:
// - We require AggregatorIDs (coinmarketcap, etc) for the feeds provider, but it does not have any.
func DropFeedsWithoutAggregatorIDs() TransformFeed {
	return func(_ context.Context, logger *zap.Logger, cfg config.GenerateConfig, feeds types.Feeds) (types.Feeds,
		types.RemovalReasons, error,
	) {
		logger.Info("dropping feeds", zap.Int("num feeds", len(feeds)))

		out := make([]types.Feed, 0, len(feeds))
		removals := types.NewRemovalReasons()
		for _, feed := range feeds {
			providerConfig := cfg.Providers[feed.ProviderConfig.Name]
			if (feed.CMCInfo.BaseID != 0 && providerConfig.RequireAggregateIDs) || !providerConfig.
				RequireAggregateIDs {
				out = append(out, feed)
			} else {
				removals.AddRemovalReasonFromFeed(feed, feed.ProviderConfig.Name,
					fmt.Sprintf("Transform DropFeedsWithoutAggregatorIDs: BaseCMCID: %d, RequireAggregateIDs: %v", feed.CMCInfo.BaseID,
						providerConfig.RequireAggregateIDs))
				logger.Info("dropping feed", zap.Any("ticker", feed.Ticker.String()), zap.Any("provider", feed.ProviderConfig.Name))
			}
		}

		logger.Info("dropped feeds", zap.Int("remaining feeds", len(out)))
		return out, removals, nil
	}
}

// InvertOrDrop attempts to invert any potential feeds that could be inverted to a desired quote config to be valid.
//
// For example:
//
// If the feed is for BTC/MOG and the list of desired quotes is [ETH, BTC, USD, SOL]
// this transform will try to invert the feed to become MOG/BTC and add the "invert"
// flag to the underlying ProviderConfig.
//
// Feeds whose base AND quote fall outside the target quotes are dropped.
func InvertOrDrop() TransformFeed {
	return func(_ context.Context, logger *zap.Logger, cfg config.GenerateConfig, feeds types.Feeds) (types.Feeds,
		types.RemovalReasons, error,
	) {
		logger.Info("inverting feeds", zap.Int("feeds", len(feeds)))

		out := make([]types.Feed, 0, len(feeds))
		removals := types.NewRemovalReasons()
		quotes := maps.Keys(cfg.Quotes)

		for _, feed := range feeds {
			ticker := feed.Ticker
			// first check if the quote is already a valid quote
			_, found := cfg.Quotes[ticker.CurrencyPair.Quote]
			if found {
				// if the quote config exists, do nothing
				out = append(out, feed)
				continue
			}

			// next, check if the base is a valid quote
			_, found = cfg.Quotes[ticker.CurrencyPair.Base]
			if found {
				logger.Debug("inverting", zap.Any("feed", feed))
				// if the base config exists, invert
				feed.ProviderConfig.Invert = true
				feed.Ticker.CurrencyPair = ticker.CurrencyPair.Invert()

				// invert the price if it is not zero
				if feed.ReferencePrice.Cmp(big.NewFloat(0)) != 0 {
					feed.ReferencePrice = new(big.Float).Quo(big.NewFloat(1), feed.ReferencePrice)
				}

				// invert the CMC IDs
				feed.CMCInfo.Invert()

				logger.Debug("inverted feed", zap.Any("feed", feed))
				out = append(out, feed)
				continue
			}

			removals.AddRemovalReasonFromFeed(feed, feed.ProviderConfig.Name, fmt.Sprintf("Transform InvertOrDrop: %s, "+
				"feed cannot be inverted to quotes: %s", feed.Ticker.String(), quotes))
			logger.Debug("dropping feed", zap.Any("feed", feed))
		}

		logger.Info("inverted", zap.Int("feeds remaining", len(out)))
		return out, removals, nil
	}
}

// PruneByLiquidity removes feeds that do not have an associated quote config.
//
// If the market has a quote config, the following checks are performed:
// - check if 24hr liquidity in USD is sufficient.
func PruneByLiquidity() TransformFeed {
	return func(_ context.Context, logger *zap.Logger, cfg config.GenerateConfig, feeds types.Feeds) (types.Feeds,
		types.RemovalReasons, error,
	) {
		out := make([]types.Feed, 0, len(feeds))
		removals := types.NewRemovalReasons()

		logger.Info("pruning feeds by liquidity", zap.Int("feeds", len(feeds)))

		for _, feed := range feeds {
			providerCfg, found := cfg.Providers[feed.ProviderConfig.Name]
			if found && providerCfg.IgnoreLiquidity {
				out = append(out, feed)
				continue
			}

			ticker := feed.Ticker
			quoteConfig, found := cfg.Quotes[ticker.CurrencyPair.Quote]
			if found && feed.LiquidityInfo.IsSufficient(quoteConfig.MinProviderLiquidity) {
				out = append(out, feed)
				continue
			}

			var reason string
			if !found {
				reason = "PruneByLiquidity: Not Found"
			} else {
				reason = fmt.Sprintf("PruneByLiquidity: NegativeDepthTwo: %f, PositiveDepthTwo: %f, "+
					"MinProviderLiquidity: %f",
					feed.LiquidityInfo.NegativeDepthTwo,
					feed.LiquidityInfo.PositiveDepthTwo,
					quoteConfig.MinProviderLiquidity,
				)
			}
			removals.AddRemovalReasonFromFeed(feed, feed.ProviderConfig.Name, reason)
			logger.Debug("dropping feed", zap.Any("feed", feed))
		}

		logger.Info("pruned feeds by liquidity", zap.Int("feeds", len(feeds)))

		return out, removals, nil
	}
}

// PruneByQuoteVolume removes feeds that do not have an associated quote config.
//
// If the market has a quote config, the following checks are performed:
// - check if 24hr quote volume is sufficient.
func PruneByQuoteVolume() TransformFeed {
	return func(_ context.Context, logger *zap.Logger, cfg config.GenerateConfig, feeds types.Feeds) (types.Feeds,
		types.RemovalReasons, error,
	) {
		logger.Info("pruning feeds by quote volume", zap.Int("feeds", len(feeds)))

		out := make([]types.Feed, 0, len(feeds))
		removals := types.NewRemovalReasons()
		for _, feed := range feeds {
			providerCfg, found := cfg.Providers[feed.ProviderConfig.Name]
			if found && providerCfg.IgnoreVolume {
				out = append(out, feed)
				continue
			}

			ticker := feed.Ticker
			quoteConfig, found := cfg.Quotes[ticker.CurrencyPair.Quote]

			dailyQuoteVolumeFloat, _ := feed.DailyQuoteVolume.Float64()
			if found && dailyQuoteVolumeFloat >= quoteConfig.MinProviderVolume {
				out = append(out, feed)
				continue
			}

			var reason string
			if !found {
				reason = "PruneByQuote: Not Found"
			} else {
				reason = fmt.Sprintf("PruneByQuote: DailyQuoteVolume: %f, MinProviderVolume: %f", feed.DailyQuoteVolume, quoteConfig.MinProviderVolume)
			}
			removals.AddRemovalReasonFromFeed(feed, feed.ProviderConfig.Name, reason)
			logger.Debug("dropping feed", zap.Any("feed", feed))
		}

		logger.Info("pruned feeds by quote volume", zap.Int("feeds remaining", len(out)))
		return out, removals, nil
	}
}

// ResolveNamingAliases chooses a canonical set of Feeds that have the same TickerString()
//
// Group all feeds with the same TickerString together:
// - differentiate between the feeds using CoinMarketCap identifiers.
// - choose one CoinMarketCap identifier group per TickerString()
func ResolveNamingAliases() TransformFeed {
	return func(_ context.Context, logger *zap.Logger, _ config.GenerateConfig, feeds types.Feeds) (types.Feeds,
		types.RemovalReasons, error,
	) {
		logger.Info("resolving ticker string naming aliases", zap.Int("feeds", len(feeds)))

		// "BASE/QUOTE" -> BaseCMCID-QuoteCMCID -> []Feeds
		feedGroupsPerTicker := make(map[string]map[string]types.Feeds)
		for _, feed := range feeds {
			if _, ok := feedGroupsPerTicker[feed.TickerString()]; !ok {
				feedGroupsPerTicker[feed.TickerString()] = make(map[string]types.Feeds)
			}
			feedGroupsPerTicker[feed.TickerString()][feed.UniqueID()] = append(feedGroupsPerTicker[feed.TickerString()][feed.UniqueID()], feed)
		}

		removals := types.NewRemovalReasons()
		out := make(types.Feeds, 0)

		// choose the "best" asset for the given TickerString
		for tickerString, feedGroups := range feedGroupsPerTicker {
			logger.Debug("resolving ticker string naming aliases", zap.String("ticker", tickerString))

			bestGroupID, err := getHighestRankFeedGroup(feedGroups)
			if err != nil {
				logger.Info("no group found for ticker", zap.String("ticker", tickerString), zap.Error(err))
				continue
			}

			out = append(out, feedGroups[bestGroupID]...)

			// remove feeds for conflicting tickers
			for id, feeds := range feedGroups {
				if id == bestGroupID {
					continue
				}
				for _, feed := range feeds {
					removals.AddRemovalReasonFromFeed(
						feed,
						feed.ProviderConfig.Name,
						fmt.Sprintf(
							"removing due to naming alias for ticker %s, pair %s, CMC pair %s chosen instead",
							tickerString,
							feed.UniqueID(),
							bestGroupID,
						),
					)
				}
			}
		}

		out.Sort()
		logger.Info("resolved ticker string naming aliases", zap.Int("feeds", len(out)))

		return out, removals, nil
	}
}

// getHighestRankFeedGroup uses CMC Rank information to choose which set of feeds
// is "best" for generation.
// The input data to this function should be a map[CMCIds] -> feeds with the _same_ CMC Info.
// The set of feeds with the _lowest_ BaseAssetRank will be chosen.
func getHighestRankFeedGroup(feedGroups map[string]types.Feeds) (string, error) {
	if len(feedGroups) == 0 {
		return "", fmt.Errorf("no feed groups found")
	}

	bestGroup := ""
	bestRankBase := int64(math.MaxInt64)
	bestRankQuote := int64(math.MaxInt64)
	for groupID, group := range feedGroups {
		if len(group) == 0 {
			return "", fmt.Errorf("no feeds found in group %s", groupID)
		}

		feed := group[0]

		// if we don't have ranking info, don't consider
		if !feed.CMCInfo.HasRank() {
			continue
		}

		// all items in this group have the same CMC Rank, so we just use item 0
		if feed.CMCInfo.BaseRank < bestRankBase {
			bestGroup = groupID
			bestRankBase = feed.CMCInfo.BaseRank
			bestRankQuote = feed.CMCInfo.QuoteRank
		} else if feed.CMCInfo.BaseRank == bestRankBase {
			// compare quoteRank
			if feed.CMCInfo.QuoteRank < bestRankQuote {
				bestGroup = groupID
				bestRankQuote = feed.CMCInfo.QuoteRank
			}
		}
	}

	if bestGroup == "" {
		return "", fmt.Errorf("no feed valid groups found in feeds")
	}

	return bestGroup, nil
}

// TopFeedsForProvider chooses only the top N feeds for a provider if it has a filter set.
// The feeds are sorted by the base asset's CMC rank and then the top N are chosen.
// If no filter is set, the feeds are sorted, but no feeds will be removed.
func TopFeedsForProvider() TransformFeed {
	return func(_ context.Context, logger *zap.Logger, cfg config.GenerateConfig, feeds types.Feeds,
	) (types.Feeds, types.RemovalReasons, error) {
		provFeeds := feeds.ToProviderFeeds()

		removals := types.NewRemovalReasons()

		for provider, feedsForProvider := range provFeeds {
			provConfig, ok := cfg.Providers[provider]
			if !ok {
				return nil, nil, fmt.Errorf("provider %s not found", provider)
			}

			if len(feedsForProvider) == 0 {
				continue
			}

			numFeedsToRetain := provConfig.Filters.TopMarkets
			if numFeedsToRetain == 0 || uint64(len(feedsForProvider)) <= numFeedsToRetain {
				// in this case, we have fewer feeds than we are trying to prune to, so just keep them all
				continue
			}

			logger.Info("filtering top markets per provider", zap.String("provider", provider), zap.Int("feeds", len(feedsForProvider)))

			// sort the feeds based on CMC rank, then take the top N
			// this will sort the feeds where feeds[0] has the best CMC rank
			feedsForProvider.Sort()

			// after sorting, only take top N
			provFeeds[provider] = feedsForProvider[:numFeedsToRetain]

			// add removal reasons for all markets to be removed
			for _, feed := range feedsForProvider[numFeedsToRetain:] {
				logger.Debug("removing feed", zap.Any("feed", feed))
				removals.AddRemovalReasonFromFeed(feed, provider,
					fmt.Sprintf("only selecting top %d feeds for this provider", numFeedsToRetain))
			}
		}

		return provFeeds.ToFeeds(), removals, nil
	}
}

// PruneByProviderLiquidity removes feeds that don't meet provider-specific liquidity thresholds.
// Each provider can specify a min_provider_liquidity threshold in the config.
func PruneByProviderLiquidity() TransformFeed {
	return func(_ context.Context, logger *zap.Logger, cfg config.GenerateConfig, feeds types.Feeds) (types.Feeds, types.RemovalReasons, error) {
		logger.Info("pruning by provider liquidity", zap.Int("feeds", len(feeds)))

		out := make([]types.Feed, 0, len(feeds))
		removals := types.NewRemovalReasons()

		for _, feed := range feeds {
			providerName := feed.ProviderConfig.Name
			providerConfig, found := cfg.Providers[providerName]

			// Skip if provider ignores liquidity
			if found && providerConfig.IgnoreLiquidity {
				out = append(out, feed)
				continue
			}

			if found && feed.LiquidityInfo.IsSufficient(providerConfig.MinProviderLiquidity) {
				out = append(out, feed)
				continue
			}

			var reason string
			if !found {
				reason = "PruneByProviderLiquidity: Not Found"
			} else if !feed.LiquidityInfo.IsSufficient(providerConfig.MinProviderLiquidity) {
				reason = fmt.Sprintf("PruneByProviderLiquidity: NegativeDepthTwo: %f, PositiveDepthTwo: %f, "+
					"MinProviderLiquidity: %f",
					feed.LiquidityInfo.NegativeDepthTwo,
					feed.LiquidityInfo.PositiveDepthTwo,
					providerConfig.MinProviderLiquidity,
				)
			}
			removals.AddRemovalReasonFromFeed(feed, providerName, reason)
			logger.Debug("dropping feed", zap.Any("feed", feed))
		}

		logger.Info("pruned feeds by provider liquidity", zap.Int("feeds remaining", len(out)))
		return out, removals, nil
	}
}

// PruneByProviderUsdVolume removes feeds that don't meet provider-specific USD volume thresholds.
// Each provider can specify a min_provider_volume threshold in the config.
func PruneByProviderUsdVolume() TransformFeed {
	return func(_ context.Context, logger *zap.Logger, cfg config.GenerateConfig, feeds types.Feeds) (types.Feeds, types.RemovalReasons, error) {
		logger.Info("pruning by provider volume", zap.Int("feeds", len(feeds)))

		out := make([]types.Feed, 0, len(feeds))
		removals := types.NewRemovalReasons()

		for _, feed := range feeds {
			providerName := feed.ProviderConfig.Name
			providerCfg, found := cfg.Providers[providerName]
			if found && providerCfg.IgnoreVolume {
				out = append(out, feed)
				continue
			}

			dailyUsdVolumeFloat, _ := feed.DailyUsdVolume.Float64()
			if found && dailyUsdVolumeFloat >= providerCfg.MinProviderVolume {
				out = append(out, feed)
				continue
			}

			var reason string
			if !found {
				reason = "PruneByProviderUsdVolume: Not Found"
			} else if dailyUsdVolumeFloat < providerCfg.MinProviderVolume {
				reason = fmt.Sprintf("PruneByProviderUsdVolume: Volume24H: %f, MinProviderVolume: %f",
					dailyUsdVolumeFloat,
					providerCfg.MinProviderVolume,
				)
			}
			removals.AddRemovalReasonFromFeed(feed, providerName, reason)
			logger.Debug("dropping feed", zap.Any("feed", feed))
		}

		logger.Info("pruned feeds by provider volume", zap.Int("feeds remaining", len(out)))
		return out, removals, nil
	}
}

func keyCurrencyPairProviderName(cp, provider string) string {
	return strings.Join([]string{provider, cp}, "_")
}
