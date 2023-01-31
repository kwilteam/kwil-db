package txsvc

import (
	txpb "kwil/api/protobuf/tx/v0/gen/go"
	"kwil/kwil/repository"
	"kwil/x/execution/executor"
	"kwil/x/logx"
	"kwil/x/pricing/pricer"
)

type Service struct {
	txpb.UnimplementedTxServiceServer

	log logx.Logger

	dao repository.Queries

	executor executor.Executor
	pricing  pricer.Pricer
}

func NewService(queries repository.Queries, exec executor.Executor) *Service {
	return &Service{
		log:      logx.New().Named("txsvc"),
		dao:      queries,
		executor: exec,
		pricing:  pricer.NewPricer(),
	}
}
