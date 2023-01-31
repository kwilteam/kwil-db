package accountsvc

import (
	pb "kwil/api/protobuf/account/v0/gen/go"
	"kwil/kwil/repository"
	"kwil/x/logx"
)

type Service struct {
	pb.UnimplementedAccountServiceServer

	dao repository.Queries
	log logx.Logger
}

func NewService(queries repository.Queries) *Service {
	return &Service{
		log: logx.New().Named("accountsvc"),
		dao: queries,
	}
}
