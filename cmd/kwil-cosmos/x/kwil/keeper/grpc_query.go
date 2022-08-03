package keeper

import (
	"github.com/kwilteam/kwil-db/cmd/kwil-cosmos/x/kwil/types"
)

var _ types.QueryServer = Keeper{}
