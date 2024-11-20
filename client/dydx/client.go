package dydx

import (
	"context"

	"github.com/skip-mev/connect-mmu/lib/http"
)

var _ Client = &HTTPClient{}

//go:generate mockery --name Client  --filename mock_dydx_client.go
type Client interface {
	AllPerpetuals(ctx context.Context) (*AllPerpetualsResponse, error)
}

// HTTPClient represents a client for interacting with the dYdX API.
type HTTPClient struct {
	BaseURL string
	client  *http.Client
}

// NewHTTPClient creates a new dYdX API client.
func NewHTTPClient(baseURL string) Client {
	return &HTTPClient{
		BaseURL: baseURL,
		client:  http.NewClient(),
	}
}
