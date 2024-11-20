package generator

import (
	"context"
	"errors"
	"fmt"

	cmttypes "github.com/cometbft/cometbft/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	mmtypes "github.com/skip-mev/connect/v2/x/marketmap/types"
	slinkymmtypes "github.com/skip-mev/slinky/x/marketmap/types"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/skip-mev/connect-mmu/config"
	"github.com/skip-mev/connect-mmu/signing"
)

var _ TransactionGenerator = &SigningTransactionGenerator{}

// SigningTransactionGenerator is a transaction provider that creates + signs market-update txs
// with the signer key.
type SigningTransactionGenerator struct {
	coreGenerator
}

// NewSigningTransactionGeneratorFromConfig creates a new SigningTransactionGenerator from a config.
func NewSigningTransactionGeneratorFromConfig(
	cfg config.DispatchConfig,
	chainCfg config.ChainConfig,
	signingAgent signing.SigningAgent,
	logger *zap.Logger,
) (TransactionGenerator, error) {
	// create tx config
	cdc, err := signing.Codec(chainCfg.Prefix)
	if err != nil {
		return nil, fmt.Errorf("failed to create interface registry: %w", err)
	}

	// set up connection to chain for gas estimation
	chainGRPC, err := grpc.NewClient(chainCfg.GRPCAddress, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to create chain gRPC client: %w", err)
	}

	gasEstimator := NewSimulationGasEstimator(chainGRPC, logger)

	return NewSigningTransactionGenerator(
		cdc,
		cfg,
		chainCfg,
		logger,
		gasEstimator,
		signingAgent,
	)
}

// NewSigningTransactionGenerator creates a new SigningTransactionGenerator.
func NewSigningTransactionGenerator(
	codec codec.Codec,
	cfg config.DispatchConfig,
	chainCfg config.ChainConfig,
	logger *zap.Logger,
	gasEstimator GasEstimator,
	signingAgent signing.SigningAgent,
) (TransactionGenerator, error) {
	sdkTxConfig := signing.TxConfig(codec)

	return &SigningTransactionGenerator{
		coreGenerator: coreGenerator{
			sdkTxConfig:   sdkTxConfig,
			logger:        logger,
			txConfig:      cfg.TxConfig,
			signingConfig: cfg.SigningConfig,
			chainConfig:   chainCfg,
			gasEstimator:  gasEstimator,
			signingAgent:  signingAgent,
		},
	}, nil
}

// GenerateTransactions generates and signs a set of transactions from the given set of markets using its internally
// configured wallet. Simulate can be set to true which will simulate the execution of the transactions, but will not
// sign and dispatch them.
func (s *SigningTransactionGenerator) GenerateTransactions(
	ctx context.Context,
	msgs []sdk.Msg,
) ([]cmttypes.Tx, error) {
	// get the account
	acc, err := s.coreGenerator.signingAgent.GetSigningAccount(ctx)
	if err != nil {
		s.logger.Error("failed to get signing account", zap.Error(err))
		return nil, err
	}

	s.logger.Info("account used to submit txs", zap.Any("account", acc))

	if acc.GetPubKey() == nil {
		return nil, errors.New("cannot find public key for signing account")
	}

	address, err := signing.PubKeyBech32(s.chainConfig.Prefix, acc.GetPubKey())
	if err != nil {
		s.logger.Error("failed to bech32ify address", zap.Error(err))
		return nil, NewInvalidSignerPubkeyError(err)
	}

	txs := make([]cmttypes.Tx, 0)
	simSequence := acc.GetSequence()

	for _, msg := range msgs {
		accSequence := acc.GetSequence()

		var upsertMsg sdk.Msg
		switch s.chainConfig.Version {
		case config.VersionConnect:
			upsert, ok := msg.(*mmtypes.MsgUpsertMarkets)
			if !ok {
				s.logger.Error("failed to cast sdk.Msg to expected type connect.MsgUpsertMarkets", zap.Any("msg", msg))
				return nil, fmt.Errorf("failed to cast sdk.Msg to expected type connect.MsgUpsertMarkets")
			}
			// ensure that the message authority is the signer key bech32 address for the chain
			upsert.Authority = address

			upsertMsg = upsert
		case config.VersionSlinky:
			upsert, ok := msg.(*slinkymmtypes.MsgUpsertMarkets)
			if !ok {
				s.logger.Error("failed to cast sdk.Msg to expected type slinky.MsgUpsertMarkets", zap.Any("msg", msg))
				return nil, fmt.Errorf("failed to cast sdk.Msg to expected type slinky.MsgUpsertMarkets")
			}
			// ensure that the message authority is the signer key bech32 address for the chain
			upsert.Authority = address

			upsertMsg = upsert
		default:
			return nil, fmt.Errorf("unsupported version: %s", s.chainConfig.Version)
		}

		txb, err := s.estimateUnsignedTx(upsertMsg, accSequence, simSequence)
		if err != nil {
			s.logger.Error("failed to estimate tx", zap.Error(err))
			return nil, err
		}

		tx, err := s.signingAgent.Sign(ctx, txb)
		if err != nil {
			s.logger.Error("failed to sign tx", zap.Error(err))
			return nil, err
		}

		// update the account sequence
		err = acc.SetSequence(accSequence + 1)
		if err != nil {
			return nil, err
		}

		txs = append(txs, tx)
	}

	s.logger.Info("generated txs", zap.Int("num tx", len(txs)))
	return txs, nil
}
