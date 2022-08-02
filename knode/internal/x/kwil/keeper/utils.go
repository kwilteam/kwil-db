package keeper

import sdk "github.com/cosmos/cosmos-sdk/types"

// TODO: Write a test for this
func (k msgServer) IsDbOwner(ctx sdk.Context, dbName string, address string) bool {
	// Get the database
	database, isFound := k.GetDatabases(ctx, dbName)

	// If the db is not found, it is not owned
	if !isFound {
		return false
	}

	// If the passedd address is not the owner, then return false
	if database.Owner != address {
		return false
	}
	return true
}
