package keeper

import (
	"context"

	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	"github.com/kwilteam/kwil-db/knode/internal/x/kwil/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (k Keeper) DdlAll(c context.Context, req *types.QueryAllDdlRequest) (*types.QueryAllDdlResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	var ddls []types.Ddl
	ctx := sdk.UnwrapSDKContext(c)

	store := ctx.KVStore(k.storeKey)
	ddlStore := prefix.NewStore(store, types.KeyPrefix(types.DdlKeyPrefix))

	pageRes, err := query.Paginate(ddlStore, req.Pagination, func(key []byte, value []byte) error {
		var ddl types.Ddl
		if err := k.cdc.Unmarshal(value, &ddl); err != nil {
			return err
		}

		ddls = append(ddls, ddl)
		return nil
	})

	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QueryAllDdlResponse{Ddl: ddls, Pagination: pageRes}, nil
}

func (k Keeper) Ddl(c context.Context, req *types.QueryGetDdlRequest) (*types.QueryGetDdlResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	ctx := sdk.UnwrapSDKContext(c)

	val, found := k.GetDdl(
		ctx,
		req.Index,
	)
	if !found {
		return nil, status.Error(codes.NotFound, "not found")
	}

	return &types.QueryGetDdlResponse{Ddl: val}, nil
}
