package pricingsvc

import (
	pb "kwil/api/protobuf/pricing/v0/gen/go"
	"kwil/x/pricing/pricer"
)

type Service struct {
	pb.UnimplementedPricingServiceServer

	pricer pricer.Pricer
}

func NewService() *Service {
	pricer := pricer.NewPricer()
	return &Service{
		pricer: pricer,
	}
}
