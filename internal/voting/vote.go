package voting

import (
	"context"
	"fmt"
	"math/big"
	"sort"

	"github.com/kwilteam/kwil-db/core/log"
	"github.com/kwilteam/kwil-db/internal/accounts"
	"github.com/kwilteam/kwil-db/internal/sql"
)

var registeredPayloads = make(map[string]ResolutionPayload)

func RegisterPayload(payload ResolutionPayload) error {
	_, ok := registeredPayloads[payload.Type()]
	if ok {
		return fmt.Errorf("payload %s already registered", payload.Type())
	}

	registeredPayloads[payload.Type()] = payload
	return nil
}

// ResolutionPayload is a payload that can be used as the body of a resolution
type ResolutionPayload interface {
	// Type returns the type of the payload.
	// This should be constant for a given payload implementation.
	Type() string
	// UnmarshalBinary unmarshals the payload from binary data.
	UnmarshalBinary(data []byte) error

	// MarshalBinary marshals the payload into binary data.
	MarshalBinary() ([]byte, error)

	// Apply is called when a resolution is approved. Voters is the list of all voters voted for the resolution, including the proposer.
	// Ensure that all changes to the datastores should be deterministic, else it will lead to consensus failures.
	Apply(ctx context.Context, datastores Datastores, proposer []byte, voters []Voter, logger log.Logger) error
}

type Voter struct {
	PubKey []byte
	Power  int64
}

// Datastores provides implementers of ResolutionPayload with access
// to different datastore interfaces
type Datastores struct {
	Accounts  AccountStore
	Databases Datasets
}

type AccountStore interface {
	// Account gets an account by its identifier
	GetAccount(ctx context.Context, identifier []byte) (*accounts.Account, error)

	// Credit credits an account with a given amount
	Credit(ctx context.Context, account []byte, amount *big.Int) error
}

type Datasets interface {
	// Execute executes a statement with the given arguments.
	// Execute(ctx context.Context, dbid string, stmt string, args ...any) (*sql.ResultSet, error)
	// NOT USE by VoteProcessor for now.

	// Query executes a query with the given arguments.
	// It will not read uncommitted data.
	Query(ctx context.Context, dbid string, query string, args ...any) (*sql.ResultSet, error)
}

func OrderedListOfVoters(voters map[string]int64) []string {
	var orderedVoters []string
	for voter := range voters {
		orderedVoters = append(orderedVoters, voter)
	}
	sort.Strings(orderedVoters)
	return orderedVoters
}
