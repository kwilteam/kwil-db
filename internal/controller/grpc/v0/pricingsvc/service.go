package pricingsvc

import (
	pb "kwil/api/protobuf/kwil/pricing/v0/gen/go"
	"kwil/pkg/pricing/pricer"
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
