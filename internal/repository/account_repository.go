package repository

import (
	"context"
	"fmt"
	"kwil/internal/repository/gen"
	accounts2 "kwil/pkg/fund/accounts"
	"kwil/pkg/sql/errors"
	bigutil "kwil/pkg/utils/big"
	"strings"
)

type Accounter interface {
	UpdateAccount(ctx context.Context, account *accounts2.Account) error
	GetAccount(ctx context.Context, address string) (*accounts2.Account, error)
	Spend(ctx context.Context, spend *accounts2.Spend) error
}

func (q *queries) UpdateAccount(ctx context.Context, account *accounts2.Account) error {
	return q.gen.UpdateAccountByAddress(ctx, &gen.UpdateAccountByAddressParams{
		AccountAddress: strings.ToLower(account.Address),
		Spent:          account.Spent,
		Balance:        account.Balance,
		Nonce:          account.Nonce,
	})
}

func (q *queries) GetAccount(ctx context.Context, address string) (*accounts2.Account, error) {
	addr := strings.ToLower(address)
	acc, err := q.gen.GetAccount(ctx, addr)
	if err != nil {
		if errors.IsNoRowsInResult(err) {
			return accounts2.EmptyAccount(addr), nil
		}
		return nil, err
	}

	return &accounts2.Account{
		Address: addr,
		Nonce:   acc.Nonce,
		Balance: acc.Balance,
		Spent:   acc.Spent,
	}, nil
}

func (q *queries) Spend(ctx context.Context, spend *accounts2.Spend) error {
	addr := strings.ToLower(spend.Address)
	// get the config
	acc, err := q.gen.GetAccount(ctx, addr)
	if err != nil {
		if errors.IsNoRowsInResult(err) {
			return accounts2.ErrAccountNotRegistered
		}
		return fmt.Errorf("error getting config for address %s: %d", addr, err)
	}

	// check the nonce
	if acc.Nonce+1 != spend.Nonce {
		return accounts2.ErrInvalidNonce
	}

	remaining, err := bigutil.BigStr(acc.Balance).Sub(spend.Amount)
	if err != nil {
		return fmt.Errorf("error subtracting amount from balance: %d", err)
	}
	if remaining.Sign() < 0 {
		return accounts2.ErrInsufficientFunds
	}

	// calculate the new spent
	newSpent, err := bigutil.BigStr(acc.Spent).Add(spend.Amount)
	if err != nil {
		return fmt.Errorf("error adding amount to spent: %d", err)
	}

	// update the config
	return q.gen.UpdateAccountById(ctx, &gen.UpdateAccountByIdParams{
		ID:      acc.ID,
		Spent:   newSpent.String(),
		Balance: remaining.String(),
		Nonce:   spend.Nonce,
	})
}
