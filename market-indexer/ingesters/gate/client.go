package gate

import (
	"context"
	"encoding/json"

	"github.com/skip-mev/connect-mmu/lib/http"
)

const (
	EndpointTickers = "https://api.gateio.ws/api/v4/spot/tickers"
)

var _ Client = &httpClient{}

// Client is an interface for a client that can interact with
// the Gate.io api
//
//go:generate mockery --name Client --filename mock_gate_client.go
type Client interface {
	// Tickers returns all tickers on the
	// Gate.io.
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

	var tickers []TickerData
	if err := json.NewDecoder(resp.Body).Decode(&tickers); err != nil {
		return nil, err
	}

	return tickers, nil
}
