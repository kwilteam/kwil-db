package accountsvc

import (
	"context"
	"fmt"
	"kwil/x/proto/accountspb"
	"kwil/x/sqlx/errors"
	"strings"
)

func (s *Service) GetAccount(ctx context.Context, req *accountspb.GetAccountRequest) (*accountspb.GetAccountResponse, error) {
	acc, err := s.dao.GetAccount(ctx, strings.ToLower(req.Address))
	if err != nil {
		if errors.IsNoRowsInResult(err) {
			return &accountspb.GetAccountResponse{
				Address: req.Address,
				Nonce:   0,
				Balance: "0",
				Spent:   "0",
			}, nil
		}
		return nil, fmt.Errorf("error getting account for address %s: %d", req.Address, err)
	}
	return &accountspb.GetAccountResponse{
		Address: req.Address,
		Nonce:   acc.Nonce,
		Balance: acc.Balance,
		Spent:   acc.Spent,
	}, nil
}
