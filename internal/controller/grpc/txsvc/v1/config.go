package txsvc

import (
	"context"
	txpb "github.com/kwilteam/kwil-db/api/protobuf/tx/v1"
)

func (s *Service) GetConfig(ctx context.Context, req *txpb.GetConfigRequest) (*txpb.GetConfigResponse, error) {

	return &txpb.GetConfigResponse{
		ChainCode:       int64(s.cfg.Deposits.ChainCode),
		PoolAddress:     s.cfg.Deposits.PoolAddress,
		ProviderAddress: s.providerAddress,
	}, nil
}
