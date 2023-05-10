package balances_test

import (
	"fmt"
	"github.com/kwilteam/kwil-db/pkg/balances"
	"math/big"
	"os"
	"testing"
)

var testPath string

func init() {
	dirname, err := os.UserHomeDir()
	if err != nil {
		dirname = "/tmp"
	}

	testPath = fmt.Sprintf("%s/.kwil_test/sqlite/", dirname)
}

func Test_AccountStore(t *testing.T) {
	as, err := balances.NewAccountStore(balances.Wipe(), balances.WithPath(testPath))
	if err != nil {
		t.Errorf("error creating account store: %v", err)
	}
	defer as.Close()

	// try to get an account that doesn't exist
	_, err = as.GetAccount("0x123")
	if err == nil {
		t.Errorf("expected error getting non-existent account")
	}

	// spend for an account that doesn't exist
	spend := balances.Spend{
		AccountAddress: "0x123",
		Amount:         big.NewInt(100),
		Nonce:          1,
	}

	err = as.Spend(&spend)
	if err == nil {
		t.Errorf("expected error spending from non-existent account")
	}

	// credit an account
	credit := balances.Credit{
		AccountAddress: "0x123",
		Amount:         big.NewInt(100),
	}

	err = as.Credit(&credit)
	if err != nil {
		t.Errorf("error crediting account: %v", err)
	}

	// get the account
	account, err := as.GetAccount("0x123")
	if err != nil {
		t.Errorf("error getting account: %v", err)
	}

	// check the balance
	if account.Balance.Cmp(big.NewInt(100)) != 0 {
		t.Errorf("expected balance of 100, got %v", account.Balance)
	}

	// check the nonce
	if account.Nonce != 0 {
		t.Errorf("expected nonce of 0, got %v", account.Nonce)
	}

	// spend from the account
	err = as.Spend(&spend)
	if err != nil {
		t.Errorf("error spending from account: %v", err)
	}

	// get the account
	account, err = as.GetAccount("0x123")
	if err != nil {
		t.Errorf("error getting account: %v", err)
	}

	// check the balance
	if account.Balance.Cmp(big.NewInt(0)) != 0 {
		t.Errorf("expected balance of 0, got %v", account.Balance)
	}

	// check the nonce
	if account.Nonce != 1 {
		t.Errorf("expected nonce of 1, got %v", account.Nonce)
	}
}

func Test_BatchSpendAndCredit(t *testing.T) {
	as, err := balances.NewAccountStore(balances.Wipe(), balances.WithPath(testPath))
	if err != nil {
		t.Errorf("error creating account store: %v", err)
	}
	defer as.Close()

	// batch of spends
	spendList := []*balances.Spend{
		{
			AccountAddress: "0x123",
			Amount:         big.NewInt(100),
			Nonce:          1,
		},
		{
			AccountAddress: "0x456",
			Amount:         big.NewInt(100),
			Nonce:          1,
		},
		{
			AccountAddress: "0x123",
			Amount:         big.NewInt(100),
			Nonce:          2,
		},
	}

	// batch of credits
	creditList := []*balances.Credit{
		{
			AccountAddress: "0x123",
			Amount:         big.NewInt(100),
		},
		{
			AccountAddress: "0x456",
			Amount:         big.NewInt(100),
		},
		{
			AccountAddress: "0x123",
			Amount:         big.NewInt(200),
		},
	}

	// start chain at height 0
	const someChainCode = 0
	err = as.CreateChain(someChainCode, 0)
	if err != nil {
		t.Errorf("error setting height: %v", err)
	}

	chainConifg := &balances.ChainConfig{
		ChainCode: someChainCode,
		Height:    1,
	}

	// try spend
	err = as.BatchSpend(spendList, chainConifg)
	if err == nil {
		t.Errorf("expected error spending from non-existent accounts")
	}

	// try credit
	err = as.BatchCredit(creditList, chainConifg)
	if err != nil {
		t.Errorf("error crediting accounts: %v", err)
	}

	// try spend
	err = as.BatchSpend(spendList, nil)
	if err != nil {
		t.Errorf("error spending from accounts: %v", err)
	}

	// get the accounts
	account1, err := as.GetAccount("0x123")
	if err != nil {
		t.Errorf("error getting account: %v", err)
	}

	account2, err := as.GetAccount("0x456")
	if err != nil {
		t.Errorf("error getting account: %v", err)
	}

	// check the balances
	if account1.Balance.Cmp(big.NewInt(100)) != 0 {
		t.Errorf("expected balance of 100, got %v", account1.Balance)
	}

	if account2.Balance.Cmp(big.NewInt(0)) != 0 {
		t.Errorf("expected balance of 0, got %v", account2.Balance)
	}

	// check the nonces
	if account1.Nonce != 2 {
		t.Errorf("expected nonce of 2, got %v", account1.Nonce)
	}

	if account2.Nonce != 1 {
		t.Errorf("expected nonce of 1, got %v", account2.Nonce)
	}

	// check the height
	height, err := as.GetHeight(someChainCode)
	if err != nil {
		t.Errorf("error getting height: %v", err)
	}

	if height != 1 {
		t.Errorf("expected height of 1, got %v", height)
	}
}

func Test_NonexistentChain(t *testing.T) {
	as, err := balances.NewAccountStore(balances.Wipe(), balances.WithPath(testPath))
	if err != nil {
		t.Errorf("error creating account store: %v", err)
	}
	defer as.Close()

	height, err := as.GetHeight(0)
	if err != nil {
		t.Errorf("nonexistent chain should return height of 0, got error: %v", err)
	}

	if height != 0 {
		t.Errorf("expected height of 0, got %v", height)
	}
}
