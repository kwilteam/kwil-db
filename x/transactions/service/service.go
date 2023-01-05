package service

import (
	"context"
	accounts "kwil/x/accounts/service"
	execution "kwil/x/execution/service"
	pricing "kwil/x/pricing/service"
	"kwil/x/transactions/dto"
)

type TransactionService interface {
	DeployDatabase(ctx context.Context, tx *dto.Transaction) (*dto.Response, error)
	DropDatabase(ctx context.Context, tx *dto.Transaction) (*dto.Response, error)
	ExecuteQuery(ctx context.Context, tx *dto.Transaction) (*dto.Response, error)
}

type service struct {
	execution execution.ExecutionService
	accounts  accounts.AccountsService
	pricing   pricing.PricingService
}

func NewService(ex execution.ExecutionService, acc accounts.AccountsService, pr pricing.PricingService) TransactionService {
	return &service{
		execution: ex,
		accounts:  acc,
		pricing:   pr,
	}
}
