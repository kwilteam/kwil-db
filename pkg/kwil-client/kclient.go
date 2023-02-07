package kwil_client

import (
	"context"
	"kwil/pkg/types/contracts/escrow"
	"kwil/pkg/types/data_types/any_type"
	"kwil/pkg/types/databases"
	"kwil/pkg/types/transactions"
	"math/big"
)

type KClient interface {
	DepositFund(ctx context.Context, to string, amount *big.Int) (*escrow.DepositResponse, error)
	GetDatabaseSchema(ctx context.Context, owner string, dbName string) (*databases.Database[[]byte], error)
	DeployDatabase(ctx context.Context, db *databases.Database[[]byte]) (*transactions.Response, error)
	DropDatabase(ctx context.Context, dbName string) (*transactions.Response, error)
	ExecuteDatabase(ctx context.Context, dbName string, queryName string, queryInputs []anytype.KwilAny) (*transactions.Response, error)
}
