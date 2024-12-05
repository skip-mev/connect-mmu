package generator

import (
	"fmt"

	"go.uber.org/zap"

	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"google.golang.org/grpc"
)

type SimulationGasEstimator struct {
	conn   *grpc.ClientConn
	logger *zap.Logger
}

var _ GasEstimator = &SimulationGasEstimator{}

func NewSimulationGasEstimator(conn *grpc.ClientConn, logger *zap.Logger) GasEstimator {
	return &SimulationGasEstimator{
		conn:   conn,
		logger: logger,
	}
}

// Estimate uses a node to run a simulation of a transaction and adjusts the GasUsed for more headroom.
func (s *SimulationGasEstimator) Estimate(txf tx.Factory, msgs []sdk.Msg, gasAdjust float64) (uint64, error) {
	if s.conn == nil {
		return 0, fmt.Errorf("grpc conn not initialized")
	}

	if gasAdjust < 1 {
		return 0, fmt.Errorf("gasAdjust must be >= 1")
	}

	resp, _, err := tx.CalculateGas(s.conn, txf, msgs...)
	if err != nil {
		if resp != nil {
			s.logger.Error("failed to calculate gas estimation", zap.Error(err), zap.String("txResult", resp.String()))
		} else {
			s.logger.Error("failed to calculate gas estimation", zap.Error(err))
		}
		return 0, fmt.Errorf("failed to calculate gas: %w", err)
	}

	if resp == nil {
		return 0, fmt.Errorf("nil response from gasEstimator")
	}

	mul := float64(resp.GasInfo.GasUsed) * gasAdjust
	return uint64(mul), nil
}
