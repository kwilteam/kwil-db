package kclient

import (
	"context"
	"kwil/pkg/contracts/escrow/types"
	"kwil/pkg/databases"
	gclient "kwil/pkg/grpc/client"
	"kwil/pkg/types/data_types/any_type"
	txs "kwil/pkg/types/transactions"
	"math/big"
)

// @yaiba TODO: delcare KClient input and output types here?

type KClient interface {
	DepositFund(ctx context.Context, amount *big.Int) (*types.DepositResponse, error)
	GetDatabaseSchema(ctx context.Context, dbName string) (*databases.Database[[]byte], error)
	DeployDatabase(ctx context.Context, db *databases.Database[[]byte]) (*txs.Response, error)
	DropDatabase(ctx context.Context, dbName string) (*txs.Response, error)
	ExecuteDatabase(ctx context.Context, dbName string, queryName string, queryInputs []anytype.KwilAny) (*txs.Response, error)
	GetServiceConfig(ctx context.Context) (gclient.SvcConfig, error)
}
