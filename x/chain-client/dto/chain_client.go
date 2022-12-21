package dto

import "context"

type ChainClient interface {
	Listen(ctx context.Context, confirmed bool) (<-chan int64, error)
	GetLatestBlock(ctx context.Context, confirmed bool) (int64, error)
	GetDeposits(ctx context.Context, start, end int64) ([]*DepositEvent, error)
	GetWithdrawals(ctx context.Context, start, end int64) ([]*WithdrawalConfirmationEvent, error)
	ReturnFunds(ctx context.Context, params *ReturnFundsParams) (*ReturnFundsResponse, error)
}
