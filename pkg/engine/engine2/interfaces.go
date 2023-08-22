package engine2

import (
	"context"

	"github.com/kwilteam/kwil-db/pkg/engine/execution"
)

// An Executor processes a set of statements within a procedure, and returns the results of those statements
type Executor interface {
	Close() error
	ExecuteProcedure(ctx context.Context, name string, args []any, opts ...execution.ExecutionOpt) ([]map[string]any, error)
}
