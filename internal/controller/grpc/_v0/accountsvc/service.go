package accountsvc

import (
	accountspb "github.com/kwilteam/kwil-db/api/protobuf/accounts/v0"
	"github.com/kwilteam/kwil-db/internal/repository"
	"github.com/kwilteam/kwil-db/pkg/log"
)

type Service struct {
	accountspb.UnimplementedAccountServiceServer

	dao repository.Queries
	log log.Logger
}

func NewService(queries repository.Queries, logger log.Logger) *Service {
	return &Service{
		log: *logger.Named("accountsvc"),
		dao: queries,
	}
}
