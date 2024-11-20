package util_test

import (
	"testing"

	"github.com/cosmos/gogoproto/proto"
	connecttypes "github.com/skip-mev/connect/v2/pkg/types"
	"github.com/skip-mev/connect/v2/x/marketmap/types"
	"github.com/stretchr/testify/require"

	"github.com/skip-mev/connect-mmu/util"
)

var (
	cpBTCUSD = connecttypes.CurrencyPair{
		Base:  "BTC",
		Quote: "USD",
	}

	tickerBTCUSD = types.Ticker{
		CurrencyPair:     cpBTCUSD,
		Decimals:         10,
		MinProviderCount: 3,
		Enabled:          false,
		Metadata_JSON:    "",
	}

	marketBTCUSD = types.Market{
		Ticker: tickerBTCUSD,
		ProviderConfigs: []types.ProviderConfig{
			{
				Name:            "binance",
				OffChainTicker:  "btc-usd",
				NormalizeByPair: nil,
				Invert:          false,
				Metadata_JSON:   "",
			},
		},
	}
)

func TestGenerateUpdate(t *testing.T) {
	tests := []struct {
		name      string
		mm        types.MarketMap
		authority string
		want      proto.Message
	}{
		{
			name:      "empty",
			mm:        types.MarketMap{},
			authority: "test",
			want:      &types.MsgUpdateMarkets{Authority: "test", UpdateMarkets: []types.Market{}},
		},
		{
			name: "one market",
			mm: types.MarketMap{
				Markets: map[string]types.Market{
					cpBTCUSD.String(): marketBTCUSD,
				},
			},
			authority: "test",
			want: &types.MsgUpdateMarkets{Authority: "test", UpdateMarkets: []types.Market{
				marketBTCUSD,
			}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.want, util.GenerateUpdate(tt.mm, tt.authority))
		})
	}
}

func TestGenerateUpsert(t *testing.T) {
	tests := []struct {
		name      string
		mm        types.MarketMap
		authority string
		want      proto.Message
	}{
		{
			name:      "empty",
			mm:        types.MarketMap{},
			authority: "test",
			want:      &types.MsgUpsertMarkets{Authority: "test", Markets: []types.Market{}},
		},
		{
			name: "one market",
			mm: types.MarketMap{
				Markets: map[string]types.Market{
					cpBTCUSD.String(): marketBTCUSD,
				},
			},
			authority: "test",
			want: &types.MsgUpsertMarkets{Authority: "test", Markets: []types.Market{
				marketBTCUSD,
			}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Run(tt.name, func(t *testing.T) {
				require.Equal(t, tt.want, util.GenerateUpsert(tt.mm, tt.authority))
			})
		})
	}
}

func TestGenerateCreate(t *testing.T) {
	tests := []struct {
		name      string
		mm        types.MarketMap
		authority string
		want      proto.Message
	}{
		{
			name:      "empty",
			mm:        types.MarketMap{},
			authority: "test",
			want:      &types.MsgCreateMarkets{Authority: "test", CreateMarkets: []types.Market{}},
		},
		{
			name: "one market",
			mm: types.MarketMap{
				Markets: map[string]types.Market{
					cpBTCUSD.String(): marketBTCUSD,
				},
			},
			authority: "test",
			want: &types.MsgCreateMarkets{Authority: "test", CreateMarkets: []types.Market{
				marketBTCUSD,
			}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Run(tt.name, func(t *testing.T) {
				require.Equal(t, tt.want, util.GenerateCreate(tt.mm, tt.authority))
			})
		})
	}
}
