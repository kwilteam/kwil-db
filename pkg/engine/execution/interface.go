package execution

import (
	"context"
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
}

type PreparedStatement interface {
	// Execute executes a prepared statement with the given arguments.
	Execute(ctx context.Context, args map[string]any) ([]map[string]any, error)

	// Close closes the statement.
	Close() error
}
