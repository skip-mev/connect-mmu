package config

import (
	"errors"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// TransactionConfig is the necessary configuration for a MarketUpdateTransactionProvider.
type TransactionConfig struct {
	// MaxBytesPerTx is the maximum number of bytes allowed in a single transaction
	MaxBytesPerTx int `json:"max_bytes_per_tx"`

	// MaxGas is the maximum amount of gas allowed in a single transaction
	MaxGas uint64 `json:"max_gas"`

	// GasAdjustment is the gas adjustment multiplier to apply when estimating gas.
	GasAdjustment float64 `json:"gas_adjustment"`

	// MinGasPrice is the min gas prices used for the transaction
	MinGasPrice sdk.DecCoin `json:"min_gas_price"`
}

func DefaultTxConfig() TransactionConfig {
	return TransactionConfig{
		MaxBytesPerTx: 100000,
		MaxGas:        800000000,
		GasAdjustment: 1.5,
		MinGasPrice: sdk.DecCoin{
			Denom:  "utoken",
			Amount: math.LegacyNewDec(20000000000),
		},
	}
}

// ValidateBasic validates the configuration.
func (c *TransactionConfig) ValidateBasic() error {
	if c.MaxBytesPerTx <= 0 {
		return ErrInvalidMaxBytesPerTx
	}

	if c.MaxGas <= 0 {
		return ErrInvalidMaxGas
	}

	if !c.MinGasPrice.IsValid() {
		return ErrInvalidTxFee
	}

	if c.GasAdjustment < 1 {
		return ErrInvalidGasAdjustment
	}

	return nil
}

var (
	// ErrInvalidMaxBytesPerTx is thrown when the max bytes per tx is invalid.
	ErrInvalidMaxBytesPerTx = errors.New("max bytes per tx must be greater than 0")

	// ErrInvalidMaxGas is thrown when the max gas is invalid.
	ErrInvalidMaxGas = errors.New("max gas must be greater than 0")

	// ErrInvalidTxFee is thrown when the tx fee is invalid.
	ErrInvalidTxFee = errors.New("tx fee must be greater than or equal to 0")

	ErrInvalidGasAdjustment = errors.New("gas adjustment must be greater than or equal to 1")
)
