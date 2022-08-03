package keeper

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"strconv"
	"strings"

	"github.com/kwilteam/kwil-db/cmd/kwil-cosmos/x/kwil/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

func (k msgServer) DefineQuery(goCtx context.Context, msg *types.MsgDefineQuery) (*types.MsgDefineQueryResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// First, I will check to ensure the dbid exists

	db, isFound := k.GetDatabases(ctx, msg.DbId)
	if !isFound {
		return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "Specified DBid does not exist")
	}
	if db.Owner != msg.Creator {
		return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "Caller is not the owner of this DB")
	}

	// Adding the publicity to the ParQuer to ensure that they give different query indices
	var QueryToBeHashed strings.Builder
	QueryToBeHashed.WriteString(msg.ParQuer)
	QueryToBeHashed.WriteString(strconv.FormatBool(msg.Publicity))

	// Create the queryid
	queryIDBytes := sha256.Sum256([]byte(QueryToBeHashed.String()))
	queryID := hex.EncodeToString(queryIDBytes[:])

	// Create the index which is the dbid, queryid, and publicity concatenated
	var sb strings.Builder
	sb.WriteString(msg.DbId)
	sb.WriteString(queryID)

	// Now check the KV to ensure this db id does not exist
	_, isFound = k.GetQueryids(ctx, sb.String())
	if isFound {
		return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "This query already exists for this database")
	}

	// Set the new paramaterized query
	newParQuer := types.Queryids{
		Index:     sb.String(),
		Queryid:   queryID,
		Query:     msg.ParQuer,
		Dbid:      msg.DbId,
		Publicity: strconv.FormatBool(msg.Publicity),
	}
	// Store it
	k.SetQueryids(ctx, newParQuer)

	return &types.MsgDefineQueryResponse{Id: queryID}, nil
}
