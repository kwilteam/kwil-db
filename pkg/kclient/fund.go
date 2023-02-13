package kclient

import (
	"context"
	"fmt"
	"kwil/pkg/contracts/escrow/types"
	"math/big"
)

func (c *Client) DepositFund(ctx context.Context, amount *big.Int) (*types.DepositResponse, error) {
	fundingCfg, err := c.Kwil.GetFundingServiceConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get funding service config: %w", err)
	}
	return c.Fund.DepositFund(ctx, fundingCfg.ProviderAddress, amount)
}
