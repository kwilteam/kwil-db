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

func (k Keeper) QueryidsAll(c context.Context, req *types.QueryAllQueryidsRequest) (*types.QueryAllQueryidsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	var queryidss []types.Queryids
	ctx := sdk.UnwrapSDKContext(c)

	store := ctx.KVStore(k.storeKey)
	queryidsStore := prefix.NewStore(store, types.KeyPrefix(types.QueryidsKeyPrefix))

	pageRes, err := query.Paginate(queryidsStore, req.Pagination, func(key []byte, value []byte) error {
		var queryids types.Queryids
		if err := k.cdc.Unmarshal(value, &queryids); err != nil {
			return err
		}

		queryidss = append(queryidss, queryids)
		return nil
	})

	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QueryAllQueryidsResponse{Queryids: queryidss, Pagination: pageRes}, nil
}

func (k Keeper) Queryids(c context.Context, req *types.QueryGetQueryidsRequest) (*types.QueryGetQueryidsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	ctx := sdk.UnwrapSDKContext(c)

	val, found := k.GetQueryids(
		ctx,
		req.Index,
	)
	if !found {
		return nil, status.Error(codes.NotFound, "not found")
	}

	return &types.QueryGetQueryidsResponse{Queryids: val}, nil
}
