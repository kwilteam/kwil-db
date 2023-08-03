package txsvc

import (
	"context"

	txpb "github.com/kwilteam/kwil-db/api/protobuf/tx/v1"
)

func (s *Service) GetAccount(ctx context.Context, req *txpb.GetAccountRequest) (*txpb.GetAccountResponse, error) {
	acc, err := s.executor.GetAccount(ctx, req.Address)
	if err != nil {
		return nil, err
	}

	return &txpb.GetAccountResponse{
		Account: &txpb.Account{
			Address: acc.Address,
			Nonce:   acc.Nonce,
			Balance: acc.Balance,
		},
	}, nil
}
