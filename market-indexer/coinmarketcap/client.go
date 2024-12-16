package coinmarketcap

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/skip-mev/connect-mmu/lib/http"
)

const (
	EndpointCryptoMap       = "https://pro-api.coinmarketcap.com/v1/cryptocurrency/map"
	EndpointExchangeMap     = "https://pro-api.coinmarketcap.com/v1/exchange/map"
	EndpointExchangeAssets  = "https://pro-api.coinmarketcap.com/v1/exchange/assets?id=%d"
	EndpointExchangeMarkets = "https://pro-api.coinmarketcap.com/v1/exchange/market-pairs/latest?id=%d&limit=5000"
	EndpointFiatMap         = "https://pro-api.coinmarketcap.com/v1/fiat/map"
	EndpointQuote           = "https://pro-api.coinmarketcap.com/v2/cryptocurrency/quotes/latest?id=%s"
	EndpointInfo            = "https://pro-api.coinmarketcap.com/v2/cryptocurrency/info?id=%s"
)

var _ Client = &httpClient{}

// Client is an interface for getting data from CoinMarketCap.
//
//go:generate mockery --name Client --filename mock_coinmarketcap_client.go
type Client interface {
	// CryptoIDMap gets the full currency ID Map from CoinMarketMap.
	CryptoIDMap(ctx context.Context) (CryptoIDMapResponse, error)

	// ExchangeIDMap gets the full exchange ID Map from CoinMarketMap.
	ExchangeIDMap(ctx context.Context) (ExchangeIDMapResponse, error)

	// ExchangeAssets gets the full list of assets for the given exchange.
	ExchangeAssets(ctx context.Context, exchange int) (ExchangeAssetsResponse, error)

	// ExchangeMarkets gets the full list of markets for the given exchange.
	ExchangeMarkets(ctx context.Context, exchange int) (ExchangeMarketsResponse, error)

	// FiatMap gets the full fiat asset map from CoinMarketCap.
	FiatMap(ctx context.Context) (FiatResponse, error)

	// Quote gets the quote for the provided ID from CoinMarketCap.
	Quote(ctx context.Context, id int64) (QuoteResponse, error)

	// Quotes gets the quotes for all provided IDs from CoinMarketCap.
	Quotes(ctx context.Context, ids []int64) (QuoteResponse, error)

	// Info gets all info for the given ID from CoinMarketCap.
	Info(ctx context.Context, ids []int64) (InfoResponse, error)
}

type httpClient struct {
	client *http.Client
	apiKey string
}

func NewHTTPClient(apiKey string) Client {
	return &httpClient{
		client: http.NewClient(),
		apiKey: apiKey,
	}
}

// CryptoIDMap gets the cryptocurrency ID Map from CoinMarketCap using the HTTP client.
// It handles pagination internally and returns all available cryptocurrency mappings.
func (h *httpClient) CryptoIDMap(ctx context.Context) (CryptoIDMapResponse, error) {
	var response CryptoIDMapResponse
	var allData []CryptoIDMapData

	start := 1
	limit := 10000 // Current page size from CMC

	for {
		opts := []http.GetOptions{
			http.WithHeader("X-CMC_PRO_API_KEY", h.apiKey),
			http.WithJSONAccept(),
			http.WithQueryParam("start", fmt.Sprintf("%d", start)),
		}

		resp, err := h.client.GetWithContext(ctx, EndpointCryptoMap, opts...)
		if err != nil {
			return response, err
		}

		var pageResponse CryptoIDMapResponse
		if err := json.NewDecoder(resp.Body).Decode(&pageResponse); err != nil {
			resp.Body.Close()
			return response, err
		}
		resp.Body.Close()

		allData = append(allData, pageResponse.Data...)

		// If we got less than the limit, we've reached the end
		if len(pageResponse.Data) < limit {
			break
		}

		start += limit
	}

	response.Data = allData
	return response, nil
}

// ExchangeIDMap gets the exchange ID Map from CoinMarketCap using the HTTP client.
func (h *httpClient) ExchangeIDMap(ctx context.Context) (ExchangeIDMapResponse, error) {
	var response ExchangeIDMapResponse

	opts := []http.GetOptions{
		http.WithHeader("X-CMC_PRO_API_KEY", h.apiKey),
		http.WithJSONAccept(),
	}

	resp, err := h.client.GetWithContext(ctx, EndpointExchangeMap, opts...)
	if err != nil {
		return response, err
	}
	defer resp.Body.Close()

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return response, err
	}

	return response, nil
}

// ExchangeAssets gets the full list of assets for the given exchange.
func (h *httpClient) ExchangeAssets(ctx context.Context, exchange int) (ExchangeAssetsResponse, error) {
	var response ExchangeAssetsResponse

	opts := []http.GetOptions{
		http.WithHeader("X-CMC_PRO_API_KEY", h.apiKey),
		http.WithJSONAccept(),
	}

	resp, err := h.client.GetWithContext(ctx, fmt.Sprintf(EndpointExchangeAssets, exchange), opts...)
	if err != nil {
		return response, err
	}
	defer resp.Body.Close()

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return response, err
	}

	return response, nil
}

// ExchangeMarkets gets the full list of markets for the given exchange.
func (h *httpClient) ExchangeMarkets(ctx context.Context, exchange int) (ExchangeMarketsResponse, error) {
	var response ExchangeMarketsResponse

	opts := []http.GetOptions{
		http.WithHeader("X-CMC_PRO_API_KEY", h.apiKey),
		http.WithJSONAccept(),
		http.WithQueryParam("category", "spot"),
	}

	resp, err := h.client.GetWithContext(ctx, fmt.Sprintf(EndpointExchangeMarkets, exchange), opts...)
	if err != nil {
		return response, err
	}
	defer resp.Body.Close()

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return response, err
	}

	return response, nil
}

// FiatMap gets the full fiat asset map from CoinMarketCap.
func (h *httpClient) FiatMap(ctx context.Context) (FiatResponse, error) {
	var response FiatResponse

	opts := []http.GetOptions{
		http.WithHeader("X-CMC_PRO_API_KEY", h.apiKey),
		http.WithJSONAccept(),
	}

	resp, err := h.client.GetWithContext(ctx, EndpointFiatMap, opts...)
	if err != nil {
		return response, err
	}
	defer resp.Body.Close()

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return response, err
	}

	return response, nil
}

func (h *httpClient) Quotes(ctx context.Context, ids []int64) (QuoteResponse, error) {
	var response QuoteResponse

	opts := []http.GetOptions{
		http.WithHeader("X-CMC_PRO_API_KEY", h.apiKey),
		http.WithJSONAccept(),
	}

	strIDs := make([]string, 0, len(ids))
	for _, id := range ids {
		strIDs = append(strIDs, strconv.FormatInt(id, 10))
	}

	resp, err := h.client.GetWithContext(ctx, fmt.Sprintf(EndpointQuote, strings.Join(strIDs, ",")), opts...)
	if err != nil {
		return response, err
	}
	defer resp.Body.Close()

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return response, err
	}

	return response, nil
}

// Quote gets the quotes for the provided IDs from CoinMarketCap.
func (h *httpClient) Quote(ctx context.Context, id int64) (QuoteResponse, error) {
	var response QuoteResponse

	opts := []http.GetOptions{
		http.WithHeader("X-CMC_PRO_API_KEY", h.apiKey),
		http.WithJSONAccept(),
	}

	resp, err := h.client.GetWithContext(ctx, fmt.Sprintf(EndpointQuote, fmt.Sprintf("%d", id)), opts...)
	if err != nil {
		return response, err
	}
	defer resp.Body.Close()

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return response, err
	}

	return response, nil
}

// Info gets the info for the provided IDs from CoinMarketCap.
func (h *httpClient) Info(ctx context.Context, ids []int64) (InfoResponse, error) {
	var response InfoResponse

	opts := []http.GetOptions{
		http.WithHeader("X-CMC_PRO_API_KEY", h.apiKey),
		http.WithJSONAccept(),
	}

	// Convert each integer to a string
	strSlice := make([]string, len(ids))
	for i, num := range ids {
		strSlice[i] = fmt.Sprintf("%d", num)
	}

	// Join the string slice with a comma separator
	idsString := strings.Join(strSlice, ",")

	resp, err := h.client.GetWithContext(ctx, fmt.Sprintf(EndpointInfo, idsString), opts...)
	if err != nil {
		return response, err
	}
	defer resp.Body.Close()

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return response, err
	}

	return response, nil
}
