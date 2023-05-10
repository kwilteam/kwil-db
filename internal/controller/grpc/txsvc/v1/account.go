package txsvc

import (
	"context"
	txpb "github.com/kwilteam/kwil-db/api/protobuf/tx/v1"
	"github.com/kwilteam/kwil-db/internal/entity"
	"github.com/kwilteam/kwil-db/pkg/utils/serialize"
)

func (s *Service) GetAccount(ctx context.Context, req *txpb.GetAccountRequest) (*txpb.GetAccountResponse, error) {
	acc, err := s.executor.GetAccount(req.Address)
	if err != nil {
		return nil, err
	}

	pbAcc, err := serialize.Convert[entity.Account, txpb.Account](acc)
	if err != nil {
		return nil, err
	}

	return &txpb.GetAccountResponse{
		Account: pbAcc,
	}, nil
}
