package client

import (
	"context"
	"fmt"
	cfgpb "kwil/api/protobuf/config/v0"
)

func (c *Client) GetServiceConfig(ctx context.Context) (SvcConfig, error) {
	resp, err := c.cfgClt.GetAll(ctx, &cfgpb.GetCfgRequest{})
	if err != nil {
		return SvcConfig{}, fmt.Errorf("failed to get service config: %w", err)
	}

	return SvcConfig{
		Funding: SvcFundingConfig{
			ChainCode:       resp.Funding.ChainCode,
			PoolAddress:     resp.Funding.PoolAddress,
			ProviderAddress: resp.Funding.ProviderAddress,
		},
	}, nil
}

func (c *Client) GetFundingServiceConfig(ctx context.Context) (SvcFundingConfig, error) {
	resp, err := c.cfgClt.GetFunding(ctx, &cfgpb.GetFundingCfgRequest{})
	if err != nil {
		return SvcFundingConfig{}, fmt.Errorf("failed to get funding service config: %w", err)
	}

	return SvcFundingConfig{
		ChainCode:       resp.ChainCode,
		PoolAddress:     resp.PoolAddress,
		ProviderAddress: resp.ProviderAddress,
	}, nil
}
