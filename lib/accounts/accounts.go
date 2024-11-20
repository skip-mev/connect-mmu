package accounts

import (
	"encoding/hex"
	"fmt"

	"github.com/cosmos/cosmos-sdk/types/bech32"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
)

// GetModuleAddress gets the given module address for a chain.
func GetModuleAddress(bech32Prefix, moduleName string) (string, error) {
	if bech32Prefix == "" || moduleName == "" {
		return "", fmt.Errorf("must provide bech32 prefix and module name, got bech32 prefix %s, module name %s", bech32Prefix, moduleName)
	}

	// Create the module address
	moduleAddr := authtypes.NewModuleAddress(moduleName)

	// Convert the module address to bytes
	addrBytes := moduleAddr.Bytes()

	// Convert to bech32 address
	bech32Addr, err := bech32.ConvertAndEncode(bech32Prefix, addrBytes)
	if err != nil {
		return "", err
	}

	return bech32Addr, nil
}

// HexAddressToValoperAddress converts the given hex address to a bech32 validator consensus address.
func HexAddressToValoperAddress(bech32Prefix, hexAddress string) (string, error) {
	if bech32Prefix == "" || hexAddress == "" {
		return "", fmt.Errorf("must provide bech32 prefix and hex address, got bech32 prefix %s, hex address %s", bech32Prefix, hexAddress)
	}

	addressBytes, err := hex.DecodeString(hexAddress)
	if err != nil {
		return "", err
	}
	bech32Address, err := bech32.ConvertAndEncode(bech32Prefix+"valcons", addressBytes)
	if err != nil {
		return "", err
	}
	return bech32Address, nil
}
