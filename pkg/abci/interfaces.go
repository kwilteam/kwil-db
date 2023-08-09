package abci

import (
	"context"

	"github.com/kwilteam/kwil-db/pkg/engine/types"
	"github.com/kwilteam/kwil-db/pkg/tx"
)

type DatabaseModule interface {
	Deploy(ctx context.Context, schema *types.Schema, tx *tx.Transaction) (*tx.ExecutionResponse, error)
	Drop(ctx context.Context, dbid string, tx *tx.Transaction) (*tx.ExecutionResponse, error)
	Execute(ctx context.Context, dbid string, action string, params []map[string]any, tx *tx.Transaction) (*tx.ExecutionResponse, error)
}

// Should be implemented in pkg/modules/validators
type ValidatorModule interface {
	ValidatorJoin(ctx context.Context, address string, tx *tx.Transaction) (*tx.ExecutionResponse, error)
	ValidatorApprove(ctx context.Context, address string, approvedBy string, tx *tx.Transaction) (*tx.ExecutionResponse, error)
}

// AtomicCommitter is an interface for a struct that implements atomic commits across multiple stores
type AtomicCommitter interface {
	Begin(ctx context.Context) error
	Commit(ctx context.Context, applyCallback func(error)) (commitId []byte, err error)
}
