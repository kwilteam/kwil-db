package voting

import (
	"context"
	"fmt"
	"math/big"

	"github.com/kwilteam/kwil-db/core/types"
	"github.com/kwilteam/kwil-db/internal/sql"
)

var registeredPayloads = make(map[string]ResolutionPayload)

func RegisterPaylod(name string, payload ResolutionPayload) error {
	_, ok := registeredPayloads[name]
	if ok {
		return fmt.Errorf("payload %s already registered", name)
	}

	registeredPayloads[name] = payload
	return nil
}

// A ResolutionPayload is a payload that can be used as the body of a resolution
type ResolutionPayload interface {
	MarshalBinary() ([]byte, error)
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
	Account(ctx context.Context, identifier []byte) (*types.Account, error)

	// Credit credits an account with a given amount
	Credit(ctx context.Context, account []byte, amount *big.Int) error

	// Debit debits an account with a given amount
	Debit(ctx context.Context, account []byte, amount *big.Int) error

	// Transfer transfers an amount from one account to another
	Transfer(ctx context.Context, from []byte, to []byte, amount *big.Int) error
}

type Datastore interface {
	// Execute executes a statement with the given arguments.
	Execute(ctx context.Context, stmt string, args map[string]any) (*sql.ResultSet, error)
	// Query executes a query with the given arguments.
	// It will not read uncommitted data.
	Query(ctx context.Context, query string, args map[string]any) (*sql.ResultSet, error)
	// Savepoint creates a savepoint.
	Savepoint() (sql.Savepoint, error)
}
