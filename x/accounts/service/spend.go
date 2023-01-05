package service

import (
	"context"
	"fmt"
	"kwil/x/accounts/dto"
	"kwil/x/accounts/repository"
)

func (s *accountsService) Spend(ctx context.Context, spend *dto.Spend) error {
	err := s.dao.Spend(ctx, &repository.SpendParams{
		AccountAddress: spend.Address,
		Balance:        spend.Amount,
		Nonce:          spend.Nonce,
	})
	if err != nil {
		return fmt.Errorf("failed to spend: %w", err)
	}

	return nil
}
