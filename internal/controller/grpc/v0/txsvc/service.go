package txsvc

import (
	txpb "kwil/api/protobuf/tx/v0"
	"kwil/internal/repository"
	"kwil/internal/usecases/executor"
	"kwil/pkg/log"
	"kwil/pkg/pricing/pricer"
)

type Service struct {
	txpb.UnimplementedTxServiceServer

	log log.Logger

	dao repository.Queries

	executor executor.Executor
	pricing  pricer.Pricer
}

func NewService(queries repository.Queries, exec executor.Executor, logger log.Logger) *Service {
	return &Service{
		log:      logger.Named("txsvc"),
		dao:      queries,
		executor: exec,
		pricing:  pricer.NewPricer(),
	}
}
