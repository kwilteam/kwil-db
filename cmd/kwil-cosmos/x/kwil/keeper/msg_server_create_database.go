package keeper

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"github.com/kwilteam/kwil-db/internal/ctx"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/kwilteam/kwil-db/cmd/kwil-cosmos/x/kwil/types"
)

func (k msgServer) CreateDatabase(goCtx context.Context, msg *types.MsgCreateDatabase) (*types.MsgCreateDatabaseResponse, error) {
	kctx := ctx.Unwrap(goCtx)
	c := sdk.UnwrapSDKContext(goCtx)

	// First, must combine the sender and block height.
	// TODO: Right now the uniqueness is determined by a seed that is provided by the user.
	// This is not an ideal system, and is only used for prototyping right now.
	dbIDBytes := sha256.Sum256([]byte(msg.Creator + msg.Seed))

	// Type cast to string
	dbName := hex.EncodeToString(dbIDBytes[:])

	// Check if db ID already exists
	_, isFound := k.GetDatabases(c, dbName)

	// Return an error if it exists
	if isFound {
		return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "Database for that user with that seed already exists, please use a different seed")
	}

	// Create the new DB
	newDB := types.Databases{
		Index: dbName,
		Dbid:  dbName,
		Owner: msg.Creator,
	}

	// Store the DB
	k.SetDatabases(c, newDB)

	// TODO: We probably want to allow people to define any CREATE DATABASE esq statement, so that this natively supports non-SQL dbs
	// This change would require changing the protobufs and messages, so I'm not going to worry about it for now
	var createStatement strings.Builder
	createStatement.WriteString("CREATE DATABASE ")
	createStatement.WriteString(dbName)

	// Create new DDL
	newDDL := types.Ddl{
		Index:     dbName,
		Statement: createStatement.String(),
		Position:  0,
		Final:     true,
	}

	//Set DDL in the database
	k.SetDdl(c, newDDL)

	newDDLIndex := types.Ddlindex{
		Index:    dbName,
		Name:     dbName,
		Position: 0,
	}
	k.SetDdlindex(c, newDDLIndex)

	if !c.IsCheckTx() {
		fmt.Println("appending database")
		err := kctx.Wal().AppendCreateDatabase(dbName, createStatement.String())
		if err != nil {
			return nil, err
		}
	}

	return &types.MsgCreateDatabaseResponse{Id: dbName}, nil
}
