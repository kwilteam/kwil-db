package client

import (
	"context"
	"fmt"
	cfgpb "github.com/kwilteam/kwil-db/api/protobuf/config/v0"
)

func (c *Client) GetServiceConfig(ctx context.Context) (SvcConfig, error) {
	resp, err := c.cfgClt.GetAll(ctx, &cfgpb.GetCfgRequest{})
	if err != nil {
		return SvcConfig{}, fmt.Errorf("failed to get service config: %w", err)
	}

	return SvcConfig{
		Funding: SvcFundingConfig{
			ChainCode:       resp.Funding.GetChainCode(),
			PoolAddress:     resp.Funding.GetPoolAddress(),
			ProviderAddress: resp.Funding.GetProviderAddress(),
			RpcUrl:          resp.Funding.GetRpcUrl(),
		},
		Gateway: SvcGatewayConfig{
			GraphqlUrl: resp.Gateway.GetGraphqlUrl(),
		},
	}, nil
}

func (c *Client) GetFundingServiceConfig(ctx context.Context) (SvcFundingConfig, error) {
	resp, err := c.cfgClt.GetFunding(ctx, &cfgpb.GetFundingCfgRequest{})
	if err != nil {
		return SvcFundingConfig{}, fmt.Errorf("failed to get funding service config: %w", err)
	}

	return SvcFundingConfig{
		ChainCode:       resp.GetChainCode(),
		PoolAddress:     resp.GetPoolAddress(),
		ProviderAddress: resp.GetProviderAddress(),
		RpcUrl:          resp.GetRpcUrl(),
	}, nil
}

func (c *Client) GetGatewayServiceConfig(ctx context.Context) (SvcGatewayConfig, error) {
	resp, err := c.cfgClt.GetGateway(ctx, &cfgpb.GetGatewayCfgRequest{})
	if err != nil {
		return SvcGatewayConfig{}, fmt.Errorf("failed to get gateway service config: %w", err)
	}

	return SvcGatewayConfig{
		GraphqlUrl: resp.GetGraphqlUrl(),
	}, nil
}
