package local

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/skip-mev/connect-mmu/config"
)

func TestNewSigningAgent(t *testing.T) {
	tests := []struct {
		name        string
		config      SigningAgentConfig
		chainConfig config.ChainConfig
		wantErr     bool
	}{
		{
			name:    "empty configs will fail",
			wantErr: true,
		},
		{
			name: "empty config will fail",
			chainConfig: config.ChainConfig{
				RPCAddress:  "test",
				GRPCAddress: "test",
				RESTAddress: "test",
				ChainID:     "test",
				DYDX:        false,
				Version:     "test",
				Prefix:      "test",
			},
			wantErr: true,
		},
		{
			name: "empty chain config will fail",
			config: SigningAgentConfig{
				PrivateKeyFile: "file",
			},
			wantErr: true,
		},
		{
			name: "invalid non-existent file",
			config: SigningAgentConfig{
				PrivateKeyFile: "file",
			},
			chainConfig: config.ChainConfig{
				RPCAddress:  "test",
				GRPCAddress: "test",
				RESTAddress: "test",
				ChainID:     "test",
				DYDX:        false,
				Version:     "test",
				Prefix:      "test",
			},
			wantErr: true,
		},
		{
			name: "invalid privkey in file",
			config: SigningAgentConfig{
				PrivateKeyFile: "../../local/fixtures/testdata/invalid.privkey",
			},
			chainConfig: config.ChainConfig{
				RPCAddress:  "test",
				GRPCAddress: "test",
				RESTAddress: "test",
				ChainID:     "test",
				DYDX:        false,
				Version:     "test",
				Prefix:      "test",
			},
			wantErr: true,
		},
		{
			name: "valid",
			config: SigningAgentConfig{
				PrivateKeyFile: "../../local/fixtures/testdata/valid.privkey",
			},
			chainConfig: config.ChainConfig{
				RPCAddress:  "test",
				GRPCAddress: "test",
				RESTAddress: "test",
				ChainID:     "test",
				DYDX:        false,
				Version:     "test",
				Prefix:      "test",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewSigningAgent(tt.config, tt.chainConfig)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
		})
	}
}
