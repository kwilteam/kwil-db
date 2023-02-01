package txsvc

import (
	txpb "kwil/api/protobuf/tx/v0/gen/go"
	"kwil/kwil/repository"
	"kwil/pkg/logger"
	"kwil/pkg/pricing/pricer"
	"kwil/x/execution/executor"
)

type Service struct {
	txpb.UnimplementedTxServiceServer

	log logger.Logger

	dao repository.Queries

	executor executor.Executor
	pricing  pricer.Pricer
}

func NewService(queries repository.Queries, exec executor.Executor) *Service {
	return &Service{
		log:      logger.New().Named("txsvc"),
		dao:      queries,
		executor: exec,
		pricing:  pricer.NewPricer(),
	}
}
