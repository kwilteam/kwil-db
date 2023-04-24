package client

import (
	"context"
	"fmt"
	"kwil/pkg/chain/contracts/escrow/types"
	"math/big"
)

func (c *Client) ApproveDeposit(ctx context.Context, amount *big.Int) (string, error) {
	if err := c.ensureTokenContractInitialized(ctx); err != nil {
		return "", fmt.Errorf("failed to ensure token contract initialized: %w", err)
	}

	res, err := c.tokenContract.Approve(ctx, c.PoolAddress, amount, c.PrivateKey)
	if err != nil {
		return "", fmt.Errorf("failed to approve deposit: %w", err)
	}

	return res.TxHash, nil
}

func (c *Client) Deposit(ctx context.Context, amount *big.Int) (string, error) {
	if err := c.ensurePoolContractInitialized(ctx); err != nil {
		return "", fmt.Errorf("failed to ensure pool contract initialized: %w", err)
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

// GetApprovedAmount returns the amount of tokens that the owner has allowed the escrow to withdraw.
// It optionally takes an address to check the allowance for. If no address is provided, it will use the
// client's address.
func (c *Client) GetApprovedAmount(ctx context.Context, address ...string) (*big.Int, error) {
	if err := c.ensureTokenContractInitialized(ctx); err != nil {
		return nil, fmt.Errorf("failed to ensure token contract initialized: %w", err)
	}

	addr, err := c.resolveAddress(address...)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve address: %w", err)
	}

	return c.tokenContract.Allowance(addr, c.PoolAddress)
}

func (c *Client) ensureTokenContractInitialized(ctx context.Context) error {
	if c.tokenContract != nil {
		return nil
	}

	if err := c.initTokenContract(ctx); err != nil {
		return fmt.Errorf("failed to init token contract: %w", err)
	}

	return nil
}

func (c *Client) ensurePoolContractInitialized(ctx context.Context) error {
	if c.poolContract != nil {
		return nil
	}

	if err := c.initPoolContract(ctx); err != nil {
		return fmt.Errorf("failed to init pool contract: %w", err)
	}

	return nil
}

func (c *Client) resolveAddress(address ...string) (string, error) {
	if len(address) == 0 {
		return c.getAddress()
	}

	return address[0], nil
}

func (c *Client) GetOnChainBalance(ctx context.Context, addr ...string) (*big.Int, error) {
	if err := c.ensureTokenContractInitialized(ctx); err != nil {
		return nil, fmt.Errorf("failed to ensure token contract initialized: %w", err)
	}

	address, err := c.resolveAddress(addr...)
	if err != nil {
		return nil, fmt.Errorf("failed to get address: %w", err)
	}

	return c.tokenContract.BalanceOf(address)
}

func (c *Client) GetDepositedAmount(ctx context.Context, addr ...string) (*big.Int, error) {
	if err := c.ensurePoolContractInitialized(ctx); err != nil {
		return nil, fmt.Errorf("failed to ensure pool contract initialized: %w", err)
	}

	address, err := c.resolveAddress(addr...)
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
