package client

import (
	"context"
	"fmt"
	"kwil/pkg/chain/contracts/escrow/types"
	"math/big"
)

func (c *Client) ApproveDeposit(ctx context.Context, amount *big.Int) (string, error) {
	if c.tokenContract == nil {
		err := c.initTokenContract(ctx)
		if err != nil {
			return "", fmt.Errorf("failed to init token contract: %w", err)
		}
	}

	res, err := c.tokenContract.Approve(ctx, c.PoolAddress, amount, c.PrivateKey)
	if err != nil {
		return "", fmt.Errorf("failed to approve deposit: %w", err)
	}

	return res.TxHash, nil
}

func (c *Client) Deposit(ctx context.Context, amount *big.Int) (string, error) {
	if c.poolContract == nil {
		err := c.initPoolContract(ctx)
		if err != nil {
			return "", fmt.Errorf("failed to init pool contract: %w", err)
		}
	}

	res, err := c.poolContract.Deposit(ctx, &types.DepositParams{
		Validator: c.ProviderAddress,
		Amount:    amount,
	}, c.PrivateKey)
	if err != nil {
		return "", fmt.Errorf("failed to deposit: %w", err)
	}

	return res.TxHash, nil
}

func (c *Client) GetApprovedAmount(ctx context.Context) (*big.Int, error) {
	if c.tokenContract == nil {
		err := c.initTokenContract(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to init token contract: %w", err)
		}
	}

	address, err := c.getAddress()
	if err != nil {
		return nil, fmt.Errorf("failed to get address: %w", err)
	}

	return c.tokenContract.Allowance(address, c.PoolAddress)
}

func (c *Client) GetBalance(ctx context.Context) (*big.Int, error) {
	if c.tokenContract == nil {
		err := c.initTokenContract(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to init token contract: %w", err)
		}
	}

	address, err := c.getAddress()
	if err != nil {
		return nil, fmt.Errorf("failed to get address: %w", err)
	}

	return c.tokenContract.BalanceOf(address)
}

func (c *Client) GetDepositBalance(ctx context.Context) (*big.Int, error) {
	if c.poolContract == nil {
		err := c.initPoolContract(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to init pool contract: %w", err)
		}
	}

	address, err := c.getAddress()
	if err != nil {
		return nil, fmt.Errorf("failed to get address: %w", err)
	}

	res, err := c.poolContract.Balance(ctx, &types.DepositBalanceParams{
		Validator: c.ProviderAddress,
		Address:   address,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get deposit balance: %w", err)
	}

	return res.Balance, nil
}
