package signing

import (
	"context"

	cmttypes "github.com/cometbft/cometbft/types"
	"github.com/cosmos/cosmos-sdk/client"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// SigningAgent is an interface for signing.
//
//nolint:revive
//go:generate mockery --name=SigningAgent --filename=mock_signing_agent.go --case=underscore
type SigningAgent interface {
	Sign(ctx context.Context, txb client.TxBuilder) (cmttypes.Tx, error)

	GetSigningAccount(ctx context.Context) (sdk.AccountI, error)
}
