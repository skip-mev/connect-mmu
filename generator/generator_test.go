package generator_test

import (
	"context"
	"encoding/json"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest"

	"github.com/skip-mev/connect-mmu/config"
	"github.com/skip-mev/connect-mmu/generator"
	"github.com/skip-mev/connect-mmu/lib/file"
	"github.com/skip-mev/connect-mmu/store/provider"
)

const (
	localIndexedMarketsFile     = "../local/fixtures/generator/indexed-markets.json"
	dydxTestnetGenerationConfig = "../local/fixtures/generator/gen_dydx.json"
)

const numIters = 20

func TestGenerationDeterminisimFromIndexedFile(t *testing.T) {
	logger := zaptest.NewLogger(t, zaptest.Level(zap.InfoLevel))

	providerStore, err := provider.NewMemoryStoreFromFile(localIndexedMarketsFile)
	require.NoError(t, err)

	gen := generator.New(logger, providerStore)

	bz, err := os.ReadFile(dydxTestnetGenerationConfig)
	require.NoError(t, err)

	var generationConfig config.GenerateConfig
	require.NoError(t, json.Unmarshal(bz, &generationConfig))

	require.NoError(t, generationConfig.Validate())

	// generate twice and check determinism
	mm1, removals1, err := gen.GenerateMarketMap(context.Background(), generationConfig)
	require.NoError(t, err)

	for range numIters {
		mm2, removals2, err := gen.GenerateMarketMap(context.Background(), generationConfig)
		require.NoError(t, err)

		if !mm1.Equal(mm2) {
			require.NoError(t, file.WriteJSONToFile("mm1.json", mm1))
			require.NoError(t, file.WriteJSONToFile("removals1.json", removals1))
			require.NoError(t, file.WriteJSONToFile("mm2.json", mm2))
			require.NoError(t, file.WriteJSONToFile("removals2.json", removals2))
		}

		require.Equal(t, len(mm1.Markets), len(mm2.Markets))
		require.True(t, mm1.Equal(mm2))
	}
}
