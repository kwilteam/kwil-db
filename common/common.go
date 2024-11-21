package common

import (
	"context"
	"math/big"
	"strings"

	"github.com/kwilteam/kwil-db/core/log"
	"github.com/kwilteam/kwil-db/core/types"
	"github.com/kwilteam/kwil-db/node/types/sql"
)

// Service provides access to general application information to
// extensions.
type Service struct {
	// Logger is a logger for the application
	Logger log.Logger
	// ExtensionConfigs is a map of the nodes extensions and local
	// configurations.
	// It maps: extension_name -> config_key -> config_value
	//
	// DEPRECATED: Use LocalConfig.AppCfg.Extensions instead.
	ExtensionConfigs map[string]map[string]string

	// GenesisConfig is the genesis configuration of the network.
	// GenesisConfig *chain.GenesisConfig

	// LocalConfig is the local configuration of the node.
	//LocalConfig *config.KwildConfig

	// Identity is the node/validator identity (pubkey).
	Identity []byte
}

// NameLogger returns a new Service with the logger named.
// Every other field is the same pointer as the original.
func (s *Service) NamedLogger(name string) *Service {
	return &Service{
		Logger:           s.Logger.New(name),
		ExtensionConfigs: s.ExtensionConfigs,
		// GenesisConfig:    s.GenesisConfig,
		// LocalConfig:      s.LocalConfig,
		Identity: s.Identity,
	}
}

// App is an application that can modify and query the local database
// instance.
type App struct {
	// Service is the base application
	Service *Service
	// DB is a connection to the underlying Postgres database
	DB sql.DB
	// Engine is the underlying KwilDB engine, capable of storing and
	// executing against Kuneiform schemas
	Engine Engine
	// Accounts is the account manager for the application
	Accounts Accounts
	// Validators is the validator manager for the application
	Validators Validators
}

// TxContext is contextual information provided to a transaction execution Route
// handler. This is defined in common as it is used by both the internal txapp
// router and extension implementations in extensions/consensus.
type TxContext struct {
	Ctx context.Context
	// BlockContext is the context of the current block.
	BlockContext *BlockContext
	// TxID is the ID of the current transaction.
	TxID string
	// Signer is the public key of the transaction signer.
	Signer []byte
	// Caller is the string identifier of the transaction signer.
	// It is derived from the signer's registered authenticator.
	Caller string
	// Authenticator is the authenticator used to sign the transaction.
	Authenticator string
}

// Engine is an interface for the main database engine that is responsible for deploying
// and executing Kuneiform datasets.
type Engine interface {
	SchemaGetter
	// CreateDataset deploys a new dataset from a schema.
	// The dataset will be owned by the caller.
	CreateDataset(ctx *TxContext, tx sql.DB, schema *types.Schema) error
	// DeleteDataset deletes a dataset.
	// The caller must be the owner of the dataset.
	DeleteDataset(ctx *TxContext, tx sql.DB, dbid string) error
	// Procedure executes a procedure in a dataset. It can be given
	// either a readwrite or readonly database transaction. If it is
	// given a read-only transaction, it will not be able to execute
	// any procedures that are not `view`.
	Procedure(ctx *TxContext, tx sql.DB, options *ExecutionData) (*sql.ResultSet, error)
	// ListDatasets returns a list of all datasets on the network.
	ListDatasets(caller []byte) ([]*types.DatasetIdentifier, error)
	// Execute executes a SQL statement on a dataset.
	// It uses Kwil's SQL dialect.
	Execute(ctx *TxContext, tx sql.DB, dbid, query string, values map[string]any) (*sql.ResultSet, error)
	// Reload reloads the engine with the latest db state
	Reload(ctx context.Context, tx sql.Executor) error
}

// Accounts is an interface for managing accounts on the Kwil network. It
// should be used to credit, debit, and transfer funds between Kwil accounts.
type Accounts interface {
	// Credit credits an account with the given amount. If the account
	// does not exist, it will be created. A negative amount will be
	// treated as a debit. Accounts cannot have negative balances, and
	// will return an error if the amount would cause the balance to go
	// negative.
	Credit(ctx context.Context, tx sql.Executor, address []byte, balance *big.Int) error
	// Transfer transfers an amount from one account to another. If the
	// from account does not have enough funds to transfer the amount,
	// it will fail. If the to account does not exist, it will be
	// created. The amount must be greater than 0.
	Transfer(ctx context.Context, tx sql.TxMaker, from, to []byte, amt *big.Int) error
	// GetAccount retrieves the account with the given identifier. If the
	// account does not exist, it will return an account with a balance
	// of 0 and a nonce of 0.
	GetAccount(ctx context.Context, tx sql.Executor, account []byte) (*types.Account, error)
	// ApplySpend applies a spend to the account. If the account does not
	// exist, it will be created. If the account does not have enough
	// funds to spend the amount, the entire balance will be spent and
	// the spend will fail.
	ApplySpend(ctx context.Context, tx sql.Executor, account []byte, amount *big.Int, nonce int64) error
}

// Validators is an interface for managing validators on the Kwil network.
type Validators interface {
	// GetValidatorPower retrieves the power of the given validator. If
	// the validator does not exist, it will return 0.
	GetValidatorPower(ctx context.Context, tx sql.Executor, validator []byte) (int64, error)
	// GetValidators retrieves all validators.
	GetValidators() []*types.Validator
	// SetValidatorPower sets the power of a validator. If the target
	// validator does not exist, it will be created with the given power.
	// If set to 0, the target validator will be deleted, and will no
	// longer be considered a validator. It will return an error if a
	// negative power is given.
	SetValidatorPower(ctx context.Context, tx sql.Executor, validator []byte, power int64) error
}

// SchemaGetter is an interface for getting the schema of a dataset.
type SchemaGetter interface {
	// GetSchema returns the schema of a dataset.
	// It will return an error if the dataset does not exist.
	GetSchema(dbid string) (*types.Schema, error)
}

// ExecutionOptions is contextual data that is passed to a procedure
// during call / execution. It is scoped to the lifetime of a single
// execution.
type ExecutionData struct {
	//TxCtx *TxContext
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

	// MigrationStatus determines the status of the migration.
	MigrationStatus types.MigrationStatus

	// MaxVotesPerTx is the maximum number of votes that can be included in a
	// single transaction.
	MaxVotesPerTx int64
}
