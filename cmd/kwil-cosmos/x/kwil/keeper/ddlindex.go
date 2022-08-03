package keeper

import (
	"strconv"

	"github.com/kwilteam/kwil-db/cmd/kwil-cosmos/x/kwil/types"

	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// SetDdlindex set a specific ddlindex in the store from its index
func (k Keeper) SetDdlindex(ctx sdk.Context, ddlindex types.Ddlindex) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.DdlindexKeyPrefix))
	b := k.cdc.MustMarshal(&ddlindex)
	store.Set(types.DdlindexKey(
		ddlindex.Index,
	), b)

	// Now, we go to the old position and set final to false if this is not the first one

	if ddlindex.Position > 0 {

		formerDDLKey := ddlindex.Index + strconv.Itoa(int(ddlindex.Position-1))

		formerDDLValue, isFound := k.GetDdl(ctx, formerDDLKey)
		if !isFound {
			panic("Error finding ddl when updating 'final' to false")
		}
		formerDDLValue.Final = false
		k.SetDdl(ctx, formerDDLValue)

	}
}

// GetDdlindex returns a ddlindex from its index
func (k Keeper) GetDdlindex(
	ctx sdk.Context,
	index string,

) (val types.Ddlindex, found bool) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.DdlindexKeyPrefix))

	b := store.Get(types.DdlindexKey(
		index,
	))
	if b == nil {
		return val, false
	}

	k.cdc.MustUnmarshal(b, &val)
	return val, true
}

// RemoveDdlindex removes a ddlindex from the store
func (k Keeper) RemoveDdlindex(
	ctx sdk.Context,
	index string,

) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.DdlindexKeyPrefix))
	store.Delete(types.DdlindexKey(
		index,
	))
}

// GetAllDdlindex returns all ddlindex
func (k Keeper) GetAllDdlindex(ctx sdk.Context) (list []types.Ddlindex) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.DdlindexKeyPrefix))
	iterator := sdk.KVStorePrefixIterator(store, []byte{})

	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var val types.Ddlindex
		k.cdc.MustUnmarshal(iterator.Value(), &val)
		list = append(list, val)
	}

	return
}
