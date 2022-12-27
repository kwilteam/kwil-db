package app

import (
	"context"
	"fmt"
	"kwil/x/proto/depositsvc"
)

func (s *Service) GetBalance(ctx context.Context, req *depositsvc.GetBalanceRequest) (*depositsvc.GetBalanceResponse, error) {
	wallet, err := s.service.GetBalancesAndSpent(ctx, req.Wallet)
	if err != nil {
		return nil, fmt.Errorf("error getting balance and spent for wallet %s: %d", req.Wallet, err)
	}

	return &depositsvc.GetBalanceResponse{
		Balance: wallet.Balance,
		Spent:   wallet.Spent,
	}, nil
}
