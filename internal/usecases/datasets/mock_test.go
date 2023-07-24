package datasets_test

import (
	"fmt"
	"math/big"

	"github.com/kwilteam/kwil-db/pkg/balances"
)

type mockAccountStore struct{}

func (m *mockAccountStore) BatchCredit(creditList []*balances.Credit, chain *balances.ChainConfig) error {
	return nil
}

func (m *mockAccountStore) Close() error {
	return nil
}

func (m *mockAccountStore) GetAccount(address string) (*balances.Account, error) {
	bal, ok := new(big.Int).SetString("10000000000000000000000", 10)
	if !ok {
		return nil, fmt.Errorf("error parsing balance")
	}

	return &balances.Account{
		Address: address,
		Balance: bal,
	}, nil
}

func (m *mockAccountStore) Spend(spend *balances.Spend) error {
	return nil
}

func (a *mockAccountStore) UpdateGasCosts(gas_enabled bool) {
}

func (a *mockAccountStore) GasEnabled() bool {
	return false
}
