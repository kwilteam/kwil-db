package types

import (
	"context"
	"math/big"

	"github.com/kwilteam/kwil-db/pkg/balances"
	"github.com/kwilteam/kwil-db/pkg/transactions"
	"github.com/kwilteam/kwil-db/pkg/validators"
)

// TransportClient abstracts the communication with a kwil-db node, either via
// gRPC or HTTP.
type TransportClient interface {
	Close() error
	Call(ctx context.Context, req *transactions.SignedMessage) ([]map[string]any, error)
	TxQuery(ctx context.Context, txHash []byte) (*TcTxQueryResponse, error)
	GetTarget() string
	GetSchema(ctx context.Context, dbid string) (*transactions.Schema, error)
	Query(ctx context.Context, dbid string, query string) ([]map[string]any, error)
	ListDatabases(ctx context.Context, ownerPubKey []byte) ([]string, error)
	GetAccount(ctx context.Context, pubKey []byte) (*balances.Account, error)
	Broadcast(ctx context.Context, tx *transactions.Transaction) ([]byte, error)
	Ping(ctx context.Context) (string, error)
	EstimateCost(ctx context.Context, tx *transactions.Transaction) (*big.Int, error)
	ValidatorJoinStatus(ctx context.Context, pubKey []byte) (*validators.JoinRequest, error)
	CurrentValidators(ctx context.Context) ([]*validators.Validator, error)
}

// Should we define types TransportClient need?

// TcTxQueryResponse is the response type for the TxQuery method of the transport client.
type TcTxQueryResponse struct {
	Hash     []byte
	Height   int64
	Tx       transactions.Transaction
	TxResult transactions.TransactionResult
}
