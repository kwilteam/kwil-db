package atomic

import (
	"context"
	"io"

	"github.com/kwilteam/kwil-db/pkg/sql"
)

// Database is a connection to a database.  It is not safe for concurrent use.
type Database interface {
	Prepare(stmt string) (sql.Statement, error)
	CreateSession() (sql.Session, error)
	Savepoint() (sql.Savepoint, error)
	Execute(ctx context.Context, stmt string, args map[string]any) error
	TableExists(ctx context.Context, table string) (bool, error)
	ApplyChangeset(reader io.Reader) error
	CheckpointWal() error
}

type DatabaseOpener interface {
	// OpenDatabase returns a database connection for the given dbid
	OpenDatabase(ctx context.Context, dbid string) (Database, error)

	// DeleteDatabase deletes a database from disk.  This is not reversible.
	// if no database exists with the given dbid, this method does nothing.
	DeleteDatabase(ctx context.Context, dbid string) error
}

// StatementParser is a function that takes a statement and returns a new statement and an error.
// It attempts to rewrite the statement to be deterministic, if it is not already.  If it cannot
// do so, it returns an error.
// It also performs checks like protecting against cartesian products.
type StatementParser func(stmt string) (newStatement string, err error)

// DeterministicEncoderDecoder is responsible for deterministically encoding and decoding
// arbitrary data/
type DeterministicEncoderDecoder interface {
	Encode(val any) ([]byte, error)
	Decode(data []byte, val any) error
}
