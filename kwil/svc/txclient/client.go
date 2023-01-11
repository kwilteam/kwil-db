package txclient

import (
	"context"
	"kwil/x/proto/txpb"
	"kwil/x/types/databases"
	"kwil/x/types/execution"
	"kwil/x/types/transactions"

	"google.golang.org/grpc"
)

type TxClient interface {
	Broadcast(ctx context.Context, tx *transactions.Transaction) (*transactions.Response, error)
	GetSchema(ctx context.Context, db *databases.DatabaseIdentifier) (*databases.Database, error)
	GetSchemaById(ctx context.Context, id string) (*databases.Database, error)
	ListDatabases(ctx context.Context, address string) ([]string, error)
	GetExecutables(ctx context.Context, db *databases.DatabaseIdentifier) ([]*execution.Executable, error)
	GetExecutablesById(ctx context.Context, id string) ([]*execution.Executable, error)
}

type client struct {
	txs txpb.TxServiceClient
}

func New(cc *grpc.ClientConn) TxClient {
	return &client{
		txs: txpb.NewTxServiceClient(cc),
	}
}
