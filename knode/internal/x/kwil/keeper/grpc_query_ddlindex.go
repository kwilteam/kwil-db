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

func (k Keeper) DdlindexAll(c context.Context, req *types.QueryAllDdlindexRequest) (*types.QueryAllDdlindexResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	var ddlindexs []types.Ddlindex
	ctx := sdk.UnwrapSDKContext(c)

	store := ctx.KVStore(k.storeKey)
	ddlindexStore := prefix.NewStore(store, types.KeyPrefix(types.DdlindexKeyPrefix))

	pageRes, err := query.Paginate(ddlindexStore, req.Pagination, func(key []byte, value []byte) error {
		var ddlindex types.Ddlindex
		if err := k.cdc.Unmarshal(value, &ddlindex); err != nil {
			return err
		}

		ddlindexs = append(ddlindexs, ddlindex)
		return nil
	})

	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QueryAllDdlindexResponse{Ddlindex: ddlindexs, Pagination: pageRes}, nil
}

func (k Keeper) Ddlindex(c context.Context, req *types.QueryGetDdlindexRequest) (*types.QueryGetDdlindexResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	ctx := sdk.UnwrapSDKContext(c)

	val, found := k.GetDdlindex(
		ctx,
		req.Index,
	)
	if !found {
		return nil, status.Error(codes.NotFound, "not found")
	}

	return &types.QueryGetDdlindexResponse{Ddlindex: val}, nil
}
