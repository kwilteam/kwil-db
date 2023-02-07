package ethereum

import (
	"context"
	"fmt"
	escrow2 "kwil/pkg/types/contracts/escrow"
	"math/big"
)

// DepositFund deposits funds to the escrow contract
func (c *Client) DepositFund(ctx context.Context, to string, amount *big.Int) (*escrow2.DepositResponse, error) {
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

	depoistRes, err := c.Escrow.Deposit(ctx, &escrow2.DepositParams{
		Validator: to,
		Amount:    amount,
	})
	if err != nil {
		return nil, err
	}

	return depoistRes, nil
}

func (c *Client) GetDepositBalance(ctx context.Context, validator string, wallet string) (*big.Int, error) {
	balanceRes, err := c.Escrow.Balance(ctx, &escrow2.DepositBalanceParams{
		Validator: validator,
		Address:   wallet,
	})
	if err != nil {
		return nil, err
	}
	return balanceRes.Balance, nil
}
