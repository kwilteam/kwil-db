package service

import (
	"context"
	"database/sql"
	"kwil/x/accounts/dto"
	"kwil/x/accounts/repository"
	"kwil/x/sqlx/sqlclient"
)

type AccountsService interface {
	// WithTx returns a new service that enables AccountsService
	// methods to be used within a transaction.
	WithTx(tx *sql.Tx) AccountsService

	// Spend deducts the amount from the account's balance,
	// adds the amount to the account's spent, and increments the account's nonce.
	Spend(ctx context.Context, spend dto.Spend) error

	// IncreaseBalance increases the account's balance by the given amount.
	// It does not modify the account's nonce or spent.
	IncreaseBalance(ctx context.Context, address string, amount string) error

	// DecreaseBalance decreases the account's balance by the given amount.
	// It does not modify the account's nonce or spent.
	DecreaseBalance(ctx context.Context, address string, amount string) error

	// GetAccount returns the account for the given address.
	GetAccount(ctx context.Context, address string) (*dto.Account, error)
}

type accountsService struct {
	dao *repository.Queries
	db  *sqlclient.DB
}

func NewService(db *sqlclient.DB) AccountsService {
	return &accountsService{
		dao: repository.New(db),
		db:  db,
	}
}
