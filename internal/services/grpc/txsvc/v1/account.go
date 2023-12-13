package txsvc

import (
	"context"

	txpb "github.com/kwilteam/kwil-db/core/rpc/protobuf/tx/v1"
)

func (s *Service) GetAccount(ctx context.Context, req *txpb.GetAccountRequest) (*txpb.GetAccountResponse, error) {
	// Status is presently just 0 for confirmed and 1 for pending, but there may
	// be others such as finalized and safe.
	uncommitted := false
	if req.Status != nil && *req.Status > 0 {
		uncommitted = true
	}

	balance, nonce, err := s.nodeApp.AccountInfo(ctx, req.Identifier, uncommitted)
	if err != nil {
		return nil, err
	}

	ident := []byte(nil)
	if nonce > 0 { // return nil pubkey for non-existent account
		ident = req.Identifier
	}

	return &txpb.GetAccountResponse{
		Account: &txpb.Account{
			Identifier: ident, // nil for non-existent account
			Nonce:      nonce,
			Balance:    balance.String(),
		},
	}, nil
}
