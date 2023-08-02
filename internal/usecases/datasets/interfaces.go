package datasets

import (
	"context"
	"io"
	"math/big"

	"github.com/kwilteam/kwil-db/internal/entity"
	"github.com/kwilteam/kwil-db/pkg/balances"
	"github.com/kwilteam/kwil-db/pkg/tx"
)

// DatasetUseCaseInterface is the interface for the dataset use case

type DatasetUseCaseInterface interface {
	Close() error
	Deploy(ctx context.Context, deployment *entity.DeployDatabase) (rec *tx.Receipt, err error)
	Drop(ctx context.Context, drop *entity.DropDatabase) (txReceipt *tx.Receipt, err error)
	Execute(ctx context.Context, action *entity.ExecuteAction) (rec *tx.Receipt, err error)
	GetAccount(ctx context.Context, address string) (*entity.Account, error)
	GetSchema(ctx context.Context, dbid string) (*entity.Schema, error)
	ListDatabases(ctx context.Context, owner string) ([]string, error)
	PriceDeploy(deployment *entity.DeployDatabase) (*big.Int, error)
	PriceDrop(drop *entity.DropDatabase) (*big.Int, error)
	PriceExecute(action *entity.ExecuteAction) (*big.Int, error)
	Query(ctx context.Context, query *entity.DBQuery) ([]byte, error)
	// Call calls a read-only action on a database
<<<<<<< HEAD
	Call(ctx context.Context, action *entity.CallAction) ([]map[string]any, error)
=======
	Spend(ctx context.Context, address string, amount string, nonce int64) error
	Call(ctx context.Context, action *entity.CallAction) ([]map[string]any, error)
	StartBlockSession() error
	EndBlockSession() ([]byte, error)
	InitalizeAppHash(appHash []byte)
>>>>>>> dc1f6266 (added ABCI to handle all mutative interactions (deploy, drop, execute).)
}

type AccountStore interface {
	GetAccount(ctx context.Context, address string) (*balances.Account, error)
	Spend(ctx context.Context, spend *balances.Spend) error
	Close() error
	Savepoint() (balances.Savepoint, error)
	CreateSession() (balances.Session, error)
	ApplyChangeset(changeset io.Reader) error
}
