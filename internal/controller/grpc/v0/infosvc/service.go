package infosvc

import (
	pb "kwil/api/protobuf/info/v0/gen/go"
	"kwil/kwil/repository"
	"kwil/pkg/logger"
)

type Service struct {
	pb.UnsafeInfoServiceServer

	dao repository.Queries
	log logger.Logger
}

func NewService(queries repository.Queries, logger logger.Logger) *Service {
	return &Service{
		log: logger.Named("infosvc"),
		dao: queries,
	}
}
