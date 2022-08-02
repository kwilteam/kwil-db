package txsvc

import (
	txpb "github.com/kwilteam/kwil-db/api/protobuf/tx/v0"
	"github.com/kwilteam/kwil-db/internal/repository"
	"github.com/kwilteam/kwil-db/internal/usecases/executor"
	"github.com/kwilteam/kwil-db/pkg/log"
	"github.com/kwilteam/kwil-db/pkg/pricing/pricer"
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
		log:      *logger.Named("txsvc"),
		dao:      queries,
		executor: exec,
		pricing:  pricer.NewPricer(),
	}
}
