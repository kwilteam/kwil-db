package types

import (
	"context"
	"fmt"
	"math/big"
)

var registeredPayloads = make(map[string]ResolutionPayload)

// TODO: Where should this go?
func RegisterPaylod(payload ResolutionPayload) error {
	_, ok := registeredPayloads[payload.Type()]
	if ok {
		return fmt.Errorf("payload %s already registered", payload.Type())
	}

	registeredPayloads[payload.Type()] = payload
	return nil
}

func RegisteredPayloads() map[string]ResolutionPayload {
	return registeredPayloads
}

// A ResolutionPayload is a payload that can be used as the body of a resolution
type ResolutionPayload interface {
	// Type returns the type of the payload.
	// This should be constant for a given payload implementation.
	Type() string
	// UnmarshalBinary unmarshals the payload from binary data.
	UnmarshalBinary(data []byte) error
	// Apply is called when a resolution is approved.
	Apply(ctx context.Context, datastores *Datastores) error
}

// Datastores provides implementers of ResolutionPayload with access
// to different datastore interfaces
type Datastores struct {
	Accounts  AccountStore
	Databases Datastore
}

type AccountStore interface {
	// Account gets an account by its identifier
	// GetAccount(ctx context.Context, identifier []byte) (*Account, error)

	AccountInfo(ctx context.Context, account []byte) (balance *big.Int, nonce int64, err error)

	// Credit credits an account with a given amount
	Credit(ctx context.Context, account []byte, amount *big.Int) error
}

type Datastore interface {
	// Execute executes a statement with the given arguments.
	Execute(ctx context.Context, dbid string, stmt string, args map[string]any) (ResultSet, error)
	// Query executes a query with the given arguments.
	// It will not read uncommitted data.
	Query(ctx context.Context, dbid string, query string, args map[string]any) (ResultSet, error)
}

type Savepoint interface {
	Rollback() error
	Commit() error
}

type ResultSet interface {
	// Columns returns the columns that are returned by the result.
	Columns() []string

	Rows() [][]any

	// Next returns whether or not there is another row to be read.
	// If there is, it will increment the row index, which can be used
	// to get the values for the current row.
	Next() (rowReturned bool)
	// Err returns the error that occurred during the query.
	Err() error
	// Values returns the values for the current row.
	// The values are returned in the same order as the columns.
	Values() ([]any, error)

	Map() []map[string]any
}
