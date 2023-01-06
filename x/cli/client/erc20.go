package client

import (
	"context"
	"kwil/x/contracts/token/dto"
	"math/big"
)

func (c *client) GetBalance() (*big.Int, error) {
	return c.UnconnectedClient.Token.BalanceOf(c.UnconnectedClient.Address.Hex())
}

func (c *client) Approve(ctx context.Context, spender string, amount *big.Int) (*dto.ApproveResponse, error) {
	return c.UnconnectedClient.Token.Approve(ctx, spender, amount)
}
