package master

import (
	"context"
)

type Datastore interface {
	Close() error
	Execute(ctx context.Context, stmt string, args map[string]any) error
	Query(ctx context.Context, query string, args map[string]any) ([]map[string]any, error)
	TableExists(ctx context.Context, table string) (bool, error)
	Delete() error
	Savepoint() (Savepoint, error)
}

type Savepoint interface {
	Rollback() error
	Commit() error
}
