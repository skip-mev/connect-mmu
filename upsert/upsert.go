package upsert

import (
	"errors"
	"fmt"
	"slices"

	"github.com/skip-mev/connect/v2/x/marketmap/types"
	"go.uber.org/zap"

	"github.com/skip-mev/connect-mmu/config"
	"github.com/skip-mev/connect-mmu/upsert/strategy"
)

// Generator is a type that facilitates generating market upserts.
type Generator struct {
	logger      *zap.Logger
	cfg         config.UpsertConfig
	generatedMM types.MarketMap
	currentMM   types.MarketMap
}

// New returns a new upsert generator.
func New(
	logger *zap.Logger,
	cfg config.UpsertConfig,
	generated, current types.MarketMap,
) (*Generator, error) {
	var err error
	current, err = current.GetValidSubset()
	if err != nil {
		return nil, fmt.Errorf("failed to get valid subset of markets from on-chain marketmap: %w", err)
	}

	generated, err = generated.GetValidSubset()
	if err != nil {
		return nil, fmt.Errorf("failed to get valid subset of markets from generated marketmap: %w", err)
	}

	return &Generator{
		logger:      logger,
		cfg:         cfg,
		generatedMM: generated,
		currentMM:   current,
	}, nil
}

// GenerateUpserts generates a slice of market upserts.
func (d *Generator) GenerateUpserts() ([]types.Market, error) {
	upserts, err := strategy.GetMarketMapUpserts(d.logger, d.currentMM, d.generatedMM)
	if err != nil {
		d.logger.Error("failed to determine upserts", zap.Error(err))
		return nil, err
	}
	upserts = removeFromUpserts(upserts, d.cfg.RestrictedMarkets)
	d.logger.Info("determined upserts", zap.Int("upserts", len(upserts)))

	// reorder so that any new normalize by markets are first
	upserts, err = orderNormalizeMarketsFirst(upserts)
	if err != nil {
		d.logger.Error("failed to reorder upserts", zap.Error(err))
		return nil, err
	}

	// early exit if there are no upserts
	if len(upserts) == 0 {
		d.logger.Info("no upserts found - returning")
		return upserts, nil
	}

	errs := make([]error, 0)
	for _, upsert := range upserts {
		if err := upsert.ValidateBasic(); err != nil {
			errs = append(errs, err)
		}
	}
	if len(errs) > 0 {
		return nil, fmt.Errorf("generated %d invalid market(s): %w", len(errs), errors.Join(errs...))
	}

	if err := validateUpserts(d.currentMM, upserts); err != nil {
		return nil, fmt.Errorf("generated invalid upserts in marketmap: %w", err)
	}

	return upserts, nil
}

// validateUpserts adds the upserts to a marketmap, and validates the configuration.
func validateUpserts(currentMM types.MarketMap, upserts []types.Market) error {
	for _, upsert := range upserts {
		currentMM.Markets[upsert.Ticker.String()] = upsert
	}
	return currentMM.ValidateBasic()
}

// removeFromUpserts removes the specified markets from the upserts slice.
func removeFromUpserts(upserts []types.Market, remove []string) []types.Market {
	if len(remove) == 0 {
		return upserts
	}

	if len(upserts) == 0 {
		return nil
	}

	filtered := make([]types.Market, 0)
	for _, upsert := range upserts {
		if !slices.Contains(remove, upsert.Ticker.String()) {
			filtered = append(filtered, upsert)
		}
	}
	return filtered
}

// orderNormalizeMarketsFirst reorders markets such that markets that are used in "normalize_by_pair" are ordered first.
func orderNormalizeMarketsFirst(upserts []types.Market) ([]types.Market, error) {
	output := make([]types.Market, 0, len(upserts))

	// create map for checking
	upsertsAsMap := make(map[string]types.Market)
	for _, upsert := range upserts {
		upsertsAsMap[upsert.Ticker.String()] = upsert
	}

	seenNormalizeBys := make(map[string]struct{})
	for _, upsert := range upserts {
		for _, pc := range upsert.ProviderConfigs {
			if pc.NormalizeByPair != nil {
				ticker := pc.NormalizeByPair.String()

				// if the normalize pair exists in our upserts, assume it is newly added
				// it may be just being updated, but moving it to the front will have no side effects
				if market, found := upsertsAsMap[ticker]; found {
					if _, ok := seenNormalizeBys[ticker]; !ok {
						// push back this pair to our array if we have not seen it yet
						output = append(output, market)
						seenNormalizeBys[ticker] = struct{}{}
					}
				}
			}
		}
	}

	// push back remaining markets
	for ticker, market := range upsertsAsMap {
		if _, ok := seenNormalizeBys[ticker]; !ok {
			output = append(output, market)
		}
	}

	if len(output) != len(upserts) {
		return nil, fmt.Errorf("invalid reorder: expected %d outputs, got %d", len(upserts), len(output))
	}

	return output, nil
}
