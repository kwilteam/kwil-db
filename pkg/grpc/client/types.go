package client

import (
	"context"
	"kwil/pkg/accounts"
	"kwil/pkg/databases"
	"kwil/pkg/databases/executables"
)

type SvcConfig struct {
	Funding SvcFundingConfig
	Gateway SvcGatewayConfig
}

type SvcFundingConfig struct {
	ChainCode       int64
	PoolAddress     string
	ProviderAddress string
	RpcUrl          string
}

type SvcGatewayConfig struct {
	GraphqlUrl string
}

type GrpcClient interface {
	ListDatabases(ctx context.Context, address string) ([]string, error)
	GetQueries(ctx context.Context, id string) ([]*executables.QuerySignature, error)
	GetSchema(ctx context.Context, id string) (*databases.Database[[]byte], error)
	EstimateCost(ctx context.Context, tx *accounts.Transaction) (string, error)
	Broadcast(ctx context.Context, tx *accounts.Transaction) (*accounts.Response, error)
	Ping(ctx context.Context) (string, error)
	GetAccount(ctx context.Context, address string) (*accounts.Account, error)
	Close() error
	GetServiceConfig(ctx context.Context) (SvcConfig, error)
	GetFundingServiceConfig(ctx context.Context) (SvcFundingConfig, error)
}
