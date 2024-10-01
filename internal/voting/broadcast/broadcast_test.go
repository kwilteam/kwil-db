package broadcast_test

import (
	"context"
	"testing"

	"math/big"

	cmtCoreTypes "github.com/cometbft/cometbft/rpc/core/types"
	"github.com/stretchr/testify/require"

	"github.com/kwilteam/kwil-db/common"
	"github.com/kwilteam/kwil-db/common/sql"
	"github.com/kwilteam/kwil-db/core/crypto"
	"github.com/kwilteam/kwil-db/core/crypto/auth"
	"github.com/kwilteam/kwil-db/core/log"
	"github.com/kwilteam/kwil-db/core/types"
	"github.com/kwilteam/kwil-db/core/types/transactions"
	"github.com/kwilteam/kwil-db/extensions/resolutions"
	dbtest "github.com/kwilteam/kwil-db/internal/sql/pg/test"
	"github.com/kwilteam/kwil-db/internal/voting"
	"github.com/kwilteam/kwil-db/internal/voting/broadcast"
)

const maxVoteIDsPerTx = 100

const testType = "test"

func init() {
	err := resolutions.RegisterResolution(testType, resolutions.ModAdd, resolutions.ResolutionConfig{})
	if err != nil {
		panic(err)
	}
}

func Test_Broadcaster(t *testing.T) {
	type testCase struct {
		name          string
		events        []string
		expectedNonce int
		didBroadcast  bool
		balance       *big.Int

		broadcaster *broadcaster        // optional
		txapp       *mockTxApp          // optional
		v           *mockValidatorStore // optional
		err         error               // optional
	}

	tests := []testCase{
		{
			name:          "no events",
			expectedNonce: -1,
			didBroadcast:  false,
			v:             &mockValidatorStore{isValidator: true},
		},
		{
			name: "has events, not validator",
			events: []string{
				"hello",
			},
			expectedNonce: -1,
			didBroadcast:  false,
		},
		{
			name:          "single event",
			events:        []string{"hello"},
			expectedNonce: 1,
			didBroadcast:  true,
			balance:       big.NewInt(voting.ValidatorVoteIDPrice),
			v:             &mockValidatorStore{isValidator: true},
		},
		{
			name: "multiple events",
			events: []string{
				"hello",
				"world",
			},
			balance:       big.NewInt(voting.ValidatorVoteIDPrice * 2),
			expectedNonce: 1, // should broadcast all of them at once
			didBroadcast:  true,
			v:             &mockValidatorStore{isValidator: true},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			v := tc.v
			if v == nil {
				v = &mockValidatorStore{}
			} else {
				v.pubkey = validatorSigner().Identity()
			}
			ctx := context.Background()

			db, cleanup, err := dbtest.NewTestPool(ctx, []string{"kwild_events", "kwild_voting"}) // db is the event store specific connection
			require.NoError(t, err)
			defer cleanup()

			_, err = voting.NewEventStore(ctx, db)
			require.NoError(t, err)

			// create a second db connection to emulate the consensus db
			consensusDB, cleanup2, err := dbtest.NewTestPool(ctx, nil) // don't need to delete schema since we will never commit
			require.NoError(t, err)
			defer cleanup2()

			consensusTx, err := consensusDB.BeginTx(ctx)
			require.NoError(t, err)
			defer consensusTx.Rollback(ctx) // always rollback, to clean up

			txapp := tc.txapp
			if txapp == nil {
				txapp = &mockTxApp{
					balance: tc.balance,
				}
			}
			txapp.mockValidatorStore = *v

			e := &mockEventStore{}
			for _, event := range tc.events {
				e.events = append(e.events, &types.VotableEvent{
					Body: []byte(event),
					Type: "test",
				})
			}

			didBroadcast := false
			usedNonce := int(-1)

			b := tc.broadcaster
			if b == nil {
				b = &broadcaster{
					broadcastFn: func(ctx context.Context, tx []byte, sync uint8) (res *cmtCoreTypes.ResultBroadcastTx, err error) {
						didBroadcast = true
						receivedTx := &transactions.Transaction{}
						err = receivedTx.UnmarshalBinary(tx)
						require.NoError(t, err)

						usedNonce = int(receivedTx.Body.Nonce)
						return nil, nil
					},
				}
			}

			bc := broadcast.NewEventBroadcaster(e, b, txapp, validatorSigner(), "test-chain", maxVoteIDsPerTx, log.NewNoOp())

			// create resolutions for the events
			for _, event := range e.events {
				err = voting.CreateResolution(ctx, consensusTx, event, 10000, []byte("proposer"))
				require.NoError(t, err)
			}

			err = bc.RunBroadcast(ctx, &mockDB{}, &common.BlockContext{
				Proposer: []byte("proposer"),
			})
			if tc.err != nil {
				require.Equal(t, tc.err, err)
				return
			}
			require.NoError(t, err)

			require.Equal(t, tc.didBroadcast, didBroadcast)
			require.Equal(t, tc.expectedNonce, usedNonce)
		})
	}
}

type mockEventStore struct {
	events []*types.VotableEvent
}

func (m *mockEventStore) MarkBroadcasted(ctx context.Context, ids []*types.UUID) error {
	return nil
}

func (m *mockEventStore) GetUnbroadcastedEvents(ctx context.Context) ([]*types.UUID, error) {
	var ids []*types.UUID
	for _, event := range m.events {
		ids = append(ids, event.ID())
	}
	return ids, nil
}

type broadcaster struct {
	broadcastFn func(ctx context.Context, tx []byte, sync uint8) (res *cmtCoreTypes.ResultBroadcastTx, err error)
}

func (b *broadcaster) BroadcastTx(ctx context.Context, tx []byte, sync uint8) (res *cmtCoreTypes.ResultBroadcastTx, err error) {
	return b.broadcastFn(ctx, tx, sync)
}

type mockTxApp struct {
	balance *big.Int // the balance to return for AccountInfo
	nonce   int64    // the nonce to return for AccountInfo

	price *big.Int // the price to return for Price

	mockValidatorStore
}

func (m *mockTxApp) AccountInfo(ctx context.Context, db sql.DB, acctID []byte, getUncommitted bool) (balance *big.Int, nonce int64, err error) {
	return m.balance, m.nonce, nil
}

func (m *mockTxApp) Price(ctx context.Context, db sql.DB, tx *transactions.Transaction, c *common.ChainContext) (*big.Int, error) {
	if m.price == nil {
		return big.NewInt(0), nil
	}
	return m.price, nil
}

func validatorSigner() *auth.Ed25519Signer {
	pk, err := crypto.Ed25519PrivateKeyFromHex("7c67e60fce0c403ff40193a3128e5f3d8c2139aed36d76d7b5f1e70ec19c43f00aa611bf555596912bc6f9a9f169f8785918e7bab9924001895798ff13f05842")
	if err != nil {
		panic(err)
	}

	return &auth.Ed25519Signer{
		Ed25519PrivateKey: *pk,
	}
}

type mockValidatorStore struct {
	isValidator bool
	pubkey      []byte
}

func (m *mockValidatorStore) GetValidators(ctx context.Context, db sql.DB) ([]*types.Validator, error) {
	if m.isValidator {
		return []*types.Validator{
			{
				PubKey: m.pubkey,
				Power:  1,
			},
		}, nil
	}
	return nil, nil
}

type mockDB struct{}

func (m *mockDB) Execute(ctx context.Context, stmt string, args ...any) (*sql.ResultSet, error) {
	return nil, nil
}

func (m *mockDB) BeginTx(ctx context.Context) (sql.Tx, error) {
	return &mockTX{}, nil
}

type mockTX struct{}

func (m *mockTX) Rollback(ctx context.Context) error {
	return nil
}

func (m *mockTX) Commit(ctx context.Context) error {
	return nil
}

func (m *mockTX) Execute(ctx context.Context, stmt string, args ...any) (*sql.ResultSet, error) {
	return nil, nil
}

func (m *mockTX) BeginTx(ctx context.Context) (sql.Tx, error) {
	return &mockTX{}, nil
}
