package repository

import (
	"context"
	"fmt"
	"kwil/kwil/repository/gen"
	"kwil/pkg/sql/errors"
	accountTypes "kwil/x/types/accounts"
	bigutil "kwil/x/utils/big"
	"strings"
)

type Accounter interface {
	UpdateAccount(ctx context.Context, account *accountTypes.Account) error
	GetAccount(ctx context.Context, address string) (*accountTypes.Account, error)
	Spend(ctx context.Context, spend *accountTypes.Spend) error
}

func (q *queries) UpdateAccount(ctx context.Context, account *accountTypes.Account) error {
	return q.gen.UpdateAccountByAddress(ctx, &gen.UpdateAccountByAddressParams{
		AccountAddress: strings.ToLower(account.Address),
		Spent:          account.Spent,
		Balance:        account.Balance,
		Nonce:          account.Nonce,
	})
}

func (q *queries) GetAccount(ctx context.Context, address string) (*accountTypes.Account, error) {
	addr := strings.ToLower(address)
	acc, err := q.gen.GetAccount(ctx, addr)
	if err != nil {
		if errors.IsNoRowsInResult(err) {
			return accountTypes.EmptyAccount(addr), nil
		}
		return nil, err
	}

	return &accountTypes.Account{
		Address: addr,
		Nonce:   acc.Nonce,
		Balance: acc.Balance,
		Spent:   acc.Spent,
	}, nil
}

func (q *queries) Spend(ctx context.Context, spend *accountTypes.Spend) error {
	addr := strings.ToLower(spend.Address)
	// get the info
	acc, err := q.gen.GetAccount(ctx, addr)
	if err != nil {
		if errors.IsNoRowsInResult(err) {
			return accountTypes.ErrAccountNotRegistered
		}
		return fmt.Errorf("error getting info for address %s: %d", addr, err)
	}

	// check the nonce
	if acc.Nonce+1 != spend.Nonce {
		return accountTypes.ErrInvalidNonce
	}

	remaining, err := bigutil.BigStr(acc.Balance).Sub(spend.Amount)
	if err != nil {
		return fmt.Errorf("error subtracting amount from balance: %d", err)
	}
	if remaining.Sign() < 0 {
		return accountTypes.ErrInsufficientFunds
	}

	// calculate the new spent
	newSpent, err := bigutil.BigStr(acc.Spent).Add(spend.Amount)
	if err != nil {
		return fmt.Errorf("error adding amount to spent: %d", err)
	}

	// update the info
	return q.gen.UpdateAccountById(ctx, &gen.UpdateAccountByIdParams{
		ID:      acc.ID,
		Spent:   newSpent.String(),
		Balance: remaining.String(),
		Nonce:   spend.Nonce,
	})
}
