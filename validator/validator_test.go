package validator

import (
	"testing"

	mmtypes "github.com/skip-mev/connect/v2/x/marketmap/types"
	"github.com/stretchr/testify/require"

	"github.com/skip-mev/connect-mmu/validator/types"
)

func TestMissingReports(t *testing.T) {
	health := make(types.MarketHealth)
	health["BAR/FOO"] = types.ProviderCounts{
		"meow": &types.Counts{},
	}
	mm := mmtypes.MarketMap{Markets: make(map[string]mmtypes.Market)}
	mm.Markets["FOO/BAR"] = mmtypes.Market{
		ProviderConfigs: []mmtypes.ProviderConfig{{Name: "foobar"}},
	}

	v := New(mm)
	missing := v.MissingReports(health)
	_, ok := missing["FOO/BAR"]
	require.True(t, ok)
	_, ok = missing["BAR/FOO"]
	require.False(t, ok)
}
