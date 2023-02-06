package accountsvc

import (
	pb "kwil/api/protobuf/kwil/account/v0/gen/go"
	"kwil/kwil/repository"
	"kwil/pkg/log"
)

type Service struct {
	pb.UnimplementedAccountServiceServer

	dao repository.Queries
	log log.Logger
}

func NewService(queries repository.Queries, logger log.Logger) *Service {
	return &Service{
		log: logger.Named("accountsvc"),
		dao: queries,
	}
}
