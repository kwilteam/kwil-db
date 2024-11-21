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
)

// Client defines methods are used to talk to a Kwil provider.
type Client interface {
	// CallAction. Deprecated: Use Call instead.
	CallAction(ctx context.Context, dbid string, action string, inputs []any) (*Records, error)
	Call(ctx context.Context, dbid string, procedure string, inputs []any) (*CallResult, error)
	ChainID() string
	ChainInfo(ctx context.Context) (*types.ChainInfo, error)
	// DeployDatabase(ctx context.Context, payload *types.Schema, opts ...TxOpt) (types.Hash, error)
	// DropDatabase(ctx context.Context, name string, opts ...TxOpt) (types.Hash, error)
	// DropDatabaseID(ctx context.Context, dbid string, opts ...TxOpt) (types.Hash, error)
	// DEPRECATED: Use Execute instead.
	// ExecuteAction(ctx context.Context, dbid string, action string, tuples [][]any, opts ...TxOpt) (types.Hash, error)
	Execute(ctx context.Context, dbid string, action string, tuples [][]any, opts ...TxOpt) (types.Hash, error)
	GetAccount(ctx context.Context, pubKey []byte, status types.AccountStatus) (*types.Account, error)
	GetSchema(ctx context.Context, dbid string) (*types.Schema, error)
	ListDatabases(ctx context.Context, owner []byte) ([]*types.DatasetIdentifier, error)
	Ping(ctx context.Context) (string, error)
	Query(ctx context.Context, dbid string, query string) (*Records, error)
	TxQuery(ctx context.Context, txHash types.Hash) (*types.TcTxQueryResponse, error)
	WaitTx(ctx context.Context, txHash types.Hash, interval time.Duration) (*types.TcTxQueryResponse, error)
	Transfer(ctx context.Context, to []byte, amount *big.Int, opts ...TxOpt) (types.Hash, error)
}

// CallResult is the result of a call to a procedure.
type CallResult struct {
	Records *Records `json:"records"`
	Logs    []string `json:"logs,omitempty"`
}
