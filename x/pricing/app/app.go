package app

import (
	"kwil/x/logx"
	"kwil/x/pricing/service"
	"kwil/x/proto/pricingpb"
)

type Service struct {
	pricingpb.UnimplementedPricingServiceServer

	log     logx.Logger
	pricing service.PricingService
}

func NewService() *Service {
	return &Service{
		log:     logx.New(),
		pricing: service.NewService(),
	}
}
