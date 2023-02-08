package client

import (
	"context"
	"fmt"
	pb "kwil/api/protobuf/kwil/configuration/v0/gen/go"
)

func (c *Gr) GetServiceConfig(ctx context.Context) (SvcConfig, error) {
	resp, err := c.cfgClt.GetAll(ctx, &pb.GetCfgRequest{})
	if err != nil {
		return SvcConfig{}, fmt.Errorf("failed to get service config: %w", err)
	}

	return SvcConfig{
		Funding: SvcFundingConfig{
			ChainCode:        resp.Funding.ChainCode,
			PoolAddress:      resp.Funding.PoolAddress,
			ValidatorAccount: resp.Funding.ValidatorAccount,
		},
	}, nil
}

func (c *Gr) GetFundingServiceConfig(ctx context.Context) (SvcFundingConfig, error) {
	resp, err := c.cfgClt.GetFunding(ctx, &pb.GetFundingCfgRequest{})
	if err != nil {
		return SvcFundingConfig{}, fmt.Errorf("failed to get funding service config: %w", err)
	}

	return SvcFundingConfig{
		ChainCode:        resp.ChainCode,
		PoolAddress:      resp.PoolAddress,
		ValidatorAccount: resp.ValidatorAccount,
	}, nil
}
