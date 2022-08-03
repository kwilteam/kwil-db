package keeper_test

import (
	"strconv"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	keepertest "github.com/kwilteam/kwil-db/cmd/kwil-cosmos/testutil/keeper"
	"github.com/kwilteam/kwil-db/cmd/kwil-cosmos/testutil/nullify"
	"github.com/kwilteam/kwil-db/cmd/kwil-cosmos/x/kwil/types"
)

// Prevent strconv unused error
var _ = strconv.IntSize

func TestQueryidsQuerySingle(t *testing.T) {
	keeper, ctx := keepertest.KwilKeeper(t)
	wctx := sdk.WrapSDKContext(ctx)
	msgs := createNQueryids(keeper, ctx, 2)
	for _, tc := range []struct {
		desc     string
		request  *types.QueryGetQueryidsRequest
		response *types.QueryGetQueryidsResponse
		err      error
	}{
		{
			desc: "First",
			request: &types.QueryGetQueryidsRequest{
				Index: msgs[0].Index,
			},
			response: &types.QueryGetQueryidsResponse{Queryids: msgs[0]},
		},
		{
			desc: "Second",
			request: &types.QueryGetQueryidsRequest{
				Index: msgs[1].Index,
			},
			response: &types.QueryGetQueryidsResponse{Queryids: msgs[1]},
		},
		{
			desc: "KeyNotFound",
			request: &types.QueryGetQueryidsRequest{
				Index: strconv.Itoa(100000),
			},
			err: status.Error(codes.NotFound, "not found"),
		},
		{
			desc: "InvalidRequest",
			err:  status.Error(codes.InvalidArgument, "invalid request"),
		},
	} {
		t.Run(tc.desc, func(t *testing.T) {
			response, err := keeper.Queryids(wctx, tc.request)
			if tc.err != nil {
				require.ErrorIs(t, err, tc.err)
			} else {
				require.NoError(t, err)
				require.Equal(t,
					nullify.Fill(tc.response),
					nullify.Fill(response),
				)
			}
		})
	}
}

func TestQueryidsQueryPaginated(t *testing.T) {
	keeper, ctx := keepertest.KwilKeeper(t)
	wctx := sdk.WrapSDKContext(ctx)
	msgs := createNQueryids(keeper, ctx, 5)

	request := func(next []byte, offset, limit uint64, total bool) *types.QueryAllQueryidsRequest {
		return &types.QueryAllQueryidsRequest{
			Pagination: &query.PageRequest{
				Key:        next,
				Offset:     offset,
				Limit:      limit,
				CountTotal: total,
			},
		}
	}
	t.Run("ByOffset", func(t *testing.T) {
		step := 2
		for i := 0; i < len(msgs); i += step {
			resp, err := keeper.QueryidsAll(wctx, request(nil, uint64(i), uint64(step), false))
			require.NoError(t, err)
			require.LessOrEqual(t, len(resp.Queryids), step)
			require.Subset(t,
				nullify.Fill(msgs),
				nullify.Fill(resp.Queryids),
			)
		}
	})
	t.Run("ByKey", func(t *testing.T) {
		step := 2
		var next []byte
		for i := 0; i < len(msgs); i += step {
			resp, err := keeper.QueryidsAll(wctx, request(next, 0, uint64(step), false))
			require.NoError(t, err)
			require.LessOrEqual(t, len(resp.Queryids), step)
			require.Subset(t,
				nullify.Fill(msgs),
				nullify.Fill(resp.Queryids),
			)
			next = resp.Pagination.NextKey
		}
	})
	t.Run("Total", func(t *testing.T) {
		resp, err := keeper.QueryidsAll(wctx, request(nil, 0, 0, true))
		require.NoError(t, err)
		require.Equal(t, len(msgs), int(resp.Pagination.Total))
		require.ElementsMatch(t,
			nullify.Fill(msgs),
			nullify.Fill(resp.Queryids),
		)
	})
	t.Run("InvalidRequest", func(t *testing.T) {
		_, err := keeper.QueryidsAll(wctx, nil)
		require.ErrorIs(t, err, status.Error(codes.InvalidArgument, "invalid request"))
	})
}
