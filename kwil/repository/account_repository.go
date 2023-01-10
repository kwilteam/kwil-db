package repository

import (
	"context"
	"fmt"
	"kwil/kwil/repository/gen"
	"kwil/x/sqlx/errors"
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
	// get the account
	acc, err := q.gen.GetAccount(ctx, addr)
	if err != nil {
		if errors.IsNoRowsInResult(err) {
			return fmt.Errorf("address not registered on server %s", addr)
		}
		return fmt.Errorf("error getting account for address %s: %d", addr, err)
	}

	// check the nonce
	if acc.Nonce+1 != spend.Nonce {
		return fmt.Errorf("invalid nonce for address %s.  expected: %d. received: %d", addr, acc.Nonce+1, spend.Nonce)
	}

	remaining, err := bigutil.BigStr(acc.Balance).Sub(spend.Amount)
	if err != nil {
		return fmt.Errorf("error subtracting amount from balance: %d", err)
	}
	if remaining.Sign() < 0 {
		return fmt.Errorf("insufficient funds to spend for address %s", addr)
	}

	// calculate the new spent
	newSpent, err := bigutil.BigStr(acc.Spent).Add(spend.Amount)
	if err != nil {
		return fmt.Errorf("error adding amount to spent: %d", err)
	}

	// update the account
	return q.gen.UpdateAccountById(ctx, &gen.UpdateAccountByIdParams{
		ID:      acc.ID,
		Spent:   newSpent.String(),
		Balance: remaining.String(),
		Nonce:   spend.Nonce,
	})
}
