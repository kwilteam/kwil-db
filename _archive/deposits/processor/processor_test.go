package processor_test

import (
	"kwil/_archive/deposits/processor"
	dt "kwil/x/deposits/types"
	"kwil/x/logx"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestProcessor(t *testing.T) {
	l := logx.New()

	p := processor.NewProcessor(l)

	// Test deposit
	md := dt.Deposit{
		Caller: "bennan",
		Amount: "100",
	}
	err := p.ProcessDeposit(&md)
	if err != nil {
		panic(err)
	}

	bal := p.GetBalance(md.Caller)

	assert.Equal(t, "100", bal.String())

	// test spend
	ms := dt.Spend{
		Caller: "bennan",
		Amount: "50",
	}

	err = p.ProcessSpend(&ms)
	if err != nil {
		panic(err)
	}

	bal = p.GetBalance(ms.Caller)
	assert.Equal(t, "50", bal.String())

	// check spent
	spent := p.GetSpent(ms.Caller)
	assert.Equal(t, "50", spent.String())

	// run gc and check balance and spent again
	p.RunGC()

	bal = p.GetBalance(ms.Caller)
	assert.Equal(t, "50", bal.String())

	// check spent
	spent = p.GetSpent(ms.Caller)
	assert.Equal(t, "50", spent.String())

	// test withdrawal request
	mwr := dt.WithdrawalRequest{
		Wallet:     "bennan",
		Amount:     "50",
		Cid:        "n1",
		Expiration: 100,
	}

	err = p.ProcessWithdrawalRequest(&mwr)
	if err != nil {
		panic(err)
	}

	bal = p.GetBalance(mwr.Wallet)
	assert.Equal(t, "0", bal.String())

	// now deposit some more
	err = p.ProcessDeposit(&md)
	if err != nil {
		panic(err)
	}

	// run the gc
	p.RunGC()

	// check the balance
	bal = p.GetBalance(mwr.Wallet)
	assert.Equal(t, "100", bal.String())

	// now withdraw 200
	mwr = dt.WithdrawalRequest{
		Wallet:     "bennan",
		Amount:     "200",
		Cid:        "n2",
		Expiration: 100,
	}

	err = p.ProcessWithdrawalRequest(&mwr)
	if err != nil {
		panic(err)
	}

	bal = p.GetBalance(mwr.Wallet)
	assert.Equal(t, "0", bal.String())

	// there should now be two withdrawals
	// lets try getting them

	exs := p.NonceExist("n1")
	assert.True(t, exs)

	exs = p.NonceExist("n2")
	assert.True(t, exs)

	// try spending money I don't have
	err = p.ProcessSpend(&ms)
	assert.Equal(t, processor.ErrInsufficientBalance, err)

	// test withdraw confirmation
	mwc := dt.WithdrawalConfirmation{
		Cid: "n1",
	}

	p.ProcessWithdrawalConfirmation(&mwc)

	// now test with a nonexistent nonce
	mwc = dt.WithdrawalConfirmation{
		Cid: "ncewc3",
	}

	// this will throw memory error if it fails
	p.ProcessWithdrawalConfirmation(&mwc)
}

func TestBadParsing(t *testing.T) {
	l := logx.New()

	p := processor.NewProcessor(l)

	// test depositing a bad amount
	md := dt.Deposit{
		Caller: "bennan",
		Amount: "1d0",
	}

	err := p.ProcessDeposit(&md)
	assert.Equal(t, processor.ErrCantParseAmount, err)

	// test spending a bad amount
	ms := dt.Spend{
		Caller: "bennan",
		Amount: "1fr4e0",
	}
	err = p.ProcessSpend(&ms)
	assert.Equal(t, processor.ErrCantParseAmount, err)

	// test withdrawal request with bad amount
	mwr := dt.WithdrawalRequest{
		Wallet:     "bennan",
		Amount:     "1fr4e0",
		Cid:        "n1",
		Expiration: 100,
	}

	err = p.ProcessWithdrawalRequest(&mwr)
	assert.Equal(t, processor.ErrCantParseAmount, err)
}

func TestWithdrawNoMoney(t *testing.T) {
	l := logx.New()

	p := processor.NewProcessor(l)

	mwr := dt.WithdrawalRequest{
		Wallet:     "bennan",
		Amount:     "100",
		Cid:        "n1",
		Expiration: 100,
	}

	// make sure i have no balance and spent
	bal := p.GetBalance(mwr.Wallet)
	assert.Equal(t, "0", bal.String())

	spent := p.GetSpent(mwr.Wallet)
	assert.Equal(t, "0", spent.String())

	err := p.ProcessWithdrawalRequest(&mwr)
	assert.Equal(t, processor.ErrInsufficientBalance, err)
}

func TestExpiredWithdrawals(t *testing.T) {
	l := logx.New()

	p := processor.NewProcessor(l)

	// deposit some funds
	// Test deposit
	md := dt.Deposit{
		Caller: "bennan",
		Amount: "1000",
	}
	err := p.ProcessDeposit(&md)
	if err != nil {
		panic(err)
	}

	// spend some funds
	ms := dt.Spend{
		Caller: "bennan",
		Amount: "20",
	}
	err = p.ProcessSpend(&ms)
	if err != nil {
		panic(err)
	}

	// test withdrawal request
	mwr := dt.WithdrawalRequest{
		Wallet:     "bennan",
		Amount:     "50",
		Cid:        "n1",
		Expiration: 50,
	}
	err = p.ProcessWithdrawalRequest(&mwr)
	if err != nil {
		panic(err)
	}

	// another
	mwr = dt.WithdrawalRequest{
		Wallet:     "bennan",
		Amount:     "200",
		Cid:        "n2",
		Expiration: 100,
	}
	err = p.ProcessWithdrawalRequest(&mwr)
	if err != nil {
		panic(err)
	}

	// now expire first one
	meb := dt.EndBlock{
		Height: 50,
	}

	err = p.ProcessEndBlock(&meb)
	if err != nil {
		panic(err)
	}

	// n1 should not exist, n2 should
	exs := p.NonceExist("n1")
	assert.False(t, exs)

	exs = p.NonceExist("n2")
	assert.True(t, exs)

	// now add another
	mwr = dt.WithdrawalRequest{
		Wallet:     "bennan",
		Amount:     "200",
		Cid:        "n3",
		Expiration: 150,
	}

	err = p.ProcessWithdrawalRequest(&mwr)
	if err != nil {
		panic(err)
	}

	// now expire all
	meb = dt.EndBlock{
		Height: 200,
	}

	err = p.ProcessEndBlock(&meb)
	if err != nil {
		panic(err)
	}

	// n1 should not exist, n2 should not, n3 should not
	exs = p.NonceExist("n1")
	assert.False(t, exs)

	exs = p.NonceExist("n2")
	assert.False(t, exs)

	exs = p.NonceExist("n3")
	assert.False(t, exs)

}
