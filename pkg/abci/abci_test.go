package abci

import (
	"bytes"
	"math/big"
	"testing"

	"github.com/kwilteam/kwil-db/pkg/auth"
	"github.com/kwilteam/kwil-db/pkg/log"
	"github.com/kwilteam/kwil-db/pkg/transactions"
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
