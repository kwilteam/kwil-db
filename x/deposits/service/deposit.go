package service

import (
	"context"
	"kwil/x/deposits/dto"
	"kwil/x/deposits/repository"
)

func (s *depositsService) Deposit(ctx context.Context, deposit dto.Deposit) error {
	return s.doa.Deposit(ctx, &repository.DepositParams{
		Wallet: deposit.Wallet,
		Amount: deposit.Amount,
		TxHash: deposit.TxHash,
		Height: deposit.Height,
	})
}
