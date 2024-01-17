package deposit_oracle

import (
	"context"
	"encoding/hex"
	"math/big"
	"testing"

	"github.com/kwilteam/kwil-db/core/log"
	"github.com/kwilteam/kwil-db/internal/accounts"
	"github.com/kwilteam/kwil-db/internal/voting"
	"github.com/stretchr/testify/require"
)

const (
	testAccount     = "0xabcd"
	trimTestAccount = "abcd"
)

var (
	validAC = &AccountCredit{
		Account:   testAccount,
		Amount:    big.NewInt(100),
		TxHash:    "test-txhash",
		BlockHash: "test-blockhash",
		ChainID:   "test-chainid",
	}
)

func Test_ValidPayload(t *testing.T) {
	ds := &voting.Datastores{
		Accounts: &mockAccountStore{
			accounts: make(map[string]*accounts.Account),
		},
		Databases: nil,
	}

	ac := validAC

	err := ac.Apply(context.Background(), ds, log.NewStdOut(log.InfoLevel))
	require.NoError(t, err)

	// check that the account was credited
	bts, err := hex.DecodeString(trimTestAccount)
	require.NoError(t, err)

	acc, err := ds.Accounts.GetAccount(context.Background(), bts)
	require.NoError(t, err)
	require.Equal(t, big.NewInt(100), acc.Balance)

}

func Test_WithoutDatastores(t *testing.T) {
	ac := validAC

	err := ac.Apply(context.Background(), nil, log.NewStdOut(log.InfoLevel))
	require.Error(t, err, "datastores not initialized")
}

func Test_WithoutAccountstore(t *testing.T) {
	ds := &voting.Datastores{
		Accounts:  nil,
		Databases: nil,
	}

	ac := validAC

	err := ac.Apply(context.Background(), ds, log.NewStdOut(log.InfoLevel))
	require.Error(t, err, "accountstore not initialized")
}

type mockAccountStore struct {
	accounts map[string]*accounts.Account
}

func (mas *mockAccountStore) GetAccount(ctx context.Context, identifier []byte) (*accounts.Account, error) {
	id := hex.EncodeToString(identifier)
	acct, ok := mas.accounts[id]
	if !ok {
		return nil, accounts.ErrAccountNotFound
	}
	return acct, nil
}

func (mas *mockAccountStore) Credit(ctx context.Context, identifier []byte, amount *big.Int) error {
	id := hex.EncodeToString(identifier)
	acct, ok := mas.accounts[id]
	if !ok {
		acct = &accounts.Account{
			Identifier: identifier,
			Balance:    big.NewInt(0),
			Nonce:      0,
		}
		mas.accounts[id] = acct
	}

	acct.Balance.Add(acct.Balance, amount)

	return nil
}
