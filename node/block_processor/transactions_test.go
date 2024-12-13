package blockprocessor

import (
	"bytes"
	"context"
	"math/big"
	"testing"

	"github.com/kwilteam/kwil-db/common"
	"github.com/kwilteam/kwil-db/config"
	"github.com/kwilteam/kwil-db/core/crypto"
	"github.com/kwilteam/kwil-db/core/crypto/auth"
	"github.com/kwilteam/kwil-db/core/log"
	"github.com/kwilteam/kwil-db/core/types"
	ktypes "github.com/kwilteam/kwil-db/core/types"
	"github.com/kwilteam/kwil-db/node/txapp"
	"github.com/kwilteam/kwil-db/node/types/sql"
	"github.com/stretchr/testify/require"
)

func marshalTx(t *testing.T, tx *types.Transaction) []byte {
	b, err := tx.MarshalBinary()
	if err != nil {
		t.Fatalf("could not marshal transaction! %v", err)
	}
	return b
}

func cloneTx(tx *types.Transaction) *types.Transaction {
	sig := make([]byte, len(tx.Signature.Data))
	copy(sig, tx.Signature.Data)
	sender := make([]byte, len(tx.Sender))
	copy(sender, tx.Sender)
	body := *tx.Body // same nonce
	body.Fee = big.NewInt(0).Set(tx.Body.Fee)
	body.Payload = make([]byte, len(tx.Body.Payload))
	copy(body.Payload, tx.Body.Payload)
	return &types.Transaction{
		Signature: &auth.Signature{
			Data: sig,
			Type: tx.Signature.Type,
		},
		Body:          &body,
		Serialization: tx.Serialization,
		Sender:        sender,
	}
}

func secp256k1Signer(t *testing.T) *auth.EthPersonalSigner {
	privKey, _, err := crypto.GenerateSecp256k1Key(nil)
	require.NoError(t, err)

	privKeyBytes := privKey.Bytes()
	k, err := crypto.UnmarshalSecp256k1PrivateKey(privKeyBytes)
	require.NoError(t, err)

	return &auth.EthPersonalSigner{Key: *k}
}

func TestPrepareMempoolTxns(t *testing.T) {
	// To make these tests deterministic, we manually craft certain misorderings
	// and the known expected orderings. Also include some malformed
	// transactions that fail to unmarshal, which really shouldn't happen if the
	// initial check passed but there is graceful handling of this in the code.

	// tA is the template transaction. Several fields may not be nil because of
	// a legacy RLP issue where objects may be encoded that cannot be decoded.

	bp := &BlockProcessor{
		db:     &mockDB{},
		log:    log.DiscardLogger,
		signer: secp256k1Signer(t),
		chainCtx: &common.ChainContext{
			ChainID: "test",
			NetworkParameters: &common.NetworkParameters{
				MaxBlockSize:     6 * 1024 * 1024,
				MaxVotesPerTx:    100,
				DisabledGasCosts: true,
			},
		},
		txapp: &mockTxApp{},
	}

	tA := &types.Transaction{
		Signature: &auth.Signature{
			Data: []byte{},
			Type: auth.Ed25519Auth,
		},
		Body: &types.TransactionBody{
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

	nextTx := func(tx *types.Transaction) *types.Transaction {
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
	tProposer.Sender = bp.signer.Identity()
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

	ctx := context.Background()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			getEvents = func(_ context.Context, _ sql.Executor) ([]*types.VotableEvent, error) {
				return nil, nil
			}

			got, _, err := bp.prepareBlockTransactions(ctx, tt.txs)
			require.NoError(t, err)
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

var (
	evt1 = &types.VotableEvent{
		Type: "test",
		Body: []byte("test"),
	}
	evt2 = &types.VotableEvent{
		Type: "test",
		Body: []byte("test2"),
	}
	evt3 = &types.VotableEvent{
		Type: "test",
		Body: []byte("test3"),
	}
)

func TestPrepareVoteIDTx(t *testing.T) {
	leader := secp256k1Signer(t)
	validator := secp256k1Signer(t)
	sentry := secp256k1Signer(t)
	valSet := []*ktypes.Validator{
		{
			PubKey: validator.Identity(),
			Power:  1,
		},
		{
			PubKey: leader.Identity(),
			Power:  1,
		},
	}
	genCfg := config.DefaultGenesisConfig()
	genCfg.Leader = leader.Identity()

	bp := &BlockProcessor{
		db:     &mockDB{},
		log:    log.DiscardLogger,
		signer: secp256k1Signer(t),
		chainCtx: &common.ChainContext{
			ChainID: "test",
			NetworkParameters: &common.NetworkParameters{
				MaxBlockSize:     6 * 1024 * 1024,
				MaxVotesPerTx:    100,
				DisabledGasCosts: true,
			},
		},
		txapp:         &mockTxApp{},
		genesisParams: genCfg,
	}

	testcases := []struct {
		name    string
		signer  auth.Signer
		events  []*types.VotableEvent
		cleanup func()
		fn      func(context.Context, *BlockProcessor, sql.DB, *mockEventStore) error
	}{
		{
			name:   "no voteIDs to broadcast",
			events: []*types.VotableEvent{}, // no events
			signer: validator,
			fn: func(ctx context.Context, bp *BlockProcessor, db sql.DB, es *mockEventStore) error {
				bp.signer = validator
				tx, ids, err := bp.PrepareValidatorVoteIDTx(ctx, db)
				require.NoError(t, err)
				require.Nil(t, tx)
				require.Nil(t, ids)
				return nil
			},
		},
		{
			name:   "leader not allowed to broadcast voteIDs",
			events: []*types.VotableEvent{evt1, evt2},
			signer: leader,
			fn: func(ctx context.Context, bp *BlockProcessor, db sql.DB, es *mockEventStore) error {
				tx, ids, err := bp.PrepareValidatorVoteIDTx(ctx, db)
				require.NoError(t, err)
				require.Nil(t, tx)
				require.Nil(t, ids)
				return nil
			},
		},
		{
			name:   "sentry node not allowed to broadcast voteIDs",
			events: []*types.VotableEvent{evt1, evt2},
			signer: sentry,
			fn: func(ctx context.Context, bp *BlockProcessor, db sql.DB, es *mockEventStore) error {
				tx, ids, err := bp.PrepareValidatorVoteIDTx(ctx, db)
				require.NoError(t, err)
				require.Nil(t, tx)
				require.Nil(t, ids)
				return nil
			},
		},
		{
			name:   "validator broadcasts voteIDs in gasless mode",
			signer: validator,
			events: []*types.VotableEvent{evt1, evt2},
			fn: func(ctx context.Context, bp *BlockProcessor, db sql.DB, es *mockEventStore) error {
				tx, ids, err := bp.PrepareValidatorVoteIDTx(ctx, db)
				require.NoError(t, err)
				require.NotNil(t, tx)
				require.NotNil(t, ids)
				require.Len(t, ids, 2)
				return nil
			},
		},
		{
			name:   "insufficient gas to broadcast voteIDs",
			signer: validator,
			events: []*types.VotableEvent{evt1, evt2},
			fn: func(ctx context.Context, bp *BlockProcessor, db sql.DB, es *mockEventStore) error {
				bp.chainCtx.NetworkParameters.DisabledGasCosts = false
				// set price of tx high: 1000
				price = big.NewInt(1000)
				tx, ids, err := bp.PrepareValidatorVoteIDTx(ctx, db)

				require.NoError(t, err)
				require.Nil(t, tx)
				require.Nil(t, ids)
				return nil
			},
			cleanup: func() {
				price = big.NewInt(0)
				bp.chainCtx.NetworkParameters.DisabledGasCosts = true
			},
		},
		{
			name:   "validator has sufficient gas to broadcast voteIDs",
			signer: validator,
			events: []*types.VotableEvent{evt1, evt2},
			fn: func(ctx context.Context, bp *BlockProcessor, db sql.DB, es *mockEventStore) error {
				bp.chainCtx.NetworkParameters.DisabledGasCosts = false
				// set price of tx low: 1
				price = big.NewInt(1000)
				accountBalance = big.NewInt(1000)
				tx, ids, err := bp.PrepareValidatorVoteIDTx(ctx, db)

				require.NoError(t, err)
				require.NotNil(t, tx)
				require.NotNil(t, ids)
				require.Len(t, ids, 2)
				return nil
			},
			cleanup: func() {
				price = big.NewInt(0)
				accountBalance = big.NewInt(0)
				bp.chainCtx.NetworkParameters.DisabledGasCosts = true
			},
		},
		{
			name:   "mark broadcasted for broadcasted voteIDs",
			signer: validator,
			events: []*types.VotableEvent{evt1, evt2},
			fn: func(ctx context.Context, bp *BlockProcessor, db sql.DB, es *mockEventStore) error {
				tx, ids, err := bp.PrepareValidatorVoteIDTx(ctx, db)
				require.NoError(t, err)
				require.NotNil(t, tx)
				require.NotNil(t, ids)
				require.Len(t, ids, 2)

				err = bp.events.MarkBroadcasted(ctx, ids)
				require.NoError(t, err)

				// now no more events to broadcast
				tx, ids, err = bp.PrepareValidatorVoteIDTx(ctx, db)
				require.NoError(t, err)
				require.Nil(t, tx)
				require.Nil(t, ids)

				// add more events
				es.addEvent(evt3)
				tx, ids, err = bp.PrepareValidatorVoteIDTx(ctx, db)
				require.NoError(t, err)
				require.NotNil(t, tx)
				require.NotNil(t, ids)
				require.Len(t, ids, 1)

				return nil
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			defer func() {
				if tc.cleanup != nil {
					tc.cleanup()
				}
			}()

			db := &mockDB{}
			bp.signer = tc.signer
			es := newMockEventStore(tc.events)
			bp.events = es
			bp.validators = newValidatorStore(valSet)

			ctx := context.Background()
			err := tc.fn(ctx, bp, db, es) // run the test function
			require.NoError(t, err)
		})
	}
}

func TestPrepareVoteBodyTx(t *testing.T) {
	signer := secp256k1Signer(t)
	genCfg := config.DefaultGenesisConfig()
	genCfg.Leader = signer.Identity()

	bp := &BlockProcessor{
		db:     &mockDB{},
		log:    log.DiscardLogger,
		signer: signer,
		chainCtx: &common.ChainContext{
			ChainID: "test",
			NetworkParameters: &common.NetworkParameters{
				MaxBlockSize:     6 * 1024 * 1024,
				MaxVotesPerTx:    100,
				DisabledGasCosts: true,
			},
		},
		txapp:         &mockTxApp{},
		genesisParams: genCfg,
	}

	testcases := []struct {
		name    string
		events  []*types.VotableEvent
		cleanup func()
		fn      func(context.Context, *BlockProcessor, *mockEventStore) error
	}{
		{
			name:   "No events to broadcast(gasless mode)",
			events: []*types.VotableEvent{},
			fn: func(ctx context.Context, bp *BlockProcessor, es *mockEventStore) error {
				tx, err := bp.prepareValidatorVoteBodyTx(ctx, 1, bp.chainCtx.NetworkParameters.MaxBlockSize)
				require.NoError(t, err)
				require.Nil(t, tx)

				return nil
			},
		},
		{
			name:   "No events to broadcast (gas mode)",
			events: []*types.VotableEvent{},
			cleanup: func() {
				bp.chainCtx.NetworkParameters.DisabledGasCosts = true
			},
			fn: func(ctx context.Context, bp *BlockProcessor, es *mockEventStore) error {
				bp.chainCtx.NetworkParameters.DisabledGasCosts = false

				tx, err := bp.prepareValidatorVoteBodyTx(ctx, 1, bp.chainCtx.NetworkParameters.MaxBlockSize)
				require.NoError(t, err)
				require.Nil(t, tx)

				return nil
			},
		},
		{
			name:   "atleast 1 event to broadcast",
			events: []*types.VotableEvent{evt1, evt2},
			fn: func(ctx context.Context, bp *BlockProcessor, es *mockEventStore) error {
				tx, err := bp.prepareValidatorVoteBodyTx(ctx, 1, bp.chainCtx.NetworkParameters.MaxBlockSize)
				require.NoError(t, err)
				require.NotNil(t, tx)

				txn := &types.Transaction{}
				err = txn.UnmarshalBinary(tx)
				require.NoError(t, err)

				var payload = &ktypes.ValidatorVoteBodies{}
				err = payload.UnmarshalBinary(txn.Body.Payload)
				require.NoError(t, err)

				require.Len(t, payload.Events, 2)
				return nil
			},
		},
		{
			name:   "enforce maxVotesPerTx limit",
			events: []*types.VotableEvent{evt1, evt2, evt3},
			fn: func(ctx context.Context, bp *BlockProcessor, es *mockEventStore) error {
				bp.chainCtx.NetworkParameters.MaxVotesPerTx = 1

				tx, err := bp.prepareValidatorVoteBodyTx(ctx, 1, bp.chainCtx.NetworkParameters.MaxBlockSize)
				require.NoError(t, err)
				require.NotNil(t, tx)

				txn := &types.Transaction{}
				err = txn.UnmarshalBinary(tx)
				require.NoError(t, err)

				var payload = &ktypes.ValidatorVoteBodies{}
				err = payload.UnmarshalBinary(txn.Body.Payload)
				require.NoError(t, err)

				require.Len(t, payload.Events, 1)
				return nil
			},
			cleanup: func() {
				bp.chainCtx.NetworkParameters.MaxVotesPerTx = 100
			},
		},
		{
			name:   "enforce maxSizePerTx limit",
			events: []*types.VotableEvent{evt1, evt2, evt3},
			fn: func(ctx context.Context, bp *BlockProcessor, es *mockEventStore) error {

				emptyTxSize, err := bp.emptyVodeBodyTxSize()
				require.NoError(t, err)

				// support evt1
				txSize := emptyTxSize + int64(len(evt1.Body)+len(evt1.Type)+8)
				bp.chainCtx.NetworkParameters.MaxBlockSize = txSize + 10 /* buffer */

				tx, err := bp.prepareValidatorVoteBodyTx(ctx, 1, bp.chainCtx.NetworkParameters.MaxBlockSize)
				require.NoError(t, err)
				require.NotNil(t, tx)

				txn := &types.Transaction{}
				err = txn.UnmarshalBinary(tx)
				require.NoError(t, err)

				var payload = &ktypes.ValidatorVoteBodies{}
				err = payload.UnmarshalBinary(txn.Body.Payload)
				require.NoError(t, err)

				require.Len(t, payload.Events, 1)
				return nil

			},
			cleanup: func() {
				bp.chainCtx.NetworkParameters.MaxBlockSize = 6 * 1024 * 1024
			},
		},
		{
			name:   "insufficient funds",
			events: []*types.VotableEvent{evt1, evt2},
			fn: func(ctx context.Context, bp *BlockProcessor, es *mockEventStore) error {
				bp.chainCtx.NetworkParameters.DisabledGasCosts = false
				accountBalance = big.NewInt(0)
				price = big.NewInt(1000)

				tx, err := bp.prepareValidatorVoteBodyTx(ctx, 1, bp.chainCtx.NetworkParameters.MaxBlockSize)
				require.NoError(t, err)
				require.Nil(t, tx)

				return nil
			},
			cleanup: func() {
				bp.chainCtx.NetworkParameters.DisabledGasCosts = true
			},
		},
		{
			name:   "have sufficient funds",
			events: []*types.VotableEvent{evt1, evt2},
			fn: func(ctx context.Context, bp *BlockProcessor, es *mockEventStore) error {
				bp.chainCtx.NetworkParameters.DisabledGasCosts = false
				accountBalance = big.NewInt(1000)
				price = big.NewInt(1000)

				tx, err := bp.prepareValidatorVoteBodyTx(ctx, 1, bp.chainCtx.NetworkParameters.MaxBlockSize)
				require.NoError(t, err)
				require.NotNil(t, tx)

				txn := &types.Transaction{}
				err = txn.UnmarshalBinary(tx)
				require.NoError(t, err)

				var payload = &ktypes.ValidatorVoteBodies{}
				err = payload.UnmarshalBinary(txn.Body.Payload)
				require.NoError(t, err)

				require.Len(t, payload.Events, 2)
				return nil
			},
			cleanup: func() {
				bp.chainCtx.NetworkParameters.DisabledGasCosts = true
				accountBalance = big.NewInt(0)
				price = big.NewInt(0)
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			// cleanup
			defer func() {
				if tc.cleanup != nil {
					tc.cleanup()
				}
			}()

			es := newMockEventStore(tc.events)
			bp.events = es

			getEvents = func(_ context.Context, _ sql.Executor) ([]*types.VotableEvent, error) {
				return es.getEvents(), nil
			}

			if tc.fn != nil {
				tc.fn(context.Background(), bp, es)
			}
		})
	}

}

type mockTxApp struct{}

var accountBalance = big.NewInt(0)

func (m *mockTxApp) AccountInfo(ctx context.Context, db sql.DB, acctID []byte, getUncommitted bool) (balance *big.Int, nonce int64, err error) {
	return accountBalance, 0, nil
}

func (m *mockTxApp) ApplyMempool(ctx *common.TxContext, db sql.DB, tx *types.Transaction) error {
	return nil
}

func (m *mockTxApp) Begin(ctx context.Context, height int64) error {
	return nil
}

func (m *mockTxApp) Commit() error {
	return nil
}

func (m *mockTxApp) Rollback() {}

func (m *mockTxApp) Execute(ctx *common.TxContext, db sql.DB, tx *types.Transaction) *txapp.TxResponse {
	return nil
}

func (m *mockTxApp) Finalize(ctx context.Context, db sql.DB, block *common.BlockContext) (validatorUpgrades []*types.Validator, err error) {
	return nil, nil
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

var price = big.NewInt(0)

func (m *mockTxApp) Price(ctx context.Context, db sql.DB, tx *types.Transaction, c *common.ChainContext) (*big.Int, error) {
	return price, nil
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

type event struct {
	evt         *types.VotableEvent
	broadcasted bool
}

type mockEventStore struct {
	events map[string]event
}

func newMockEventStore(events []*types.VotableEvent) *mockEventStore {
	es := &mockEventStore{events: make(map[string]event)}
	for _, e := range events {
		es.events[e.ID().String()] = event{evt: e, broadcasted: false}
	}
	return es
}

func (m *mockEventStore) addEvent(evt *types.VotableEvent) {
	m.events[evt.ID().String()] = event{evt: evt, broadcasted: false}
}

func (m *mockEventStore) getEvents() []*types.VotableEvent {
	var events []*types.VotableEvent
	for _, e := range m.events {
		events = append(events, e.evt)
	}
	return events
}
func (m *mockEventStore) GetUnbroadcastedEvents(ctx context.Context) ([]*types.UUID, error) {
	var ids []*types.UUID
	for _, e := range m.events {
		if !e.broadcasted {
			ids = append(ids, e.evt.ID())
		}
	}
	return ids, nil
}

func (m *mockEventStore) MarkBroadcasted(ctx context.Context, ids []*types.UUID) error {
	for _, id := range ids {
		if e, ok := m.events[id.String()]; ok {
			e.broadcasted = true
			m.events[id.String()] = e
		}
	}
	return nil
}

type mockValidatorStore struct {
	valSet []*ktypes.Validator
}

func newValidatorStore(valSet []*ktypes.Validator) *mockValidatorStore {
	return &mockValidatorStore{
		valSet: valSet,
	}
}

func (v *mockValidatorStore) GetValidators() []*ktypes.Validator {
	return v.valSet
}

func (v *mockValidatorStore) ValidatorUpdates() map[string]*ktypes.Validator {
	return nil
}
