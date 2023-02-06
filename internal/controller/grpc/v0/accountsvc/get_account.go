package accountsvc

import (
	"context"
	"fmt"
	pb "kwil/api/protobuf/account/v0/gen/go"
	"kwil/pkg/sql/errors"
	"strings"
)

func (s *Service) GetAccount(ctx context.Context, req *pb.GetAccountRequest) (*pb.GetAccountResponse, error) {
	acc, err := s.dao.GetAccount(ctx, strings.ToLower(req.Address))
	if err != nil {
		if errors.IsNoRowsInResult(err) {
			return &pb.GetAccountResponse{
				Account: &pb.Account{
					Address: req.Address,
					Nonce:   0,
					Balance: "0",
					Spent:   "0",
				},
			}, nil
		}
		return nil, fmt.Errorf("error getting config for address %s: %d", req.Address, err)
	}
	return &pb.GetAccountResponse{
		Account: &pb.Account{
			Address: req.Address,
			Nonce:   acc.Nonce,
			Balance: acc.Balance,
			Spent:   acc.Spent,
		},
	}, nil
}
