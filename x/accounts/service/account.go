package service

import (
	"context"
	"fmt"
	"kwil/x/accounts/dto"
)

func (s *accountsService) GetAccount(ctx context.Context, address string) (*dto.Account, error) {
	acc, err := s.dao.GetAccount(ctx, address)
	if err != nil {
		return nil, fmt.Errorf("failed to get account from database: %w", err)
	}

	return &dto.Account{
		Address: address,
		Nonce:   acc.Nonce,
		Balance: acc.Balance,
		Spent:   acc.Spent,
	}, nil
}
