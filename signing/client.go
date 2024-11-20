package signing

import (
	"context"

	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"google.golang.org/grpc"
)

// AuthClient is the expected interface that a client of the x/auth module's grpc service
// must implement
//
//go:generate mockery --name=AuthClient --filename=mock_auth_client.go --case=underscore
type AuthClient interface {
	Account(ctx context.Context, req *authtypes.QueryAccountRequest, opts ...grpc.CallOption) (*authtypes.QueryAccountResponse, error)
}
