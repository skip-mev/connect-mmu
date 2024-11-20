package gecko

import (
	"context"
	http2 "net/http"
	"time"

	"go.uber.org/zap"

	"github.com/skip-mev/connect-mmu/lib/http"
)

const (
	BaseEndpoint = "https://api.geckoterminal.com/api/v2"

	// https://www.geckoterminal.com/dex-api
	// we can only make 30 calls a minute.
	maxCalls     = 30
	callInterval = 1 * time.Minute
)

//go:generate mockery --name Client --filename mock_gecko_client.go
type Client interface {
	// MultiToken gets fetches data for the given token addresses on a given network.
	MultiToken(ctx context.Context, network string, tokens []string) (*TokensMultiResponse, error)
	// TopPools fetches the top pools from a given network and dex.
	// Note: this query uses pagination, and can query between 1 and 10 pages.
	TopPools(ctx context.Context, network, dex string, page int) (*PoolsResponse, error)
}

type geckoClientImpl struct {
	client       *http.Client
	logger       *zap.Logger
	baseEndpoint string

	limiter *APIRateLimiter
}

func newClient(logger *zap.Logger, baseEndpoint string) Client {
	if baseEndpoint == "" {
		baseEndpoint = BaseEndpoint
	}
	return &geckoClientImpl{
		client:       http.NewClient(),
		logger:       logger,
		baseEndpoint: baseEndpoint,
		limiter:      newRateLimiter(maxCalls, callInterval),
	}
}

func (c *geckoClientImpl) GetWithContext(ctx context.Context, endpoint string) (*http2.Response, error) {
	c.limiter.WaitForNextAvailableCall()
	return c.client.GetWithContext(ctx, endpoint)
}
