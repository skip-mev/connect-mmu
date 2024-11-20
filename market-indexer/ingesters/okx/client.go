package okx

import (
	"context"
	"encoding/json"

	"github.com/skip-mev/connect-mmu/lib/http"
)

const (
	EndpointInstruments = "https://www.okx.com/api/v5/public/instruments?instType=SPOT"
	EndpointTickers     = "https://www.okx.com/api/v5/market/tickers?instType=SPOT"
)

var _ Client = &httpClient{}

// Client is an interface for a client that can interact with
// the Okx api
//
//go:generate mockery --name Client --filename mock_okx_client.go
type Client interface {
	// Instruments returns all instruments on okx.
	Instruments(context.Context) (InstrumentsResponse, error)

	// Tickers returns all tickers on okx.
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

// Instruments returns all instruments on the Okx API.
func (h *httpClient) Instruments(ctx context.Context) (InstrumentsResponse, error) {
	resp, err := h.client.GetWithContext(ctx, EndpointInstruments)
	if err != nil {
		return InstrumentsResponse{}, err
	}
	defer resp.Body.Close()

	var instrumentsResp InstrumentsResponse
	if err := json.NewDecoder(resp.Body).Decode(&instrumentsResp); err != nil {
		return InstrumentsResponse{}, err
	}

	if err := instrumentsResp.Validate(); err != nil {
		return InstrumentsResponse{}, err
	}

	return instrumentsResp, nil
}

// Tickers returns all tickers on the Okx API.
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

	if err := tickersResp.Validate(); err != nil {
		return TickersResponse{}, err
	}

	return tickersResp, nil
}
