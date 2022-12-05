package apisvc

import (
	"kwil/x/deposits"
	"kwil/x/logx"
	"kwil/x/pricing"
	"kwil/x/proto/apipb"
	"kwil/x/sqlx/executor"
)

type Service struct {
	apipb.UnimplementedKwilServiceServer

	ds    deposits.Deposits
	log   logx.Logger
	p     pricing.Service
	exctr executor.Executor
}

func NewService(ds deposits.Deposits, exctr executor.Executor) *Service {
	return &Service{
		ds:    ds,
		log:   logx.New(),
		p:     pricing.NewService(),
		exctr: exctr,
	}
}
