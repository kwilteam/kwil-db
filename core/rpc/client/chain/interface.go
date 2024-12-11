package chain

import (
	"context"

	"github.com/kwilteam/kwil-db/core/types"
	chaintypes "github.com/kwilteam/kwil-db/core/types/chain"
)

type Client interface {
	Version(ctx context.Context) (string, error)
	BlockByHeight(ctx context.Context, height int64) (*chaintypes.Block, error)
	BlockByHash(ctx context.Context, hash types.Hash) (*chaintypes.Block, error)
	BlockResultByHeight(ctx context.Context, height int64) (*chaintypes.BlockResult, error)
	BlockResultByHash(ctx context.Context, hash types.Hash) (*chaintypes.BlockResult, error)
	Tx(ctx context.Context, hash types.Hash) (*chaintypes.Tx, error)
	Genesis(ctx context.Context) (*chaintypes.Genesis, error)
	ConsensusParams(ctx context.Context) (*types.ConsensusParams, error)
	Validators(ctx context.Context) (height int64, validators []*types.Validator, err error)
	UnconfirmedTxs(ctx context.Context) (total int64, tx chaintypes.NamedTx, err error)
}
