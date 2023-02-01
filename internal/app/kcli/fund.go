package kcli

import (
	"context"
	"kwil/x/types/contracts/escrow"
	"math/big"
)

func (c *KwilClient) DepositFund(ctx context.Context, to string, amount *big.Int) (*escrow.DepositResponse, error) {
	return c.Fund.DepositFund(ctx, to, amount)
}
