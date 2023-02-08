package ethereum

import (
	"context"
	"fmt"
	"kwil/pkg/fund"
	"kwil/pkg/log"
	"math/big"
	"sync"
)

// Driver is a driver for the chain client for integration tests
type Driver struct {
	RpcUrl           string
	ValidatorAddress string

	logger     log.Logger
	connOnce   sync.Once
	Fund       fund.IFund
	fundConfig *fund.Config
}

func New(rpcUrl string, validatorAddr string, logger log.Logger) *Driver {
	return &Driver{
		logger:           logger,
		RpcUrl:           rpcUrl,
		ValidatorAddress: validatorAddr,
	}
}

func (d *Driver) DepositFund(ctx context.Context, amount *big.Int) error {
	client, err := d.GetClient()
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}

	_, err = client.DepositFund(ctx, d.ValidatorAddress, amount)

	return err
}

func (d *Driver) GetDepositBalance(ctx context.Context) (*big.Int, error) {
	client, err := d.GetClient()
	if err != nil {
		return nil, fmt.Errorf("failed to create client: %w", err)
	}

	return client.GetDepositBalance(ctx, d.ValidatorAddress)
}

func (d *Driver) ApproveToken(ctx context.Context, spender string, amount *big.Int) error {
	client, err := d.GetClient()
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}

	_, err = client.ApproveToken(ctx, spender, amount)

	return err
}

func (d *Driver) GetAllowance(ctx context.Context, from string, spender string) (*big.Int, error) {
	client, err := d.GetClient()
	if err != nil {
		return nil, fmt.Errorf("failed to create client: %w", err)
	}

	return client.GetAllowance(ctx, from, spender)
}

func (d *Driver) GetFundConfig() *fund.Config {
	return d.fundConfig
}

func (d *Driver) SetFundConfig(cfg *fund.Config) {
	d.fundConfig = cfg
}

func (d *Driver) GetClient() (fund.IFund, error) {
	var err error
	d.connOnce.Do(func() {
		d.Fund, err = NewClient(d.fundConfig, d.logger)
	})

	return d.Fund, err
}
