package gate

import (
	"context"

	"go.uber.org/zap"

	"github.com/skip-mev/connect-mmu/market-indexer/ingesters"
	"github.com/skip-mev/connect-mmu/market-indexer/ingesters/types"
	"github.com/skip-mev/connect-mmu/store/provider"
)

const (
	Name         = "gate"
	ProviderName = Name + types.ProviderNameSuffixWS
)

var _ ingesters.Ingester = &Ingester{}

// Ingester is the gate implementation of a market data Ingester.
type Ingester struct {
	logger *zap.Logger

	client Client
}

// New creates a new gate Ingester.
func New(logger *zap.Logger) *Ingester {
	if logger == nil {
		panic("cannot set nil logger")
	}

	return &Ingester{
		logger: logger.With(zap.String("ingester", Name)),
		client: NewClient(),
	}
}

// NewWithClient creates a new gate Ingester with the given Client.
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
	tickers, err := i.client.Tickers(ctx)
	if err != nil {
		return nil, err
	}

	pms := make([]provider.CreateProviderMarket, 0, len(tickers))
	for _, ticker := range tickers {
		i.logger.Debug("parsing", zap.Any("ticker", ticker))

		if !ticker.isSpot() {
			i.logger.Debug("ignoring non-spot market", zap.Any("ticker", ticker))
			continue
		}

		pm, err := ticker.toProviderMarket()
		if err != nil {
			i.logger.Error("failed to convert ticker to providerMarket", zap.Error(err))
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
