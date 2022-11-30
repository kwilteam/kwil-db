package apisvc

import (
	"kwil/x/deposits"
	"kwil/x/logx"
	"kwil/x/pricing"
	"kwil/x/proto/apipb"
)

type Service struct {
	apipb.UnimplementedKwilServiceServer

	ds  deposits.Deposits
	log logx.Logger
	p   pricing.Service
}

func NewService(ds deposits.Deposits) *Service {
	return &Service{
		ds:  ds,
		log: logx.New(),
		p:   pricing.NewService(),
	}
}
