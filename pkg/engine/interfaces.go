package engine

import (
	"context"
	"io"

	"github.com/kwilteam/kwil-db/pkg/engine/dataset"
	"github.com/kwilteam/kwil-db/pkg/engine/types"
	"github.com/kwilteam/kwil-db/pkg/sql"
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
	RegisterDataset(ctx context.Context, name string, owner types.UserIdentifier) error
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

func newDatasetUser(u types.UserIdentifier) (*datasetUser, error) {
	bts, err := u.MarshalBinary()
	if err != nil {
		return nil, err
	}

	pub, err := u.PubKey()
	if err != nil {
		return nil, err
	}

	return &datasetUser{
		pubkeyBts:     pub.Bytes(),
		marshalledBts: bts,
	}, nil
}

// implements datasets.UserIdentifier
type datasetUser struct {
	pubkeyBts []byte
	// marshalledBts are the marshalled bytes of the user
	marshalledBts []byte
}

func (u *datasetUser) Bytes() []byte {
	return u.marshalledBts
}

func (u *datasetUser) PubKey() []byte {
	return u.pubkeyBts
}
