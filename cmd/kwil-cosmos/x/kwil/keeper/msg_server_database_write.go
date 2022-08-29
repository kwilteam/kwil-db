package keeper

import (
	"context"
	"strconv"
	"strings"

	"github.com/kwilteam/kwil-db/cmd/kwil-cosmos/x/kwil/types"
	"github.com/kwilteam/kwil-db/internal/ctx"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

func (k msgServer) DatabaseWrite(goCtx context.Context, msg *types.MsgDatabaseWrite) (*types.MsgDatabaseWriteResponse, error) {
	c := sdk.UnwrapSDKContext(goCtx)

	// We must check if the dbid exists, if the queryid exists, and if it is public.
	db, isFound := k.GetDatabases(c, msg.Database)
	if !isFound {
		return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "Specified database does not exist")
	}

	// Now checking for the registered query
	// First construct the queryid index
	var queryIndex strings.Builder
	queryIndex.WriteString(msg.Database)
	queryIndex.WriteString(msg.ParQuer)

	// Retrieve from KV
	parQuer, isFound := k.GetQueryids(c, queryIndex.String())
	// Check if it was found
	if !isFound {
		return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "Specified paramaterized query does not exist")
	}
	// Check if it is public
	publicity, err := strconv.ParseBool(parQuer.Publicity)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "There was an error finding the query's publicity")
	}

	if !publicity {
		// If not public, check if this caller is the owner
		if msg.Creator != db.Owner {
			return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "Specified paramaterized query is not public")
		}
	}

	if c.IsCheckTx() {
		return &types.MsgDatabaseWriteResponse{}, nil
	}

	kctx := ctx.Unwrap(goCtx)
	if kctx == nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "Could not unwrap kwil context")
	}
	// TODO: Write to the database here

	return &types.MsgDatabaseWriteResponse{}, nil
}
