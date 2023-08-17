package accountsvc

import (
	"context"
	"fmt"
	"strings"

	accountspb "github.com/kwilteam/kwil-db/api/protobuf/accounts/v0"
	commonpb "github.com/kwilteam/kwil-db/api/protobuf/common/v0"
	"github.com/kwilteam/kwil-db/pkg/sql/errors"
)

func (s *Service) GetAccount(ctx context.Context, req *accountspb.GetAccountRequest) (*accountspb.GetAccountResponse, error) {
	acc, err := s.dao.GetAccount(ctx, strings.ToLower(req.Address))
	if err != nil {
		if errors.IsNoRowsInResult(err) {
			return &accountspb.GetAccountResponse{
				Account: &commonpb.Account{
					Address: req.Address,
					Nonce:   0,
					Balance: "0",
					Spent:   "0",
				},
			}, nil
		}
		return nil, fmt.Errorf("error getting config for address %s: %d", req.Address, err)
	}
	return &accountspb.GetAccountResponse{
		Account: &commonpb.Account{
			Address: req.Address,
			Nonce:   acc.Nonce,
			Balance: acc.Balance,
			Spent:   acc.Spent,
		},
	}, nil
}
