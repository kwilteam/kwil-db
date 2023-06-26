package engine

import (
	"context"
	"io"

	"github.com/kwilteam/kwil-db/pkg/engine/dataset3"
	"github.com/kwilteam/kwil-db/pkg/engine/types"
)

type Datastore interface{}

type Dataset interface {
	Close() error
	Procedures() []*types.Procedure
	Tables(ctx context.Context) []*types.Table
	Delete() error
	Query(ctx context.Context, stmt string, args map[string]any) (io.Reader, error)
	Execute(ctx context.Context, stmt string, args map[string]any, opts *dataset3.TxOpts) (io.Reader, error)
}
