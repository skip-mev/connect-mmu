package dydx

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/cosmos/cosmos-sdk/types/query"
)

const (
	// PERPETUAL_MARKET_TYPE_UNSPECIFIED is an unspecified market type.
	PERPETUAL_MARKET_TYPE_UNSPECIFIED = "PERPETUAL_MARKET_TYPE_UNSPECIFIED"
	// PERPETUAL_MARKET_TYPE_CROSS is a Market type for cross margin perpetual markets.
	PERPETUAL_MARKET_TYPE_CROSS = "PERPETUAL_MARKET_TYPE_CROSS"
	// PERPETUAL_MARKET_TYPE_ISOLATED is a Market type for isolated margin perpetual markets.
	PERPETUAL_MARKET_TYPE_ISOLATED = "PERPETUAL_MARKET_TYPE_ISOLATED"
)

// Perpetual represents a perpetual on the dYdX exchange.
type Perpetual struct {
	Params       PerpetualParams `json:"params"`
	FundingIndex string          `json:"funding_index"`
	OpenInterest string          `json:"open_interest"`
}

// PerpetualParams represents the parameters of a perpetual on the dYdX exchange.
type PerpetualParams struct {
	Ticker     string `json:"ticker"`
	MarketType string `json:"market_type"`
}

// AllPerpetualsResponse is the response type for the AllPerpetuals RPC method.
type AllPerpetualsResponse struct {
	Perpetuals []Perpetual `json:"perpetual"`
	Pagination struct {
		NextKey string `json:"next_key"`
		Total   string `json:"total"`
	} `json:"pagination"`
}

// GetPagination implements the saurongrpc ResponseWithPagination interface.
func (r *AllPerpetualsResponse) GetPagination() *query.PageResponse {
	// unmarshal base64 next key
	nextKey, err := base64.StdEncoding.DecodeString(r.Pagination.NextKey)
	if err != nil {
		return nil
	}

	total, err := strconv.ParseUint(r.Pagination.Total, 10, 64)
	if err != nil {
		return nil
	}

	return &query.PageResponse{
		NextKey: nextKey,
		Total:   total,
	}
}

// AllPerpetuals retrieves all perpetuals from the dYdX API.
func (c *HTTPClient) AllPerpetuals(ctx context.Context) (*AllPerpetualsResponse, error) {
	baseURL := fmt.Sprintf("%s/dydxprotocol/perpetuals/perpetual", c.BaseURL)
	resp, err := c.client.GetWithContext(ctx, baseURL)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var actualResult AllPerpetualsResponse
	if err := json.NewDecoder(resp.Body).Decode(&actualResult); err != nil {
		return nil, fmt.Errorf("error decoding response: %w", err)
	}

	lastKey := ""
	nextKey := actualResult.Pagination.NextKey
	for nextKey != "" {
		paginatedResponse, err := c.client.GetWithContext(ctx, baseURL+"?pagination.key="+nextKey)
		if err != nil {
			return nil, err
		}

		if paginatedResponse.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("unexpected status code: %d", paginatedResponse.StatusCode)
		}

		var result AllPerpetualsResponse
		if err := json.NewDecoder(paginatedResponse.Body).Decode(&result); err != nil {
			return nil, fmt.Errorf("error decoding response: %w", err)
		}

		actualResult.Perpetuals = append(actualResult.Perpetuals, result.Perpetuals...)

		lastKey = nextKey
		nextKey = result.Pagination.NextKey

		if lastKey == nextKey {
			return nil, fmt.Errorf("error saw repeat pagination key: %s", nextKey)
		}
	}

	return &actualResult, nil
}
