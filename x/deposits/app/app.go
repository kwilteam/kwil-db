package app

import (
	"context"
	deposits "kwil/x/deposits/service"
	"kwil/x/logx"
	"kwil/x/proto/depositsvc"
)

type Service struct {
	depositsvc.UnimplementedKwilServiceServer

	log     logx.Logger
	service deposits.DepositsService
}

func NewService(svc deposits.DepositsService) *Service {
	return &Service{
		log:     logx.New(),
		service: svc,
	}
}

func (s *Service) Sync(ctx context.Context) error {
	return s.service.Sync(ctx)
}
