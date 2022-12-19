package service

import (
	"context"
	"kwil/x/deposits/dto"
	"kwil/x/deposits/repository"
)

func (s *depositsService) Spend(ctx context.Context, spend dto.Spend) error {
	return s.dao.Spend(ctx, &repository.SpendParams{
		Wallet:  spend.Wallet,
		Balance: spend.Amount,
	})
}

func (s *depositsService) GetBalancesAndSpent(ctx context.Context, wallet string) (*dto.Balance, error) {
	res, err := s.dao.GetBalanceAndSpent(ctx, wallet)
	if err != nil {
		return nil, err
	}

	return &dto.Balance{
		Balance: res.Balance,
		Spent:   res.Spent,
	}, nil
}
