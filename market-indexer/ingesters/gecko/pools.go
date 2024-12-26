package gecko

import (
	"context"
	"encoding/json"
	"fmt"
	"math/big"
	"strconv"
	"strings"
	"time"
	"unicode"

	"go.uber.org/zap"

	"github.com/skip-mev/connect-mmu/lib/symbols"
	"github.com/skip-mev/connect-mmu/market-indexer/ingesters/types"
)

// https://www.geckoterminal.com/dex-api
// /networks/{network}/dexes/{dex}/pools
//
// see testdata/pools_response_example.json for example json response.

// PoolData is the data for a dex pool.
type PoolData struct {
	ID         string `json:"id"`
	Type       string `json:"type"`
	Attributes struct {
		BaseTokenPriceUsd             string    `json:"base_token_price_usd"`
		BaseTokenPriceNativeCurrency  string    `json:"base_token_price_native_currency"`
		QuoteTokenPriceUsd            string    `json:"quote_token_price_usd"`
		QuoteTokenPriceNativeCurrency string    `json:"quote_token_price_native_currency"`
		BaseTokenPriceQuoteToken      string    `json:"base_token_price_quote_token"`
		QuoteTokenPriceBaseToken      string    `json:"quote_token_price_base_token"`
		Address                       string    `json:"address"`
		Name                          string    `json:"name"`
		PoolCreatedAt                 time.Time `json:"pool_created_at"`
		FdvUsd                        string    `json:"fdv_usd"`
		MarketCapUsd                  any       `json:"market_cap_usd"`
		PriceChangePercentage         struct {
			M5  string `json:"m5"`
			H1  string `json:"h1"`
			H6  string `json:"h6"`
			H24 string `json:"h24"`
		} `json:"price_change_percentage"`
		Transactions struct {
			M5 struct {
				Buys    int `json:"buys"`
				Sells   int `json:"sells"`
				Buyers  int `json:"buyers"`
				Sellers int `json:"sellers"`
			} `json:"m5"`
			M15 struct {
				Buys    int `json:"buys"`
				Sells   int `json:"sells"`
				Buyers  int `json:"buyers"`
				Sellers int `json:"sellers"`
			} `json:"m15"`
			M30 struct {
				Buys    int `json:"buys"`
				Sells   int `json:"sells"`
				Buyers  int `json:"buyers"`
				Sellers int `json:"sellers"`
			} `json:"m30"`
			H1 struct {
				Buys    int `json:"buys"`
				Sells   int `json:"sells"`
				Buyers  int `json:"buyers"`
				Sellers int `json:"sellers"`
			} `json:"h1"`
			H24 struct {
				Buys    int `json:"buys"`
				Sells   int `json:"sells"`
				Buyers  int `json:"buyers"`
				Sellers int `json:"sellers"`
			} `json:"h24"`
		} `json:"transactions"`
		VolumeUsd struct {
			M5  string `json:"m5"`
			H1  string `json:"h1"`
			H6  string `json:"h6"`
			H24 string `json:"h24"`
		} `json:"volume_usd"`
		ReserveInUsd string `json:"reserve_in_usd"`
	} `json:"attributes"`
	Relationships struct {
		BaseToken struct {
			Data struct {
				ID   string `json:"id"`
				Type string `json:"type"`
			} `json:"data"`
		} `json:"base_token"`
		QuoteToken struct {
			Data struct {
				ID   string `json:"id"`
				Type string `json:"type"`
			} `json:"data"`
		} `json:"quote_token"`
		Dex struct {
			Data struct {
				ID   string `json:"id"`
				Type string `json:"type"`
			} `json:"data"`
		} `json:"dex"`
	} `json:"relationships"`
}

// PoolsResponse is the underlying response format from the /networks/{network}/dexes/{dex}/pools query.
type PoolsResponse struct {
	Data []PoolData `json:"data"`
}

const maxPages = 10

func (c *geckoClientImpl) TopPools(ctx context.Context, network, dex string, page int) (*PoolsResponse, error) {
	if network == "" {
		return nil, fmt.Errorf("network is required")
	}
	if dex == "" {
		return nil, fmt.Errorf("dex is required")
	}
	if page <= 0 || page > maxPages {
		return nil, fmt.Errorf("page must be between 1 and %d", maxPages)
	}

	sortOption := "h24_volume_usd_desc"
	endpoint := fmt.Sprintf("%s/networks/%s/dexes/%s/pools?page=%d&sort=%s", c.baseEndpoint, network, dex, page, sortOption)
	c.logger.Debug("getting top pools", zap.String("network", network), zap.String("dex", dex), zap.Int("page", page))
	resp, err := c.GetWithContext(ctx, endpoint)
	if err != nil {
		return nil, fmt.Errorf("gecko geckoClientImpl: failed to fetch TopPools: %w", err)
	}

	var poolsRes PoolsResponse
	if err := json.NewDecoder(resp.Body).Decode(&poolsRes); err != nil {
		return nil, fmt.Errorf("gecko geckoClientImpl: failed to JSON decode TopPools response: %w", err)
	}

	return &poolsRes, nil
}

// UsdVolume returns the 24h USD volume.
func (p *PoolData) UsdVolume() (float64, error) {
	h24VolUSD, ok := new(big.Float).SetString(p.Attributes.VolumeUsd.H24)
	if !ok {
		return 0, fmt.Errorf("unable to convert VolumeUsd.H24 to big.Float: %s", p.Attributes.VolumeUsd.H24)
	}

	vol, _ := h24VolUSD.Float64()
	return vol, nil
}

// QuoteVolume returns the 24h quote volume.
// Formula: 24h volume in usd / quote price in USD.
func (p *PoolData) QuoteVolume() (*big.Float, error) {
	h24VolUSD, _ := new(big.Float).SetString(p.Attributes.VolumeUsd.H24)
	if h24VolUSD == nil {
		return nil, fmt.Errorf("unable to convert VolumeUsd.H24 to big.Float: %s", p.Attributes.VolumeUsd.H24)
	}
	priceInUSD, ok := new(big.Float).SetString(p.Attributes.QuoteTokenPriceUsd)
	if !ok {
		return nil, fmt.Errorf("unable to convert QuoteTokenPriceUsd to big.Float: %s", p.Attributes.QuoteTokenPriceUsd)
	}

	quoteVolume := new(big.Float).Quo(h24VolUSD, priceInUSD)
	return quoteVolume, nil
}

// GetBaseAndQuoteTokenAddress gets the token address for the base and quote tokens.
func (p *PoolData) GetBaseAndQuoteTokenAddress() (base string, quote string) {
	return p.BaseAddress(), p.QuoteAddress()
}

func (p *PoolData) BaseAddress() string {
	unformattedBase := p.Relationships.BaseToken.Data.ID
	base := getAfterUnderscore(unformattedBase)
	return base
}

func (p *PoolData) QuoteAddress() string {
	unformattedQuote := p.Relationships.QuoteToken.Data.ID
	quote := getAfterUnderscore(unformattedQuote)
	return quote
}

func (p *PoolData) NetworkName() string {
	idx := strings.Index(p.ID, "_")
	if idx == -1 {
		return p.ID
	}
	return p.ID[:idx]
}

func (p *PoolData) Symbol() string {
	return strings.ReplaceAll(p.Attributes.Name, " ", "")
}

func (p *PoolData) Venue() string {
	return p.Relationships.Dex.Data.ID
}

func (p *PoolData) VenueAddress() string {
	return p.Attributes.Address
}

// Base returns the properly formated base symbol.
func (p *PoolData) Base() (string, error) {
	split, err := p.SplitSymbol()
	if err != nil {
		return "", err
	}

	s, err := symbols.ToTickerString(split[0])
	if err != nil {
		return "", err
	}

	// remove all numbers and decimals. Gecko pairs are returned as QNT/WETH0.3, and we want QNT/WETH.
	return removeTrailingNumbers(s), nil
}

func (p *PoolData) SplitSymbol() ([]string, error) {
	split := strings.Split(p.Symbol(), "/")
	if len(split) != 2 {
		return nil, fmt.Errorf("invalid symbol: %s", p.Symbol())
	}

	return split, nil
}

// Quote returns the properly formated quote symbol.
func (p *PoolData) Quote() (string, error) {
	split, err := p.SplitSymbol()
	if err != nil {
		return "", err
	}

	s, err := symbols.ToTickerString(split[1])
	if err != nil {
		return "", err
	}

	// remove all numbers and decimals. Gecko pairs are returned as QNT/WETH0.3, and we want QNT/WETH.
	return removeTrailingNumbers(s), nil
}

func (p *PoolData) ReferencePrice() (float64, error) {
	ref, err := strconv.ParseFloat(p.Attributes.BaseTokenPriceQuoteToken, 64)
	if err != nil {
		return 0, fmt.Errorf("geckoClientImpl: failed to parse reference price: %w", err)
	}
	return ref, nil
}

func (p *PoolData) Liquidity() (float64, error) {
	liquidity, err := strconv.ParseFloat(p.Attributes.ReserveInUsd, 64)
	if err != nil {
		return 0, fmt.Errorf("geckoClientImpl: failed to parse liquidity: %w", err)
	}
	return liquidity, nil
}

func (p *PoolData) OffChainTicker() (string, error) {
	targetBase, err := p.Base()
	if err != nil {
		return "", err
	}

	targetQuote, err := p.Quote()
	if err != nil {
		return "", err
	}

	targetBaseOffchain := strings.Join([]string{
		targetBase,
		geckoDexToConnectTickerVenue(p.Venue()),
		p.BaseAddress(),
	}, types.DefiTickerDelimiter)

	targetQuoteOffchain := strings.Join([]string{
		targetQuote,
		geckoDexToConnectTickerVenue(p.Venue()),
		p.QuoteAddress(),
	}, types.DefiTickerDelimiter)

	return strings.ToUpper(strings.Join([]string{
		targetBaseOffchain,
		targetQuoteOffchain,
	}, types.TickerSeparator)), nil
}

func removeTrailingNumbers(s string) string {
	// TrimRightFunc will remove characters from the end of the string
	// as long as the condition is true. Here, we check if the character is a digit.
	return strings.TrimRightFunc(s, func(r rune) bool {
		if r == '.' {
			return true
		}

		return unicode.IsDigit(r)
	})
}
