package kclient

import (
	"context"
	"kwil/pkg/contracts/escrow/types"
	"math/big"
)

func (c *Client) DepositFund(ctx context.Context, to string, amount *big.Int) (*types.DepositResponse, error) {
	return c.Fund.DepositFund(ctx, to, amount)
}
