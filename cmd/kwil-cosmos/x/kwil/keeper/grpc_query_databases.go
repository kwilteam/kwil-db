package keeper

import (
	"context"

	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	"github.com/kwilteam/kwil-db/cmd/kwil-cosmos/x/kwil/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (k Keeper) DatabasesAll(c context.Context, req *types.QueryAllDatabasesRequest) (*types.QueryAllDatabasesResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	var databasess []types.Databases
	ctx := sdk.UnwrapSDKContext(c)

	store := ctx.KVStore(k.storeKey)
	databasesStore := prefix.NewStore(store, types.KeyPrefix(types.DatabasesKeyPrefix))

	pageRes, err := query.Paginate(databasesStore, req.Pagination, func(key []byte, value []byte) error {
		var databases types.Databases
		if err := k.cdc.Unmarshal(value, &databases); err != nil {
			return err
		}

		databasess = append(databasess, databases)
		return nil
	})

	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QueryAllDatabasesResponse{Databases: databasess, Pagination: pageRes}, nil
}

func (k Keeper) Databases(c context.Context, req *types.QueryGetDatabasesRequest) (*types.QueryGetDatabasesResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	ctx := sdk.UnwrapSDKContext(c)

	val, found := k.GetDatabases(
		ctx,
		req.Index,
	)
	if !found {
		return nil, status.Error(codes.NotFound, "not found")
	}

	return &types.QueryGetDatabasesResponse{Databases: val}, nil
}
