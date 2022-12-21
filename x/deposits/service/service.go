package service

import (
	"context"
	"kwil/x/deposits/chain"
	"kwil/x/deposits/dto"
	"kwil/x/deposits/repository"
	"kwil/x/logx"
	"kwil/x/sqlx/sqlclient"
)

type DepositsService interface {
	Spend(ctx context.Context, spend dto.Spend) error
	GetBalancesAndSpent(ctx context.Context, wallet string) (*dto.Balance, error)
	Deposit(ctx context.Context, deposit dto.Deposit) error
	startWithdrawal(ctx context.Context, withdrawal dto.StartWithdrawal) error
}

type chainWriter interface {
	ReturnFunds(ctx context.Context, params *chain.ReturnFundsParams) (*chain.ReturnFundsResponse, error)
}

// in the future we can make things like expirationPeriod and chunkSize configurable, but these values are good enough for now
type depositsService struct {
	dao              *repository.Queries
	db               *sqlclient.DB
	log              logx.SugaredLogger
	expirationPeriod int64
	chainWriter      chainWriter
}

func NewService(db *sqlclient.DB) DepositsService {
	return &depositsService{
		dao:              repository.New(db),
		db:               db,
		expirationPeriod: 100,
		log:              logx.New().Named("deposits-service").Sugar(),
		chainWriter:      nil,
	}
}
