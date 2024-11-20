package gecko

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPoolsData(t *testing.T) {
	_, filename, _, _ := runtime.Caller(0)
	currentDir := filepath.Dir(filename)
	poolsResponseBz, err := os.ReadFile(filepath.Join(currentDir, "testdata/pools_response_example.json"))
	require.NoError(t, err)

	var pools PoolsResponse
	err = json.Unmarshal(poolsResponseBz, &pools)
	require.NoError(t, err)

	pool := pools.Data[0]

	ticker, err := pool.OffChainTicker()
	require.NoError(t, err)
	// see: testdata/pools_response_example.json
	expected := strings.ToUpper("WETH,uniswap_v3,0xc02aaa39b223fe8d0a0e5c4f27ead9083c756cc2/USDC,uniswap_v3," +
		"0xa0b86991c6218b36c1d19d4a2e9eb0ce3606eb48")
	require.Equal(t, ticker, expected)
}
