package configsvc

import (
	"context"
	pb "kwil/api/protobuf/config/v0"
	"kwil/internal/app/kwild/config"
	"kwil/pkg/log"
)

type Service struct {
	pb.UnimplementedConfigServiceServer
	log log.Logger

	cfg *config.AppConfig
}

func NewService(cfg *config.AppConfig, logger log.Logger) *Service {
	return &Service{
		cfg: cfg,
		log: *logger.Named("configsvc"),
	}
}

func (s *Service) GetAll(context.Context, *pb.GetCfgRequest) (*pb.GetCfgResponse, error) {
	return &pb.GetCfgResponse{
		Funding: &pb.GetFundingCfgResponse{
			ChainCode:       s.cfg.Fund.Chain.ChainCode,
			PoolAddress:     s.cfg.Fund.PoolAddress,
			ProviderAddress: s.cfg.Fund.GetAccountAddress(),
			RpcUrl:          s.cfg.Fund.Chain.PublicRpcUrl,
		},
		Gateway: &pb.GetGatewayCfgResponse{
			GraphqlUrl: s.cfg.Gateway.GetGraphqlUrl(),
		},
	}, nil
}

func (s *Service) GetFunding(context.Context, *pb.GetFundingCfgRequest) (*pb.GetFundingCfgResponse, error) {
	return &pb.GetFundingCfgResponse{
		ChainCode:       s.cfg.Fund.Chain.ChainCode,
		PoolAddress:     s.cfg.Fund.PoolAddress,
		ProviderAddress: s.cfg.Fund.GetAccountAddress(),
		RpcUrl:          s.cfg.Fund.Chain.PublicRpcUrl,
		// @yaiba TODO: get token address and name
		//TokenAddress: "",
		//TokenName:    "",
	}, nil
}

func (s *Service) GetGateway(context.Context, *pb.GetGatewayCfgRequest) (*pb.GetGatewayCfgResponse, error) {
	return &pb.GetGatewayCfgResponse{
		GraphqlUrl: s.cfg.Gateway.GetGraphqlUrl(),
	}, nil
}
