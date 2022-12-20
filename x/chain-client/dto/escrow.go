package dto

import "context"

type EscrowContract interface {
	GetDeposits(ctx context.Context, start, end int64) ([]*DepositEvent, error)
	GetWithdrawals(ctx context.Context, start, end int64) ([]*WithdrawalConfirmationEvent, error)
	ReturnFunds(ctx context.Context, params *ReturnFundsParams) (*ReturnFundsResponse, error)
}
