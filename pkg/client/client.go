package client

import (
	"context"
	"fmt"
	cc "kwil/pkg/chain/client"
	chainTypes "kwil/pkg/chain/types"
	grpc "kwil/pkg/grpc/client"
	"strings"
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

type KwilClient struct {
	grpc                  *grpc.Client
	chainClient           cc.ChainClient
	dbis                  map[string]dbi // maps the db name to its queries
	usingServiceCfg       bool
	chainRpcUrl           *string
	ProviderAddress       string
	EscrowContractAddress string
	ChainCode             chainTypes.ChainCode
}

func New(ctx context.Context, rpcUrl string, opts ...ClientOption) (*KwilClient, error) {
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
	// @yaiba TODO: option allow only chain interaction, no grpc interaction, avoid grpc connection
	c := &KwilClient{
		dbis:                  make(map[string]dbi),
		usingServiceCfg:       true,
		chainRpcUrl:           nil,
		ProviderAddress:       DefaultProviderAddress,
		EscrowContractAddress: DefaultEscrowAddress,
		ChainCode:             DefaultChainCode,
	}
	for _, opt := range opts {
		opt(c)
	}

	var err error
	c.grpc, err = grpc.New(ctx, rpcUrl)
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

	c.ChainCode = chainTypes.ChainCode(cfg.Funding.ChainCode)
	c.ProviderAddress = cfg.Funding.ProviderAddress
	c.EscrowContractAddress = cfg.Funding.PoolAddress
	c.chainRpcUrl = &cfg.Funding.RpcUrl

	// reapply opts since service config may have changed them if they were specified
	for _, opt := range opts {
		opt(c)
	}

	return c, nil
}

func (c *KwilClient) GetServiceConfig(ctx context.Context) (grpc.SvcConfig, error) {
	return c.grpc.GetServiceConfig(ctx)
}

func (c *KwilClient) SetChainRpcUrl(url string) {
	c.chainRpcUrl = &url
}

func (c *KwilClient) ListDatabases(ctx context.Context, owner string) ([]string, error) {
	return c.grpc.ListDatabases(ctx, strings.ToLower(owner))
}

func (c *KwilClient) Ping(ctx context.Context) (string, error) {
	return c.grpc.Ping(ctx)
}
