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
