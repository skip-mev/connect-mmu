package generator_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	mmtypes "github.com/skip-mev/connect/v2/x/marketmap/types"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"

	"github.com/skip-mev/connect-mmu/config"
	"github.com/skip-mev/connect-mmu/dispatcher/transaction/generator"
	"github.com/skip-mev/connect-mmu/testutil/markets"
)

func TestConvertUpsertsToMessages(t *testing.T) {
	tests := []struct {
		name    string
		cfg     config.TransactionConfig
		upserts []mmtypes.Market
		want    []sdk.Msg
		wantErr bool
	}{
		{
			name: "empty upserts",
			cfg: config.TransactionConfig{
				MaxBytesPerTx: 2000,
			},
			want: make([]sdk.Msg, 0),
		},
		{
			name: "fail due to invalid tx size",
			cfg: config.TransactionConfig{
				MaxBytesPerTx: 0,
			},
			upserts: []mmtypes.Market{
				markets.UsdtUsd,
			},
			want:    make([]sdk.Msg, 0),
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := generator.ConvertUpsertsToMessages(zaptest.NewLogger(t), tt.cfg, config.VersionConnect, "", tt.upserts)
			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.Equal(t, tt.want, got)
		})
	}
}
