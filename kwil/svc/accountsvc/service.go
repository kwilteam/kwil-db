package accountsvc

import (
	"kwil/kwil/repository"
	"kwil/x/logx"
	"kwil/x/proto/accountspb"
)

type Service struct {
	accountspb.UnimplementedAccountServiceServer

	dao repository.Queries
	log logx.Logger
}

func NewService(queries repository.Queries) *Service {
	return &Service{
		log: logx.New().Named("accountsvc"),
		dao: queries,
	}
}
