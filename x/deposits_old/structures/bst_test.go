package structures_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"kwil/x/deposits/structures"
)

type mockWithdrawal struct {
	exp    int64
	nonce  string
	wallet string
	amount string
	spent  string
}

func (m *mockWithdrawal) Expiration() int64 {
	return m.exp
}

func (m *mockWithdrawal) Nonce() string {
	return m.nonce
}

func (m *mockWithdrawal) Wallet() string {
	return m.wallet
}

func (m *mockWithdrawal) Amount() string {
	return m.amount
}

func (m *mockWithdrawal) Spent() string {
	return m.spent
}

func newMockWithdrawal(exp int64, nonce string) *mockWithdrawal {
	return &mockWithdrawal{
		exp:    exp,
		nonce:  nonce,
		wallet: "wallet",
		amount: "10",
		spent:  "5",
	}
}

func TestBST(t *testing.T) {
	bst := structures.NewBST()
	bst.Insert(1, newMockWithdrawal(1, "1"))
	bst.Insert(2, newMockWithdrawal(2, "2"))
	bst.Insert(3, newMockWithdrawal(3, "3"))
	bst.Insert(4, newMockWithdrawal(4, "4"))
	bst.Insert(5, newMockWithdrawal(5, "5"))
	bst.Insert(6, newMockWithdrawal(8, "8"))
	bst.Insert(7, newMockWithdrawal(7, "7"))
	bst.Insert(8, newMockWithdrawal(6, "6"))

	assert.Equal(t, int64(1), bst.Get(1).Item().Expiration())
	assert.Equal(t, int64(2), bst.Get(2).Item().Expiration())

	min := bst.Min()
	assert.Equal(t, int64(1), min.Key())
	bst.Remove(1)
	min = bst.Min()
	assert.Equal(t, int64(2), min.Key())
	assert.Equal(t, int64(2), min.Item().Expiration())

}
