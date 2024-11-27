package dispatcher

import (
	"context"
	"encoding/hex"
	"fmt"

	cmthttp "github.com/cometbft/cometbft/rpc/client/http"
	cmttypes "github.com/cometbft/cometbft/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"go.uber.org/zap"

	"github.com/skip-mev/connect-mmu/config"
	"github.com/skip-mev/connect-mmu/dispatcher/transaction/generator"
	"github.com/skip-mev/connect-mmu/dispatcher/transaction/submitter"
	"github.com/skip-mev/connect-mmu/signing"
)

// ServiceLabel is the label used to identify the dispatcher service in logs.
const (
	ServiceLabel = "dispatcher"
)

// Dispatcher provides functionality for generating transactions from a set of messages and a tx configuration,
// and submitting transactions sequentially to a chain via comet rpc.
type Dispatcher struct {
	// logger is the logger used by the dispatcher
	logger *zap.Logger

	// transactionGenerator is the client that provides the transaction that upserts a set of markets
	transactionGenerator generator.TransactionGenerator

	// transactionClient is the client that communicates with the market-map module of a chain
	transactionClient submitter.TransactionSubmitter

	txConfig      config.TransactionConfig
	signingConfig config.SigningConfig
}

// New creates a new Dispatcher from a configuration.
func New(
	cfg config.DispatchConfig,
	chainCfg config.ChainConfig,
	signer signing.SigningAgent,
	logger *zap.Logger,
) (*Dispatcher, error) {
	// create a comet-rpc client
	rpcClient, err := cmthttp.New(chainCfg.RPCAddress, "") // ignore websocket for now
	if err != nil {
		return nil, err
	}

	// transaction submitter
	txSubmitter := submitter.NewTransactionSubmitter(rpcClient, cfg.SubmitterConfig, logger)

	// transaction provider
	txProvider, err := generator.NewSigningTransactionGeneratorFromConfig(
		cfg,
		chainCfg,
		signer,
		logger,
	)
	if err != nil {
		return nil, err
	}

	return NewFromClients(
		txProvider,
		txSubmitter,
		logger,
		cfg,
	), nil
}

// NewFromClients creates a new Dispatcher.
func NewFromClients(
	txg generator.TransactionGenerator,
	tc submitter.TransactionSubmitter,
	logger *zap.Logger,
	cfg config.DispatchConfig,
) *Dispatcher {
	return &Dispatcher{
		transactionGenerator: txg,
		transactionClient:    tc,
		logger:               logger.With(zap.String("service", ServiceLabel)),
		txConfig:             cfg.TxConfig,
		signingConfig:        cfg.SigningConfig,
	}
}

// GenerateTransactions generates transactions for a given set of messages based on the dispatcher's tx configuration.
func (d *Dispatcher) GenerateTransactions(ctx context.Context, msgs []sdk.Msg) ([]cmttypes.Tx, error) {
	// retrieve set of transactions necessary for submitting upserts
	txs, err := d.transactionGenerator.GenerateTransactions(ctx, msgs)
	if err != nil {
		d.logger.Error("failed to generate transactions", zap.Error(err))
		return nil, err
	}

	d.logger.Info("successfully simulated and generated transactions", zap.Int("transactions", len(txs)))
	return txs, nil
}

// SubmitTransactions submits and verifies inclusion of transactions, one by one.
// If a transaction fails, the function returns the error and does not continue submitting the others.
func (d *Dispatcher) SubmitTransactions(ctx context.Context, txs []cmttypes.Tx) error {
	for _, tx := range txs {
		// submit the transaction
		if err := d.transactionClient.Submit(ctx, tx); err != nil {
			d.logger.Error("failed to submit transaction", zap.Error(err))
			return fmt.Errorf("failed to submit transaction: %w", err)
		}
		d.logger.Info("submitted transaction successfully", zap.String("tx", hex.EncodeToString(tx.Hash())))
	}

	d.logger.Info("successfully submitted all transactions", zap.Int("transactions", len(txs)))
	return nil
}
