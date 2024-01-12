package voting

import (
	"context"
	"fmt"
	"math/big"

	"github.com/kwilteam/kwil-db/core/log"
	"github.com/kwilteam/kwil-db/internal/accounts"
	"github.com/kwilteam/kwil-db/internal/sql"
)

var registeredPayloads = make(map[string]ResolutionPayload)

func RegisterPaylod(payload ResolutionPayload) error {
	_, ok := registeredPayloads[payload.Type()]
	if ok {
		return fmt.Errorf("payload %s already registered", payload.Type())
	}

	registeredPayloads[payload.Type()] = payload
	return nil
}

// A ResolutionPayload is a payload that can be used as the body of a resolution
type ResolutionPayload interface {
	// Type returns the type of the payload.
	// This should be constant for a given payload implementation.
	Type() string
	// UnmarshalBinary unmarshals the payload from binary data.
	UnmarshalBinary(data []byte) error
	// Apply is called when a resolution is approved.
	Apply(ctx context.Context, datastores *Datastores, logger log.Logger) error
}

// Datastores provides implementers of ResolutionPayload with access
// to different datastore interfaces
type Datastores struct {
	Accounts  AccountStore
	Databases Datastore
}

type AccountStore interface {
	// Account gets an account by its identifier
	GetAccount(ctx context.Context, identifier []byte) (*accounts.Account, error)

	// Credit credits an account with a given amount
	Credit(ctx context.Context, account []byte, amount *big.Int) error
}

type Datastore interface {
	// Execute executes a statement with the given arguments.
	Execute(ctx context.Context, dbid string, stmt string, args map[string]any) (*sql.ResultSet, error)
	// Query executes a query with the given arguments.
	// It will not read uncommitted data.
	Query(ctx context.Context, dbid string, query string, args map[string]any) (*sql.ResultSet, error)
}
