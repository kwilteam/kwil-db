package common

import (
	"context"
	"strings"

	"github.com/kwilteam/kwil-db/common/sql"
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
	// Identity is the node/validator identity (pubkey).
	Identity []byte
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

// TxContext is contextual information provided to a transaction execution Route
// handler. This is defined in common as it is used by both the internal txapp
// router and extension implementations in extensions/consensus.
type TxContext struct {
	Ctx context.Context
	// BlockContext is the context of the current block.
	BlockContext *BlockContext
	// TxID is the ID of the current transaction.
	TxID []byte
}

type Engine interface {
	SchemaGetter
	// CreateDataset deploys a new dataset from a schema.
	// The dataset will be owned by the caller.
	CreateDataset(ctx context.Context, tx sql.DB, schema *types.Schema, txdata *TransactionData) error
	// DeleteDataset deletes a dataset.
	// The caller must be the owner of the dataset.
	DeleteDataset(ctx context.Context, tx sql.DB, dbid string, txdata *TransactionData) error
	// Procedure executes a procedure in a dataset. It can be given
	// either a readwrite or readonly database transaction. If it is
	// given a read-only transaction, it will not be able to execute
	// any procedures that are not `view`.
	Procedure(ctx context.Context, tx sql.DB, options *ExecutionData) (*sql.ResultSet, error)
	// ListDatasets returns a list of all datasets on the network.
	ListDatasets(caller []byte) ([]*types.DatasetIdentifier, error)
	// Execute executes a SQL statement on a dataset.
	// It uses Kwil's SQL dialect.
	Execute(ctx context.Context, tx sql.DB, dbid, query string, values map[string]any) (*sql.ResultSet, error)
	// Reload reloads the engine with the latest db state
	Reload(ctx context.Context, tx sql.Executor) error
}

type SchemaGetter interface {
	// GetSchema returns the schema of a dataset.
	// It will return an error if the dataset does not exist.
	GetSchema(dbid string) (*types.Schema, error)
}

// TransactionData holds contextual data about the transaction that
// called the procedure.
type TransactionData struct {
	// Signer is the address of public key that signed the incoming
	// transaction.
	Signer []byte

	// Caller is a string identifier for the signer.
	// It is derived from the signer's registered authenticator.
	// It is injected as a variable for usage in the query, under
	// the variable name "@caller".
	Caller string

	// TxID is the transaction ID of the incoming transaction.
	TxID string

	// Height is the block height of the incoming transaction.
	Height int64
}

// ExecutionOptions is contextual data that is passed to a procedure
// during call / execution. It is scoped to the lifetime of a single
// execution.
type ExecutionData struct {
	TransactionData
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
}

func (e *ExecutionData) Clean() error {
	e.Procedure = strings.ToLower(e.Procedure)
	return nil
}

// NetworkParameters are network level configurations that can be
// evolved over the lifetime of a network.
type NetworkParameters struct {
	// MaxBlockSize is the maximum size of a block in bytes.
	MaxBlockSize int64
	// JoinExpiry is the number of blocks after which the validators
	// join request expires if not approved.
	JoinExpiry int64
	// VoteExpiry is the default number of blocks after which the validators
	// vote expires if not approved.
	VoteExpiry int64
	// DisabledGasCosts dictates whether gas costs are disabled.
	DisabledGasCosts bool
	// InMigration is true if the network is being migrated to a new network.
	// Once this is set to true, it can never be set to false. If true,
	// new databases cannot be created, old databases cannot be deleted,
	// balances cannot be transferred
	// and the vote store is paused.
	InMigration bool
	// MaxVotesPerTx is the maximum number of votes that can be included in a
	// single transaction.
	MaxVotesPerTx int64
}
