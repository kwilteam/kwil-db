//go:build pglive

package accounts_test

import (
	"context"
	"errors"
	"math/big"
	"testing"

	"github.com/kwilteam/kwil-db/internal/accounts"
	"github.com/kwilteam/kwil-db/internal/sql/pg"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	account1 = "account1"
	account2 = "account2"

	schemaName = `kwild_accts` // not exported from accounts but we want to clean up
)

func Test_Accounts(t *testing.T) {
	cfg := &pg.DBConfig{
		PoolConfig: pg.PoolConfig{
			ConnConfig: pg.ConnConfig{
				Host:   "127.0.0.1",
				Port:   "5432",
				User:   "kwild",
				Pass:   "kwild", // would be ignored if pg_hba.conf set with trust
				DBName: "kwil_test_db",
			},
			MaxConns: 11,
		},
	}

	/*
		The order of operations for each testcase is:
		- `credit`: credits the accounts with the given amount
		- `balanceAfterCredit`: checks that the accounts have the expected balance after the credit
		- `spend`: spends the amount from each account
		- `balanceAfterSpend`: checks that the accounts have the expected balance after the spend
		- `secondCredit`: credits the accounts with the given amount
		- `balanceAfterSecondCredit`: checks that the accounts have the expected balance after the second credit
	*/

	type testCase struct {
		name string

		gasOn bool

		credit             map[string]*big.Int // to test credit new/non-existent accounts
		creditErr          error
		balanceAfterCredit map[string]*big.Int

		spends            []*accounts.Spend
		spendErr          error // the error must be triggered once
		spendErrOuter     *accounts.SpendError
		balanceAfterSpend map[string]*accounts.Account

		secondCredit             map[string]*big.Int // to test credit existing accounts
		secondCreditErr          error
		balanceAfterSecondCredit map[string]*big.Int
	}

	// once we have a way to increase balances in accounts, we will have to add tests
	// for spending a valid amount
	testCases := []testCase{
		{
			name: "gas off, nonces on",
			spends: []*accounts.Spend{
				newSpend(account1, 100, 1),
				newSpend(account1, 100, 2),
				newSpend(account2, -100, 1),
			},
			gasOn: false,
			balanceAfterSpend: map[string]*accounts.Account{
				account1: newAccount(account1, 0, 2),
				account2: newAccount(account2, 0, 1),
			},
			spendErr: nil,
		},
		{
			name: "gas and nonces on, no account",
			spends: []*accounts.Spend{
				newSpend(account1, 100, 1),
			},
			gasOn:             true,
			balanceAfterSpend: map[string]*accounts.Account{},
			spendErr:          accounts.ErrAccountNotFound,
			spendErrOuter:     nil, // not for account not found
		},
		{
			name: "gas and nonces on, no funds",
			credit: map[string]*big.Int{
				account1: big.NewInt(1),
			},
			creditErr: nil,
			spends: []*accounts.Spend{
				newSpend(account1, 100, 1),
			},
			gasOn:             true,
			balanceAfterSpend: map[string]*accounts.Account{},
			spendErr:          accounts.ErrInsufficientFunds,
			spendErrOuter:     accounts.NewSpendError(accounts.ErrInsufficientFunds, big.NewInt(1), 0),
		},
		{
			name: "gas and nonces on, credits",
			credit: map[string]*big.Int{
				account1: big.NewInt(123),
			},
			creditErr: nil,
			balanceAfterCredit: map[string]*big.Int{
				account1: big.NewInt(123),
				account2: big.NewInt(0), // getting a non-existent account should return 0
			},
			spends: []*accounts.Spend{
				newSpend(account1, 100, 1),
			},
			gasOn: true,
			balanceAfterSpend: map[string]*accounts.Account{
				account1: newAccount(account1, 23, 1),
			},
			spendErr: nil,
			secondCredit: map[string]*big.Int{
				account1: big.NewInt(27),
				account2: big.NewInt(42),
			},
			secondCreditErr: nil,
			balanceAfterSecondCredit: map[string]*big.Int{
				account1: big.NewInt(50),
				account2: big.NewInt(42),
			},
		},
		{
			name:   "no account, gas off",
			spends: []*accounts.Spend{},
			gasOn:  false,
			balanceAfterSpend: map[string]*accounts.Account{
				account1: newAccount(account1, 0, 0),
			},
			spendErr: nil,
		},
		{
			name: "invalid nonce",
			spends: []*accounts.Spend{
				newSpend(account1, 100, 1),
				newSpend(account1, 100, 3),
				newSpend(account2, -100, 1),
			},
			gasOn: false,
			balanceAfterSpend: map[string]*accounts.Account{
				account1: newAccount(account1, 0, 1),
				account2: newAccount(account2, 0, 1),
			},
			spendErr:      accounts.ErrInvalidNonce,
			spendErrOuter: accounts.NewSpendError(accounts.ErrInvalidNonce, big.NewInt(0), 1),
		},
		{
			name: "Insufficient funds",
			credit: map[string]*big.Int{
				account1: big.NewInt(120),
			},
			creditErr: nil,
			spends: []*accounts.Spend{
				newSpend(account1, 100, 1),
				newSpend(account1, 100, 2),
			},
			gasOn: true,
			balanceAfterSpend: map[string]*accounts.Account{
				account1: newAccount(account1, 0, 2),
			},
			spendErr: accounts.ErrInsufficientFunds,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()

			db, err := pg.NewDB(ctx, cfg)
			require.NoError(t, err)
			defer db.Close()
			tx, err := db.BeginTx(ctx)

			require.NoError(t, err)
			defer tx.Rollback(ctx) // always rollback to avoid cleanup

			defer db.Execute(ctx, `DROP SCHEMA IF EXISTS `+schemaName+` CASCADE;`)

			opts := []accounts.AccountStoreOpts{}
			if tc.gasOn {
				opts = append(opts, accounts.WithGasCosts(true))
			}
			// if tc.noncesOn {
			// 	opts = append(opts, accounts.WithNonces(true))
			// }

			ar, err := accounts.NewAccountStore(ctx, tx, opts...)
			require.NoError(t, err)

			for acct, amt := range tc.credit {
				err := ar.Credit(ctx, tx, []byte(acct), amt)
				assert.ErrorIs(t, err, tc.creditErr)
			}

			for acct, amt := range tc.balanceAfterCredit {
				account, err := ar.GetAccount(ctx, tx, []byte(acct))
				require.NoError(t, err) // require to avoid panic
				if account.Balance.Cmp(amt) != 0 {
					t.Fatalf("expected balance %s, got %s", amt, account.Balance)
				}
			}

			var errs error
			for _, spend := range tc.spends {
				err := ar.Spend(ctx, tx, spend)
				if err != nil {
					errs = errors.Join(errs, err)
				}
			}
			assert.ErrorIs(t, errs, tc.spendErr)
			if tc.spendErrOuter != nil {
				serr := new(accounts.SpendError)
				if !errors.As(errs, &serr) {
					t.Fatal("not a spend error")
				}
				assert.ErrorIs(t, tc.spendErrOuter.Err, tc.spendErr)
				assert.Equal(t, tc.spendErrOuter.Balance, serr.Balance)
				assert.Equal(t, tc.spendErrOuter.Nonce, serr.Nonce)
			}

			for address, expectedBalance := range tc.balanceAfterSpend {
				account, err := ar.GetAccount(ctx, tx, []byte(address))
				if err != nil {
					t.Fatalf("unexpected error: %s", err)
				}

				assert.Equal(t, expectedBalance.Balance, account.Balance, "expected balance %s, got %s", expectedBalance, account.Balance)
				assert.Equal(t, expectedBalance.Nonce, account.Nonce, "expected nonce %d, got %d", expectedBalance.Nonce, account.Nonce)
			}

			for acct, amt := range tc.secondCredit {
				err := ar.Credit(ctx, tx, []byte(acct), amt)
				assert.ErrorIs(t, err, tc.secondCreditErr)
			}

			for acct, amt := range tc.balanceAfterSecondCredit {
				account, err := ar.GetAccount(ctx, tx, []byte(acct))
				assert.NoError(t, err)
				if account.Balance.Cmp(amt) != 0 {
					t.Fatalf("expected balance %s, got %s", amt, account.Balance)
				}
			}
		})
	}
}

func newSpend(address string, amount int64, nonce int64) *accounts.Spend {
	return &accounts.Spend{
		AccountID: []byte(address),
		Amount:    big.NewInt(amount),
		Nonce:     nonce,
	}
}

func newAccount(address string, balance int64, nonce int64) *accounts.Account {
	return &accounts.Account{
		Identifier: []byte(address),
		Balance:    big.NewInt(balance),
		Nonce:      nonce,
	}
}
