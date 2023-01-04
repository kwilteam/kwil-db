package app

import (
	"kwil/x/accounts/service"
	"kwil/x/logx"
	"kwil/x/proto/accountspb"
)

type Service struct {
	accountspb.UnimplementedAccountServiceServer

	service service.AccountsService
	log     logx.Logger
}

func NewService(svc service.AccountsService) *Service {
	return &Service{
		log:     logx.New(),
		service: svc,
	}
}
