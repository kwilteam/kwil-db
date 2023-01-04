package app

import (
	"context"
	"fmt"
	"kwil/x/proto/accountspb"
)

func (s *Service) GetAccount(ctx context.Context, req *accountspb.GetAccountRequest) (*accountspb.GetAccountResponse, error) {
	acc, err := s.service.GetAccount(ctx, req.Address)
	if err != nil {
		return nil, fmt.Errorf("error getting account for address %s: %d", req.Address, err)
	}

	return &accountspb.GetAccountResponse{
		Address: acc.Address,
		Nonce:   acc.Nonce,
		Balance: acc.Balance,
		Spent:   acc.Spent,
	}, nil
}
