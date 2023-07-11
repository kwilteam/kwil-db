package datasets

import (
	"context"
	"math/big"

	"github.com/kwilteam/kwil-db/internal/entity"
	"github.com/kwilteam/kwil-db/pkg/balances"
	"github.com/kwilteam/kwil-db/pkg/tx"
)

// DatasetUseCaseInterface is the interface for the dataset use case
type DatasetUseCaseInterface interface {
	// Deploy deploys a new database
	Deploy(context.Context, *entity.DeployDatabase) (*tx.Receipt, error)

	//PriceDeploy returns the price to deploy a database
	PriceDeploy(*entity.DeployDatabase) (*big.Int, error)

	// Drop drops a database
	Drop(context.Context, *entity.DropDatabase) (*tx.Receipt, error)

	// PriceDrop returns the price to drop a database
	PriceDrop(*entity.DropDatabase) (*big.Int, error)

	// Execute executes an action on a database
	Execute(context.Context, *entity.ExecuteAction) (*tx.Receipt, error)

	// PriceExecute returns the price to execute an action on a database
	PriceExecute(*entity.ExecuteAction) (*big.Int, error)

	// Query queries a database
	Query(context.Context, *entity.DBQuery) ([]byte, error)

	// GetAccount returns the account of the given address
	GetAccount(string) (*entity.Account, error)

	// ListDatabases returns a list of all databases deployed by the given address
	ListDatabases(context.Context, string) ([]string, error)

	// GetSchema returns the schema of the given database
	GetSchema(context.Context, string) (*entity.Schema, error)

	// UpdateGasCosts updates the gas costs of the use case
	UpdateGasCosts(bool)

	// GasEnabled Checks if gas costs are enabled
	GasEnabled() bool
}

type AccountStore interface {
	GetAccount(address string) (*balances.Account, error)
	Spend(spend *balances.Spend) error
	BatchCredit(creditList []*balances.Credit, chain *balances.ChainConfig) error
	Close() error
}
