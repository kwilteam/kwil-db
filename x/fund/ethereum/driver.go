package ethereum

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"kwil/x/fund"
	"math/big"
	"sync"
)

// Driver is a driver for the chain client for integration tests
type Driver struct {
	Addr string

	connOnce sync.Once
	Chain    fund.IFund
}

func (d *Driver) DepositFund(ctx context.Context, from *ecdsa.PrivateKey, to string, amount *big.Int) error {
	client, err := d.GetClient()
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}

	_, err = client.DepositFund(ctx, from, to, amount)

	return err
}

func (d *Driver) GetDepositBalance(ctx context.Context, from string, to string) (*big.Int, error) {
	client, err := d.GetClient()
	if err != nil {
		return nil, fmt.Errorf("failed to create client: %w", err)
	}

	return client.GetDepositBalance(ctx, from, to)
}

func (d *Driver) ApproveToken(ctx context.Context, from *ecdsa.PrivateKey, spender string, amount *big.Int) error {
	client, err := d.GetClient()
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}

	_, err = client.ApproveToken(ctx, from, spender, amount)

	return err
}

func (d *Driver) GetAllowance(ctx context.Context, from string, spender string) (*big.Int, error) {
	client, err := d.GetClient()
	if err != nil {
		return nil, fmt.Errorf("failed to create client: %w", err)
	}

	return client.GetAllowance(ctx, from, spender)
}

func (d *Driver) GetConfig() *fund.Config {
	return d.Chain.GetConfig()
}

func (d *Driver) GetClient() (fund.IFund, error) {
	var err error
	d.connOnce.Do(func() {
		chainCfg, err := fund.NewConfig()
		if err != nil {
			err = fmt.Errorf("failed to create chain config: %w", err)
			return
		}

		chainClient, err := NewClient(chainCfg)
		if err != nil {
			err = fmt.Errorf("failed to create chain client: %w", err)
			return
		}

		d.Chain = chainClient
	})

	return d.Chain, err
}
