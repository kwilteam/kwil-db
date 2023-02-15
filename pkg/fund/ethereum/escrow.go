package ethereum

import (
	"context"
	"fmt"
	types2 "kwil/pkg/chain/contracts/escrow/types"
	"math/big"
)

// DepositFund deposits funds to the escrow contract
func (c *Client) DepositFund(ctx context.Context, to string, amount *big.Int) (*types2.DepositResponse, error) {
	account := c.Config.GetAccountAddress()
	allowance, err := c.Token.Allowance(account, c.Config.PoolAddress)
	if err != nil {
		return nil, fmt.Errorf("failed to get allowance: %v", err)
	}

	// check if allowance >= amount
	if allowance.Cmp(amount) < 0 {
		return nil, fmt.Errorf("not enough tokens to deposit %s (allowance %s)", amount.String(), allowance.String())
	}

	balance, err := c.Token.BalanceOf(account)
	if err != nil {
		return nil, err
	}

	if balance.Cmp(amount) < 0 {
		return nil, fmt.Errorf("not enough tokens to deposit %s (balance %s)", amount.String(), balance.String())
	}

	depoistRes, err := c.Escrow.Deposit(ctx, &types2.DepositParams{
		Validator: to,
		Amount:    amount,
	}, c.Config.Wallet)
	if err != nil {
		return nil, err
	}

	return depoistRes, nil
}

func (c *Client) GetDepositBalance(ctx context.Context, validator string) (*big.Int, error) {
	balanceRes, err := c.Escrow.Balance(ctx, &types2.DepositBalanceParams{
		Validator: validator,
		Address:   c.Config.GetAccountAddress(),
	})
	if err != nil {
		return nil, err
	}
	return balanceRes.Balance, nil
}
