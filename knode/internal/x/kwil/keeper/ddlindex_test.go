package keeper_test

import (
	"strconv"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/kwilteam/kwil-db/knode/internal/x/kwil/keeper"
	"github.com/kwilteam/kwil-db/knode/internal/x/kwil/types"
	keepertest "github.com/kwilteam/kwil-db/knode/testutil/keeper"
	"github.com/kwilteam/kwil-db/knode/testutil/nullify"
	"github.com/stretchr/testify/require"
)

// Prevent strconv unused error
var _ = strconv.IntSize

func createNDdlindex(keeper *keeper.Keeper, ctx sdk.Context, n int) []types.Ddlindex {
	items := make([]types.Ddlindex, n)
	for i := range items {
		items[i].Index = strconv.Itoa(i)

		keeper.SetDdlindex(ctx, items[i])
	}
	return items
}

func TestDdlindexGet(t *testing.T) {
	keeper, ctx := keepertest.KwilKeeper(t)
	items := createNDdlindex(keeper, ctx, 10)
	for _, item := range items {
		rst, found := keeper.GetDdlindex(ctx,
			item.Index,
		)
		require.True(t, found)
		require.Equal(t,
			nullify.Fill(&item),
			nullify.Fill(&rst),
		)
	}
}
func TestDdlindexRemove(t *testing.T) {
	keeper, ctx := keepertest.KwilKeeper(t)
	items := createNDdlindex(keeper, ctx, 10)
	for _, item := range items {
		keeper.RemoveDdlindex(ctx,
			item.Index,
		)
		_, found := keeper.GetDdlindex(ctx,
			item.Index,
		)
		require.False(t, found)
	}
}

func TestDdlindexGetAll(t *testing.T) {
	keeper, ctx := keepertest.KwilKeeper(t)
	items := createNDdlindex(keeper, ctx, 10)
	require.ElementsMatch(t,
		nullify.Fill(items),
		nullify.Fill(keeper.GetAllDdlindex(ctx)),
	)
}
