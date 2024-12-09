package file

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	connecttypes "github.com/skip-mev/connect/v2/pkg/types"
	"github.com/skip-mev/connect/v2/x/marketmap/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestReadJSONFromFile(t *testing.T) {
	market := types.Market{
		Ticker: types.Ticker{CurrencyPair: connecttypes.CurrencyPair{Base: "FOO", Quote: "BAR"}},
	}
	markets := []types.Market{market}

	bz, err := json.Marshal(markets)
	require.NoError(t, err)

	f, err := os.CreateTemp("", "testdata.json")
	require.NoError(t, err)

	err = os.WriteFile(f.Name(), bz, 0o600)
	require.NoError(t, err)

	m, err := ReadJSONFromFile[[]types.Market](f.Name())
	require.NoError(t, err)

	require.Equal(t, markets, m)
}

func TestWriteJSONToFile(t *testing.T) {
	dir := t.TempDir()
	x := types.Market{ProviderConfigs: []types.ProviderConfig{{Name: "foo"}}}
	filePath := filepath.Join(dir, "data.json")
	err := WriteJSONToFile(filePath, x)
	require.NoError(t, err)

	y, err := ReadJSONFromFile[types.Market](filePath)
	require.NoError(t, err)
	assert.Equal(t, x.ProviderConfigs, y.ProviderConfigs)
}
