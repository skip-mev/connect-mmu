package types

import (
	"fmt"
	"math/big"
	"strconv"
	"strings"

	mmtypes "github.com/skip-mev/connect/v2/x/marketmap/types"
	"golang.org/x/exp/slices"

	"github.com/skip-mev/connect-mmu/types"
)

// RemovalReasons contains identifying information around a removed Ticker and the reasons for its removal.
// Used for debugging why a market is not considered valid/supported by the MMU.
type RemovalReasons map[string][]RemovalReason

// NewRemovalReasons creates a new RemovalReasons map.
func NewRemovalReasons() RemovalReasons {
	return make(RemovalReasons)
}

// AddRemovalReasonFromFeed adds a reason for a provider to remove a Feed.
func (r RemovalReasons) AddRemovalReasonFromFeed(feed Feed, provider string, reason string) {
	if _, found := r[feed.Ticker.String()]; !found {
		r[feed.Ticker.String()] = []RemovalReason{}
	}
	r[feed.Ticker.String()] = append(r[feed.Ticker.String()], RemovalReason{
		Provider: provider,
		Reason:   reason,
		Feed:     feed,
	})
}

// AddRemovalReasonFromMarket adds a reason for a provider to remove a Market.
func (r RemovalReasons) AddRemovalReasonFromMarket(market mmtypes.Market, provider string, reason string) {
	if _, found := r[market.Ticker.String()]; !found {
		r[market.Ticker.String()] = []RemovalReason{}
	}
	r[market.Ticker.String()] = append(r[market.Ticker.String()], RemovalReason{
		Provider: provider,
		Reason:   reason,
		Market:   market,
	})
}

// Merge merges two RemovalReasons maps.
func (r RemovalReasons) Merge(other RemovalReasons) {
	for ticker, reasons := range other {
		r[ticker] = append(r[ticker], reasons...)
	}
}

// RemovalReason is a struct containing the reason for a (provider, market) / market removal.
type RemovalReason struct {
	Reason   string         `json:"reason"`
	Provider string         `json:"provider"`
	Market   mmtypes.Market `json:"market,omitempty"`
	Feed     Feed           `json:"feed,omitempty"`
}

// Feed is a wrapper around a Market that includes additional data for filtering such as
// 24hr volume, liquidity, etc..
type Feed struct {
	// Ticker is the ticker that this Feed refers to.
	Ticker mmtypes.Ticker
	// ProviderConfig is the provider config that this Feed refers to.
	ProviderConfig mmtypes.ProviderConfig
	// DailyLiquidity is the 24-hour volume in terms of the quote.
	DailyQuoteVolume *big.Float
	// DailyUsdVolume is the 24-hour volume in USD
	DailyUsdVolume *big.Float
	// ReferencePrice is the reference price of the base asset in terms of the quote.
	ReferencePrice *big.Float
	// CMCInfo contains coinmarketcap specific information
	CMCInfo types.CoinMarketCapInfo
	// LiquidityInfo contains buy and sell side liquidity denominated in USD.
	LiquidityInfo types.LiquidityInfo
}

func NewFeed(
	t mmtypes.Ticker,
	pc mmtypes.ProviderConfig,
	quoteVolume, usdVolume, referencePrice float64,
	liquidityInfo types.LiquidityInfo,
	cmcInfo types.CoinMarketCapInfo,
) Feed {
	cmcInfo.BaseID = resolveWrappedAssetAliases(cmcInfo.BaseID)
	cmcInfo.QuoteID = resolveWrappedAssetAliases(cmcInfo.QuoteID)

	return Feed{
		Ticker:           t,
		ProviderConfig:   pc,
		DailyQuoteVolume: big.NewFloat(quoteVolume),
		DailyUsdVolume:   big.NewFloat(usdVolume),
		LiquidityInfo:    liquidityInfo,
		ReferencePrice:   big.NewFloat(referencePrice),
		CMCInfo:          cmcInfo,
	}
}

// TickerString returns the string representation of the Feed's Market's Ticker.
func (f *Feed) TickerString() string { return f.Ticker.String() }

// knownWrappedAssetAliases maps the CMC ids of wrapped assets to their native asset CMC ID.
// This is used to resolve wrapped assets to their native assets to assert that these are
// _essentially_ the same asset.
var knownWrappedAssetAliases = map[int64]int64{
	// Wrapped SOL -> SOL
	// - SOL:  https://coinmarketcap.com/currencies/solana/
	// - Wrapped SOL: https://coinmarketcap.com/currencies/wrapped-solana/
	16116: 5426,
}

func resolveWrappedAssetAliases(id int64) int64 {
	if nativeAssetID, found := knownWrappedAssetAliases[id]; found {
		return nativeAssetID
	}

	return id
}

// UniqueID returns an ID that uniquely identifies the asset pair that is being represented using CoinMarketCap IDs
// ID is of form: "BaseAssetID-QuoteAssetID".
func (f *Feed) UniqueID() string {
	resolvedBaseID := resolveWrappedAssetAliases(f.CMCInfo.BaseID)
	resolvedQuoteID := resolveWrappedAssetAliases(f.CMCInfo.QuoteID)

	return strconv.FormatInt(resolvedBaseID, 10) + "-" + strconv.FormatInt(resolvedQuoteID, 10)
}

// Compare compares two Feeds
// If Volume is equal, false a is returned.
// A bool is returned indicating:
// If false, a is "greater than" b
// If true, b is "greater than a".
//
// Comparison is done as follows:
// 1. attempt to compare CMC ranks of the feeds
// 2. else, attempt to compare liquidity of the feeds
// 3. fall back to quote volume of the assets.
func Compare(a, b Feed) bool {
	if a.CMCInfo.BaseRank != b.CMCInfo.BaseRank {
		return a.CMCInfo.BaseRank > b.CMCInfo.BaseRank
	}

	if a.LiquidityInfo.TotalLiquidity() != b.LiquidityInfo.TotalLiquidity() {
		return a.LiquidityInfo.TotalLiquidity() < b.LiquidityInfo.TotalLiquidity()
	}

	cmp := a.DailyUsdVolume.Cmp(b.DailyUsdVolume)
	return cmp < 0
}

// Sort follows the following rules:
//
// # CMC rank
// if feed A has a CMC rank and feed B does not, feedA is higher than feedB
// if feed A has a lower CMC rank than feed b, feedA is higher than feed B
//
// # Liquidity
// if feed A and feed B have non-zero, non-equal liquidity, if feed A has more liquidity than feed B, feed A is higher than feed B
//
// # QuoteVolume
// if feed A has more daily quote volume than feed B, feed A is higher than feed B
func (f Feeds) Sort() {
	slices.SortFunc(f, func(a, b Feed) int {
		bIsBetter := Compare(a, b)
		if bIsBetter {
			return 1
		}

		return -1
	})
}

// ProviderFeeds is a type alias for a map of ProviderName -> []Feed.
type ProviderFeeds map[string]Feeds

// ToFeeds flattens a ProviderFeeds map to a Feeds slice.
func (pf ProviderFeeds) ToFeeds() Feeds {
	feeds := make(Feeds, 0, len(pf))

	for _, provFeeds := range pf {
		feeds = append(feeds, provFeeds...)
	}

	feeds.Sort()
	return feeds
}

// Feeds is a type alias for a slice of Feeds.
type Feeds []Feed

func (f Feeds) ToProviderFeeds() ProviderFeeds {
	providerFeeds := make(ProviderFeeds, len(f))

	for _, feed := range f {
		feeds, found := providerFeeds[feed.ProviderConfig.Name]
		if !found {
			providerFeeds[feed.ProviderConfig.Name] = make(Feeds, 0)
		}

		providerFeeds[feed.ProviderConfig.Name] = append(feeds, feed)
	}

	return providerFeeds
}

// ToMarketMap translates the set of feeds to a valid MarketMap by:
// - converting Feed objects to Markets or appending them to existing markets.
// - removing markets that have providers below MinProviderCount.
// Returns an error if the resulting marketmap is invalid.
func (f Feeds) ToMarketMap() (mmtypes.MarketMap, error) {
	// calculate total liquidity per market
	liquidityPerMarket := make(map[string]float64, len(f))
	for _, feed := range f {
		liquidityPerMarket[feed.UniqueID()] += feed.LiquidityInfo.TotalLiquidity()
	}

	avgRefPrices, err := CalculateAverageReferencePrices(f)
	if err != nil {
		return mmtypes.MarketMap{}, err
	}

	mm := mmtypes.MarketMap{Markets: make(map[string]mmtypes.Market)}

	for _, feed := range f {
		if mmMarket, found := mm.Markets[feed.TickerString()]; found {
			// always use the highest MinProviderCount available if combining
			if feed.Ticker.MinProviderCount > mmMarket.Ticker.MinProviderCount {
				mmMarket.Ticker.MinProviderCount = feed.Ticker.MinProviderCount
			}

			mmMarket.ProviderConfigs = append(mmMarket.ProviderConfigs, feed.ProviderConfig)
			mm.Markets[feed.TickerString()] = mmMarket

			continue
		}

		tickerMD, err := ToTickerMetadataJSON(feed, avgRefPrices[feed.TickerString()], liquidityPerMarket[feed.UniqueID()])
		if err != nil {
			return mmtypes.MarketMap{}, err
		}

		mm.Markets[feed.TickerString()] = mmtypes.Market{
			Ticker: mmtypes.Ticker{
				CurrencyPair:     feed.Ticker.CurrencyPair,
				Decimals:         types.DecimalPlacesFromPrice(feed.ReferencePrice),
				MinProviderCount: feed.Ticker.MinProviderCount,
				Enabled:          false,
				Metadata_JSON:    tickerMD,
			},
			ProviderConfigs: []mmtypes.ProviderConfig{
				feed.ProviderConfig,
			},
		}
	}

	// sort all provider configs per market so that the output is deterministic
	// the ordering is not important beyond having a stable output
	for name, market := range mm.Markets {
		providerConfigs := market.ProviderConfigs

		slices.SortFunc(providerConfigs, func(a, b mmtypes.ProviderConfig) int {
			return strings.Compare(a.Name, b.Name)
		})

		market.ProviderConfigs = providerConfigs
		mm.Markets[name] = market
	}

	return mm, nil
}

func (f *Feed) Equal(feedB Feed) bool {
	if !f.Ticker.Equal(feedB.Ticker) {
		return false
	}

	if !f.ProviderConfig.Equal(feedB.ProviderConfig) {
		return false
	}

	if f.CMCInfo != feedB.CMCInfo {
		return false
	}

	if f.LiquidityInfo != feedB.LiquidityInfo {
		return false
	}

	if f.ReferencePrice.Cmp(feedB.ReferencePrice) != 0 {
		return false
	}

	return true
}

// Equal checks equality between Feed arrays per element doing a deep equality check.
func (f Feeds) Equal(feedsB Feeds) bool {
	if len(f) != len(feedsB) {
		return false
	}

	for i, feedA := range f {
		if !feedA.Equal(feedsB[i]) {
			return false
		}
	}

	return true
}

func CalculateAverageReferencePrices(feeds Feeds) (map[string]*big.Float, error) {
	// ticker -> sum of all reference prices
	feedReferencePriceSum := make(map[string]*big.Float)
	// ticker -> number of occurrences of this ticker ( for averaging )
	feedReferencePriceOccurrences := make(map[string]int)

	for _, feed := range feeds {
		ticker := feed.TickerString()

		feedReferencePriceOccurrences[ticker]++
		current := feedReferencePriceSum[ticker]
		if current == nil {
			current = big.NewFloat(0)
		}
		feedReferencePriceSum[ticker] = current.Add(current, feed.ReferencePrice)
	}

	// ticker -> average reference price
	feedAverageReferencePrice := make(map[string]*big.Float)
	for ticker, sum := range feedReferencePriceSum {
		occurrences, ok := feedReferencePriceOccurrences[ticker]
		if !ok {
			return nil, fmt.Errorf("no occurrences for ticker: %s", ticker)
		}

		feedAverageReferencePrice[ticker] = sum.Quo(sum, big.NewFloat(float64(occurrences)))
	}

	return feedAverageReferencePrice, nil
}
