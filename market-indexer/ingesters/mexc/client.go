package mexc

import (
	"context"
	"encoding/json"

	"github.com/skip-mev/connect-mmu/lib/http"
)

const (
	EndpointTickers = "https://api.mexc.com/api/v3/ticker/24hr"
)

var _ Client = &httpClient{}

// Client is an interface for a client that can interact with
// the Mexc api
//
//go:generate mockery --name Client --filename mock_mexc_client.go
type Client interface {
	// Tickers returns all tickers on the
	// Mexc.
	Tickers(context.Context) ([]TickerData, error)
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

func (h *httpClient) Tickers(ctx context.Context) ([]TickerData, error) {
	resp, err := h.client.GetWithContext(ctx, EndpointTickers)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var tickersResp []TickerData
	if err := json.NewDecoder(resp.Body).Decode(&tickersResp); err != nil {
		return nil, err
	}

	return tickersResp, nil
}
