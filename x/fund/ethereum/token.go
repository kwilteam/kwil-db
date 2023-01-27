package ethereum

import (
	"context"
	"crypto/ecdsa"
	"kwil/x/types/contracts/token"
	"math/big"
)

func (c *Client) ApproveToken(ctx context.Context, pk *ecdsa.PrivateKey, spender string, amount *big.Int) (*token.ApproveResponse, error) {
	return c.Token.Approve(ctx, spender, amount)
}

func (c *Client) GetAllowance(ctx context.Context, from string, spender string) (*big.Int, error) {
	return c.Token.Allowance(from, spender)
}

func (c *Client) GetBalance(ctx context.Context, account string) (*big.Int, error) {
	return c.Token.BalanceOf(account)
}
