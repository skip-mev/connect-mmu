package coinbase

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/skip-mev/connect-mmu/lib/http"
)

const (
	allTickersEndpoint     = "https://api.exchange.coinbase.com/products"
	statsPerTickerEndpoint = "https://api.exchange.coinbase.com/products/stats"
)

// Client is an interface for a client that can interact with
// the coinbase api
//
//go:generate mockery --name Client --filename mock_coinbase_client.go
type Client interface {
	// Products returns the list of all products from the coinbase api
	// in accordance with the products api
	Products(context.Context) (Products, error)

	// Stats returns the stats for all markets from the coinbase api
	// in accordance with the products/stats api
	Stats(context.Context) (Stats, error)
}

// NewHTTPCoinbaseClient creates a new coinbase client that interacts with
// the coinbase api over HTTP.
func NewHTTPCoinbaseClient() Client {
	return &httpCoinbaseClient{
		client: http.NewClient(),
	}
}

var _ Client = &httpCoinbaseClient{}

// httpCoinbaseClient is the default implementation of the coinbaseClient
// over HTTP.
type httpCoinbaseClient struct {
	// client is the http client used to make requests to the coinbase api
	client *http.Client
}

func (c *httpCoinbaseClient) Products(ctx context.Context) (Products, error) {
	// query the all tickers endpoint
	resp, err := c.client.GetWithContext(ctx, allTickersEndpoint)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var products Products
	if err := json.NewDecoder(resp.Body).Decode(&products); err != nil {
		return nil, fmt.Errorf("failed to decode response from %s-ingester: %w", Name, err)
	}

	return products, nil
}

func (c *httpCoinbaseClient) Stats(ctx context.Context) (Stats, error) {
	// query the stats per ticker endpoint
	resp, err := c.client.GetWithContext(ctx, statsPerTickerEndpoint)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var stats Stats
	if err := json.NewDecoder(resp.Body).Decode(&stats); err != nil {
		return nil, fmt.Errorf("failed to decode response from %s-ingester: %w", Name, err)
	}

	return stats, nil
}
