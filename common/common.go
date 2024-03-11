package common

import (
	"context"

	sql "github.com/kwilteam/kwil-db/common/sql"
	"github.com/kwilteam/kwil-db/core/log"
	"github.com/kwilteam/kwil-db/core/types"
)

// Service provides access to general application information to
// extensions.
type Service struct {
	// Logger is a logger for the application
	Logger log.SugaredLogger
	// ExtensionConfigs is a map of the nodes extensions and local
	// configurations.
	// It maps: extension_name -> config_key -> config_value
	ExtensionConfigs map[string]map[string]string
}

// App is an application that can modify and query the local database
// instance.
type App struct {
	// Service is the base application
	Service *Service
	// DB is a connection to the underlying Postgres database
	DB sql.DB
	// Engine is the underlying KwilDB engine, capable of storing and
	// executing against
	// Kuneiform schemas
	Engine Engine
}

type Engine interface {
	// CreateDataset deploys a new dataset from a schema.
	// The dataset will be owned by the caller.
	CreateDataset(ctx context.Context, tx sql.DB, schema *Schema, caller []byte) error
	// DeleteDataset deletes a dataset.
	// The caller must be the owner of the dataset.
	DeleteDataset(ctx context.Context, tx sql.DB, dbid string, caller []byte) error
	// Procedure executes a procedure in a dataset. It can be given
	// either a readwrite or readonly database transaction. If it is
	// given a read-only transaction, it will not be able to execute
	// any procedures that are not `view`.
	Procedure(ctx context.Context, tx sql.DB, options *ExecutionData) (*sql.ResultSet, error)
	// GetSchema returns the schema of a dataset.
	// It will return an error if the dataset does not exist.
	GetSchema(ctx context.Context, dbid string) (*Schema, error)
	// ListDatasets returns a list of all datasets on the network.
	ListDatasets(ctx context.Context, caller []byte) ([]*types.DatasetIdentifier, error)
	// Execute executes a SQL statement on a dataset.
	// It uses Kwil's SQL dialect.
	Execute(ctx context.Context, tx sql.DB, dbid, query string, values map[string]any) (*sql.ResultSet, error)
}

// ExecutionOptions is contextual data that is passed to a procedure
// during call / execution. It is scoped to the lifetime of a single
// execution.
type ExecutionData struct {
	// Dataset is the DBID of the dataset that was called.
	// Even if a procedure in another dataset is called, this will
	// always be the original dataset.
	Dataset string

	// Procedure is the original procedure that was called.
	// Even if a nested procedure is called, this will always be the
	// original procedure.
	Procedure string

	// Args are the arguments that were passed to the procedure.
	// Currently these are all string or untyped nil values.
	Args []any

	// Signer is the address of public key that signed the incoming
	// transaction.
	Signer []byte

	// Caller is a string identifier for the signer.
	// It is derived from the signer's registered authenticator.
	// It is injected as a variable for usage in the query, under
	// the variable name "@caller".
	Caller string
}

func (e *ExecutionData) Clean() error {
	return runCleans(
		cleanDBID(&e.Dataset),
		cleanIdent(&e.Procedure),
	)
}
