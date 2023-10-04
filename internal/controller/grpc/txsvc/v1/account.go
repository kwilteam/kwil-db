package txsvc

import (
	"context"

	txpb "github.com/kwilteam/kwil-db/api/protobuf/tx/v1"
)

func (s *Service) GetAccount(ctx context.Context, req *txpb.GetAccountRequest) (*txpb.GetAccountResponse, error) {
	acc, err := s.accountStore.GetAccount(ctx, req.PublicKey)
	if err != nil {
		return nil, err
	}

	return &txpb.GetAccountResponse{
		Account: &txpb.Account{
			PublicKey: acc.PublicKey, // nil for non-existent account
			Nonce:     acc.Nonce,
			Balance:   acc.Balance.String(),
		},
	}, nil
}
