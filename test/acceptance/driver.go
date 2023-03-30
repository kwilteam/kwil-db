package acceptance

import (
	"context"
	"kwil/pkg/databases"
	grpc "kwil/pkg/grpc/client/v1"
	"math/big"
)

type KwilACTDriver interface {
	DepositFund(ctx context.Context, amount *big.Int) error
	GetDepositBalance(ctx context.Context) (*big.Int, error)
	ApproveToken(ctx context.Context, spender string, amount *big.Int) error
	GetAllowance(ctx context.Context, from string, spender string) (*big.Int, error)
	DeployDatabase(ctx context.Context, db *databases.Database[[]byte]) error
	DatabaseShouldExists(ctx context.Context, owner string, dbName string) error
	ExecuteQuery(ctx context.Context, dbName string, queryName string, queryInputs []string) error
	DropDatabase(ctx context.Context, dbName string) error
	GetServiceConfig(ctx context.Context) (grpc.SvcConfig, error)
	GetUserAddress() string
	QueryDatabase(ctx context.Context, query string) ([]byte, error)
}
