package service

import (
	"context"
	"kwil/x/deposits/dto"
	"kwil/x/deposits/repository"
	"kwil/x/sqlx/sqlclient"
)

type DepositsService interface {
	Spend(ctx context.Context, spend dto.Spend) error
	GetBalancesAndSpent(ctx context.Context, wallet string) (*dto.Balance, error)
	Deposit(ctx context.Context, deposit dto.Deposit) error
	StartWithdrawal(ctx context.Context, withdrawal dto.StartWithdrawal) error
}

type depositsService struct {
	dao              *repository.Queries
	db               sqlclient.DB
	expirationPeriod int64
}

func NewService(db sqlclient.DB) DepositsService {
	return &depositsService{
		dao:              repository.New(db),
		db:               db,
		expirationPeriod: 100,
	}
}
