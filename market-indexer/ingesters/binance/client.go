package binance

import (
	"context"
	"encoding/json"

	"github.com/skip-mev/connect-mmu/lib/http"
)

const (
	EndpointTickers = "https://api.binance.com/api/v3/ticker/24hr"
)

var _ Client = &httpClient{}

// Client is an interface for getting data from Binance.
//
//go:generate mockery --name Client --filename mock_binance_client.go
type Client interface {
	// Tickers gets all tickers from Binance.
	Tickers(ctx context.Context) ([]TickerData, error)
}

type httpClient struct {
	client *http.Client
}

func NewHTTPClient() Client {
	return &httpClient{
		client: http.NewClient(),
	}
}

// Tickers gets all tickers from Binance using the HTTP client.
func (h *httpClient) Tickers(ctx context.Context) ([]TickerData, error) {
	resp, err := h.client.GetWithContext(ctx, EndpointTickers)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var tickers []TickerData
	if err := json.NewDecoder(resp.Body).Decode(&tickers); err != nil {
		return nil, err
	}

	return tickers, nil
}
