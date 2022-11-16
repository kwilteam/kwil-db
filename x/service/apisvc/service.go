package apisvc

import (
	"kwil/x/deposits"
	"kwil/x/execution"
	"kwil/x/logx"
	"kwil/x/metadata"
	"kwil/x/pricing"
	"kwil/x/proto/apipb"
)

type Service struct {
	apipb.UnimplementedKwilServiceServer

	ds  deposits.Deposits
	log logx.Logger
	p   pricing.Service
	md  metadata.Service
	e   execution.Service
}

func NewService(ds deposits.Deposits, md metadata.Service, e execution.Service) *Service {
	return &Service{
		ds:  ds,
		log: logx.New(),
		md:  md,
		p:   pricing.NewService(),
		e:   e,
	}
}
