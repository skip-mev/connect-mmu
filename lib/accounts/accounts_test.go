package accounts_test

import (
	"encoding/hex"
	"testing"

	"github.com/cosmos/cosmos-sdk/types/bech32"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	"github.com/stretchr/testify/require"

	"github.com/skip-mev/connect-mmu/lib/accounts"
)

func TestGetModuleAddress(t *testing.T) {
	tests := []struct {
		name         string
		bech32Prefix string
		moduleName   string
		want         string
		wantErr      bool
	}{
		{
			name:    "empty args returns error",
			wantErr: true,
		},
		{
			name:       "empty bech32 returns error",
			moduleName: govtypes.ModuleName,
			wantErr:    true,
		},
		{
			name:         "empty moduleName returns error",
			bech32Prefix: "cosmos",
			wantErr:      true,
		},
		{
			name:         "gov moduleName proper address",
			bech32Prefix: "dydx",
			moduleName:   govtypes.ModuleName,
			want:         "dydx10d07y265gmmuvt4z0w9aw880jnsr700jnmapky",
			wantErr:      false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := accounts.GetModuleAddress(tt.bech32Prefix, tt.moduleName)
			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestHexAddressToValoperAddress(t *testing.T) {
	_, bz, err := bech32.DecodeAndConvert("dydx10d07y265gmmuvt4z0w9aw880jnsr700jnmapky")
	require.NoError(t, err)

	validHex := hex.EncodeToString(bz)

	tests := []struct {
		name         string
		bech32Prefix string
		hexAddress   string
		want         string
		wantErr      bool
	}{
		{
			name:    "empty args returns error",
			wantErr: true,
		},
		{
			name:       "empty bech32 returns error",
			hexAddress: validHex,
			wantErr:    true,
		},
		{
			name:         "empty hexAddress returns error",
			bech32Prefix: "cosmos",
			wantErr:      true,
		},
		{
			name:         "invalid hex returns error",
			bech32Prefix: "dydx",
			hexAddress:   govtypes.ModuleName,
			wantErr:      true,
		},
		{
			name:         "valid hex returns consaddr",
			bech32Prefix: "dydx",
			hexAddress:   validHex,
			want:         "dydxvalcons10d07y265gmmuvt4z0w9aw880jnsr700jzkct35",
			wantErr:      false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := accounts.HexAddressToValoperAddress(tt.bech32Prefix, tt.hexAddress)
			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			require.Equal(t, tt.want, got)
		})
	}
}
