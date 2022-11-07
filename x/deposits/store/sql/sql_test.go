package sql_test

import (
	"fmt"
	"math/big"
	"math/rand"
	"testing"

	ks "kwil/x/deposits/store/sql"
)

func Test_SQLStore(t *testing.T) {
	db, err := ks.TestDB()
	if err != nil {
		t.Error(err)
	}
	defer db.Close()

	// test to make sure the table deposits exists
	res, err := db.Query("SELECT tablename FROM pg_catalog.pg_tables WHERE tablename = 'deposits'")
	if err != nil {
		t.Error(err)
	}

	if !res.Next() {
		t.Error("table deposits does not exist")
	}

	// test height
	err = db.SetHeight(123)
	if err != nil {
		t.Error(err)
	}

	height, err := db.GetHeight()
	if err != nil {
		t.Error(err)
	}

	if height != 123 {
		t.Error("height is not 123")
	}

	// create random wallet address
	addr := generateNonce(30)
	fmt.Println(addr)
	err = db.RemoveWallet(addr)
	if err != nil {
		t.Error(err)
	}

	// check that balance and spent are 0
	balance, err := db.GetBalance(addr)
	if err != nil {
		t.Error(err)
	}

	if balance.Cmp(big.NewInt(0)) != 0 {
		t.Error("balance is not 0")
	}

	spent, err := db.GetSpent(addr)
	if err != nil {
		t.Error(err)
	}

	if spent.Cmp(big.NewInt(0)) != 0 {
		t.Error("spent is not 0, it is ", spent)
	}

	/* now we make four deposits for 100 dollars each.
	one of them will be for height 122, one for 123, one for 124, and one for 125.
	*/

	// deposit 1
	err = db.Deposit("txid1", addr, "100", 122)
	if err != nil {
		t.Error(err)
	}

	// deposit 2
	err = db.Deposit("txid2", addr, "100", 123)
	if err != nil {
		t.Error(err)
	}

	// deposit 3
	err = db.Deposit("txid3", addr, "100", 124)
	if err != nil {
		t.Error(err)
	}

	// deposit 4
	err = db.Deposit("txid4", addr, "100", 125)
	if err != nil {
		t.Error(err)
	}

	// now commit height
	err = db.CommitHeight(123)
	if err != nil {
		t.Error(err)
	}

	// check height is 124
	err = db.QueryRow("SELECT get_height()").Scan(&height)
	if err != nil {
		t.Error(err)
	}

	if height != 124 {
		t.Error("height is not 124")
	}

	// check balance is 200
	balance, err = db.GetBalance(addr)
	if err != nil {
		t.Error(err)
	}

	if balance.Cmp(big.NewInt(200)) != 0 {
		t.Error("balance is not 200")
	}

	// check spent is 0
	spent, err = db.GetSpent(addr)
	if err != nil {
		t.Error(err)
	}

	if spent.Cmp(big.NewInt(0)) != 0 {
		t.Error("spent is not 0, it is ", spent)
	}

	// spend 100
	err = db.Spend(addr, "100")
	if err != nil {
		t.Error(err)
	}

	// check balance is 100
	balance, err = db.GetBalance(addr)
	if err != nil {
		t.Error(err)
	}

	if balance.Cmp(big.NewInt(100)) != 0 {
		t.Error("balance is not 100")
	}

	// check spent is 100
	spent, err = db.GetSpent(addr)
	if err != nil {
		t.Error(err)
	}

	if spent.Cmp(big.NewInt(100)) != 0 {
		t.Error("spent is not 100, it is", spent)
	}

	// withdraw 50
	nce := generateNonce(10)
	nw, err := db.StartWithdrawal(nce, addr, "50", 130)
	if err != nil {
		t.Error(err)
	}

	if nw.Cid != nce {
		t.Errorf("cid is not %s, received %s", nce, nw.Cid)
	}

	if nw.Wallet != addr {
		t.Errorf("address is not %s, received %s", addr, nw.Wallet)
	}

	if nw.Amount != "50" {
		t.Errorf("amount is not 50, received %s", nw.Amount)
	}

	if nw.Expiration != 130 {
		t.Errorf("expiry is not 130, received %d", nw.Expiration)
	}

	if nw.Fee != "100" {
		t.Errorf("fee is not 100, received %s", nw.Fee)
	}

	// check balance is 50
	balance, err = db.GetBalance(addr)
	if err != nil {
		t.Error(err)
	}

	if balance.Cmp(big.NewInt(50)) != 0 {
		t.Error("balance is not 50")
	}

	// check spent is 0
	spent, err = db.GetSpent(addr)
	if err != nil {
		t.Error(err)
	}

	if spent.Cmp(big.NewInt(0)) != 0 {
		t.Error("spent is not 0")
	}

	// withdraw another 50 with 150 expiry
	_, err = db.StartWithdrawal(generateNonce(10), addr, "50", 150)
	if err != nil {
		t.Error(err)
	}

	// check balance is 0
	balance, err = db.GetBalance(addr)
	if err != nil {
		t.Error(err)
	}

	if balance.Cmp(big.NewInt(0)) != 0 {
		t.Error("balance is not 0")
	}

	// finish the first withdrawal
	exists, err := db.FinishWithdrawal(nce)
	if err != nil {
		t.Error(err)
	}

	if !exists {
		t.Error("withdrawal does not exist")
	}

	// commit height 160
	err = db.CommitHeight(160) // this also processes the 2 pending deposits, worth $100 each
	if err != nil {
		t.Error(err)
	}

	// check balance is 50
	balance, err = db.GetBalance(addr)
	if err != nil {
		t.Error(err)
	}

	if balance.Cmp(big.NewInt(250)) != 0 {
		t.Error("balance is not 250.  it is", balance)
	}

	// check spent is 0
	spent, err = db.GetSpent(addr)
	if err != nil {
		t.Error(err)
	}

	if spent.Cmp(big.NewInt(0)) != 0 {
		t.Error("spent is not 0")
	}

	// spend 300
	err = db.Spend(addr, "300")
	if err == nil {
		t.Error("spending more than balance should have failed")
	}

	// ensure balance still 250
	balance, err = db.GetBalance(addr)
	if err != nil {
		t.Error(err)
	}

	if balance.Cmp(big.NewInt(250)) != 0 {
		t.Error("balance is not 250")
	}

}

func Test_Withdrawal(t *testing.T) {
	db, err := ks.TestDB()
	if err != nil {
		t.Error(err)
	}
	defer db.Close()

	// test to make sure the table deposits exists
	res, err := db.Query("SELECT tablename FROM pg_catalog.pg_tables WHERE tablename = 'deposits'")
	if err != nil {
		t.Error(err)
	}

	if !res.Next() {
		t.Error("table deposits does not exist")
	}

	// create random wallet address
	addr := generateNonce(30)
	fmt.Println(addr)
	err = db.RemoveWallet(addr)
	if err != nil {
		t.Error(err)
	}

	// check balance is 0
	balance, err := db.GetBalance(addr)
	if err != nil {
		t.Error(err)
	}

	if balance.Cmp(big.NewInt(0)) != 0 {
		t.Error("balance is not 0")
	}

	// check spent is 0
	spent, err := db.GetSpent(addr)
	if err != nil {
		t.Error(err)
	}

	if spent.Cmp(big.NewInt(0)) != 0 {
		t.Error("spent is not 0")
	}

	// now let's give them some money and no spend
	err = db.Deposit("txid1", addr, "100", 122)
	if err != nil {
		t.Error(err)
	}

	// commit height 122
	err = db.CommitHeight(122)
	if err != nil {
		t.Error(err)
	}

	// check balance is 00
	balance, err = db.GetBalance(addr)
	if err != nil {
		t.Error(err)
	}

	if balance.Cmp(big.NewInt(100)) != 0 {
		t.Error("balance is not 100")
	}

	// start withdrawal
	nce := generateNonce(10)
	nw, err := db.StartWithdrawal(nce, addr, "100", 130)
	if err != nil {
		t.Error(err)
	}

	if nw.Cid != nce {
		t.Errorf("cid is not %s, received %s", nce, nw.Cid)
	}

	if nw.Wallet != addr {
		t.Errorf("address is not %s, received %s", addr, nw.Wallet)
	}

	if nw.Amount != "100" {
		t.Errorf("amount is not 100, received %s", nw.Amount)
	}

	if nw.Expiration != 130 {
		t.Errorf("expiry is not 130, received %d", nw.Expiration)
	}

	if nw.Fee != "0" {
		t.Errorf("fee is not 0, received %s", nw.Fee)
	}

	// check balance is 0
	balance, err = db.GetBalance(addr)
	if err != nil {
		t.Error(err)
	}

	if balance.Cmp(big.NewInt(0)) != 0 {
		t.Error("balance is not 0")
	}

	// deposit more money
	err = db.Deposit("txid2", addr, "100", 123)
	if err != nil {
		t.Error(err)
	}

	// commit height 123

	err = db.CommitHeight(123)
	if err != nil {
		t.Error(err)
	}

	// check balance is 100
	balance, err = db.GetBalance(addr)
	if err != nil {
		t.Error(err)
	}

	if balance.Cmp(big.NewInt(100)) != 0 {
		t.Error("balance is not 100")
	}

	// spend some
	err = db.Spend(addr, "50")
	if err != nil {
		t.Error(err)
	}

	// check balance is 50
	balance, err = db.GetBalance(addr)
	if err != nil {
		t.Error(err)
	}

	if balance.Cmp(big.NewInt(50)) != 0 {
		t.Error("balance is not 50")
	}

	// check spent is 50
	spent, err = db.GetSpent(addr)
	if err != nil {
		t.Error(err)
	}

	if spent.Cmp(big.NewInt(50)) != 0 {
		t.Error("spent is not 50")
	}

	// get balance and spent
	bal, sp, err := db.GetBalanceAndSpent(addr)
	if err != nil {
		t.Error(err)
	}

	if bal != "50" {
		t.Error("balance is not 50")
	}

	if sp != "50" {
		t.Error("spent is not 50")
	}

	// start withdrawal
	nce = generateNonce(10)
	nw, err = db.StartWithdrawal(nce, addr, "100", 130)
	if err != nil {
		t.Error(err)
	}

	if nw.Cid != nce {
		t.Errorf("cid is not %s, received %s", nce, nw.Cid)
	}

	if nw.Wallet != addr {
		t.Errorf("address is not %s, received %s", addr, nw.Wallet)
	}

	if nw.Amount != "50" {
		t.Errorf("amount is not 100, received %s", nw.Amount)
	}

	if nw.Expiration != 130 {
		t.Errorf("expiry is not 130, received %d", nw.Expiration)
	}

	if nw.Fee != "50" {
		t.Errorf("fee is not 0, received %s", nw.Fee)
	}

	// check balance is 0
	balance, err = db.GetBalance(addr)
	if err != nil {
		t.Error(err)
	}

	if balance.Cmp(big.NewInt(0)) != 0 {
		t.Error("balance is not 0")
	}

	// check spent is 0
	spent, err = db.GetSpent(addr)
	if err != nil {
		t.Error(err)
	}

	if spent.Cmp(big.NewInt(0)) != 0 {
		t.Error("spent is not 0")
	}

	// add tx to second withdrawal
	err = db.AddTx(nce, "0x132")
	if err != nil {
		t.Error(err)
	}

	// test get withdrawals for wallet
	wdr, err := db.GetWithdrawalsForWallet(addr)
	if err != nil {
		t.Error(err)
	}

	if len(wdr) != 2 {
		t.Errorf("expected 2 withdrawals, received %d", len(wdr))
	}

	// check withdrawal 1
	if wdr[1].Cid != nce {
		t.Errorf("cid is not %s, received %s", nce, wdr[1].Cid)
	}

	if wdr[1].Wallet != addr || wdr[0].Wallet != addr {
		t.Errorf("address is not %s, received %s", addr, wdr[1].Wallet)
	}

	if wdr[1].Amount != "50" || wdr[0].Amount != "100" {
		t.Errorf("amount is not 100, received %s", wdr[1].Amount)
	}

	// commit block 130
	err = db.CommitHeight(130)
	if err != nil {
		t.Error(err)
	}

	// balance should be 150, spent should be 50 since both withdrawals expired
	balance, err = db.GetBalance(addr)
	if err != nil {
		t.Error(err)
	}

	if balance.Cmp(big.NewInt(150)) != 0 {
		t.Error("balance is not 150")
	}

	// check spent is 50
	spent, err = db.GetSpent(addr)
	if err != nil {
		t.Error(err)
	}

	if spent.Cmp(big.NewInt(50)) != 0 {
		t.Error("spent is not 50")
	}
}

func generateNonce(l uint8) string {
	const chars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	result := make([]byte, l)
	for i := uint8(0); i < l; i++ {
		result[i] = chars[rand.Intn(len(chars))]
	}
	return string(result)
}
