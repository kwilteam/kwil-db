package broadcast_test

import (
	"context"
	"testing"

	"math/big"

	cmtCoreTypes "github.com/cometbft/cometbft/rpc/core/types"
	"github.com/stretchr/testify/require"

	"github.com/kwilteam/kwil-db/core/crypto"
	"github.com/kwilteam/kwil-db/core/crypto/auth"
	"github.com/kwilteam/kwil-db/core/types"
	"github.com/kwilteam/kwil-db/core/types/transactions"
	"github.com/kwilteam/kwil-db/internal/events/broadcast"
)

func Test_Broadcaster(t *testing.T) {
	type testCase struct {
		name          string
		events        []string
		expectedNonce int
		didBroadcast  bool

		broadcaster *broadcaster        // optional
		ai          *mockAccountInfoer  // optional
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
			v:             &mockValidatorStore{isValidator: true},
		},
		{
			name: "multiple events",
			events: []string{
				"hello",
				"world",
			},
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
			}

			ai := tc.ai
			if ai == nil {
				ai = &mockAccountInfoer{}
			}

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

			bc := broadcast.NewEventBroadcaster(e, b, ai, v, validatorSigner(), "test-chain")

			err := bc.RunBroadcast(context.Background(), []byte("proposer"))
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

func (m *mockEventStore) GetUnreceivedEvents(ctx context.Context) ([]*types.VotableEvent, error) {
	return m.events, nil
}

type broadcaster struct {
	broadcastFn func(ctx context.Context, tx []byte, sync uint8) (res *cmtCoreTypes.ResultBroadcastTx, err error)
}

func (b *broadcaster) BroadcastTx(ctx context.Context, tx []byte, sync uint8) (res *cmtCoreTypes.ResultBroadcastTx, err error) {
	return b.broadcastFn(ctx, tx, sync)
}

type mockAccountInfoer struct {
	balance *big.Int
	nonce   int64
}

func (m *mockAccountInfoer) AccountInfo(ctx context.Context, acctID []byte, getUncommitted bool) (balance *big.Int, nonce int64, err error) {
	return m.balance, m.nonce, nil
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
}

func (m *mockValidatorStore) IsCurrent(ctx context.Context, validator []byte) (bool, error) {
	return m.isValidator, nil
}
