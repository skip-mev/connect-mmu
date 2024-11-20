//nolint:revive
package crypto_com

import (
	"context"
	"encoding/json"

	"github.com/skip-mev/connect-mmu/lib/http"
)

const (
	EndpointInstruments = "https://api.crypto.com/exchange/v1/public/get-instruments"
	EndpointTickers     = "https://api.crypto.com/exchange/v1/public/get-tickers"
)

// Client is an interface for a client that can interact with
// the crypto.com api.
//
//go:generate mockery --name Client --filename mock_crypto_com_client.go
type Client interface {
	// Instruments returns the list of all instruments from crypto.com.
	Instruments(context.Context) (InstrumentsResponse, error)

	// Tickers returns the list of all tickers from crypto.com.
	Tickers(context.Context) (TickersResponse, error)
}

// NewHTTPClient creates a new coinbase client that interacts with
// the crypto.com api over HTTP.
func NewHTTPClient() Client {
	return &httpClient{
		client: http.NewClient(),
	}
}

var _ Client = &httpClient{}

// httpClient is the default implementation of the Client
// over HTTP.
type httpClient struct {
	client *http.Client
}

// Instruments queries the crypto.com API for instruments using its http client.
func (h *httpClient) Instruments(ctx context.Context) (InstrumentsResponse, error) {
	resp, err := h.client.GetWithContext(ctx, EndpointInstruments)
	if err != nil {
		return InstrumentsResponse{}, err
	}
	defer resp.Body.Close()

	var getInstrResp InstrumentsResponse
	if err := json.NewDecoder(resp.Body).Decode(&getInstrResp); err != nil {
		return InstrumentsResponse{}, err
	}

	return getInstrResp, nil
}

// Tickers queries the crypto.com API for tickers using its http client.
func (h *httpClient) Tickers(ctx context.Context) (TickersResponse, error) {
	resp, err := h.client.GetWithContext(ctx, EndpointTickers)
	if err != nil {
		return TickersResponse{}, err
	}
	defer resp.Body.Close()

	var getTickersResp TickersResponse
	if err := json.NewDecoder(resp.Body).Decode(&getTickersResp); err != nil {
		return TickersResponse{}, err
	}

	return getTickersResp, nil
}
