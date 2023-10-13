package sql

import (
	"context"
	"io"

	"github.com/kwilteam/kwil-db/core/log"
)

type Opener interface {
	Open(fileName string, logger log.Logger) (Database, error)
}

type Database interface {
	ApplyChangeset(reader io.Reader) error
	CheckpointWal() error
	Close() error
	CreateSession() (Session, error)
	Delete() error
	Execute(ctx context.Context, stmt string, args map[string]any) error
	Prepare(stmt string) (Statement, error)
	Query(ctx context.Context, query string, args map[string]any) ([]map[string]any, error)
	Savepoint() (Savepoint, error)
	TableExists(ctx context.Context, table string) (bool, error)
	EnableForeignKey() error
	DisableForeignKey() error
	// this should get deleted once we fix the engine
	QueryUnsafe(ctx context.Context, query string, args map[string]any) ([]map[string]any, error)
}

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
	GenerateChangeset() (Changeset, error)
}

type Changeset interface {
	// Export gets the changeset as a byte array.
	Export() ([]byte, error)

	// ID generates a deterministic ID for the changeset.
	// TODO: this is not deterministic yet
	ID() ([]byte, error)

	Close() error
}
