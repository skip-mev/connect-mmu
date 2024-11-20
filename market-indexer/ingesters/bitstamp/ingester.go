package bitstamp

import (
	"context"

	"go.uber.org/zap"

	"github.com/skip-mev/connect-mmu/market-indexer/ingesters"
	"github.com/skip-mev/connect-mmu/market-indexer/ingesters/types"
	"github.com/skip-mev/connect-mmu/store/provider"
)

const (
	Name         = "bitstamp"
	ProviderName = Name + types.ProviderNameSuffixAPI
)

var _ ingesters.Ingester = &Ingester{}

// Ingester is the bitstamp implementation of a market data Ingester.
type Ingester struct {
	logger *zap.Logger

	client Client
}

// New creates a new bitstamp Ingester.
func New(logger *zap.Logger) *Ingester {
	if logger == nil {
		panic("cannot set nil logger")
	}

	return &Ingester{
		logger: logger.With(zap.String("ingester", Name)),
		client: NewClient(),
	}
}

// NewWithClient creates a new okx Ingester with the given Client.
func NewWithClient(logger *zap.Logger, client Client) *Ingester {
	if logger == nil {
		panic("cannot set nil logger")
	}

	return &Ingester{
		logger: logger.With(zap.String("ingester", Name)),
		client: client,
	}
}

func (ig *Ingester) GetProviderMarkets(ctx context.Context) ([]provider.CreateProviderMarket, error) {
	ig.logger.Info("fetching data")

	tickers, err := ig.client.Tickers(ctx)
	if err != nil {
		ig.logger.Error("failed to fetch tickers", zap.Error(err))
		return nil, err
	}

	ig.logger.Info("fetched data", zap.Int("count", len(tickers)))

	pms := make([]provider.CreateProviderMarket, 0, len(tickers))
	for _, ticker := range tickers {
		ig.logger.Debug("parsing", zap.Any("ticker", ticker))

		pm, err := ticker.toProviderMarket()
		if err != nil {
			return nil, err
		}

		pms = append(pms, pm)
	}

	ig.logger.Info("creates", zap.Int("count", len(pms)))

	return pms, nil
}

// Name returns the Ingester's human-readable name.
func (ig *Ingester) Name() string {
	return Name
}
