package client

import (
	"context"
	"kwil/pkg/databases"
	"kwil/pkg/types/accounts"
	"kwil/pkg/types/execution"
	"kwil/pkg/types/transactions"
)

type SvcConfig struct {
	Funding SvcFundingConfig
}

type SvcFundingConfig struct {
	ChainCode        int64
	PoolAddress      string
	ValidatorAccount string
}

type GrpcClient interface {
	ListDatabases(ctx context.Context, address string) ([]string, error)
	GetExecutablesById(ctx context.Context, id string) ([]*execution.Executable, error)
	GetSchema(ctx context.Context, owner string, dbName string) (*databases.Database[[]byte], error)
	GetSchemaById(ctx context.Context, id string) (*databases.Database[[]byte], error)
	EstimateCost(ctx context.Context, tx *transactions.Transaction) (string, error)
	Broadcast(ctx context.Context, tx *transactions.Transaction) (*transactions.Response, error)
	Ping(ctx context.Context) (string, error)
	GetAccount(ctx context.Context, address string) (accounts.Account, error)
	GetServiceConfig(ctx context.Context) (SvcConfig, error)
	GetFundingServiceConfig(ctx context.Context) (SvcFundingConfig, error)
	Close() error
}
