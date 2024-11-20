package bybit

import (
	"context"
	"encoding/json"

	"github.com/skip-mev/connect-mmu/lib/http"
)

const (
	EndpointInstruments = "https://api.bybit.com/v5/market/instruments-info?category=spot"
	EndpointTickers     = "https://api.bybit.com/v5/market/tickers?category=spot"
)

var _ Client = &httpClient{}

// Client is an interface for getting data from ByBit.
//
//go:generate mockery --name Client --filename mock_bybit_client.go
type Client interface {
	// Instruments gets all instruments from ByBit.
	Instruments(ctx context.Context) (InstrumentsResponse, error)

	// Tickers gets all tickers from ByBit.
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
