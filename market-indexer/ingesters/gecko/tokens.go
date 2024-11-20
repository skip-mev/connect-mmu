package gecko

import (
	"context"
	"encoding/json"
	"fmt"

	"go.uber.org/zap"

	sauronslices "github.com/skip-mev/connect-mmu/lib/slices"
	sauronstrings "github.com/skip-mev/connect-mmu/lib/strings"
)

// https://www.geckoterminal.com/dex-api
// /networks/{network}/tokens/multi/{comma,separated,addresses}
//
// see testdata/tokens_multi_response.json for example json response.

// TokenData is the data for a token.
type TokenData struct {
	ID         string `json:"id"`
	Type       string `json:"type"`
	Attributes struct {
		Address           string `json:"address"`
		Name              string `json:"name"`
		Symbol            string `json:"symbol"`
		ImageURL          string `json:"image_url"`
		CoingeckoCoinID   string `json:"coingecko_coin_id"`
		Decimals          int    `json:"decimals"`
		TotalSupply       string `json:"total_supply"`
		PriceUsd          string `json:"price_usd"`
		FdvUsd            string `json:"fdv_usd"`
		TotalReserveInUsd string `json:"total_reserve_in_usd"`
		VolumeUsd         struct {
			H24 string `json:"h24"`
		} `json:"volume_usd"`
		MarketCapUsd any `json:"market_cap_usd"`
	} `json:"attributes"`
	Relationships struct {
		TopPools struct {
			Data []struct {
				ID   string `json:"id"`
				Type string `json:"type"`
			} `json:"data"`
		} `json:"top_pools"`
	} `json:"relationships"`
}

func (t TokenData) Decimals() int {
	return t.Attributes.Decimals
}

// TokensMultiResponse is the underlying response format from the /networks/{network}/tokens/multi/{addresses} query.
type TokensMultiResponse struct {
	Data []TokenData `json:"data,omitempty"`
}

// maxTokens is the maximum allowed tokens you can query at one time.
// https://www.geckoterminal.com/dex-api
// see: /networks/{network}/tokens/multi/{addresses}.
const maxTokens = 30

func (c *geckoClientImpl) MultiToken(ctx context.Context, network string, tokens []string) (*TokensMultiResponse, error) {
	if network == "" {
		return nil, fmt.Errorf("network is required")
	}
	if len(tokens) == 0 {
		return nil, fmt.Errorf("tokens must be non-zero")
	}

	// we need to chunk tokens just in case someone passes in more than 30 tokens.
	// so if someone passes in 35 tokens, it will query the first 30, then the last 5, and combine the results.
	chunkedTokens := sauronslices.Chunk(tokens, maxTokens)

	response := TokensMultiResponse{Data: make([]TokenData, 0)}
	for _, chunk := range chunkedTokens {
		endpoint := fmt.Sprintf("%s/networks/%s/tokens/multi/%s", c.baseEndpoint, network, sauronstrings.CommaSeparate(chunk))
		c.logger.Debug("getting tokens", zap.String("network", network), zap.Strings("tokens", chunk))
		res, err := c.GetWithContext(ctx, endpoint)
		if err != nil {
			return nil, fmt.Errorf("gecko client: failed to fetch tokens multi: %w", err)
		}

		var tokensMultiResponse TokensMultiResponse
		if err := json.NewDecoder(res.Body).Decode(&tokensMultiResponse); err != nil {
			return nil, fmt.Errorf("gecko client: failed to failed to JSON decode tokens multi response: %w", err)
		}

		response.Data = append(response.Data, tokensMultiResponse.Data...)
	}

	return &response, nil
}
