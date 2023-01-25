package ethereum

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"github.com/ethereum/go-ethereum/crypto"
	"kwil/x/types/contracts/token"
	"math/big"
)

func (c *Client) ApproveToken(ctx context.Context, pk *ecdsa.PrivateKey, spender string, amount *big.Int) (*token.ApproveResponse, error) {
	account := crypto.PubkeyToAddress(pk.PublicKey).Hex()

	// get balance
	balance, err := c.Token.BalanceOf(account)
	if err != nil {
		return nil, fmt.Errorf("could not get balance: %w", err)
	}

	// check if balance is less than amount
	if balance.Cmp(amount) < 0 {
		return nil, fmt.Errorf("not enough tokens to fund %s (balance %s)", amount.String(), balance.String())
	}

	return c.Token.Approve(ctx, spender, amount)
}

func (c *Client) GetAllowance(ctx context.Context, from string, spender string) (*big.Int, error) {
	return c.Token.Allowance(from, spender)
}

func (c *Client) GetBalance(ctx context.Context, account string) (*big.Int, error) {
	return c.Token.BalanceOf(account)
}
