package service

import (
	"context"
	"kwil/x/chain-client/dto"
)

type EscrowContract interface {
	ReturnFunds(ctx context.Context, params *dto.ReturnFundsParams) (*dto.ReturnFundsResponse, error)
	GetDeposits(ctx context.Context, startBlock, endBlock int64) ([]*dto.DepositEvent, error)
	GetWithdrawals(ctx context.Context, startBlock, endBlock int64) ([]*dto.WithdrawalConfirmationEvent, error)
}

func (c *chainClient) ReturnFunds(ctx context.Context, params *dto.ReturnFundsParams) (*dto.ReturnFundsResponse, error) {
	return c.client.ReturnFunds(ctx, params)
}

func (c *chainClient) GetDeposits(ctx context.Context, startBlock, endBlock int64) ([]*dto.DepositEvent, error) {
	return c.client.GetDeposits(ctx, startBlock, endBlock)
}

func (c *chainClient) GetWithdrawals(ctx context.Context, startBlock, endBlock int64) ([]*dto.WithdrawalConfirmationEvent, error) {
	return c.client.GetWithdrawals(ctx, startBlock, endBlock)
}
