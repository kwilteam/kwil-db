package keeper

import (
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/kwilteam/kwil-db/cmd/kwil-cosmos/x/kwil/types"
)

// SetQueryids set a specific queryids in the store from its index
func (k Keeper) SetQueryids(ctx sdk.Context, queryids types.Queryids) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.QueryidsKeyPrefix))
	b := k.cdc.MustMarshal(&queryids)
	store.Set(types.QueryidsKey(
		queryids.Index,
	), b)
}

// GetQueryids returns a queryids from its index
func (k Keeper) GetQueryids(
	ctx sdk.Context,
	index string,

) (val types.Queryids, found bool) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.QueryidsKeyPrefix))

	b := store.Get(types.QueryidsKey(
		index,
	))
	if b == nil {
		return val, false
	}

	k.cdc.MustUnmarshal(b, &val)
	return val, true
}

// RemoveQueryids removes a queryids from the store
func (k Keeper) RemoveQueryids(
	ctx sdk.Context,
	index string,

) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.QueryidsKeyPrefix))
	store.Delete(types.QueryidsKey(
		index,
	))
}

// GetAllQueryids returns all queryids
func (k Keeper) GetAllQueryids(ctx sdk.Context) (list []types.Queryids) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.QueryidsKeyPrefix))
	iterator := sdk.KVStorePrefixIterator(store, []byte{})

	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var val types.Queryids
		k.cdc.MustUnmarshal(iterator.Value(), &val)
		list = append(list, val)
	}

	return
}
