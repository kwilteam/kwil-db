package accountsvc

import (
	pb "kwil/api/protobuf/account/v0/gen/go"
	"kwil/kwil/repository"
	"kwil/pkg/logger"
)

type Service struct {
	pb.UnimplementedAccountServiceServer

	dao repository.Queries
	log logger.Logger
}

func NewService(queries repository.Queries, logger logger.Logger) *Service {
	return &Service{
		log: logger.Named("accountsvc"),
		dao: queries,
	}
}
