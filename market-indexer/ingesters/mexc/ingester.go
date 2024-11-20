package mexc

import (
	"context"

	"go.uber.org/zap"

	"github.com/skip-mev/connect-mmu/market-indexer/ingesters"
	"github.com/skip-mev/connect-mmu/market-indexer/ingesters/types"
	"github.com/skip-mev/connect-mmu/store/provider"
)

const (
	Name         = "mexc"
	ProviderName = Name + types.ProviderNameSuffixWS
)

var _ ingesters.Ingester = &Ingester{}

// Ingester is the kraken implementation of a market data Ingester.
type Ingester struct {
	logger *zap.Logger

	client Client
}

// New creates a new mexc Ingester.
func New(logger *zap.Logger) *Ingester {
	if logger == nil {
		panic("cannot set nil logger")
	}

	return &Ingester{
		logger: logger.With(zap.String("ingester", Name)),
		client: NewClient(),
	}
}

// NewWithClient creates a new mexc Ingester with the given Client.
func NewWithClient(logger *zap.Logger, client Client) *Ingester {
	if logger == nil {
		panic("cannot set nil logger")
	}

	return &Ingester{
		logger: logger.With(zap.String("ingester", Name)),
		client: client,
	}
}

func (i *Ingester) GetProviderMarkets(ctx context.Context) ([]provider.CreateProviderMarket, error) {
	i.logger.Info("fetching data")

	tickers, err := i.client.Tickers(ctx)
	if err != nil {
		i.logger.Error("failed to fetch tickers", zap.Error(err))
		return nil, err
	}

	i.logger.Info("fetched data", zap.Int("count", len(tickers)))

	pms := make([]provider.CreateProviderMarket, 0, len(tickers))
	for _, ticker := range tickers {
		i.logger.Debug("parsing", zap.Any("ticker", ticker))
		pm, err := ticker.toProviderMarket()
		if err != nil {
			i.logger.Error("failed to parse ticker", zap.Error(err))
			continue
		}

		pms = append(pms, pm)
	}

	return pms, nil
}

// Name returns the Ingester's human-readable name.
func (i *Ingester) Name() string {
	return Name
}
