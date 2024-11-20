package config_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/skip-mev/connect-mmu/config"
)

func TestChainConfig_ValidateBasic(t *testing.T) {
	tests := []struct {
		name    string
		config  config.ChainConfig
		wantErr bool
	}{
		{
			name: "valid non-DYDX chain config",
			config: config.ChainConfig{
				RPCAddress:  "http://rpc.example.com",
				GRPCAddress: "http://grpc.example.com",
				RESTAddress: "http://rest.example.com",
				ChainID:     "foo",
				DYDX:        false,
				Version:     config.VersionSlinky,
				Prefix:      "bar",
			},
			wantErr: false,
		},
		{
			name: "valid DYDX chain config",
			config: config.ChainConfig{
				RPCAddress:  "http://rpc.example.com",
				GRPCAddress: "http://grpc.example.com",
				RESTAddress: "http://rest.example.com",
				ChainID:     "foo",
				DYDX:        true,
				Version:     config.VersionSlinky,
				Prefix:      "bar",
			},
			wantErr: false,
		},
		{
			name: "invalid prefix",
			config: config.ChainConfig{
				RPCAddress:  "http://rpc.example.com",
				GRPCAddress: "http://grpc.example.com",
				RESTAddress: "http://rest.example.com",
				ChainID:     "foo",
				DYDX:        false,
				Version:     config.VersionSlinky,
			},
			wantErr: true,
		},
		{
			name: "invalid chain version",
			config: config.ChainConfig{
				RPCAddress:  "http://rpc.example.com",
				GRPCAddress: "http://grpc.example.com",
				RESTAddress: "http://rest.example.com",
				ChainID:     "foo",
				DYDX:        true,
				Version:     "invalid",
				Prefix:      "bar",
			},
			wantErr: true,
		},
		{
			name: "invalid DYDX chain config - missing GRPC address",
			config: config.ChainConfig{
				RPCAddress:  "http://rpc.example.com",
				GRPCAddress: "",
				RESTAddress: "http://rest.example.com",
				ChainID:     "foo",
				DYDX:        true,
				Version:     config.VersionSlinky,
				Prefix:      "bar",
			},
			wantErr: true,
		},
		{
			name: "invalid DYDX chain config - missing REST address",
			config: config.ChainConfig{
				RPCAddress:  "http://rpc.example.com",
				GRPCAddress: "http://grpc.example.com",
				RESTAddress: "",
				ChainID:     "foo",
				DYDX:        true,
				Version:     config.VersionSlinky,
				Prefix:      "bar",
			},
			wantErr: true,
		},
		{
			name: "invalid DYDX chain config - missing both GRPC and REST addresses",
			config: config.ChainConfig{
				RPCAddress:  "http://rpc.example.com",
				GRPCAddress: "",
				RESTAddress: "",
				ChainID:     "foo",
				DYDX:        true,
				Version:     config.VersionSlinky,
				Prefix:      "bar",
			},
			wantErr: true,
		},
		{
			name: "invalid missing chain id",
			config: config.ChainConfig{
				RPCAddress:  "http://rpc.example.com",
				GRPCAddress: "http://grpc.example.com",
				RESTAddress: "http://rest.example.com",
				DYDX:        false,
				Version:     config.VersionSlinky,
				Prefix:      "bar",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
