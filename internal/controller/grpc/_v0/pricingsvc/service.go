package pricingsvc

import (
	pricingpb "github.com/kwilteam/kwil-db/api/protobuf/pricing/v0"
	"github.com/kwilteam/kwil-db/internal/usecases/executor"
	"github.com/kwilteam/kwil-db/pkg/pricing/pricer"
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
