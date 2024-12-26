package gecko

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/skip-mev/connect/v2/providers/apis/defi/uniswapv3"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/skip-mev/connect-mmu/config"
	"github.com/skip-mev/connect-mmu/store/provider"
)

func TestGetProviderMarkets(t *testing.T) {
	// stand up test http server. this will return the data from the example json.
	_, filename, _, _ := runtime.Caller(0)
	currentDir := filepath.Dir(filename)
	poolsResponse, err := os.ReadFile(filepath.Join(currentDir, "testdata/pools_response_example.json"))
	require.NoError(t, err)
	tokensResponse, err := os.ReadFile(filepath.Join(currentDir, "testdata/tokens_multi_response_example.json"))
	require.NoError(t, err)
	var pools PoolsResponse
	err = json.Unmarshal(poolsResponse, &pools)
	require.NoError(t, err)
	var tokens TokensMultiResponse
	err = json.Unmarshal(tokensResponse, &tokens)
	require.NoError(t, err)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var file []byte
		switch {
		case r.URL.Path == "/networks/eth/dexes/uniswap_v3/pools":
			file = poolsResponse
		case strings.HasPrefix(r.URL.Path, "/networks/eth/tokens/multi/"):
			file = tokensResponse
		default:
			t.Logf("Unexpected request: %s", r.URL.Path)
			http.Error(w, "Not found", http.StatusNotFound)
			return
		}

		w.Write(file)
	}))
	defer server.Close()
	logger, err := zap.NewDevelopment()
	require.NoError(t, err)

	// create gecko client but with test server URL
	client := newClient(logger, server.URL)

	ingester := &Ingester{
		logger: logger,
		client: client,
		pairs:  []config.GeckoNetworkDexPair{{Network: "eth", Dex: "uniswap_v3"}},
	}

	// get the provider markets.
	markets, err := ingester.GetProviderMarkets(context.Background())
	require.NoError(t, err)

	// WETH/USDC market. USDC address is first in sorted order, so we should invert.
	metaData1 := uniswapv3.PoolConfig{
		Address:       pools.Data[0].VenueAddress(),
		BaseDecimals:  18,
		QuoteDecimals: 6,
		Invert:        true,
	}
	metaData1Bz, err := json.Marshal(metaData1)
	require.NoError(t, err)

	// PEPPER/WETH market. PEPPER address is first in sorted order, so we don't need to invert.
	metaData2 := uniswapv3.PoolConfig{
		Address:       pools.Data[1].VenueAddress(),
		BaseDecimals:  18,
		QuoteDecimals: 18,
		Invert:        false,
	}
	metaData2Bz, err := json.Marshal(metaData2)
	require.NoError(t, err)

	targetBase0, err := pools.Data[0].Base()
	require.NoError(t, err)
	targetQuote0, err := pools.Data[0].Quote()
	require.NoError(t, err)
	offChainTicker0, err := pools.Data[0].OffChainTicker()
	require.NoError(t, err)
	liquidity0, err := pools.Data[0].Liquidity()
	require.NoError(t, err)
	usdVolume0, err := pools.Data[0].UsdVolume()
	require.NoError(t, err)

	targetBase1, err := pools.Data[1].Base()
	require.NoError(t, err)
	targetQuote1, err := pools.Data[1].Quote()
	require.NoError(t, err)
	offChainTicker1, err := pools.Data[1].OffChainTicker()
	require.NoError(t, err)
	liquidity1, err := pools.Data[1].Liquidity()
	require.NoError(t, err)
	usdVolume1, err := pools.Data[1].UsdVolume()
	require.NoError(t, err)

	// should end up with these markets.
	baseMarkets := []provider.CreateProviderMarket{
		{
			Create: provider.CreateProviderMarketParams{
				TargetBase:       targetBase0,
				TargetQuote:      targetQuote0,
				OffChainTicker:   offChainTicker0,
				ProviderName:     geckoDexToConnectDex(pools.Data[0].Venue()),
				QuoteVolume:      281462633.1550315,
				UsdVolume:        usdVolume0,
				MetadataJSON:     metaData1Bz,
				ReferencePrice:   3409.83,
				NegativeDepthTwo: liquidity0 / 2,
				PositiveDepthTwo: liquidity0 / 2,
			},
			BaseAddress:  pools.Data[0].BaseAddress(),
			QuoteAddress: pools.Data[0].QuoteAddress(),
		},
		{
			Create: provider.CreateProviderMarketParams{
				TargetBase:       targetBase1,
				TargetQuote:      targetQuote1,
				OffChainTicker:   offChainTicker1,
				ProviderName:     geckoDexToConnectDex(pools.Data[1].Venue()),
				QuoteVolume:      3639.743321519964,
				UsdVolume:        usdVolume1,
				MetadataJSON:     metaData2Bz,
				ReferencePrice:   0.000000001585379138,
				NegativeDepthTwo: liquidity1 / 2,
				PositiveDepthTwo: liquidity1 / 2,
			},
			BaseAddress:  pools.Data[1].BaseAddress(),
			QuoteAddress: pools.Data[1].QuoteAddress(),
		},
	}
	// calling pools actually uses pagination. so we call pools maxPages (10) times.
	// this means we should have those 2 markets above returned maxPages (10) times (20 total).
	expectedMarkets := make([]provider.CreateProviderMarket, 0, maxPages*2)
	for i := 0; i < 10; i++ {
		expectedMarkets = append(expectedMarkets, baseMarkets...)
	}
	require.Equal(t, len(expectedMarkets), len(markets))

	// Assert the content of the returned markets
	for i, expected := range expectedMarkets {
		require.Equal(t, expected, markets[i])
	}
}
