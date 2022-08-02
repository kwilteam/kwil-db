package kwil

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/kwilteam/kwil-db/knode/internal/x/kwil/keeper"
	"github.com/kwilteam/kwil-db/knode/internal/x/kwil/types"
)

// InitGenesis initializes the capability module's state from a provided genesis
// state.
func InitGenesis(ctx sdk.Context, k keeper.Keeper, genState types.GenesisState) {
	// Set all the databases
	for _, elem := range genState.DatabasesList {
		k.SetDatabases(ctx, elem)
	}
	// Set all the ddl
	for _, elem := range genState.DdlList {
		k.SetDdl(ctx, elem)
	}
	// Set all the ddlindex
	for _, elem := range genState.DdlindexList {
		k.SetDdlindex(ctx, elem)
	}
	// Set all the queryids
	for _, elem := range genState.QueryidsList {
		k.SetQueryids(ctx, elem)
	}
	// this line is used by starport scaffolding # genesis/module/init
	k.SetParams(ctx, genState.Params)
}

// ExportGenesis returns the capability module's exported genesis.
func ExportGenesis(ctx sdk.Context, k keeper.Keeper) *types.GenesisState {
	genesis := types.DefaultGenesis()
	genesis.Params = k.GetParams(ctx)

	genesis.DatabasesList = k.GetAllDatabases(ctx)
	genesis.DdlList = k.GetAllDdl(ctx)
	genesis.DdlindexList = k.GetAllDdlindex(ctx)
	genesis.QueryidsList = k.GetAllQueryids(ctx)
	// this line is used by starport scaffolding # genesis/module/export

	return genesis
}
