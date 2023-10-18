package txsvc

import (
	"context"

	txpb "github.com/kwilteam/kwil-db/core/rpc/protobuf/tx/v1"
)

func (s *Service) GetAccount(ctx context.Context, req *txpb.GetAccountRequest) (*txpb.GetAccountResponse, error) {
	// Status is presently just 0 for confirmed and 1 for pending, but there may
	// be others such as finalized and safe.
	if req.Status != nil && *req.Status > 0 {
		// Ask the node application for account info, which includes any unconfirmed.
		balance, nonce, err := s.nodeApp.AccountInfo(ctx, req.PublicKey)
		if err != nil {
			return nil, err
		}
		var pubkey []byte
		if nonce > 0 { // return nil pubkey for non-existent account
			pubkey = req.PublicKey
		}
		return &txpb.GetAccountResponse{
			Account: &txpb.Account{
				PublicKey: pubkey, // nil for non-existent account
				Nonce:     nonce,
				Balance:   balance.String(),
			},
		}, nil
	}

	acct, err := s.accountStore.GetAccount(ctx, req.PublicKey)
	if err != nil {
		return nil, err
	}
	return &txpb.GetAccountResponse{
		Account: &txpb.Account{
			PublicKey: acct.PublicKey, // nil for non-existent account
			Nonce:     acct.Nonce,
			Balance:   acct.Balance.String(),
		},
	}, nil
}
