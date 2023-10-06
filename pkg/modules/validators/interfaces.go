package validators

import (
	"context"
	"math/big"

	"github.com/kwilteam/kwil-db/pkg/balances"
	"github.com/kwilteam/kwil-db/pkg/validators"
)

type Spender interface {
	Spend(ctx context.Context, spend *balances.Spend) error
}

type ValidatorMgr interface {
	GenesisInit(ctx context.Context, vals []*validators.Validator, blockHeight int64) error
	CurrentSet(ctx context.Context) ([]*validators.Validator, error)
	Update(ctx context.Context, validator []byte, power int64) error
	Join(ctx context.Context, joiner []byte, power int64) error
	Leave(ctx context.Context, joiner []byte) error
	Approve(ctx context.Context, joiner, approver []byte) error
	Finalize(ctx context.Context) []*validators.Validator // end of block processing requires providing list of updates to the node's consensus client
	UpdateBlockHeight(blockHeight int64)

	PriceJoin(ctx context.Context) (*big.Int, error)
	PriceApprove(ctx context.Context) (*big.Int, error)
	PriceLeave(ctx context.Context) (*big.Int, error)
}
