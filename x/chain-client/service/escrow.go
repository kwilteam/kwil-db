package service

import (
	"context"
	"kwil/x/chain-client/dto"
)

func (c *chainClient) ReturnFunds(ctx context.Context, params *dto.ReturnFundsParams) (*dto.ReturnFundsResponse, error) {
	return c.escrow.ReturnFunds(ctx, params)
}

func (c *chainClient) GetDeposits(ctx context.Context, startBlock, endBlock int64) ([]*dto.DepositEvent, error) {
	return c.escrow.GetDeposits(ctx, startBlock, endBlock)
}

func (c *chainClient) GetWithdrawals(ctx context.Context, startBlock, endBlock int64) ([]*dto.WithdrawalConfirmationEvent, error) {
	return c.escrow.GetWithdrawals(ctx, startBlock, endBlock)
}
