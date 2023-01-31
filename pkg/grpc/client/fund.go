package client

import (
	"context"
	"crypto/ecdsa"
	"kwil/x/types/contracts/escrow"
	"math/big"
)

func (c *Client) DepositFund(ctx context.Context, pk *ecdsa.PrivateKey, to string, amount *big.Int) (*escrow.DepositResponse, error) {
	return c.Chain.DepositFund(ctx, pk, to, amount)
}
