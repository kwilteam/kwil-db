// package user defines the interface for a user client transport.
// the user client is the main service for end users to interact with a Kwil network.
package user

import (
	"context"
	"math/big"

	"github.com/kwilteam/kwil-db/core/rpc/client"
	"github.com/kwilteam/kwil-db/core/types"
	"github.com/kwilteam/kwil-db/core/types/transactions"
)

// TxSvcClient is the interface for a txsvc client.
// The txsvc is the main service for end users to interact with a Kwil network.
type TxSvcClient interface {
	Broadcast(ctx context.Context, tx *transactions.Transaction, sync client.BroadcastWait) ([]byte, error)
	Call(ctx context.Context, msg *transactions.CallMessage, opts ...client.ActionCallOption) ([]map[string]any, error)
	ChainInfo(ctx context.Context) (*types.ChainInfo, error)
	EstimateCost(ctx context.Context, tx *transactions.Transaction) (*big.Int, error)
	GetAccount(ctx context.Context, pubKey []byte, status types.AccountStatus) (*types.Account, error)
	GetSchema(ctx context.Context, dbid string) (*types.Schema, error)
	ListDatabases(ctx context.Context, ownerPubKey []byte) ([]*types.DatasetIdentifier, error)
	Ping(ctx context.Context) (string, error)
	Query(ctx context.Context, dbid string, query string) ([]map[string]any, error)
	TxQuery(ctx context.Context, txHash []byte) (*transactions.TcTxQueryResponse, error)
}
