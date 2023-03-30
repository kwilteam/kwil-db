package txsvc

import (
	"context"
	txpb "kwil/api/protobuf/tx/v1"
)

func (s *Service) GetConfig(ctx context.Context, req *txpb.GetConfigRequest) (*txpb.GetConfigResponse, error) {
	return &txpb.GetConfigResponse{
		ChainCode:       s.cfg.Fund.Chain.ChainCode,
		PoolAddress:     s.cfg.Fund.PoolAddress,
		ProviderAddress: s.cfg.Fund.GetAccountAddress(),
	}, nil
}
