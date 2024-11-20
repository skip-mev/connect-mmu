package huobi

import (
	"context"
	"encoding/json"

	"github.com/skip-mev/connect-mmu/lib/http"
)

const EndpointTickers = "https://api.huobi.pro/market/tickers"

var _ Client = &httpClient{}

// Client is an interface for a client that can interact with
// the Huobi api
//
//go:generate mockery --name Client --filename mock_huobi_client.go
type Client interface {
	// Tickers gets all tickers from Huobi.
	Tickers(ctx context.Context) (TickersResponse, error)
}

type httpClient struct {
	client *http.Client
}

// NewClient is the default implementation
// of the Client using HTTP.
func NewClient() Client {
	return &httpClient{
		client: http.NewClient(),
	}
}

// Tickers returns all tickers on the Huobi API using an HTTP client.
func (h *httpClient) Tickers(ctx context.Context) (TickersResponse, error) {
	resp, err := h.client.GetWithContext(ctx, EndpointTickers)
	if err != nil {
		return TickersResponse{}, err
	}
	defer resp.Body.Close()

	var tickerResp TickersResponse
	if err := json.NewDecoder(resp.Body).Decode(&tickerResp); err != nil {
		return TickersResponse{}, err
	}

	if err := tickerResp.Validate(); err != nil {
		return TickersResponse{}, err
	}

	return tickerResp, nil
}
