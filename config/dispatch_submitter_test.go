package config_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/skip-mev/connect-mmu/config"
)

func TestSubmitterConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		cfg     config.SubmitterConfig
		wantErr bool
	}{
		{
			name:    "empty config is invalid",
			wantErr: true,
		},
		{
			name:    "default config is valid",
			cfg:     config.DefaultSubmitterConfig(),
			wantErr: false,
		},
		{
			name: "0 polling frequency is invalid",
			cfg: config.SubmitterConfig{
				PollingFrequency: 0,
				PollingDuration:  config.DefaultPollingDuration,
			},
			wantErr: true,
		},
		{
			name: "0 polling duration is invalid",
			cfg: config.SubmitterConfig{
				PollingFrequency: config.DefaultPollingFrequency,
				PollingDuration:  0,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.cfg.ValidateBasic()
			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
		})
	}
}
