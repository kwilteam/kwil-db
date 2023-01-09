package deposits

import (
	"context"
	"kwil/kwil/repository"
	"kwil/x/types/deposits"
)

func (s *depositer) Spend(ctx context.Context, spend deposits.Spend) error {
	return s.dao.Spend(ctx, &repository.SpendParams{
		AccountAddress: spend.Address,
		Balance:        spend.Amount,
	})
}

func (s *depositer) GetBalancesAndSpent(ctx context.Context, wallet string) (*deposits.Balance, error) {
	res, err := s.dao.GetAccount(ctx, wallet)
	if err != nil {
		return nil, err
	}

	return &deposits.Balance{
		Balance: res.Balance,
		Spent:   res.Spent,
	}, nil
}
