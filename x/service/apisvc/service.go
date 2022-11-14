package apisvc

import (
	"kwil/x/deposits"
	"kwil/x/execution"
	"kwil/x/logx"
	"kwil/x/pricing"
	"kwil/x/proto/apipb"
	"kwil/x/schema"
)

type Service struct {
	apipb.UnimplementedKwilServiceServer

	ds  deposits.Deposits
	log logx.Logger
	md  schema.Service
	p   pricing.Service
	e   execution.Service
}

func NewService(ds deposits.Deposits, md schema.Service, e execution.Service) *Service {
	return &Service{
		ds:  ds,
		log: logx.New(),
		md:  md,
		p:   pricing.NewService(),
		e:   e,
	}
}
