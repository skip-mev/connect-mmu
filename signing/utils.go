package signing

import (
	"context"
	"fmt"

	txsigning "cosmossdk.io/x/tx/signing"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/bech32"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	authcodec "github.com/cosmos/cosmos-sdk/x/auth/codec"
	authtx "github.com/cosmos/cosmos-sdk/x/auth/tx"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	gogoproto "github.com/cosmos/gogoproto/proto"
	mmtypes "github.com/skip-mev/connect/v2/x/marketmap/types"
	slinkymmtypes "github.com/skip-mev/slinky/x/marketmap/types"
)

// Codec returns a codec for signing with the given address prefix.
func Codec(prefix string) (*codec.ProtoCodec, error) {
	ac := authcodec.NewBech32Codec(prefix)

	ir, err := codectypes.NewInterfaceRegistryWithOptions(codectypes.InterfaceRegistryOptions{
		ProtoFiles: gogoproto.HybridResolver,
		SigningOptions: txsigning.Options{
			AddressCodec:          ac,
			ValidatorAddressCodec: ac,
		},
	})
	if err != nil {
		return nil, err
	}

	authtypes.RegisterInterfaces(ir)
	cryptocodec.RegisterInterfaces(ir)
	mmtypes.RegisterInterfaces(ir)
	slinkymmtypes.RegisterInterfaces(ir)

	return codec.NewProtoCodec(ir), nil
}

// TxConfig returns a tx config for direct signing.
func TxConfig(codec codec.Codec) client.TxConfig {
	return authtx.NewTxConfig(
		codec,
		[]signing.SignMode{signing.SignMode_SIGN_MODE_DIRECT},
	)
}

// PubKeyBech32 returns a bech32 address string given a pubkey and an address prefix.
func PubKeyBech32(addressPrefix string, pk cryptotypes.PubKey) (string, error) {
	// get the account associated with the pubkey
	return sdk.Bech32ifyAddressBytes(addressPrefix, pk.Address())
}

func GetAccountAny(ctx context.Context, authClient AuthClient, bech32Addr string) (sdk.AccountI, error) {
	// get the account associated with the address
	accountAny, err := authClient.Account(ctx, &authtypes.QueryAccountRequest{Address: bech32Addr})
	if err != nil {
		return nil, fmt.Errorf("failed to get account: %w", err)
	}

	if accountAny == nil {
		err := fmt.Errorf("nil response")
		return nil, fmt.Errorf("failed to get account: %w", err)
	}

	// get prefix from bech32 string
	prefix, _, err := bech32.DecodeAndConvert(bech32Addr)
	if err != nil {
		return nil, fmt.Errorf("failed to decode address: %w", err)
	}

	// create tx config
	cdc, err := Codec(prefix)
	if err != nil {
		return nil, fmt.Errorf("failed to create interface registry: %w", err)
	}

	// unmarshal the account
	var acc sdk.AccountI
	if err := cdc.UnpackAny(accountAny.Account, &acc); err != nil {
		return nil, fmt.Errorf("failed to get account: %w", err)
	}

	return acc, nil
}
