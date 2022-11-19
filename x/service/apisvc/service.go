package apisvc

import (
	"ksl/sqlclient"
	"kwil/x/deposits"
	"kwil/x/logx"
	"kwil/x/metadata"
	"kwil/x/pricing"
	"kwil/x/proto/apipb"
)

type Service struct {
	apipb.UnimplementedKwilServiceServer

	ds      deposits.Deposits
	log     logx.Logger
	p       pricing.Service
	md      metadata.Service
	sqlOpen sqlclient.OpenerFunc
	mp      *metadata.ConnectionProvider
}

func NewService(ds deposits.Deposits, md metadata.Service, mp *metadata.ConnectionProvider) *Service {
	return &Service{
		ds:  ds,
		log: logx.New(),
		md:  md,
		p:   pricing.NewService(),
		mp:  mp,
	}
}
