package keeper_test

import (
	"strconv"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/kwilteam/kwil-db/knode/internal/x/kwil/types"
	keepertest "github.com/kwilteam/kwil-db/knode/testutil/keeper"
	"github.com/kwilteam/kwil-db/knode/testutil/nullify"
)

// Prevent strconv unused error
var _ = strconv.IntSize

func TestDdlQuerySingle(t *testing.T) {
	keeper, ctx := keepertest.KwilKeeper(t)
	wctx := sdk.WrapSDKContext(ctx)
	msgs := createNDdl(keeper, ctx, 2)
	for _, tc := range []struct {
		desc     string
		request  *types.QueryGetDdlRequest
		response *types.QueryGetDdlResponse
		err      error
	}{
		{
			desc: "First",
			request: &types.QueryGetDdlRequest{
				Index: msgs[0].Index,
			},
			response: &types.QueryGetDdlResponse{Ddl: msgs[0]},
		},
		{
			desc: "Second",
			request: &types.QueryGetDdlRequest{
				Index: msgs[1].Index,
			},
			response: &types.QueryGetDdlResponse{Ddl: msgs[1]},
		},
		{
			desc: "KeyNotFound",
			request: &types.QueryGetDdlRequest{
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
			response, err := keeper.Ddl(wctx, tc.request)
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

func TestDdlQueryPaginated(t *testing.T) {
	keeper, ctx := keepertest.KwilKeeper(t)
	wctx := sdk.WrapSDKContext(ctx)
	msgs := createNDdl(keeper, ctx, 5)

	request := func(next []byte, offset, limit uint64, total bool) *types.QueryAllDdlRequest {
		return &types.QueryAllDdlRequest{
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
			resp, err := keeper.DdlAll(wctx, request(nil, uint64(i), uint64(step), false))
			require.NoError(t, err)
			require.LessOrEqual(t, len(resp.Ddl), step)
			require.Subset(t,
				nullify.Fill(msgs),
				nullify.Fill(resp.Ddl),
			)
		}
	})
	t.Run("ByKey", func(t *testing.T) {
		step := 2
		var next []byte
		for i := 0; i < len(msgs); i += step {
			resp, err := keeper.DdlAll(wctx, request(next, 0, uint64(step), false))
			require.NoError(t, err)
			require.LessOrEqual(t, len(resp.Ddl), step)
			require.Subset(t,
				nullify.Fill(msgs),
				nullify.Fill(resp.Ddl),
			)
			next = resp.Pagination.NextKey
		}
	})
	t.Run("Total", func(t *testing.T) {
		resp, err := keeper.DdlAll(wctx, request(nil, 0, 0, true))
		require.NoError(t, err)
		require.Equal(t, len(msgs), int(resp.Pagination.Total))
		require.ElementsMatch(t,
			nullify.Fill(msgs),
			nullify.Fill(resp.Ddl),
		)
	})
	t.Run("InvalidRequest", func(t *testing.T) {
		_, err := keeper.DdlAll(wctx, nil)
		require.ErrorIs(t, err, status.Error(codes.InvalidArgument, "invalid request"))
	})
}
