package configsvc

import (
	"context"
	pb "kwil/api/protobuf/config/v0"
	"kwil/pkg/fund"
	"kwil/pkg/log"
)

type Service struct {
	pb.UnimplementedConfigServiceServer
	log log.Logger

	fundCfg *fund.Config
}

func NewService(cfg *fund.Config, logger log.Logger) *Service {
	return &Service{
		fundCfg: cfg,
		log:     logger.Named("configsvc"),
	}
}

func (s *Service) GetAll(context.Context, *pb.GetCfgRequest) (*pb.GetCfgResponse, error) {
	return &pb.GetCfgResponse{
		Funding: &pb.GetFundingCfgResponse{
			ChainCode:       s.fundCfg.Chain.ChainCode,
			PoolAddress:     s.fundCfg.PoolAddress,
			ProviderAddress: s.fundCfg.GetAccountAddress(),
		}}, nil
}

func (s *Service) GetFunding(context.Context, *pb.GetFundingCfgRequest) (*pb.GetFundingCfgResponse, error) {
	return &pb.GetFundingCfgResponse{
		ChainCode:       s.fundCfg.Chain.ChainCode,
		PoolAddress:     s.fundCfg.PoolAddress,
		ProviderAddress: s.fundCfg.GetAccountAddress(),
		// @yaiba TODO: get token address and name
		//TokenAddress: "",
		//TokenName:    "",
	}, nil
}
