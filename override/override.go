package override

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/skip-mev/connect-mmu/generator/types"
	connecttypes "github.com/skip-mev/connect/v2/pkg/types"
	mmtypes "github.com/skip-mev/connect/v2/x/marketmap/types"
	"github.com/skip-mev/connect/v2/x/marketmap/types/tickermetadata"
	"go.uber.org/zap"

	"github.com/skip-mev/connect-mmu/client/dydx"
	libdydx "github.com/skip-mev/connect-mmu/lib/dydx"
	"github.com/skip-mev/connect-mmu/override/update"
)

// MarketMapOverride is an interface for overriding a generated marketmap with what is on-chain according to specific rules.
//
//go:generate mockery --name MarketMapOverride --filename mock_upsert_strategy.go
type MarketMapOverride interface {
	OverrideGeneratedMarkets(
		ctx context.Context,
		logger *zap.Logger,
		actual, generated mmtypes.MarketMap,
		options update.Options,
	) (mmtypes.MarketMap, error)
}

type CoreOverride struct{}

var _ MarketMapOverride = (*CoreOverride)(nil)

func NewCoreOverride() MarketMapOverride {
	return &CoreOverride{}
}

// OverrideGeneratedMarkets does the following:
// - merges disjoint markets from actual and generated
// - appends newly generated provider configs to intersecting markets in actual and generated.
func (o *CoreOverride) OverrideGeneratedMarkets(
	_ context.Context,
	logger *zap.Logger,
	actual, generated mmtypes.MarketMap,
	options update.Options,
) (mmtypes.MarketMap, error) {
	logger.Info("overriding markets", zap.Any("options", options))

	appendedMarketMap, err := update.CombineMarketMaps(logger, actual, generated, options)
	if err != nil {
		logger.Error("failed to update to market map", zap.Error(err))
		return mmtypes.MarketMap{}, fmt.Errorf("failed to update to market map: %w", err)
	}

	return appendedMarketMap, nil
}

type DyDxOverride struct {
	client dydx.Client
}

var _ MarketMapOverride = (*DyDxOverride)(nil)

func NewDyDxOverride(client dydx.Client) (MarketMapOverride, error) {
	if client == nil {
		return nil, fmt.Errorf("client cannot be nil nil")
	}

	return &DyDxOverride{
		client: client,
	}, nil
}

// OverrideGeneratedMarkets does the following:
// - merges disjoint markets from actual and generated
// - appends newly generated provider configs to intersecting markets in actual and generated
// - ensures that all CrossMargin markets on-chain are equal to the actual market map (no change).
func (o *DyDxOverride) OverrideGeneratedMarkets(
	ctx context.Context,
	logger *zap.Logger,
	actual, generated mmtypes.MarketMap,
	options update.Options,
) (mmtypes.MarketMap, error) {
	logger.Info("overriding markets for dydx", zap.Any("options", options))

	// first append to the actual market map
	combinedMarketMap, err := update.CombineMarketMaps(logger, actual, generated, options)
	if err != nil {
		logger.Error("failed to update to market map", zap.Error(err))
		return mmtypes.MarketMap{}, fmt.Errorf("failed to update to market map: %w", err)
	}

	logger.Info("combined actual and generated market maps", zap.Int("markets", len(combinedMarketMap.Markets)))

	// filter away all markets that are cross-margined
	perpsResp, err := o.client.AllPerpetuals(ctx)
	if err != nil {
		return mmtypes.MarketMap{}, err
	}

	if perpsResp == nil {
		return mmtypes.MarketMap{}, fmt.Errorf("nil perpetuals response")
	}

	logger.Info("got perpetuals", zap.Int("count", len(perpsResp.Perpetuals)))

	// for each perpetual, identify if there's a corresponding ticker in the market-map, and set it equal
	// to the corresponding market in actual
	for _, perpetual := range perpsResp.Perpetuals {
		connectTicker, err := libdydx.MarketPairToCurrencyPair(perpetual.Params.Ticker)
		if err != nil {
			return mmtypes.MarketMap{}, err
		}

		// perpetual markets should always be in the actual market map, error if they are not in correspondence
		actualMarket, ok := actual.Markets[connectTicker.String()]
		if !ok {
			logger.Error("actual market for cross-margined perpetual not found", zap.String("ticker", connectTicker.String()))
			return mmtypes.MarketMap{}, fmt.Errorf("actual market for cross-margined perpetual %s not found", connectTicker.String())
		}

		// check for the market in generated
		generatedMarket, ok := combinedMarketMap.Markets[connectTicker.String()]
		if !ok {
			logger.Debug("perpetual market not found in generated", zap.String("ticker", connectTicker.String()))
			continue
		}

		// if the market is not cross-margined, continue
		if perpetual.Params.MarketType != dydx.PERPETUAL_MARKET_TYPE_CROSS {
			logger.Debug("perpetual market is not cross-margined", zap.String("ticker", connectTicker.String()))
			continue
		}

		// ensure the generated market, and actual are equal
		if !generatedMarket.Equal(actualMarket) {
			logger.Debug(
				"generated market is not equal to actual",
				zap.String("ticker", connectTicker.String()),
				zap.String("generated", generatedMarket.String()),
				zap.String("actual", actualMarket.String()),
			)
		}

		combinedMarketMap.Markets[connectTicker.String()] = actualMarket
	}

	return combinedMarketMap, nil
}

// ConsolidateDeFiMarkets takes a generated marketmap and attempts to move any DeFi markets to normal markets if the generated market
// has the same CMC ID as a normal market in actual.
//
// example:
// generated market: FOO,UNISWAP,0XFOOBAR/USD - CMC ID 4
// actual market:    FOO/USD - CMC ID 4
//
// result: FOO,UNISWAP,0XFOOBAR/USD ---becomes---> FOO/USD.
func ConsolidateDeFiMarkets(logger *zap.Logger, generated, actual mmtypes.MarketMap) (mmtypes.MarketMap, error) {
	generatedCMCIDMapping, err := getCMCTickerMapping(generated)
	if err != nil {
		return mmtypes.MarketMap{}, fmt.Errorf("failed to get CMC ID map for generated market map: %w", err)
	}
	actualCMCIDMapping, err := getCMCTickerMapping(actual)
	if err != nil {
		return mmtypes.MarketMap{}, fmt.Errorf("failed to get CMC ID map for actual market map: %w", err)
	}

	for cmcID, generatedTicker := range generatedCMCIDMapping {
		if actualTicker, ok := actualCMCIDMapping[cmcID]; ok {
			if generatedTicker != actualTicker {
				if isDefiTicker(generatedTicker) && !isDefiTicker(actualTicker) {
					logger.Debug("consolidating ticker", zap.String("generated", generatedTicker), zap.String("actual", actualTicker))
					generatedMarket := generated.Markets[generatedTicker]
					pair, err := connecttypes.CurrencyPairFromString(actualTicker)
					if err != nil {
						return mmtypes.MarketMap{}, fmt.Errorf("failed to convert ticker %s to currency pair: %w", actualTicker, err)
					}
					generatedMarket.Ticker.CurrencyPair = pair
					generated.Markets[actualTicker] = generatedMarket
					delete(generated.Markets, generatedTicker)
				}
			}
		}
	}
	return generated, nil
}

func isDefiTicker(ticker string) bool {
	return !connecttypes.IsLegacyAssetString(ticker)
}

// getCMCTickerMapping extracts a mapping of cmc ID's to ticker from the marketmap.
func getCMCTickerMapping(mm mmtypes.MarketMap) (map[string]string, error) {
	cmcIDToTickers := make(map[string]string)
	for ticker, market := range mm.Markets {
		if market.Ticker.Metadata_JSON != "" {
			var md tickermetadata.CoreMetadata
			if err := json.Unmarshal([]byte(market.Ticker.Metadata_JSON), &md); err != nil {
				return nil, fmt.Errorf("failed to unmarshal market metadata for %q: %w", ticker, err)
			}
			for _, aggID := range md.AggregateIDs {
				if aggID.Venue == types.VenueCoinMarketcap {
					if _, ok := cmcIDToTickers[aggID.ID]; ok {
						return nil, fmt.Errorf("duplicate cmc ID %q found for ticker %q", aggID.ID, ticker)
					}
					cmcIDToTickers[aggID.ID] = ticker
					break
				}
			}
		}
	}
	return cmcIDToTickers, nil
}
