package accountsvc

import (
	accountspb "kwil/api/protobuf/accounts/v0"
	"kwil/internal/repository"
	"kwil/pkg/log"
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
