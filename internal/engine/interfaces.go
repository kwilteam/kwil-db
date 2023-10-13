package engine

import (
	"context"
	"io"

	"github.com/kwilteam/kwil-db/internal/engine/dataset"
	"github.com/kwilteam/kwil-db/internal/engine/types"
	"github.com/kwilteam/kwil-db/internal/sql"
)

type Dataset interface {
	ApplyChangeset(changeset io.Reader) error
	Call(ctx context.Context, action string, args []any, opts *dataset.TxOpts) ([]map[string]any, error)
	Close() error
	CreateSession() (sql.Session, error)
	DBID() string
	Delete() error
	Execute(ctx context.Context, action string, args [][]any, opts *dataset.TxOpts) ([]map[string]any, error)
	ListExtensions(ctx context.Context) ([]*types.Extension, error)
	ListProcedures() []*types.Procedure
	ListTables(ctx context.Context) ([]*types.Table, error)
	Metadata() (name string, owner dataset.User)
	Query(ctx context.Context, stmt string, args map[string]any) ([]map[string]any, error)
	Savepoint() (sql.Savepoint, error)
}

type MasterDB interface {
	Close() error
	ListDatasets(ctx context.Context) ([]*types.DatasetInfo, error)
	// ListDatasetsByOwner lists all datasets owned by the given owner.
	// It identifies the owner by public key.
	ListDatasetsByOwner(ctx context.Context, owner []byte) ([]string, error)
	// RegisterDataset registers a dataset to the master database
	// It tracks the desired address of the deployer, which can be used later
	RegisterDataset(ctx context.Context, name string, owner *types.User) error
	UnregisterDataset(ctx context.Context, dbid string) error
}

// CommitRegister is an interface for registering atomically committable data stores
// Any database registered to this will be atomically synced in a 2pc transaction
type CommitRegister interface {
	// Register registers a database to the commit register
	Register(ctx context.Context, name string, db sql.Database) error
	// Unregister unregisters a database from the commit register
	Unregister(ctx context.Context, name string) error
}

func (e *Engine) newDatasetUser(u *types.User) (*datasetUser, error) {
	bts, err := u.MarshalBinary()
	if err != nil {
		return nil, err
	}

	addr, err := e.addresser(u.AuthType, u.PublicKey)
	if err != nil {
		return nil, err
	}

	return &datasetUser{
		pubkeyBts:     u.PublicKey,
		marshalledBts: bts,
		address:       addr,
	}, nil
}

// implements datasets.User
type datasetUser struct {
	pubkeyBts []byte
	// marshalledBts are the marshalled bytes of the user
	marshalledBts []byte
	address       string
}

func (u *datasetUser) Bytes() []byte {
	return u.marshalledBts
}

func (u *datasetUser) PubKey() []byte {
	return u.pubkeyBts
}

func (u *datasetUser) Address() string {
	return u.address
}

// Addresser is a function that takes an address type and a public key and returns an address
type Addresser func(addressType string, pubkey []byte) (string, error)
