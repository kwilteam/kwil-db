package client

import (
	"context"
	"math/big"
	"time"

	"github.com/kwilteam/kwil-db/core/types"
	"github.com/kwilteam/kwil-db/core/types/transactions"
)

// Client defines methods are used to talk to a Kwil provider
type Client interface {
	CallAction(ctx context.Context, dbid string, action string, inputs []any) (*Records, error)
	ChainID() string
	ChainInfo(ctx context.Context) (*types.ChainInfo, error)
	DeployDatabase(ctx context.Context, payload *transactions.Schema, opts ...TxOpt) (transactions.TxHash, error)
	DropDatabase(ctx context.Context, name string, opts ...TxOpt) (transactions.TxHash, error)
	DropDatabaseID(ctx context.Context, dbid string, opts ...TxOpt) (transactions.TxHash, error)
	ExecuteAction(ctx context.Context, dbid string, action string, tuples [][]any, opts ...TxOpt) (transactions.TxHash, error)
	GetAccount(ctx context.Context, pubKey []byte, status types.AccountStatus) (*types.Account, error)
	GetSchema(ctx context.Context, dbid string) (*transactions.Schema, error)
	ListDatabases(ctx context.Context, owner []byte) ([]*types.DatasetIdentifier, error)
	Ping(ctx context.Context) (string, error)
	Query(ctx context.Context, dbid string, query string) (*Records, error)
	TxQuery(ctx context.Context, txHash []byte) (*transactions.TcTxQueryResponse, error)
	WaitTx(ctx context.Context, txHash []byte, interval time.Duration) (*transactions.TcTxQueryResponse, error)
	Transfer(ctx context.Context, to []byte, amount *big.Int, opts ...TxOpt) (transactions.TxHash, error)
}
