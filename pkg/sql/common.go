package sql

import "context"

type Statement interface {
	Execute(ctx context.Context, args map[string]any) ([]map[string]any, error)
	Close() error
}

type Savepoint interface {
	Rollback() error
	Commit() error
}

type Session interface {
	Delete() error
	GenerateChangeset() ([]byte, error)
}
