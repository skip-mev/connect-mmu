package kucoin

import (
	"context"
	"encoding/json"

	"github.com/skip-mev/connect-mmu/lib/http"
)

const (
	EndpointTickers = "https://api.kucoin.com/api/v1/market/allTickers"
)

var _ Client = &httpClient{}

// Client is an interface for a client that can interact with
// the KuCoin api.
//
//go:generate mockery --name Client --filename mock_kucoin_client.go
type Client interface {
	// Tickers returns all tickers on KuCoin.
	Tickers(context.Context) (TickersResponse, error)
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

func (h *httpClient) Tickers(ctx context.Context) (TickersResponse, error) {
	resp, err := h.client.GetWithContext(ctx, EndpointTickers)
	if err != nil {
		return TickersResponse{}, err
	}
	defer resp.Body.Close()

	var tickersResp TickersResponse
	if err := json.NewDecoder(resp.Body).Decode(&tickersResp); err != nil {
		return TickersResponse{}, err
	}

	return tickersResp, nil
}
