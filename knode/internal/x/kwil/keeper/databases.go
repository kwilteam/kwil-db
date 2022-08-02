package keeper

import (
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/kwilteam/kwil-db/knode/internal/x/kwil/types"
)

// SetDatabases set a specific databases in the store from its index
func (k Keeper) SetDatabases(ctx sdk.Context, databases types.Databases) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.DatabasesKeyPrefix))
	b := k.cdc.MustMarshal(&databases)
	store.Set(types.DatabasesKey(
		databases.Index,
	), b)
}

// GetDatabases returns a databases from its index
func (k Keeper) GetDatabases(
	ctx sdk.Context,
	index string,

) (val types.Databases, found bool) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.DatabasesKeyPrefix))

	b := store.Get(types.DatabasesKey(
		index,
	))
	if b == nil {
		return val, false
	}

	k.cdc.MustUnmarshal(b, &val)
	return val, true
}

// RemoveDatabases removes a databases from the store
func (k Keeper) RemoveDatabases(
	ctx sdk.Context,
	index string,

) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.DatabasesKeyPrefix))
	store.Delete(types.DatabasesKey(
		index,
	))
}

// GetAllDatabases returns all databases
func (k Keeper) GetAllDatabases(ctx sdk.Context) (list []types.Databases) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.DatabasesKeyPrefix))
	iterator := sdk.KVStorePrefixIterator(store, []byte{})

	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var val types.Databases
		k.cdc.MustUnmarshal(iterator.Value(), &val)
		list = append(list, val)
	}

	return
}
