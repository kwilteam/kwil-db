package app

import (
	"kwil/x/logx"
	"kwil/x/pricing/service"
	"kwil/x/proto/pricingsvc"
)

type Service struct {
	pricingsvc.UnimplementedKwilServiceServer

	log     logx.Logger
	pricing service.PricingService
}

func NewService() *Service {
	return &Service{
		log:     logx.New(),
		pricing: service.NewService(),
	}
}
