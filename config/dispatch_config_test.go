package config_test

import (
	"testing"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/skip-mev/connect-mmu/config"
)

func TestConfig_Validate(t *testing.T) {
	dummySigningConfig := config.SigningConfig{
		Type:   "foo",
		Config: map[string]any{"foo": "bar"},
	}
	tests := []struct {
		name    string
		config  config.DispatchConfig
		wantErr bool
		dryRun  bool
	}{
		{
			name:    "invalid empty",
			config:  config.DispatchConfig{},
			wantErr: true,
		},
		{
			name: "invalid submitter config - fail",
			config: config.DispatchConfig{
				TxConfig: config.TransactionConfig{
					MaxBytesPerTx: 1,
					MaxGas:        1,
					GasAdjustment: 1.5,
					MinGasPrice:   sdk.NewDecCoin("stake", math.NewInt(100)),
				},
				SigningConfig:   dummySigningConfig,
				SubmitterConfig: config.SubmitterConfig{PollingFrequency: 0},
			},
			wantErr: true,
		},
		{
			name: "invalid no signing config - fail",
			config: config.DispatchConfig{
				TxConfig: config.TransactionConfig{
					MaxBytesPerTx: 1,
					MaxGas:        1,
					GasAdjustment: 1.5,
					MinGasPrice:   sdk.NewDecCoin("stake", math.NewInt(100)),
				},
				SubmitterConfig: config.SubmitterConfig{PollingFrequency: 0},
			},
			wantErr: true,
		},
		{
			name: "valid - no restrictions",
			config: config.DispatchConfig{
				TxConfig: config.TransactionConfig{
					MaxBytesPerTx: 1,
					MaxGas:        1,
					GasAdjustment: 1.5,
					MinGasPrice:   sdk.NewDecCoin("stake", math.NewInt(100)),
				},
				SigningConfig:   dummySigningConfig,
				SubmitterConfig: config.DefaultSubmitterConfig(),
			},
			wantErr: false,
		},
		{
			name: "valid - restrictions",
			config: config.DispatchConfig{
				TxConfig: config.TransactionConfig{
					MaxBytesPerTx: 1,
					MaxGas:        1,
					GasAdjustment: 1.5,
					MinGasPrice:   sdk.NewDecCoin("stake", math.NewInt(100)),
				},
				SigningConfig:   dummySigningConfig,
				SubmitterConfig: config.DefaultSubmitterConfig(),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
		})
	}
}
