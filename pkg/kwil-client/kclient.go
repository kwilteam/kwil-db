package kwil_client

import (
	"context"
	"kwil/pkg/grpc/client"
	"kwil/x/types/contracts/escrow"
	"kwil/x/types/data_types/any_type"
	"kwil/x/types/databases"
	"kwil/x/types/transactions"
	"math/big"
)

type KClient interface {
	DepositFund(ctx context.Context, to string, amount *big.Int) (*escrow.DepositResponse, error)
	GetDatabaseSchema(ctx context.Context, owner string, dbName string) (*databases.Database[[]byte], error)
	DeployDatabase(ctx context.Context, db *databases.Database[[]byte]) (*transactions.Response, error)
	DropDatabase(ctx context.Context, dbName string) (*transactions.Response, error)
	ExecuteDatabase(ctx context.Context, dbName string, queryName string, queryInputs []anytype.KwilAny) (*transactions.Response, error)
	GetNodeInfo(ctx context.Context) (client.NodeInfo, error)
}
