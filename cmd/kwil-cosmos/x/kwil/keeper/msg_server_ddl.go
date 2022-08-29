package keeper

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/kwilteam/kwil-db/cmd/kwil-cosmos/x/kwil/types"
	"github.com/kwilteam/kwil-db/internal/ctx"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

func (k msgServer) DDL(goCtx context.Context, msg *types.MsgDDL) (*types.MsgDDLResponse, error) {
	c := sdk.UnwrapSDKContext(goCtx)

	// First check if the specified DB exists
	db, isFound := k.GetDatabases(c, msg.Dbid)
	if !isFound {
		return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "Specified DBid does not exist")
	}

	// Check if caller is the database owner
	if db.Owner != msg.Creator {
		return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "Caller is not the owner of this DB")
	}

	// Store the DDL
	// First get the correct position for this to be stored in

	ddlIndex, isFound := k.GetDdlindex(c, msg.Dbid)
	if !isFound {
		return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "Could not find index for specified DB")
	}

	// Now we check to ensure that there is NOT already DDL at the new location in the map
	newPosition := ddlIndex.Position + 1

	// Creating string builder to create the index
	var newDDLLocation strings.Builder

	// Create index
	newDDLLocation.WriteString(msg.Dbid)
	newDDLLocation.WriteString(strconv.Itoa(int(newPosition)))
	_, isFound = k.GetDdl(c, newDDLLocation.String())
	if isFound {
		fmt.Println("There was an error when adding new DDL")
		return nil, sdkerrors.Wrap(sdkerrors.ErrUnknownRequest, "Internal error: DDL already exists at this map location")
	}

	if c.IsCheckTx() {
		return &types.MsgDDLResponse{}, nil
	}

	kctx := ctx.Unwrap(goCtx)
	if kctx == nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "Could not unwrap kwil context")
	}

	// If we have reached this point, we can update the new DDL and DDL index

	newDDL := types.Ddl{
		Index:     newDDLLocation.String(),
		Statement: msg.Ddl,
		Position:  newPosition,
		Final:     true,
	}

	//Set DDL in the database
	k.SetDdl(c, newDDL)

	// Create new ddl index

	newDDLIndex := types.Ddlindex{
		Index:    msg.Dbid,
		Name:     msg.Dbid,
		Position: newPosition,
	}
	k.SetDdlindex(c, newDDLIndex)

	return &types.MsgDDLResponse{}, nil
}
