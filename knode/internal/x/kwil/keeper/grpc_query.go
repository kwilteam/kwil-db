package keeper

import (
	"github.com/kwilteam/kwil-db/knode/internal/x/kwil/types"
)

var _ types.QueryServer = Keeper{}
