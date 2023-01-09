package client

import (
	"context"
	tokenTypes "kwil/x/types/contracts/token"
	"math/big"
)

func (c *client) GetBalance() (*big.Int, error) {
	return c.UnconnectedClient.Token().BalanceOf(c.UnconnectedClient.Address())
}

func (c *client) Approve(ctx context.Context, spender string, amount *big.Int) (*tokenTypes.ApproveResponse, error) {
	return c.UnconnectedClient.Token().Approve(ctx, spender, amount)
}
