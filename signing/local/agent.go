package local

import (
	"context"
	"fmt"
	"os"

	txv1beta1 "cosmossdk.io/api/cosmos/tx/v1beta1"
	"github.com/cometbft/cometbft/types"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	authsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/mitchellh/mapstructure"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/proto"

	"github.com/skip-mev/connect-mmu/config"
	mmusigning "github.com/skip-mev/connect-mmu/signing"
)

const (
	TypeName = "local_agent"
	keyName  = "local"
)

var (
	_ mmusigning.SigningAgent = &SigningAgent{}
	_ mmusigning.Factory      = NewSigningAgent
)

// SigningAgent is a SigningAgent that can be used from a local private key file.
type SigningAgent struct {
	kr         keyring.Keyring
	authClient authtypes.QueryClient
	// txConfig is the SDK tx config used for transaction construction
	sdkTxConfig client.TxConfig
	chainConfig config.ChainConfig
}

func NewSigningAgent(config any, chainConfig config.ChainConfig) (mmusigning.SigningAgent, error) {
	var cfg SigningAgentConfig
	decoderCfg := mapstructure.DecoderConfig{
		Result:  &cfg,
		TagName: "json",
	}

	decoder, err := mapstructure.NewDecoder(&decoderCfg)
	if err != nil {
		return nil, fmt.Errorf("error creating local agent config decoder %v: %w", config, err)
	}
	err = decoder.Decode(config)
	if err != nil {
		return nil, fmt.Errorf("error decoding local agent config %v: %w", config, err)
	}
	if err = cfg.Validate(); err != nil {
		return nil, fmt.Errorf("error local agent config %v: %w", config, err)
	}
	return NewLocalSigningAgent(cfg.PrivateKeyFile, chainConfig)
}

func NewLocalSigningAgent(privKeyFile string, chainConfig config.ChainConfig) (*SigningAgent, error) {
	// create a grpc client / comet rpc client from the configured rpcs for the chain
	chainConn, err := grpc.NewClient(chainConfig.GRPCAddress, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("could not create chain connection: %w", err)
	}

	// create tx config
	cdc, err := mmusigning.Codec(chainConfig.Prefix)
	if err != nil {
		return nil, fmt.Errorf("failed to create interface registry: %w", err)
	}

	pkBz, err := os.ReadFile(privKeyFile)
	if err != nil {
		return nil, fmt.Errorf("error reading private key file: %w", err)
	}

	kr := keyring.NewInMemory(cdc)
	err = kr.ImportPrivKeyHex(keyName, string(pkBz), "secp256k1")
	if err != nil {
		return nil, fmt.Errorf("error importing private key: %w", err)
	}

	return &SigningAgent{
		kr:          kr,
		authClient:  authtypes.NewQueryClient(chainConn),
		sdkTxConfig: mmusigning.TxConfig(cdc),
		chainConfig: chainConfig,
	}, nil
}

func (s *SigningAgent) Sign(ctx context.Context, txb client.TxBuilder) (types.Tx, error) {
	acc, err := s.GetSigningAccount(ctx)
	if err != nil {
		return nil, fmt.Errorf("error getting signing account: %w", err)
	}

	// set the account number + sequence
	if err := txb.SetSignatures(signing.SignatureV2{
		PubKey: acc.GetPubKey(),
		Data: &signing.SingleSignatureData{
			SignMode:  signing.SignMode_SIGN_MODE_DIRECT,
			Signature: []byte{},
		},
		Sequence: acc.GetSequence(),
	}); err != nil {
		return nil, err
	}

	signingTx, ok := txb.GetTx().(authsigning.V2AdaptableTx)
	if !ok {
		return nil, fmt.Errorf("unexpected tx type %T", txb.GetTx())
	}
	txData := signingTx.GetSigningTxData()

	// sign the tx
	protoOpts := proto.MarshalOptions{
		Deterministic: true, // deterministic encoding for signature verification
	}

	signDocBz, err := protoOpts.Marshal(&txv1beta1.SignDoc{
		BodyBytes:     txData.BodyBytes,
		AuthInfoBytes: txData.AuthInfoBytes,
		ChainId:       s.chainConfig.ChainID,
		AccountNumber: acc.GetAccountNumber(),
	})
	if err != nil {
		return nil, err
	}

	signature, _, err := s.kr.Sign(keyName, signDocBz, signing.SignMode_SIGN_MODE_DIRECT)
	if err != nil {
		return nil, fmt.Errorf("error signing transaction: %w", err)
	}

	if err := txb.SetSignatures(signing.SignatureV2{
		PubKey: acc.GetPubKey(),
		Data: &signing.SingleSignatureData{
			SignMode:  signing.SignMode_SIGN_MODE_DIRECT,
			Signature: signature,
		},
		Sequence: acc.GetSequence(),
	}); err != nil {
		return nil, err
	}

	return s.sdkTxConfig.TxEncoder()(txb.GetTx())
}

func (s *SigningAgent) GetSigningAccount(ctx context.Context) (sdk.AccountI, error) {
	record, err := s.kr.Key(keyName)
	if err != nil {
		return nil, fmt.Errorf("error getting key record: %w", err)
	}

	pubKey, err := record.GetPubKey()
	if err != nil {
		return nil, fmt.Errorf("error getting key record pubkey: %w", err)
	}

	return s.getAccountFromPubKey(ctx, pubKey)
}

func (s *SigningAgent) getAccountFromPubKey(ctx context.Context, pubKey cryptotypes.PubKey) (sdk.AccountI, error) {
	// get the account associated with the pubkey
	address, err := mmusigning.PubKeyBech32(s.chainConfig.Prefix, pubKey)
	if err != nil {
		return nil, fmt.Errorf("failed to bech32ify address: %w", err)
	}

	acc, err := mmusigning.GetAccountAny(ctx, s.authClient, address)
	if err != nil {
		return nil, fmt.Errorf("failed to get account: %w", err)
	}

	// update the account's pk
	err = acc.SetPubKey(pubKey)
	if err != nil {
		return nil, err
	}

	return acc, nil
}
