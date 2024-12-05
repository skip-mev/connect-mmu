package generator

import (
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// GasEstimator is an interface for estimating the gas consumption of transactions.
type GasEstimator interface {
	Estimate(txf tx.Factory, msgs []sdk.Msg, gasAdjust float64) (uint64, error)
}
