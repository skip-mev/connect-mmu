package kraken

import (
	"context"
	"encoding/json"

	"github.com/skip-mev/connect-mmu/lib/http"
)

const (
	EndpointAssets  = "https://api.kraken.com/0/public/AssetPairs"
	EndpointTickers = "https://api.kraken.com/0/public/Ticker"
)

var _ Client = &httpClient{}

// Client is an interface for getting data from kraken.
//
//go:generate mockery --name Client --filename mock_kraken_client.go
type Client interface {
	// AssetPairs gets all asset pair from Kraken.
	AssetPairs(ctx context.Context) (AssetPairsResponse, error)

	// Tickers gets all ticker from Kraken.
	Tickers(ctx context.Context) (TickersResponse, error)
}

type httpClient struct {
	client *http.Client
}

func NewHTTPClient() Client {
	return &httpClient{
		client: http.NewClient(),
	}
}

func (h *httpClient) AssetPairs(ctx context.Context) (AssetPairsResponse, error) {
	resp, err := h.client.GetWithContext(ctx, EndpointAssets)
	if err != nil {
		return AssetPairsResponse{}, err
	}
	defer resp.Body.Close()

	var getAssetsResp AssetPairsResponse
	if err := json.NewDecoder(resp.Body).Decode(&getAssetsResp); err != nil {
		return AssetPairsResponse{}, err
	}

	return getAssetsResp, nil
}

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

	return tickerResp, nil
}
