package bitfinex

import (
	"context"
	"encoding/json"

	"github.com/skip-mev/connect-mmu/lib/http"
)

const (
	EndpointTickers = "https://api-pub.bitfinex.com/v2/tickers?symbols=ALL"
)

var _ Client = &httpClient{}

// Client is an interface for getting data from Bitfinex.
//
//go:generate mockery --name Client --filename mock_bitfinex_client.go
type Client interface {
	// Tickers gets all tickers from Bitfinex.
	Tickers(ctx context.Context) ([][]interface{}, error)
}

type httpClient struct {
	client *http.Client
}

func NewHTTPClient() Client {
	return &httpClient{
		client: http.NewClient(),
	}
}

// Tickers returns all tickers from Bitfinex using the REST API.
func (h *httpClient) Tickers(ctx context.Context) ([][]interface{}, error) {
	resp, err := h.client.GetWithContext(ctx, EndpointTickers)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var respI [][]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&respI); err != nil {
		return nil, err
	}

	return respI, nil
}
