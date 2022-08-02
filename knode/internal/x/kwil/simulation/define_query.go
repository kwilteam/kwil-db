package simulation

import (
	"math/rand"

	"github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/kwilteam/kwil-db/knode/internal/x/kwil/keeper"
	"github.com/kwilteam/kwil-db/knode/internal/x/kwil/types"
)

func SimulateMsgDefineQuery(
	ak types.AccountKeeper,
	bk types.BankKeeper,
	k keeper.Keeper,
) simtypes.Operation {
	return func(r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		simAccount, _ := simtypes.RandomAcc(r, accs)
		msg := &types.MsgDefineQuery{
			Creator: simAccount.Address.String(),
		}

		// TODO: Handling the DefineQuery simulation

		return simtypes.NoOpMsg(types.ModuleName, msg.Type(), "DefineQuery simulation not implemented"), nil, nil
	}
}
