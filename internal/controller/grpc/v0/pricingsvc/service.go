package pricingsvc

import (
	pricingpb "kwil/api/protobuf/pricing/v0"
	"kwil/pkg/pricing/pricer"
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
