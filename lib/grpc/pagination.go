package grpc

import (
	"context"

	"github.com/cosmos/cosmos-sdk/types/query"
)

// This is taken from https://github.com/dydxprotocol/v4-chain/blob/main/protocol/daemons/shared/paginated_grpc_request.go
const (
	// PaginatedRequestLimit is the maximum number of entries that can be returned in a paginated request.
	PaginatedRequestLimit = 10000
)

// ResponseWithPagination represents a response-type from a cosmos-module's GRPC service for entries that are paginated.
type ResponseWithPagination interface {
	GetPagination() *query.PageResponse
}

// PaginatedQuery is a function type that represents a paginated query to a cosmos-module's GRPC service.
type PaginatedQuery func(ctx context.Context, req *query.PageRequest) (ResponseWithPagination, error)

func HandlePaginatedQuery(ctx context.Context, pq PaginatedQuery, initialPagination *query.PageRequest) error {
	for {
		// make the query
		resp, err := pq(ctx, initialPagination)
		if err != nil {
			return err
		}

		// break if there is no next-key
		if resp.GetPagination() == nil || len(resp.GetPagination().NextKey) == 0 {
			return nil
		}

		// otherwise, update the next-key and continue
		initialPagination.Key = resp.GetPagination().NextKey
	}
}
