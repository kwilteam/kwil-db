package service

import (
	"context"
	"kwil/x/accounts/repository"
)

func (s *accountsService) IncreaseBalance(ctx context.Context, address string, amount string) error {
	return s.dao.IncreaseBalance(ctx, &repository.IncreaseBalanceParams{
		AccountAddress: address,
		Balance:        amount,
	})
}

func (s *accountsService) DecreaseBalance(ctx context.Context, address string, amount string) error {
	return s.dao.DecreaseBalance(ctx, &repository.DecreaseBalanceParams{
		AccountAddress: address,
		Balance:        amount,
	})
}
