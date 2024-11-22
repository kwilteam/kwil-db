package abci

import (
	"bytes"
	"context"
	"math/big"
	"testing"

	"github.com/kwilteam/kwil-db/common"
	"github.com/kwilteam/kwil-db/common/sql"
	"github.com/kwilteam/kwil-db/core/crypto"
	"github.com/kwilteam/kwil-db/core/crypto/auth"
	"github.com/kwilteam/kwil-db/core/log"
	"github.com/kwilteam/kwil-db/core/types/transactions"

	"github.com/kwilteam/kwil-db/core/types"
	"github.com/kwilteam/kwil-db/internal/txapp"
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

func newTxBts(t *testing.T, nonce uint64, signer auth.Signer) []byte {
	tx := &transactions.Transaction{
		Signature:     &auth.Signature{},
		Serialization: transactions.SignedMsgConcat,
		Body: &transactions.TransactionBody{
			Description: "test",
			Payload:     []byte(`random payload`),
			Fee:         big.NewInt(0),
			Nonce:       nonce,
		},
		Sender: signer.Identity(),
	}

	msg, err := tx.SerializeMsg()
	if err != nil {
		t.Fatalf("serialization failed: %v", err)
	}
	tx.Signature, err = signer.Sign(msg)
	if err != nil {
		t.Fatalf("signing failed: %v", err)
	}

	bts, err := tx.MarshalBinary()
	if err != nil {
		t.Fatalf("could not marshal transaction! %v", err)
	}
	return bts
}

func Test_prepareMempoolTxns(t *testing.T) {
	// To make these tests deterministic, we manually craft certain misorderings
	// and the known expected orderings. Also include some malformed
	// transactions that fail to unmarshal, which really shouldn't happen if the
	// initial check passed but there is graceful handling of this in the code.

	// tA is the template transaction. Several fields may not be nil because of
	// a legacy RLP issue where objects may be encoded that cannot be decoded.

	abciApp := &AbciApp{
		txApp: &mockTxApp{},
		db:    &mockDB{},
	}
	logger := log.NewStdOut(log.DebugLevel)

	abciApp.log = logger

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

	// proposer party
	tProposer := cloneTx(tA)
	tProposer.Sender = []byte(`proposer`)
	tProposerb := marshalTx(t, tProposer)

	invalid := []byte{9, 90, 22}

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
			[][]byte{tOtherSenderAb, tAb, tBb},
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
		{
			"multi-party,proposer in the last, reorder",
			[][]byte{tOtherSenderAb, tAb, tOtherSenderBb, tOtherSenderCb, tBb, tProposerb},
			[][]byte{tProposerb, tOtherSenderAb, tAb, tOtherSenderBb, tOtherSenderCb, tBb},
		},
		{
			"multi-party,proposer in the middle, reorder",
			[][]byte{tOtherSenderAb, tAb, tOtherSenderBb, tProposerb, tOtherSenderCb, tBb},
			[][]byte{tProposerb, tOtherSenderAb, tAb, tOtherSenderBb, tOtherSenderCb, tBb},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := abciApp.prepareBlockTransactions(context.Background(), tt.txs, &logger, 1e6, []byte("proposer"), 0)
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

func Test_ProcessProposal_UnfundedAccount(t *testing.T) {
	abciApp := &AbciApp{
		txApp: &mockTxApp{},
		cfg: AbciConfig{
			GasEnabled: true,
		},
		db: &mockDB{},
	}
	logger := log.NewStdOut(log.DebugLevel)

	abciApp.log = logger

	keyA, _ := crypto.GenerateSecp256k1Key()
	signerA := &auth.EthPersonalSigner{Key: *keyA}

	txA1 := newTxBts(t, 1, signerA)

	// Unfunded account
	txs := abciApp.prepareBlockTransactions(context.Background(), [][]byte{txA1}, &logger, 1e6, []byte("proposer"), 0)
	assert.Len(t, txs, 0)

}

func Test_ProcessProposal_TxValidation(t *testing.T) {
	ctx := context.Background()
	abciApp := &AbciApp{
		txApp: &mockTxApp{},
		db:    &mockDB{},
	}
	logger := log.NewStdOut(log.DebugLevel)

	abciApp.log = logger

	keyA, _ := crypto.GenerateSecp256k1Key()
	signerA := &auth.EthPersonalSigner{Key: *keyA}
	keyB, _ := crypto.GenerateSecp256k1Key()
	signerB := &auth.EthPersonalSigner{Key: *keyB}

	txA1 := newTxBts(t, 1, signerA)
	txA2 := newTxBts(t, 2, signerA)
	txA3 := newTxBts(t, 3, signerA)
	txA4 := newTxBts(t, 4, signerA)
	txB1 := newTxBts(t, 1, signerB)
	txB2 := newTxBts(t, 2, signerB)
	txB3 := newTxBts(t, 3, signerB)

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
			err := abciApp.validateProposalTransactions(ctx, tc.txs, nil)
			if tc.err {
				assert.Error(t, err, "expected error due to %s", tc.name)
			} else {
				assert.NoError(t, err, "TC: %s, expected no error", tc.name)
			}
		})
	}
}

type mockTxApp struct{}

func (m *mockTxApp) MarkBroadcasted(ctx context.Context, ids []types.UUID) error {
	return nil
}

func (m *mockTxApp) AccountInfo(ctx context.Context, db sql.DB, acctID []byte, getUncommitted bool) (balance *big.Int, nonce int64, err error) {
	return big.NewInt(0), 0, nil
}

func (m *mockTxApp) ApplyMempool(ctx *common.TxContext, db sql.DB, tx *transactions.Transaction) error {
	return nil
}

func (m *mockTxApp) Begin(ctx context.Context, height int64) error {
	return nil
}

func (m *mockTxApp) Commit(ctx context.Context) {}

func (m *mockTxApp) Execute(ctx *common.TxContext, db sql.DB, tx *transactions.Transaction) *txapp.TxResponse {
	return nil
}

func (m *mockTxApp) Finalize(ctx context.Context, db sql.DB, block *common.BlockContext) (validatorUpgrades []*types.Validator, approvedJoins, expiredJoins [][]byte, err error) {
	return nil, nil, nil, nil
}

func (m *mockTxApp) GenesisInit(ctx context.Context, db sql.DB, validators []*types.Validator, accounts []*types.Account,
	initialHeight int64, chain *common.ChainContext) error {
	return nil
}

func (m *mockTxApp) GetValidators(ctx context.Context, db sql.DB) ([]*types.Validator, error) {
	return nil, nil
}

func (m *mockTxApp) ProposerTxs(ctx context.Context, db sql.DB, txNonce uint64, maxTxSz int64, block *common.BlockContext) ([][]byte, error) {
	return nil, nil
}

func (m *mockTxApp) UpdateValidator(ctx context.Context, db sql.DB, validator []byte, power int64) error {
	return nil
}

func (m *mockTxApp) Reload(ctx context.Context, db sql.DB) error {
	return nil
}

func (m *mockTxApp) Price(ctx context.Context, db sql.DB, tx *transactions.Transaction, c *common.ChainContext) (*big.Int, error) {
	return big.NewInt(0), nil
}

type mockDB struct{}

func (m *mockDB) BeginPreparedTx(ctx context.Context) (sql.PreparedTx, error) {
	return &mockTx{}, nil
}

func (m *mockDB) BeginReadTx(ctx context.Context) (sql.OuterReadTx, error) {
	return &mockTx{}, nil
}

func (m *mockDB) BeginSnapshotTx(ctx context.Context) (sql.Tx, string, error) {
	return &mockTx{}, "", nil
}

func (m *mockDB) Execute(ctx context.Context, stmt string, args ...any) (*sql.ResultSet, error) {
	return nil, nil
}

func (m *mockDB) BeginTx(ctx context.Context) (sql.Tx, error) {
	return &mockTx{}, nil
}

func (m *mockDB) AutoCommit(on bool) {}

type mockTx struct{}

func (m *mockTx) Subscribe(ctx context.Context) (<-chan string, func(context.Context) error, error) {
	return make(<-chan string), func(ctx context.Context) error { return nil }, nil
}

func (m *mockTx) Rollback(ctx context.Context) error {
	return nil
}

func (m *mockTx) Commit(ctx context.Context) error {
	return nil
}

func (m *mockTx) Execute(ctx context.Context, stmt string, args ...any) (*sql.ResultSet, error) {
	return nil, nil
}

func (m *mockTx) BeginTx(ctx context.Context) (sql.Tx, error) {
	return &mockTx{}, nil
}

func (m *mockTx) Precommit(ctx context.Context, changes chan<- any) ([]byte, error) {
	return nil, nil
}
