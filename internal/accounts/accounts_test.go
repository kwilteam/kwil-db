package accounts_test

import (
	"context"
	"errors"
	"math/big"
	"os"
	"testing"

	"github.com/kwilteam/kwil-db/internal/accounts"
	"github.com/kwilteam/kwil-db/internal/sql/adapter"
	"github.com/kwilteam/kwil-db/internal/sql/sqlite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	account1 = "account1"
	account2 = "account2"
)

func Test_Accounts(t *testing.T) {
	type testCase struct {
		name string

		gasOn bool

		credit             map[string]*big.Int // to test credit new/non-existent accounts
		creditErr          error
		afterCreditBalance map[string]*big.Int

		spends        []*accounts.Spend
		finalBalances map[string]*accounts.Account
		err           error // the error must be triggered once

		postCredit             map[string]*big.Int // to test credit existing accounts
		postCreditErr          error
		afterPostCreditBalance map[string]*big.Int
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
			finalBalances: map[string]*accounts.Account{
				account1: newAccount(account1, 0, 2),
				account2: newAccount(account2, 0, 1),
			},
			err: nil,
		},
		{
			name: "gas and nonces on, no account",
			spends: []*accounts.Spend{
				newSpend(account1, 100, 1),
			},
			gasOn:         true,
			finalBalances: map[string]*accounts.Account{},
			err:           accounts.ErrAccountNotFound,
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
			gasOn:         true,
			finalBalances: map[string]*accounts.Account{},
			err:           accounts.ErrInsufficientFunds,
		},
		{
			name: "gas and nonces on, credits",
			credit: map[string]*big.Int{
				account1: big.NewInt(123),
			},
			creditErr: nil,
			afterCreditBalance: map[string]*big.Int{
				account1: big.NewInt(123),
				account2: big.NewInt(0), // same
			},
			spends: []*accounts.Spend{
				newSpend(account1, 100, 1),
			},
			gasOn: true,
			finalBalances: map[string]*accounts.Account{
				account1: newAccount(account1, 23, 1),
			},
			err: nil,
			postCredit: map[string]*big.Int{
				account1: big.NewInt(27),
				account2: big.NewInt(42),
			},
			postCreditErr: nil,
			afterPostCreditBalance: map[string]*big.Int{
				account1: big.NewInt(50),
				account2: big.NewInt(42),
			},
		},
		{
			name:   "no account, gas off",
			spends: []*accounts.Spend{},
			gasOn:  false,
			finalBalances: map[string]*accounts.Account{
				account1: newAccount(account1, 0, 0),
			},
			err: nil,
		},
		{
			name: "invalid nonce",
			spends: []*accounts.Spend{
				newSpend(account1, 100, 1),
				newSpend(account1, 100, 3),
				newSpend(account2, -100, 1),
			},
			gasOn: false,
			finalBalances: map[string]*accounts.Account{
				account1: newAccount(account1, 0, 1),
				account2: newAccount(account2, 0, 1),
			},
			err: accounts.ErrInvalidNonce,
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
			finalBalances: map[string]*accounts.Account{
				account1: newAccount(account1, 0, 2),
			},
			err: accounts.ErrInsufficientFunds,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			deleteTestDir()
			defer deleteTestDir()

			ctx := context.Background()

			pool, err := sqlite.NewPool(ctx, "./tmp/accounts_test.db", 1, 1, true)
			require.NoError(t, err)
			defer pool.Close()

			opts := []accounts.AccountStoreOpts{}
			if tc.gasOn {
				opts = append(opts, accounts.WithGasCosts(true))
			}
			// if tc.noncesOn {
			// 	opts = append(opts, accounts.WithNonces(true))
			// }

			ar, err := accounts.NewAccountStore(ctx, &adapter.PoolAdapater{Pool: pool}, &mockCommittable{skip: false}, opts...)
			require.NoError(t, err)

			for acct, amt := range tc.credit {
				err := ar.Credit(ctx, []byte(acct), amt)
				assert.ErrorIs(t, err, tc.creditErr)
			}

			for acct, amt := range tc.afterCreditBalance {
				account, err := ar.GetAccount(ctx, []byte(acct))
				assert.NoError(t, err)
				if account.Balance.Cmp(amt) != 0 {
					t.Fatalf("expected balance %s, got %s", amt, account.Balance)
				}
			}

			errs := []error{}
			for _, spend := range tc.spends {
				err := ar.Spend(ctx, spend)
				if err != nil {
					errs = append(errs, err)
				}
			}
			assertErr(t, errs, tc.err)

			for address, expectedBalance := range tc.finalBalances {
				account, err := ar.GetAccount(ctx, []byte(address))
				if err != nil {
					t.Fatalf("unexpected error: %s", err)
				}

				assert.Equal(t, expectedBalance.Balance, account.Balance, "expected balance %s, got %s", expectedBalance, account.Balance)
				assert.Equal(t, expectedBalance.Nonce, account.Nonce, "expected nonce %d, got %d", expectedBalance.Nonce, account.Nonce)
			}

			for acct, amt := range tc.postCredit {
				err := ar.Credit(ctx, []byte(acct), amt)
				// assertErr(t, []error{err}, tc.creditErr)
				assert.ErrorIs(t, err, tc.postCreditErr)
			}

			for acct, amt := range tc.afterPostCreditBalance {
				account, err := ar.GetAccount(ctx, []byte(acct))
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

func assertErr(t *testing.T, errs []error, target error) {
	t.Helper()
	if target == nil {
		if len(errs) > 0 {
			t.Fatalf("expected no error, got %s", errs)
		}
		return
	}

	contains := false
	for _, err := range errs {
		if errors.Is(err, target) {
			contains = true
		}
	}

	if !contains {
		t.Fatalf("expected error %s, got %s", target, errs)
	}
}

type mockCommittable struct {
	skip bool
}

var testDir = "./tmp"

func (m *mockCommittable) Register(value []byte) error {
	return nil
}

func (m *mockCommittable) Skip() bool {
	return m.skip
}

func deleteTestDir() {
	err := os.RemoveAll(testDir)
	if err != nil {
		panic(err)
	}
}
