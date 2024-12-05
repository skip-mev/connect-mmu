package generator

import (
	"context"
	"fmt"
	"math"

	cmttypes "github.com/cometbft/cometbft/types"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	"go.uber.org/zap"

	"github.com/skip-mev/connect-mmu/config"
	mmusigning "github.com/skip-mev/connect-mmu/signing"
)

// TransactionGenerator handles the process of transforming a set of Markets in to a
// transaction that creates a set of upserts.
type TransactionGenerator interface {
	// GenerateTransactions returns a set of transactions to upsert a set of markets.
	GenerateTransactions(ctx context.Context, msgs []sdk.Msg) ([]cmttypes.Tx, error)
}

// coreGenerator is a set of core types commonly shared by generators
type coreGenerator struct {
	// txConfig is the SDK tx config used for transaction construction
	sdkTxConfig client.TxConfig

	// logger is the logger used by the transaction provider
	logger *zap.Logger

	// config
	txConfig      config.TransactionConfig
	signingConfig config.SigningConfig
	chainConfig   config.ChainConfig

	// gasEstimator is used to simulate transactions and estimate gas costs.
	gasEstimator GasEstimator

	// signingAgent is used to sign transactions as they are being generated.
	signingAgent mmusigning.SigningAgent
}

func (c *coreGenerator) estimateUnsignedTx(
	msg sdk.Msg,
	accSequence,
	simSequence uint64,
) (client.TxBuilder, error) {
	txf := tx.Factory{}

	minGasPrice := c.txConfig.MinGasPrice

	txf = txf.WithGas(c.txConfig.MaxGas)
	txf = txf.WithSignMode(signing.SignMode_SIGN_MODE_DIRECT)
	// Set sequence for simulation.
	txf = txf.WithSequence(simSequence)
	txf = txf.WithChainID(c.chainConfig.ChainID)
	txf = txf.WithTxConfig(c.sdkTxConfig)

	c.logger.Info("estimating transaction and gas")
	gas, err := c.gasEstimator.Estimate(txf, []sdk.Msg{msg}, c.txConfig.GasAdjustment)
	if err != nil {
		return nil, fmt.Errorf("failed to estimate gas: %w", err)
	}
	if gas > c.txConfig.MaxGas {
		return nil, fmt.Errorf("gas estimation of %d exceeds max gas: %d", gas, c.txConfig.MaxGas)
	}
	c.logger.Info("gas returned from tx simulation", zap.Uint64("gas_estimation", gas))

	if gas > math.MaxInt64 {
		gas = math.MaxInt64
	}

	// create some padding
	txf = txf.WithGasPrices(minGasPrice.String())
	txf = txf.WithGas(gas)
	txf = txf.WithSequence(accSequence) // set actual sequence

	c.logger.Info("transaction configuration",
		zap.String("chain-id", txf.ChainID()),
		zap.Uint64("sequence", txf.Sequence()),
		zap.Uint64("gas_estimation", gas),
		zap.String("gas prices", txf.GasPrices().String()),
	)

	return txf.BuildUnsignedTx(msg)
}
