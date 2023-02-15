package ethereum

import (
	"context"
	"kwil/pkg/chain/contracts/token/types"
	"math/big"
)

func (c *Client) ApproveToken(ctx context.Context, spender string, amount *big.Int) (*types.ApproveResponse, error) {
	return c.Token.Approve(ctx, spender, amount, c.Config.Wallet)
}

func (c *Client) GetAllowance(ctx context.Context, from string, spender string) (*big.Int, error) {
	return c.Token.Allowance(from, spender)
}

func (c *Client) GetBalance(ctx context.Context, account string) (*big.Int, error) {
	return c.Token.BalanceOf(account)
}
