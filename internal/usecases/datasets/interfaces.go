package datasets

import (
	"github.com/kwilteam/kwil-db/internal/entity"
	"github.com/kwilteam/kwil-db/pkg/balances"
	"github.com/kwilteam/kwil-db/pkg/engine/datasets"
	"github.com/kwilteam/kwil-db/pkg/engine/models"
	"github.com/kwilteam/kwil-db/pkg/tx"
	"math/big"
)

// DatasetUseCaseInterface is the interface for the dataset use case
type DatasetUseCaseInterface interface {
	// Deploy deploys a new database
	Deploy(*entity.DeployDatabase) (*tx.Receipt, error)

	//PriceDeploy returns the price to deploy a database
	PriceDeploy(*entity.DeployDatabase) (*big.Int, error)

	// Drop drops a database
	Drop(*entity.DropDatabase) (*tx.Receipt, error)

	// PriceDrop returns the price to drop a database
	PriceDrop(*entity.DropDatabase) (*big.Int, error)

	// Execute executes an action on a database
	Execute(*entity.ExecuteAction) (*tx.Receipt, error)

	// PriceExecute returns the price to execute an action on a database
	PriceExecute(*entity.ExecuteAction) (*big.Int, error)

	// Query queries a database
	Query(*entity.DBQuery) ([]byte, error)

	// GetAccount returns the account of the given address
	GetAccount(string) (*entity.Account, error)

	// ListDatabases returns a list of all databases deployed by the given address
	ListDatabases(string) ([]string, error)

	// GetSchema returns the schema of the given database
	GetSchema(string) (*models.Dataset, error)
}

type engineInterface interface {
	Close() error
	Deploy(schema *models.Dataset) error
	DropDataset(dbid string) error
	GetDeployPrice(schema *models.Dataset) (*big.Int, error)
	GetDropPrice(dbid string) (*big.Int, error)
	ListDatabases(owner string) ([]string, error)
	GetDataset(dbid string) (*datasets.Dataset, error)
}

type accountStore interface {
	GetAccount(address string) (*balances.Account, error)
	Spend(spend *balances.Spend) error
	BatchCredit(creditList []*balances.Credit, chain *balances.ChainConfig) error
	Close() error
}
