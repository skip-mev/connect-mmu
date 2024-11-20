package simulate

import (
	"context"
	"fmt"

	cmttypes "github.com/cometbft/cometbft/types"
	"github.com/cosmos/cosmos-sdk/client"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/bech32"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/mitchellh/mapstructure"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/skip-mev/connect-mmu/config"
	"github.com/skip-mev/connect-mmu/signing"
)

const (
	TypeName = "simulate_agent"
)

var _ signing.SigningAgent = &SigningAgent{}

// SigningAgent is a SigningAgent that can be used for simulation.
type SigningAgent struct {
	address    string
	authClient authtypes.QueryClient
	// txConfig is the SDK tx config used for transaction construction
	sdkTxConfig client.TxConfig
}

var _ signing.Factory = NewSigningAgent

func NewSigningAgent(config any, chainCfg config.ChainConfig) (signing.SigningAgent, error) {
	var cfg SigningAgentConfig
	decoderCfg := mapstructure.DecoderConfig{
		Result:  &cfg,
		TagName: "json",
	}
	decoder, err := mapstructure.NewDecoder(&decoderCfg)
	if err != nil {
		return nil, fmt.Errorf("error creating simulate agent config decoder: %w", err)
	}
	err = decoder.Decode(config)
	if err != nil {
		return nil, fmt.Errorf("error decoding simulate agent config: %w", err)
	}
	if err = cfg.Validate(); err != nil {
		return nil, fmt.Errorf("error validating simulate agent config: %w", err)
	}
	return NewSimulateSigningAgent(cfg.Address, chainCfg.GRPCAddress)
}

func NewSimulateSigningAgent(address, chainGrpcAddress string) (*SigningAgent, error) {
	// create a grpc client / comet rpc client from the configured rpcs for the chain
	chainConn, err := grpc.NewClient(chainGrpcAddress, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("could not create chain connection: %w", err)
	}

	prefix, _, err := bech32.DecodeAndConvert(address)
	if err != nil {
		return nil, fmt.Errorf("failed to decode address: %w", err)
	}

	// create tx config
	cdc, err := signing.Codec(prefix)
	if err != nil {
		return nil, fmt.Errorf("failed to create interface registry: %w", err)
	}

	return &SigningAgent{
		address:     address,
		authClient:  authtypes.NewQueryClient(chainConn),
		sdkTxConfig: signing.TxConfig(cdc),
	}, nil
}

// Sign returns the encoded bytes of the simulation Tx.
//
// NOTE: this tx is not signed but can be used for inspecting the encoded bytes.
func (s *SigningAgent) Sign(_ context.Context, txb client.TxBuilder) (cmttypes.Tx, error) {
	return s.sdkTxConfig.TxEncoder()(txb.GetTx())
}

func (s *SigningAgent) GetSigningAccount(ctx context.Context) (sdk.AccountI, error) {
	acc, err := signing.GetAccountAny(ctx, s.authClient, s.address)
	if err != nil {
		return nil, fmt.Errorf("failed to get account any: %w", err)
	}

	return acc, nil
}
