package submitter

import (
	"context"
	"encoding/hex"
	"fmt"
	"time"

	cometabci "github.com/cometbft/cometbft/abci/types"
	ctypes "github.com/cometbft/cometbft/rpc/core/types"
	cmttypes "github.com/cometbft/cometbft/types"
	"go.uber.org/zap"

	"github.com/skip-mev/connect-mmu/config"
)

// TransactionSubmitter is the expected interface of a client that can submit transactions to a chain.
//
//go:generate mockery --name=TransactionSubmitter --output=mocks --case=underscore
type TransactionSubmitter interface {
	// Submit submits a transaction to the chain.
	Submit(ctx context.Context, tx cmttypes.Tx) error
}

var _ TransactionSubmitter = &CometTransactionSubmitter{}

// CometJSONRPCClient is the interface expected to be fulfilled by a comet JSON-RPC client.
//
//go:generate mockery --name CometJSONRPCClient --output=mocks --case=underscore
type CometJSONRPCClient interface {
	// BroadcastTxSync broadcasts a transaction to the chain and waits for checkTx.
	BroadcastTxSync(ctx context.Context, tx cmttypes.Tx) (*ctypes.ResultBroadcastTx, error)

	// Tx queries a tx by hash.
	Tx(ctx context.Context, hash []byte, prove bool) (*ctypes.ResultTx, error)
}

// NewTransactionSubmitter creates a new transaction submitter.
func NewTransactionSubmitter(
	client CometJSONRPCClient,
	cfg config.SubmitterConfig,
	logger *zap.Logger,
) TransactionSubmitter {
	return &CometTransactionSubmitter{
		client: client,
		logger: logger,
		cfg:    cfg,
	}
}

// CometTransactionSubmitter is a transaction submitter that submits transactions synchronously, specifically,
// it blocks on a check-tx and a transaction commit response before returning.
type CometTransactionSubmitter struct {
	client CometJSONRPCClient
	logger *zap.Logger
	cfg    config.SubmitterConfig
}

func (sts *CometTransactionSubmitter) Submit(ctx context.Context, tx cmttypes.Tx) error {
	sts.logger.Info("submitting tx", zap.String("tx", hex.EncodeToString(tx.Hash())))

	// create extended timeout to account for block inclusion of Tx
	broadcastCtx, cancel := context.WithTimeout(ctx, 20*time.Second)
	defer cancel()

	res, err := sts.client.BroadcastTxSync(broadcastCtx, tx)
	if err != nil {
		sts.logger.Error("failed to broadcast transaction", zap.Error(err))
		return NewTxBroadcastError(err)
	}

	// check for the check-tx code
	if res.Code != cometabci.CodeTypeOK {
		sts.logger.Error("transaction check-tx failed", zap.Uint32("code", res.Code), zap.String("log", res.Log))
		return NewCheckTxError(res.Log, res.Code)
	}

	sts.logger.Info("transaction submitted", zap.Uint32("code", res.Code), zap.String("tx", hex.EncodeToString(tx.Hash())))
	sts.logger.Debug("transaction log", zap.String("log", res.Log))

	// check that the tx has been included in a block
	if err := sts.checkTxInclusion(ctx, res.Hash); err != nil {
		return err
	}

	return nil
}

// checkTxInclusion checks that the given transaction has been included in a block.
func (sts *CometTransactionSubmitter) checkTxInclusion(ctx context.Context, hash []byte) error {
	ticker := time.NewTicker(sts.cfg.PollingFrequency)
	defer ticker.Stop()

	timer := time.NewTimer(sts.cfg.PollingDuration)
	defer timer.Stop()

	sts.logger.Info("checking transaction inclusion in block", zap.String("tx", hex.EncodeToString(hash)),
		zap.Duration("interval", sts.cfg.PollingFrequency), zap.Duration("polling time", sts.cfg.PollingDuration))

	for {
		select {
		case <-ticker.C:
			// try to check inclusion
			result, err := sts.client.Tx(ctx, hash, true)
			if err != nil {
				sts.logger.Debug("failed to check transaction", zap.Error(err))
				continue
			}

			// check the code of the result
			if result.TxResult.Code != cometabci.CodeTypeOK {
				sts.logger.Error("transaction tx result failed", zap.Uint32("code", result.TxResult.Code),
					zap.String("log", result.TxResult.Log), zap.String("info", result.TxResult.Info))
				return fmt.Errorf("transaction tx result failed with code: %d, log: %s", result.TxResult.Code,
					result.TxResult.Log)
			}

			sts.logger.Debug("transaction was successful", zap.Uint32("code", result.TxResult.Code),
				zap.String("log", result.TxResult.Log), zap.String("info", result.TxResult.Info))
			return nil

		case <-timer.C:
			return fmt.Errorf("timed out waiting for tx %s inclusion", hex.EncodeToString(hash))
		}
	}
}
