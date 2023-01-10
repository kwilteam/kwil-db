package txsvc

import (
	"kwil/kwil/repository"
	"kwil/x/execution/executor"
	"kwil/x/pricing/pricer"
	"kwil/x/proto/txpb"
)

type Service struct {
	txpb.UnimplementedTxServiceServer

	dao repository.Queries

	executor executor.Executor
	pricing  pricer.Pricer
}

func NewService(queries repository.Queries, exec executor.Executor) *Service {
	return &Service{
		dao:      queries,
		executor: exec,
		pricing:  pricer.NewPricer(),
	}
}
