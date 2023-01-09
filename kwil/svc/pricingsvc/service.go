package pricingsvc

import (
	"kwil/x/pricing/pricer"
	"kwil/x/proto/pricingpb"
)

type Service struct {
	pricingpb.UnimplementedPricingServiceServer

	pricer pricer.Pricer
}

func NewService() *Service {
	pricer := pricer.NewPricer()
	return &Service{
		pricer: pricer,
	}
}
