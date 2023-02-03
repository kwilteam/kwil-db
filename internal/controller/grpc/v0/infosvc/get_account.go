package infosvc

import (
	"context"
	"fmt"
	infopb "kwil/api/protobuf/info/v0/gen/go"
	"kwil/pkg/sql/errors"
	"strings"
)

func (s *Service) GetAccount(ctx context.Context, req *infopb.GetAccountRequest) (*infopb.GetAccountResponse, error) {
	acc, err := s.dao.GetAccount(ctx, strings.ToLower(req.Address))
	if err != nil {
		if errors.IsNoRowsInResult(err) {
			return &infopb.GetAccountResponse{
				Account: &infopb.Account{
					Address: req.Address,
					Nonce:   0,
					Balance: "0",
					Spent:   "0",
				},
			}, nil
		}
		return nil, fmt.Errorf("error getting info for address %s: %d", req.Address, err)
	}
	return &infopb.GetAccountResponse{
		Account: &infopb.Account{
			Address: req.Address,
			Nonce:   acc.Nonce,
			Balance: acc.Balance,
			Spent:   acc.Spent,
		},
	}, nil
}
