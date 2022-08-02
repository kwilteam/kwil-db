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

func createNDatabases(keeper *keeper.Keeper, ctx sdk.Context, n int) []types.Databases {
	items := make([]types.Databases, n)
	for i := range items {
		items[i].Index = strconv.Itoa(i)

		keeper.SetDatabases(ctx, items[i])
	}
	return items
}

func TestDatabasesGet(t *testing.T) {
	keeper, ctx := keepertest.KwilKeeper(t)
	items := createNDatabases(keeper, ctx, 10)
	for _, item := range items {
		rst, found := keeper.GetDatabases(ctx,
			item.Index,
		)
		require.True(t, found)
		require.Equal(t,
			nullify.Fill(&item),
			nullify.Fill(&rst),
		)
	}
}
func TestDatabasesRemove(t *testing.T) {
	keeper, ctx := keepertest.KwilKeeper(t)
	items := createNDatabases(keeper, ctx, 10)
	for _, item := range items {
		keeper.RemoveDatabases(ctx,
			item.Index,
		)
		_, found := keeper.GetDatabases(ctx,
			item.Index,
		)
		require.False(t, found)
	}
}

func TestDatabasesGetAll(t *testing.T) {
	keeper, ctx := keepertest.KwilKeeper(t)
	items := createNDatabases(keeper, ctx, 10)
	require.ElementsMatch(t,
		nullify.Fill(items),
		nullify.Fill(keeper.GetAllDatabases(ctx)),
	)
}
