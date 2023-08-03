package datasets

import (
	"context"
	"math/big"

	"github.com/kwilteam/kwil-db/internal/entity"
	"github.com/kwilteam/kwil-db/pkg/balances"
	"github.com/kwilteam/kwil-db/pkg/tx"
	gowal "github.com/tidwall/wal"
)

// DatasetUseCaseInterface is the interface for the dataset use case

type DatasetUseCaseInterface interface {
	ApplyChangesets(wal *gowal.Log) error
	BlockCommit(wal *gowal.Log, prevAppHash []byte) ([]byte, error)
	Close() error
	Deploy(ctx context.Context, deployment *entity.DeployDatabase) (rec *tx.Receipt, err error)
	Drop(ctx context.Context, drop *entity.DropDatabase) (txReceipt *tx.Receipt, err error)
	Execute(ctx context.Context, action *entity.ExecuteAction) (rec *tx.Receipt, err error)
	GenerateAppHash(prevAppHash []byte) []byte
	GetAccount(ctx context.Context, address string) (*entity.Account, error)
	GetSchema(ctx context.Context, dbid string) (*entity.Schema, error)
	ListDatabases(ctx context.Context, owner string) ([]string, error)
	PriceDeploy(deployment *entity.DeployDatabase) (*big.Int, error)
	PriceDrop(drop *entity.DropDatabase) (*big.Int, error)
	PriceExecute(action *entity.ExecuteAction) (*big.Int, error)
	Query(ctx context.Context, query *entity.DBQuery) ([]byte, error)
	UpdateBlockHeight(height int64)
	// Call calls a read-only action on a database
	Spend(ctx context.Context, address string, amount string, nonce int64) error
	Call(ctx context.Context, action *entity.CallAction) ([]map[string]any, error)
}

type AccountStore interface {
	GetAccount(ctx context.Context, address string) (*balances.Account, error)
	Spend(ctx context.Context, spend *balances.Spend) error
	Close() error
}
