package config_test

import (
	"testing"

	"github.com/skip-mev/connect-mmu/config"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
)

func TestTransactionConfig_ValidateBasic(t *testing.T) {
	tcs := []struct {
		name string
		cfg  config.TransactionConfig
		err  error
	}{
		{
			name: "invalid max bytes per tx",
			cfg: config.TransactionConfig{
				MaxBytesPerTx: 0,
				MaxGas:        100,
				GasAdjustment: 1.5,
				MinGasPrice:   sdk.NewDecCoin("stake", math.NewInt(100)),
			},
			err: config.ErrInvalidMaxBytesPerTx,
		},
		{
			name: "invalid max gas",
			cfg: config.TransactionConfig{
				MaxBytesPerTx: 1,
				MaxGas:        0,
				GasAdjustment: 1.5,
				MinGasPrice:   sdk.NewDecCoin("stake", math.NewInt(100)),
			},
			err: config.ErrInvalidMaxGas,
		},
		{
			name: "invalid tx fee",
			cfg: config.TransactionConfig{
				MaxBytesPerTx: 1,
				MaxGas:        1,
				GasAdjustment: 1.5,
				MinGasPrice: sdk.DecCoin{
					Denom: "",
				},
			},
			err: config.ErrInvalidTxFee,
		},
		{
			name: "invalid gas adjustment",
			cfg: config.TransactionConfig{
				MaxBytesPerTx: 1,
				MaxGas:        1,
				GasAdjustment: 0.5,
				MinGasPrice:   sdk.NewDecCoin("stake", math.NewInt(100)),
			},
			err: config.ErrInvalidGasAdjustment,
		},
		{
			name: "valid",
			cfg: config.TransactionConfig{
				MaxBytesPerTx: 1,
				MaxGas:        1,
				GasAdjustment: 1.5,
				MinGasPrice:   sdk.NewDecCoin("stake", math.NewInt(100)),
			},
			err: nil,
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			require.Equal(t, tc.err, tc.cfg.ValidateBasic())
		})
	}
}
