package submitter_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"go.uber.org/zap/zaptest"

	cmtabci "github.com/cometbft/cometbft/abci/types"
	rpctypes "github.com/cometbft/cometbft/rpc/core/types"
	cmttypes "github.com/cometbft/cometbft/types"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/skip-mev/connect-mmu/config"
	"github.com/skip-mev/connect-mmu/dispatcher/transaction/submitter"
	"github.com/skip-mev/connect-mmu/dispatcher/transaction/submitter/mocks"
)

func TestTransactionSubmitter(t *testing.T) {
	cli := mocks.NewCometJSONRPCClient(t)

	pollingFrequency := 100 * time.Millisecond
	s := submitter.NewTransactionSubmitter(
		cli,
		config.SubmitterConfig{
			PollingFrequency: pollingFrequency,
			PollingDuration:  config.DefaultPollingDuration,
		},
		zaptest.NewLogger(t),
	)

	t.Run("failure to broadcast", func(t *testing.T) {
		ctx := context.Background()
		tx := cmttypes.Tx("tx")
		err := fmt.Errorf("error broadcasting tx")

		cli.On("BroadcastTxSync", mock.Anything, tx).Return(nil, err).Once()

		actualErr := s.Submit(ctx, tx)
		require.Error(t, actualErr)
	})

	t.Run("transaction failed in check-tx", func(t *testing.T) {
		ctx := context.Background()
		tx := cmttypes.Tx("tx")

		code := uint32(1)
		log := "error in check-tx"

		cli.On("BroadcastTxSync", mock.Anything, tx).Return(&rpctypes.ResultBroadcastTx{
			Code: code,
			Log:  log,
		}, nil).Once()

		actualErr := s.Submit(ctx, tx)
		require.Error(t, actualErr)
	})

	t.Run("broadcast success but execution failure", func(t *testing.T) {
		ctx := context.Background()
		tx := cmttypes.Tx("tx")

		cli.On("BroadcastTxSync", mock.Anything, tx).Return(&rpctypes.ResultBroadcastTx{
			Code: cmtabci.CodeTypeOK,
		}, nil).Once()

		cli.On("Tx", mock.Anything, mock.Anything, true).Return(&rpctypes.ResultTx{
			Tx: tx,
			TxResult: cmtabci.ExecTxResult{
				Code: 1,
				Log:  "invalid",
			},
		}, nil).Once()

		err := s.Submit(ctx, tx)
		require.Error(t, err)
	})

	t.Run("transaction success", func(t *testing.T) {
		ctx := context.Background()
		tx := cmttypes.Tx("tx")

		cli.On("BroadcastTxSync", mock.Anything, tx).Return(&rpctypes.ResultBroadcastTx{
			Code: cmtabci.CodeTypeOK,
		}, nil).Once()

		cli.On("Tx", mock.Anything, mock.Anything, true).Return(&rpctypes.ResultTx{
			Tx: tx,
			TxResult: cmtabci.ExecTxResult{
				Code: 0,
			},
		}, nil).Once()

		err := s.Submit(ctx, tx)
		require.NoError(t, err)
	})
}
