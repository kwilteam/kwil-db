package kwil

import (
	"fmt"

	"github.com/kwilteam/kwil-db/internal/ctx"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/kwilteam/kwil-db/cmd/kwil-cosmos/x/kwil/keeper"
	"github.com/kwilteam/kwil-db/cmd/kwil-cosmos/x/kwil/types"
)

// NewHandler ...
func NewHandler(factory ctx.KwilCtxFactory, k keeper.Keeper) sdk.Handler {
	msgServer := keeper.NewMsgServerImpl(k)

	return func(ctx sdk.Context, msg sdk.Msg) (*sdk.Result, error) {
		ctx = ctx.WithContext(factory.CreateCtxFactory(ctx.Context()))
		ctx = ctx.WithEventManager(sdk.NewEventManager())

		switch msg := msg.(type) {
		case *types.MsgDatabaseWrite:
			res, err := msgServer.DatabaseWrite(sdk.WrapSDKContext(ctx), msg)
			return sdk.WrapServiceResult(ctx, res, err)
		case *types.MsgCreateDatabase:
			res, err := msgServer.CreateDatabase(sdk.WrapSDKContext(ctx), msg)
			return sdk.WrapServiceResult(ctx, res, err)
		case *types.MsgDDL:
			res, err := msgServer.DDL(sdk.WrapSDKContext(ctx), msg)
			return sdk.WrapServiceResult(ctx, res, err)
		case *types.MsgDefineQuery:
			res, err := msgServer.DefineQuery(sdk.WrapSDKContext(ctx), msg)
			return sdk.WrapServiceResult(ctx, res, err)
			// this line is used by starport scaffolding # 1
		default:
			errMsg := fmt.Sprintf("unrecognized %s message type: %T", types.ModuleName, msg)
			return nil, sdkerrors.Wrap(sdkerrors.ErrUnknownRequest, errMsg)
		}
	}
}
