package txapp

import (
	"testing"

	"context"

	"math/big"

	sql "github.com/kwilteam/kwil-db/common/sql"
	"github.com/kwilteam/kwil-db/core/crypto"
	"github.com/kwilteam/kwil-db/core/crypto/auth"
	"github.com/kwilteam/kwil-db/core/log"
	"github.com/kwilteam/kwil-db/core/types"
	"github.com/kwilteam/kwil-db/core/types/transactions"
	"github.com/kwilteam/kwil-db/extensions/resolutions"
	"github.com/kwilteam/kwil-db/internal/voting"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testType = "test"

func init() {
	err := resolutions.RegisterResolution(testType, resolutions.ResolutionConfig{})
	if err != nil {
		panic(err)
	}
}

func validatorSigner1() *auth.Ed25519Signer {
	pk, err := crypto.Ed25519PrivateKeyFromHex("7c67e60fce0c403ff40193a3128e5f3d8c2139aed36d76d7b5f1e70ec19c43f00aa611bf555596912bc6f9a9f169f8785918e7bab9924001895798ff13f05842")
	if err != nil {
		panic(err)
	}

	return &auth.Ed25519Signer{
		Ed25519PrivateKey: *pk,
	}
}
func validatorSigner2() *auth.Ed25519Signer {
	pk, err := crypto.Ed25519PrivateKeyFromHex("2b8615d7ee7b7d3fc7d6b89d9b31c045ca5c4d220c82eab25420873c99010422fb35029f37e80148ae89588710eb7d692e96a070d48e579cad51a253e9d1c030")
	if err != nil {
		panic(err)
	}

	return &auth.Ed25519Signer{
		Ed25519PrivateKey: *pk,
	}
}

func Test_Routes(t *testing.T) {

	// in this testcase we handle the router in a callback, so that
	// we can have scoped data in our mock implementations
	type testcase struct {
		name    string
		fn      func(t *testing.T, callback func(*TxApp)) // required, uses callback to allow for scoped data
		payload transactions.Payload                      // required
		fee     int64                                     // optional, if nil, will automatically use 0
		ctx     TxContext                                 // optional, if nil, will automatically create a mock
		from    auth.Signer                               // optional, if nil, will automatically use default validatorSigner1
		err     error                                     // if not nil, expect this error
	}

	// due to the relative simplicity of routes and pricing, I have only tested a few complex ones.
	// as routes / pricing becomes more complex, we should add more tests here.

	testCases := []testcase{
		{
			// this test tests vote_id, as a local validator
			// we expect that it will approve and then attempt to delete the event
			name: "validator_vote_id, as local validator",
			fee:  voting.ValidatorVoteIDPrice,
			fn: func(t *testing.T, callback func(*TxApp)) {
				approveCount := 0
				deleteCount := 0

				// override the functions with mocks
				deleteEvent = func(ctx context.Context, db sql.Executor, id types.UUID) error {
					deleteCount++

					return nil
				}

				approveResolution = func(ctx context.Context, db sql.TxMaker, resolutionID types.UUID, from []byte) error {
					approveCount++

					return nil
				}
				getVoterPower = func(ctx context.Context, db sql.Executor, identifier []byte) (int64, error) {
					return 1, nil
				}

				callback(&TxApp{
					GasEnabled: true,
				})

				assert.Equal(t, 1, approveCount)
				assert.Equal(t, 1, deleteCount)
			},
			payload: &transactions.ValidatorVoteIDs{
				ResolutionIDs: []types.UUID{
					types.NewUUIDV5([]byte("test")),
				},
			},
		},
		{
			// this test tests vote_id, as a non-local validator
			// we expect that it will approve and not attempt to delete the event
			name: "validator_vote_id, as non-local validator",
			fee:  voting.ValidatorVoteIDPrice,
			fn: func(t *testing.T, callback func(*TxApp)) {
				approveCount := 0
				deleteCount := 0

				// override the functions with mocks
				deleteEvent = func(ctx context.Context, db sql.Executor, id types.UUID) error {
					deleteCount++

					return nil
				}
				approveResolution = func(_ context.Context, _ sql.TxMaker, _ types.UUID, _ []byte) error {
					approveCount++

					return nil
				}

				getVoterPower = func(ctx context.Context, db sql.Executor, identifier []byte) (int64, error) {
					return 1, nil
				}

				callback(&TxApp{
					GasEnabled: true,
				})

				assert.Equal(t, 1, approveCount)
				assert.Equal(t, 0, deleteCount)
			},
			payload: &transactions.ValidatorVoteIDs{
				ResolutionIDs: []types.UUID{
					types.NewUUIDV5([]byte("test")),
				},
			},
			from: validatorSigner2(),
		},
		{
			// this test tests vote_id, from a non-validator
			// we expect that it will fail
			name: "validator_vote_id, as non-validator",
			fee:  voting.ValidatorVoteIDPrice,
			fn: func(t *testing.T, callback func(*TxApp)) {
				getVoterPower = func(ctx context.Context, db sql.Executor, identifier []byte) (int64, error) {
					return 0, nil
				}

				callback(&TxApp{
					GasEnabled: true,
				})
			},
			payload: &transactions.ValidatorVoteIDs{
				ResolutionIDs: []types.UUID{
					types.NewUUIDV5([]byte("test")),
				},
			},
			err: ErrCallerNotValidator,
		},
		{
			// testing validator_vote_bodies, as the proposer
			name: "validator_vote_bodies, as proposer",
			fee:  voting.ValidatorVoteIDPrice,
			fn: func(t *testing.T, callback func(*TxApp)) {
				deleteCount := 0

				// override the functions with mocks
				deleteEvent = func(ctx context.Context, db sql.Executor, id types.UUID) error {
					deleteCount++

					return nil
				}
				createResolution = func(_ context.Context, _ sql.TxMaker, _ *types.VotableEvent, _ int64, _ []byte) error {
					return nil
				}
				getVoterPower = func(ctx context.Context, db sql.Executor, identifier []byte) (int64, error) {
					return 1, nil
				}

				callback(&TxApp{
					GasEnabled: true,
				})
				assert.Equal(t, 1, deleteCount)
			},
			payload: &transactions.ValidatorVoteBodies{
				Events: []*types.VotableEvent{
					{
						Type: testType,
						Body: []byte("asdfadsf"),
					},
				},
			},
			ctx: TxContext{
				Proposer: validatorSigner1().Identity(),
			},
			from: validatorSigner1(),
		},
		{
			// testing validator_vote_bodies, as a non-proposer
			// should fail
			name: "validator_vote_bodies, as non-proposer",
			fee:  voting.ValidatorVoteIDPrice,
			fn: func(t *testing.T, callback func(*TxApp)) {
				deleteCount := 0

				deleteEvent = func(_ context.Context, _ sql.Executor, _ types.UUID) error {
					deleteCount++

					return nil
				}

				getVoterPower = func(_ context.Context, _ sql.Executor, _ []byte) (int64, error) {
					return 1, nil
				}

				callback(&TxApp{
					GasEnabled: true,
				})
				assert.Equal(t, 0, deleteCount) // 0, since this does not go through
			},
			payload: &transactions.ValidatorVoteBodies{
				Events: []*types.VotableEvent{
					{
						Type: testType,
						Body: []byte("asdfadsf"),
					},
				},
			},
			ctx: TxContext{
				Proposer: validatorSigner1().Identity(),
			},
			from: validatorSigner2(),
			err:  ErrCallerNotProposer,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.from == nil {
				tc.from = validatorSigner1()
			}

			// mock getAccount, which is func declared in interfaces.go
			getAccount = func(_ context.Context, _ sql.Executor, acctID []byte) (*types.Account, error) {
				return &types.Account{
					Identifier: acctID,
					Balance:    big.NewInt(0),
					Nonce:      0,
				}, nil
			}
			spend = func(_ context.Context, _ sql.Executor, _ []byte, _ *big.Int, _ int64) error {
				return nil
			}

			// build tx
			tx, err := transactions.CreateTransaction(tc.payload, "chainid", 1)
			require.NoError(t, err)

			tx.Body.Fee = big.NewInt(0)
			if tc.fee != 0 {
				tx.Body.Fee = big.NewInt(tc.fee)
			}

			err = tx.Sign(tc.from)
			require.NoError(t, err)

			if tc.fn == nil {
				require.Fail(t, "no callback provided")
			}

			tc.fn(t, func(app *TxApp) {
				app.currentTx = &mockOuterTx{&mockTx{&mockDb{}}} // hack to trick txapp that we are in a session

				// since every test case needs an account store, we'll just create a mock one here
				// if one isn't provided
				if app.log.L == nil {
					app.log = log.NewNoOp()
				}
				if app.signer == nil {
					app.signer = validatorSigner1()
				}

				res := app.Execute(tc.ctx, tx)
				if tc.err != nil {
					require.ErrorIs(t, tc.err, res.Error)
				} else {
					require.NoError(t, res.Error)
				}
			})

		})
	}
}
