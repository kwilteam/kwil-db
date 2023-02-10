package kclient

import (
	"context"
	txs "kwil/pkg/accounts"
	"kwil/pkg/contracts/escrow/types"
	"kwil/pkg/databases"
	"kwil/pkg/databases/spec"
	gclient "kwil/pkg/grpc/client"
	"math/big"
)

// @yaiba TODO: delcare KClient input and output types here?

type KClient interface {
	DepositFund(ctx context.Context, amount *big.Int) (*types.DepositResponse, error)
	GetDatabaseSchema(ctx context.Context, dbName string) (*databases.Database[[]byte], error)
	DeployDatabase(ctx context.Context, db *databases.Database[[]byte]) (*txs.Response, error)
	DropDatabase(ctx context.Context, dbName string) (*txs.Response, error)
	ExecuteDatabase(ctx context.Context, dbName string, queryName string, queryInputs []*spec.KwilAny) (*txs.Response, error)
	GetServiceConfig(ctx context.Context) (gclient.SvcConfig, error)
}
