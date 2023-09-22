package datasets

import (
	"context"

	"github.com/kwilteam/kwil-db/pkg/balances"
	"github.com/kwilteam/kwil-db/pkg/crypto"
	"github.com/kwilteam/kwil-db/pkg/engine"
	engineTypes "github.com/kwilteam/kwil-db/pkg/engine/types"
)

type AccountStore interface {
	Spend(ctx context.Context, spend *balances.Spend) error
}

type Engine interface {
	CreateDataset(ctx context.Context, schema *engineTypes.Schema, caller engineTypes.UserIdentifier) (dbid string, finalErr error)
	DropDataset(ctx context.Context, dbid string, sender engineTypes.UserIdentifier) error
	Execute(ctx context.Context, dbid string, procedure string, args [][]any, opts ...engine.ExecutionOpt) ([]map[string]any, error)
	ListDatasets(ctx context.Context, owner []byte) ([]string, error)
	Query(ctx context.Context, dbid string, query string) ([]map[string]any, error)
	GetSchema(ctx context.Context, dbid string) (*engineTypes.Schema, error)
}

// DatasetMessage is a message that dataset module could consume, and it can be
// verified if it's signed.
// We currently only have `transactions.CallMessage` that implements this interface.
type DatasetMessage interface {
	IsSigned() bool
	Verify() error
	GetSender() crypto.PublicKey
}
