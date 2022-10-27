package processor_test

import (
	"kwil/x/deposits/processor"
	"kwil/x/logx"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestProcessor(t *testing.T) {
	l := logx.New()

	p := processor.NewProcessor(l)

	// Test deposit
	md := mockDeposit{
		caller: "bennan",
		amount: "100",
	}
	err := p.ProcessDeposit(md)
	if err != nil {
		panic(err)
	}

	bal := p.GetBalance(md.Caller())

	assert.Equal(t, "100", bal.String())

	// test spend
	ms := mockSpend{
		caller: "bennan",
		amount: "50",
	}

	err = p.ProcessSpend(ms)
	if err != nil {
		panic(err)
	}

	bal = p.GetBalance(ms.Caller())
	assert.Equal(t, "50", bal.String())

	// check spent
	spent := p.GetSpent(ms.Caller())
	assert.Equal(t, "50", spent.String())

	// run gc and check balance and spent again
	p.RunGC()

	bal = p.GetBalance(ms.Caller())
	assert.Equal(t, "50", bal.String())

	// check spent
	spent = p.GetSpent(ms.Caller())
	assert.Equal(t, "50", spent.String())

	// test withdrawal request
	mwr := mockWithdrawalRequest{
		wallet:     "bennan",
		amount:     "50",
		nonce:      "n1",
		expiration: 100,
	}

	err = p.ProcessWithdrawalRequest(mwr)
	if err != nil {
		panic(err)
	}

	bal = p.GetBalance(mwr.Wallet())
	assert.Equal(t, "0", bal.String())

	// now deposit some more
	err = p.ProcessDeposit(md)
	if err != nil {
		panic(err)
	}

	// run the gc
	p.RunGC()

	// check the balance
	bal = p.GetBalance(mwr.Wallet())
	assert.Equal(t, "100", bal.String())

	// now withdraw 200
	mwr = mockWithdrawalRequest{
		wallet:     "bennan",
		amount:     "200",
		nonce:      "n2",
		expiration: 100,
	}

	err = p.ProcessWithdrawalRequest(mwr)
	if err != nil {
		panic(err)
	}

	bal = p.GetBalance(mwr.Wallet())
	assert.Equal(t, "0", bal.String())

	// there should now be two withdrawals
	// lets try getting them

	exs := p.NonceExist("n1")
	assert.True(t, exs)

	exs = p.NonceExist("n2")
	assert.True(t, exs)

	// try spending money I don't have
	err = p.ProcessSpend(ms)
	assert.Equal(t, processor.ErrInsufficientBalance, err)

	// test withdraw confirmation
	mwc := mockWithdrawalConfirmation{
		nonce: "n1",
	}

	p.ProcessWithdrawalConfirmation(mwc)

	// now test with a nonexistent nonce
	mwc = mockWithdrawalConfirmation{
		nonce: "ncewc3",
	}

	// this will throw memory error if it fails
	p.ProcessWithdrawalConfirmation(mwc)
}

func TestBadParsing(t *testing.T) {
	l := logx.New()

	p := processor.NewProcessor(l)

	// test depositing a bad amount
	md := mockDeposit{
		caller: "bennan",
		amount: "1d0",
	}

	err := p.ProcessDeposit(md)
	assert.Equal(t, processor.ErrCantParseAmount, err)

	// test spending a bad amount
	ms := mockSpend{
		caller: "bennan",
		amount: "1fr4e0",
	}
	err = p.ProcessSpend(ms)
	assert.Equal(t, processor.ErrCantParseAmount, err)

	// test withdrawal request with bad amount
	mwr := mockWithdrawalRequest{
		wallet:     "bennan",
		amount:     "1fr4e0",
		nonce:      "n1",
		expiration: 100,
	}

	err = p.ProcessWithdrawalRequest(mwr)
	assert.Equal(t, processor.ErrCantParseAmount, err)
}

func TestWithdrawNoMoney(t *testing.T) {
	l := logx.New()

	p := processor.NewProcessor(l)

	mwr := mockWithdrawalRequest{
		wallet:     "bennan",
		amount:     "100",
		nonce:      "n1",
		expiration: 100,
	}

	// make sure i have no balance and spent
	bal := p.GetBalance(mwr.Wallet())
	assert.Equal(t, "0", bal.String())

	spent := p.GetSpent(mwr.Wallet())
	assert.Equal(t, "0", spent.String())

	err := p.ProcessWithdrawalRequest(mwr)
	assert.Equal(t, processor.ErrInsufficientBalance, err)
}

func TestExpiredWithdrawals(t *testing.T) {
	l := logx.New()

	p := processor.NewProcessor(l)

	// deposit some funds
	// Test deposit
	md := mockDeposit{
		caller: "bennan",
		amount: "1000",
	}
	err := p.ProcessDeposit(md)
	if err != nil {
		panic(err)
	}

	// spend some funds
	ms := mockSpend{
		caller: "bennan",
		amount: "20",
	}
	err = p.ProcessSpend(ms)
	if err != nil {
		panic(err)
	}

	// test withdrawal request
	mwr := mockWithdrawalRequest{
		wallet:     "bennan",
		amount:     "50",
		nonce:      "n1",
		expiration: 50,
	}
	err = p.ProcessWithdrawalRequest(mwr)
	if err != nil {
		panic(err)
	}

	// another
	mwr = mockWithdrawalRequest{
		wallet:     "bennan",
		amount:     "200",
		nonce:      "n2",
		expiration: 100,
	}
	err = p.ProcessWithdrawalRequest(mwr)
	if err != nil {
		panic(err)
	}

	// now expire first one
	mfb := mockFinalizedBlock{
		height: 50,
	}

	err = p.ProcessFinalizedBlock(mfb)
	if err != nil {
		panic(err)
	}

	// n1 should not exist, n2 should
	exs := p.NonceExist("n1")
	assert.False(t, exs)

	exs = p.NonceExist("n2")
	assert.True(t, exs)

	// now add another
	mwr = mockWithdrawalRequest{
		wallet:     "bennan",
		amount:     "200",
		nonce:      "n3",
		expiration: 150,
	}

	err = p.ProcessWithdrawalRequest(mwr)
	if err != nil {
		panic(err)
	}

	// now expire all
	mfb = mockFinalizedBlock{
		height: 200,
	}

	err = p.ProcessFinalizedBlock(mfb)
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

type mockDeposit struct {
	caller string
	amount string
}

func (m mockDeposit) Caller() string {
	return m.caller
}

func (m mockDeposit) Amount() string {
	return m.amount
}

type mockSpend struct {
	caller string
	amount string
}

func (m mockSpend) Caller() string {
	return m.caller
}

func (m mockSpend) Amount() string {
	return m.amount
}

type mockWithdrawalRequest struct {
	wallet     string
	amount     string
	nonce      string
	expiration int64
}

func (m mockWithdrawalRequest) Wallet() string {
	return m.wallet
}

func (m mockWithdrawalRequest) Amount() string {
	return m.amount
}

func (m mockWithdrawalRequest) Nonce() string {
	return m.nonce
}

func (m mockWithdrawalRequest) Expiration() int64 {
	return m.expiration
}

type mockWithdrawalConfirmation struct {
	nonce string
}

func (m mockWithdrawalConfirmation) Nonce() string {
	return m.nonce
}

type mockFinalizedBlock struct {
	height int64
}

func (m mockFinalizedBlock) Height() int64 {
	return m.height
}
