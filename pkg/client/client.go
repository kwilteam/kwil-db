package client

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"kwil/pkg/accounts"
	cc "kwil/pkg/chain/client"
	escrowContracts "kwil/pkg/chain/contracts/escrow"
	tokenContracts "kwil/pkg/chain/contracts/token"
	"kwil/pkg/databases"
	"kwil/pkg/databases/spec"
	grpc "kwil/pkg/grpc/client"
)

const (
	// DefaultProviderAddress is the default provider address for the kwil client
	DefaultProviderAddress = "0x000"

	// DefaultEscrowAddress is the default pool address for the kwil client
	DefaultEscrowAddress = "0x000"

	// DefaultChainCode is the default chain code for the kwil client
	// Using Goerli testnet for now
	DefaultChainCode = 2
)

type client struct {
	endpoint              string
	grpc                  *grpc.Client
	chainClient           cc.ChainClient
	dbis                  map[string]dbi // maps the db name to its queries
	usingServiceCfg       bool
	chainRpcUrl           *string
	providerAddress       string
	escrowContractAddress string
	chainCode             int64
}

type KwilClient interface {
	GetSchema(ctx context.Context, owner, name string) (*databases.Database[*spec.KwilAny], error)
	GetSchemaById(ctx context.Context, id string) (*databases.Database[*spec.KwilAny], error)

	DeployDatabase(ctx context.Context, db *databases.Database[[]byte], privateKey *ecdsa.PrivateKey) (*accounts.Response, error)
	DropDatabase(ctx context.Context, dbName string, privateKey *ecdsa.PrivateKey) (*accounts.Response, error)

	ExecuteDatabase(ctx context.Context, dbOwner, dbName string, queryName string, queryInputs map[string]*spec.KwilAny, privateKey *ecdsa.PrivateKey) (*accounts.Response, error)
	ExecuteDatabaseById(ctx context.Context, id string, queryName string, queryInputs map[string]*spec.KwilAny, privateKey *ecdsa.PrivateKey) (*accounts.Response, error)

	GetServiceConfig(ctx context.Context) (grpc.SvcConfig, error)
	// SetChainRpcUrl sets the chain rpc url for the kwil client
	SetChainRpcUrl(url string)

	EscrowContract(ctx context.Context) (escrowContracts.EscrowContract, error)
	TokenContract(ctx context.Context, address string) (tokenContracts.TokenContract, error)
}

func New(ctx context.Context, rpcUrl string, opts ...ClientOption) (KwilClient, error) {
	/*
		c := &Client{
			endpoint: rpcUrl,
		}
		for _, opt := range opts {
			opt(c)
		}

		grpcClient, err := grpc.New(ctx, &grpc.Config{
			Addr: rpcUrl,
		})*/
	c := &client{
		endpoint:              rpcUrl,
		dbis:                  make(map[string]dbi),
		usingServiceCfg:       true,
		chainRpcUrl:           nil,
		providerAddress:       DefaultProviderAddress,
		escrowContractAddress: DefaultEscrowAddress,
		chainCode:             DefaultChainCode,
	}
	for _, opt := range opts {
		opt(c)
	}

	var err error
	c.grpc, err = grpc.New(ctx, &grpc.Config{
		Addr: rpcUrl,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create grpc client: %w", err)
	}

	if !c.usingServiceCfg {
		return c, nil
	}

	// apply service config
	cfg, err := c.grpc.GetServiceConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get service config from kwil provider: %w", err)
	}

	c.chainCode = cfg.Funding.ChainCode
	c.providerAddress = cfg.Funding.ProviderAddress
	c.escrowContractAddress = cfg.Funding.PoolAddress

	// reapply opts since service config may have changed them if they were specified
	for _, opt := range opts {
		opt(c)
	}

	return c, nil
}

func (c *client) GetServiceConfig(ctx context.Context) (grpc.SvcConfig, error) {
	return c.grpc.GetServiceConfig(ctx)
}

func (c *client) SetChainRpcUrl(url string) {
	c.chainRpcUrl = &url
}
