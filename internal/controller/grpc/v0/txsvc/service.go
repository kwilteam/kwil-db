package txsvc

import (
	txpb "kwil/api/protobuf/kwil/tx/v0/gen/go"
	"kwil/kwil/repository"
	"kwil/pkg/log"
	"kwil/pkg/pricing/pricer"
	"kwil/x/execution/executor"
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
