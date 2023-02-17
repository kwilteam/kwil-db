package pricingsvc

import (
	pricingpb "kwil/api/protobuf/pricing/v0"
	"kwil/internal/usecases/executor"
	"kwil/pkg/pricing/pricer"
)

type Service struct {
	pricingpb.UnimplementedPricingServiceServer

	pricer   pricer.Pricer
	executor executor.Executor
}

func NewService(exec executor.Executor) *Service {
	pricer := pricer.NewPricer()
	return &Service{
		pricer:   pricer,
		executor: exec,
	}
}
