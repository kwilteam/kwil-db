// Package client contains the shared client types, including the options used
// to construct a Client instance, and the records iterator used to represent
// the results of an action call. This package also defines the Client interface
// that should be satisfied by different implementations, such as a gateway
// client.
package client

import (
	"context"
	"math/big"
	"time"

	"github.com/kwilteam/kwil-db/core/types"
	"github.com/kwilteam/kwil-db/core/types/transactions"
)

// Client defines methods are used to talk to a Kwil provider.
type Client interface {
	// CallAction. Deprecated: Use Call instead.
	CallAction(ctx context.Context, dbid string, action string, inputs []any) (*Records, error)
	Call(ctx context.Context, dbid string, procedure string, inputs []any) (*Records, error)
	ChainID() string
	ChainInfo(ctx context.Context) (*types.ChainInfo, error)
	DeployDatabase(ctx context.Context, payload *types.Schema, opts ...TxOpt) (transactions.TxHash, error)
	DropDatabase(ctx context.Context, name string, opts ...TxOpt) (transactions.TxHash, error)
	DropDatabaseID(ctx context.Context, dbid string, opts ...TxOpt) (transactions.TxHash, error)
	// DEPRECATED: Use Execute instead.
	ExecuteAction(ctx context.Context, dbid string, action string, tuples [][]any, opts ...TxOpt) (transactions.TxHash, error)
	Execute(ctx context.Context, dbid string, action string, tuples [][]any, opts ...TxOpt) (transactions.TxHash, error)
	GetAccount(ctx context.Context, pubKey []byte, status types.AccountStatus) (*types.Account, error)
	GetSchema(ctx context.Context, dbid string) (*types.Schema, error)
	ListDatabases(ctx context.Context, owner []byte) ([]*types.DatasetIdentifier, error)
	Ping(ctx context.Context) (string, error)
	Query(ctx context.Context, dbid string, query string) (*Records, error)
	TxQuery(ctx context.Context, txHash []byte) (*transactions.TcTxQueryResponse, error)
	WaitTx(ctx context.Context, txHash []byte, interval time.Duration) (*transactions.TcTxQueryResponse, error)
	Transfer(ctx context.Context, to []byte, amount *big.Int, opts ...TxOpt) (transactions.TxHash, error)
}
