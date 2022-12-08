package apisvc

import (
	"kwil/x/logx"
	"kwil/x/pricing"
	"kwil/x/proto/apipb"
	"kwil/x/sqlx/manager"
)

type Service struct {
	apipb.UnimplementedKwilServiceServer

	log     logx.Logger
	p       pricing.Service
	manager *manager.Manager
}

func NewService(mngr *manager.Manager) *Service {
	return &Service{
		log:     logx.New(),
		p:       pricing.NewService(),
		manager: mngr,
	}
}
