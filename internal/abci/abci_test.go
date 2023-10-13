package abci

import (
	"bytes"
	"context"
	"math/big"
	"testing"

	"github.com/kwilteam/kwil-db/core/crypto/auth"
	"github.com/kwilteam/kwil-db/core/log"
	"github.com/kwilteam/kwil-db/core/types/transactions"
	"github.com/kwilteam/kwil-db/internal/accounts"

	"github.com/stretchr/testify/assert"
)

func marshalTx(t *testing.T, tx *transactions.Transaction) []byte {
	b, err := tx.MarshalBinary()
	if err != nil {
		t.Fatalf("could not marshal transaction! %v", err)
	}
	return b
}

func cloneTx(tx *transactions.Transaction) *transactions.Transaction {
	sig := make([]byte, len(tx.Signature.Signature))
	copy(sig, tx.Signature.Signature)
	sender := make([]byte, len(tx.Sender))
	copy(sender, tx.Sender)
	body := *tx.Body // same nonce
	body.Fee = big.NewInt(0).Set(tx.Body.Fee)
	body.Payload = make([]byte, len(tx.Body.Payload))
	copy(body.Payload, tx.Body.Payload)
	return &transactions.Transaction{
		Signature: &auth.Signature{
			Signature: sig,
			Type:      tx.Signature.Type,
		},
		Body:          &body,
		Serialization: tx.Serialization,
		Sender:        sender,
	}
}

type MockAccountStore struct {
}

func (m *MockAccountStore) GetAccount(ctx context.Context, pubKey []byte) (*accounts.Account, error) {
	return &accounts.Account{
		PublicKey: nil,
		Balance:   big.NewInt(0),
		Nonce:     0,
	}, nil
}

func newTxBts(t *testing.T, nonce uint64, sender string) []byte {
	tx := &transactions.Transaction{
		Signature: &auth.Signature{},
		Body: &transactions.TransactionBody{
			Description: "test",
			Payload:     []byte(`random payload`),
			Fee:         big.NewInt(0),
			Nonce:       nonce,
		},
		Sender: []byte(sender),
	}

	bts, err := tx.MarshalBinary()
	if err != nil {
		t.Fatalf("could not marshal transaction! %v", err)
	}
	return bts
}

func newTx(t *testing.T, nonce uint64, sender string) *transactions.Transaction {
	return &transactions.Transaction{
		Signature: &auth.Signature{},
		Body: &transactions.TransactionBody{
			Description: "test",
			Payload:     []byte(`random payload`),
			Fee:         big.NewInt(0),
			Nonce:       nonce,
		},
		Sender: []byte(sender),
	}
}

func Test_prepareMempoolTxns(t *testing.T) {
	// To make these tests deterministic, we manually craft certain misorderings
	// and the known expected orderings. Also include some malformed
	// transactions that fail to unmarshal, which really shouldn't happen if the
	// initial check passed but there is graceful handling of this in the code.

	// tA is the template transaction. Several fields may not be nil because of
	// a legacy RLP issue where objects may be encoded that cannot be decoded.
	tA := &transactions.Transaction{
		Signature: &auth.Signature{
			Signature: []byte{},
			Type:      auth.Ed25519Auth,
		},
		Body: &transactions.TransactionBody{
			Description: "t",
			Payload:     []byte(`x`),
			Fee:         big.NewInt(0),
			Nonce:       0,
		},
		Sender: []byte(`guy`),
	}
	tAb := marshalTx(t, tA)

	// same sender, incremented nonce
	tB := cloneTx(tA)
	tB.Body.Nonce++
	tBb := marshalTx(t, tB)

	nextTx := func(tx *transactions.Transaction) *transactions.Transaction {
		tx2 := cloneTx(tx)
		tx2.Body.Nonce++
		return tx2
	}

	// second party
	tOtherSenderA := cloneTx(tA)
	tOtherSenderA.Sender = []byte(`otherguy`)
	tOtherSenderAb := marshalTx(t, tOtherSenderA)

	// Same nonce tx, different body (diff bytes)
	tOtherSenderAbDup := cloneTx(tOtherSenderA)
	tOtherSenderAbDup.Body.Description = "dup" // not "t"
	tOtherSenderAbDupb := marshalTx(t, tOtherSenderAbDup)

	tOtherSenderB := nextTx(tOtherSenderA)
	tOtherSenderBb := marshalTx(t, tOtherSenderB)

	tOtherSenderC := nextTx(tOtherSenderB)
	tOtherSenderCb := marshalTx(t, tOtherSenderC)

	invalid := []byte{9, 90, 22}

	logger := log.NewStdOut(log.DebugLevel)
	tests := []struct {
		name string
		txs  [][]byte
		want [][]byte
	}{
		{
			"empty",
			[][]byte{},
			[][]byte{},
		},
		{
			"one and only invalid",
			[][]byte{invalid},
			[][]byte{},
		},
		{
			"one of two invalid",
			[][]byte{invalid, tBb},
			[][]byte{tBb},
		},
		{
			"one valid",
			[][]byte{tAb},
			[][]byte{tAb},
		},
		{
			"two valid",
			[][]byte{tAb, tBb},
			[][]byte{tAb, tBb},
		},
		{
			"two valid misordered",
			[][]byte{tBb, tAb},
			[][]byte{tAb, tBb},
		},
		{
			"multi-party, one misordered, stable",
			[][]byte{tOtherSenderAb, tBb, tOtherSenderBb, tAb},
			[][]byte{tOtherSenderAb, tAb, tOtherSenderBb, tBb},
		},
		{
			"multi-party, one misordered, one dup nonce, stable",
			[][]byte{tOtherSenderAb, tOtherSenderAbDupb, tBb, tAb},
			[][]byte{tOtherSenderAb, tOtherSenderAbDupb, tAb, tBb},
		},
		{
			"multi-party, both misordered, stable",
			[][]byte{tOtherSenderBb, tBb, tOtherSenderAb, tAb},
			[][]byte{tOtherSenderAb, tAb, tOtherSenderBb, tBb},
		},
		{
			"multi-party, both misordered, alt. stable",
			[][]byte{tBb, tOtherSenderBb, tOtherSenderAb, tAb},
			[][]byte{tAb, tOtherSenderAb, tOtherSenderBb, tBb},
		},
		{
			"multi-party, big, with invalid in middle",
			[][]byte{tOtherSenderCb, tBb, invalid, tOtherSenderBb, tOtherSenderAb, tAb},
			[][]byte{tOtherSenderAb, tAb, tOtherSenderBb, tOtherSenderCb, tBb},
		},
		{
			"multi-party, big, already correct",
			[][]byte{tOtherSenderAb, tAb, tOtherSenderBb, tOtherSenderCb, tBb},
			[][]byte{tOtherSenderAb, tAb, tOtherSenderBb, tOtherSenderCb, tBb},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := prepareMempoolTxns(tt.txs, 1e6, &logger)
			if len(got) != len(tt.want) {
				t.Errorf("got %d txns, expected %d", len(got), len(tt.want))
			}
			for i, txi := range got {
				if !bytes.Equal(txi, tt.want[i]) {
					t.Errorf("mismatched tx %d", i)
				}
			}
		})
	}
}

func Test_ProcessProposal_TxValidation(t *testing.T) {
	ctx := context.Background()
	abciApp := &AbciApp{
		accounts: &MockAccountStore{},
	}
	logger := log.NewStdOut(log.DebugLevel)

	abciApp.log = logger

	txA1 := newTxBts(t, 1, "A")
	txA2 := newTxBts(t, 2, "A")
	txA3 := newTxBts(t, 3, "A")
	txA4 := newTxBts(t, 4, "A")
	txB1 := newTxBts(t, 1, "B")
	txB2 := newTxBts(t, 2, "B")
	txB3 := newTxBts(t, 3, "B")

	testcases := []struct {
		name string
		txs  [][]byte
		err  bool // expect error
	}{
		{
			// Invalid ordering of nonces: 3, 1, 2 by single sender
			name: "Invalid ordering on nonces by single sender",
			txs: [][]byte{
				txA3,
				txA1,
				txA2,
			},
			err: true,
		},
		{
			// A1, A3, B3, A2, B1, B2
			name: "Invalid ordering of nonces by multiple senders",
			txs: [][]byte{
				txA1,
				txA3,
				txB3,
				txA2,
				txB1,
				txB2,
			},
			err: true,
		},
		{
			// Gaps in the ordering of nonces: 1, 3, 4  by single sender
			name: "Gaps in the ordering of nonces by single sender",
			txs: [][]byte{
				txA1,
				txA3,
				txA4,
			},
			err: true,
		},
		{
			// Gaps in the ordering of nonces by multiple senders
			name: "Gaps in the ordering of nonces by multiple senders",
			txs: [][]byte{
				txA1,
				txB1,
				txA4,
				txB3,
			},
			err: true,
		},
		{
			// Duplicate Nonces: 1, 2, 2  by single sender
			name: "Duplicate Nonces by single sender",
			txs: [][]byte{
				txA1,
				txA2,
				txA2,
			},
			err: true,
		},
		{
			// Duplicate Nonces: 1, 2, 2  by multiple senders
			name: "Duplicate Nonces by multiple senders",
			txs: [][]byte{
				txA1,
				txA2,
				txB1,
				txB2,
				txB2,
			},
			err: true,
		},
		{
			// Valid ordering of nonces: 1, 2, 3  by single sender
			name: "Valid ordering of nonces by single sender",
			txs: [][]byte{
				txA1,
				txA2,
				txA3,
			},
			err: false,
		},
		{
			// Valid ordering of nonces: 1, 2, 3  by multiple senders
			name: "Valid ordering of nonces by multiple senders",
			txs: [][]byte{
				txA1,
				txA2,
				txB1,
				txB2,
				txA3,
				txB3,
			},
			err: false,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			err := abciApp.validateProposalTransactions(ctx, tc.txs)
			if tc.err {
				assert.Error(t, err, "expected error due to %s", tc.name)
			} else {
				assert.NoError(t, err, "TC: %s, expected no error", tc.name)
			}
		})
	}
}

func Test_CheckTx(t *testing.T) {
	m := &mempool{
		accountStore: &MockAccountStore{},
		accounts:     make(map[string]*userAccount),
		nonceTracker: make(map[string]uint64),
	}
	ctx := context.Background()

	// Successful transaction A: 1
	err := m.applyTransaction(ctx, newTx(t, 1, "A"))
	assert.NoError(t, err)
	assert.EqualValues(t, m.nonceTracker["A"], 1)

	// Successful transaction A: 2
	err = m.applyTransaction(ctx, newTx(t, 2, "A"))
	assert.NoError(t, err)
	assert.EqualValues(t, m.nonceTracker["A"], 2)

	// Duplicate nonce failure
	err = m.applyTransaction(ctx, newTx(t, 2, "A"))
	assert.Error(t, err)
	assert.EqualValues(t, m.nonceTracker["A"], 2)

	// Invalid order
	err = m.applyTransaction(ctx, newTx(t, 4, "A"))
	assert.Error(t, err)
	assert.EqualValues(t, m.nonceTracker["A"], 2)

	err = m.applyTransaction(ctx, newTx(t, 3, "A"))
	assert.NoError(t, err)
	assert.EqualValues(t, m.nonceTracker["A"], 3)

	// Recheck nonce 4 transaction
	err = m.applyTransaction(ctx, newTx(t, 4, "A"))
	assert.NoError(t, err)
	assert.EqualValues(t, m.nonceTracker["A"], 4)
}
