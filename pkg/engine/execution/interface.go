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
	// Prepare will be used for RW execution.
	Prepare(ctx context.Context, query string) (PreparedStatement, error)
	// Query will be used for RO execution.
	Query(ctx context.Context, stmt string, args map[string]any) ([]map[string]any, error)
}

type PreparedStatement interface {
	// Execute executes a prepared statement with the given arguments.
	Execute(ctx context.Context, args map[string]any) ([]map[string]any, error)

	// Close closes the statement.
	Close() error

	// IsMutative returns true if the statement is mutative.
	IsMutative() bool
}

// User is an interface that can be implemented by a type to be used as a user identifier
type User interface {
	// Bytes returns a byte representation of the user identifier
	// This should follow Kwil's caller ID format
	Bytes() []byte
	// PublicKey returns the public key bytes of the user identifier
	PubKey() []byte
	// Address returns the address of the user
	Address() string
}

// noCaller is a User that is used when no user is identified
type noCaller struct{}

func (noCaller) Bytes() []byte {
	return nil
}

func (noCaller) PubKey() []byte {
	return nil
}

func (noCaller) Address() string {
	return ""
}
