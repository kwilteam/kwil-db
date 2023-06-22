package eng

import (
	"context"
	"io"
)

type InitializedExtension interface {
	Execute(ctx context.Context, method string, args ...any) ([]any, error)
}

type Initializer interface {
	Initialize(context.Context, map[string]string) (InitializedExtension, error)
}

// Datastore is an interface for a datastore, usually a sqlite DB.
type Datastore interface {
	Prepare(query string) (PreparedStatement, error)
	Savepoint() (Savepoint, error)
}

type PreparedStatement interface {
	// Execute executes a prepared statement with the given arguments.
	Execute(args map[string]any) (io.Reader, error)

	// Close closes the statement.
	Close() error
}

type Savepoint interface {
	Rollback() error
	Commit() error
}
