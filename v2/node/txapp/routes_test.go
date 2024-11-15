package txapp

import (
	"encoding/hex"
	"kwil/crypto"
	"kwil/crypto/auth"
	"kwil/extensions/resolutions"
	"kwil/log"
	"kwil/node/types/sql"
	"kwil/node/voting"
	"kwil/types"
	"testing"

	"context"

	"math/big"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testType = "test"

func init() {
	err := resolutions.RegisterResolution(testType, resolutions.ModAdd, resolutions.ResolutionConfig{})
	if err != nil {
		panic(err)
	}
}

var (
	signer1 = getSigner("7c67e60fce0c403ff40193a3128e5f3d8c2139aed36d76d7b5f1e70ec19c43f00aa611bf555596912bc6f9a9f169f8785918e7bab9924001895798ff13f05842")

	signer2 = getSigner("2b8615d7ee7b7d3fc7d6b89d9b31c045ca5c4d220c82eab25420873c99010422fb35029f37e80148ae89588710eb7d692e96a070d48e579cad51a253e9d1c030")
)

type getVoterPowerFunc func() (int64, error)

func Test_Routes(t *testing.T) {

	// in this testcase we handle the router in a callback, so that
	// we can have scoped data in our mock implementations
	type testcase struct {
		name          string
		fn            func(t *testing.T, callback func()) // required, uses callback to control when the test is run
		payload       types.Payload                       // required
		fee           int64                               // optional, if nil, will automatically use 0
		ctx           *types.TxContext                    // optional, if nil, will automatically create a mock
		from          auth.Signer                         // optional, if nil, will automatically use default validatorSigner1
		getVoterPower getVoterPowerFunc
		err           error // if not nil, expect this error
	}

	// due to the relative simplicity of routes and pricing, I have only tested a few complex ones.
	// as routes / pricing becomes more complex, we should add more tests here.

	testCases := []testcase{
		{
			// this test tests vote_id, as a local validator
			// we expect that it will approve and then attempt to delete the event
			name: "validator_vote_id, as local validator",
			fee:  voting.ValidatorVoteIDPrice,
			getVoterPower: func() (int64, error) {
				return 1, nil
			},
			fn: func(t *testing.T, callback func()) {
				approveCount := 0
				deleteCount := 0

				// override the functions with mocks
				deleteEvent = func(ctx context.Context, db sql.Executor, id *types.UUID) error {
					deleteCount++

					return nil
				}

				approveResolution = func(ctx context.Context, db sql.TxMaker, resolutionID *types.UUID, from []byte) error {
					approveCount++

					return nil
				}

				callback()

				assert.Equal(t, 1, approveCount)
				assert.Equal(t, 1, deleteCount)
			},
			payload: &types.ValidatorVoteIDs{
				ResolutionIDs: []*types.UUID{
					types.NewUUIDV5([]byte("test")),
				},
			},
		},
		{
			// this test tests vote_id, as a non-local validator
			// we expect that it will approve and not attempt to delete the event
			name: "validator_vote_id, as non-local validator",
			fee:  voting.ValidatorVoteIDPrice,
			getVoterPower: func() (int64, error) {
				return 1, nil
			},
			fn: func(t *testing.T, callback func()) {
				approveCount := 0
				deleteCount := 0

				// override the functions with mocks
				deleteEvent = func(ctx context.Context, db sql.Executor, id *types.UUID) error {
					deleteCount++

					return nil
				}
				approveResolution = func(_ context.Context, _ sql.TxMaker, _ *types.UUID, _ []byte) error {
					approveCount++

					return nil
				}
				callback()

				assert.Equal(t, 1, approveCount)
				assert.Equal(t, 0, deleteCount)
			},
			payload: &types.ValidatorVoteIDs{
				ResolutionIDs: []*types.UUID{
					types.NewUUIDV5([]byte("test")),
				},
			},
			from: signer2,
		},
		{
			// this test tests vote_id, from a non-validator
			// we expect that it will fail
			name: "validator_vote_id, as non-validator",
			fee:  voting.ValidatorVoteIDPrice,
			getVoterPower: func() (int64, error) {
				return 0, nil
			},
			fn: func(t *testing.T, callback func()) {
				callback()
			},
			payload: &types.ValidatorVoteIDs{
				ResolutionIDs: []*types.UUID{
					types.NewUUIDV5([]byte("test")),
				},
			},
			err: ErrCallerNotValidator,
		},
		{
			// testing validator_vote_bodies, as the proposer
			name: "validator_vote_bodies, as proposer",
			fee:  voting.ValidatorVoteIDPrice,
			getVoterPower: func() (int64, error) {
				return 1, nil
			},
			fn: func(t *testing.T, callback func()) {
				deleteCount := 0

				// override the functions with mocks
				deleteEvent = func(ctx context.Context, db sql.Executor, id *types.UUID) error {
					deleteCount++

					return nil
				}
				createResolution = func(_ context.Context, _ sql.TxMaker, _ *types.VotableEvent, _ int64, _ []byte) error {
					return nil
				}

				callback()
				assert.Equal(t, 1, deleteCount)
			},
			payload: &types.ValidatorVoteBodies{
				Events: []*types.VotableEvent{
					{
						Type: testType,
						Body: []byte("asdfadsf"),
					},
				},
			},
			ctx: &types.TxContext{
				BlockContext: &types.BlockContext{
					Proposer: signer1.Identity(),
				},
			},
			from: signer1,
		},
		{
			// testing validator_vote_bodies, as a non-proposer
			// should fail
			name: "validator_vote_bodies, as non-proposer",
			fee:  voting.ValidatorVoteIDPrice,
			getVoterPower: func() (int64, error) {
				return 1, nil
			},
			fn: func(t *testing.T, callback func()) {
				deleteCount := 0

				deleteEvent = func(_ context.Context, _ sql.Executor, _ *types.UUID) error {
					deleteCount++

					return nil
				}

				callback()
				assert.Equal(t, 0, deleteCount) // 0, since this does not go through
			},
			payload: &types.ValidatorVoteBodies{
				Events: []*types.VotableEvent{
					{
						Type: testType,
						Body: []byte("asdfadsf"),
					},
				},
			},
			ctx: &types.TxContext{
				BlockContext: &types.BlockContext{
					Proposer: signer1.Identity(),
				},
			},
			from: signer2,
			err:  ErrCallerNotProposer,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.from == nil {
				tc.from = signer1
			}

			// mock getAccount, which is func declared in interfaces.go
			account := &mockAccount{}
			Validators := &mockValidator{
				getVoterFn: tc.getVoterPower,
			}

			// build tx
			tx, err := types.CreateTransaction(tc.payload, "chainid", 1)
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

			tc.fn(t, func() {
				db := &mockTx{&mockDb{}}
				app := &TxApp{
					Accounts:   account,
					Validators: Validators,
				}
				if app.signer == nil {
					app.signer = signer1
				}
				if app.service == nil {
					app.service = &types.Service{
						Logger:   log.DiscardLogger,
						Identity: app.signer.Identity(),
					}
				}

				if tc.ctx == nil {
					tc.ctx = &types.TxContext{}
				}

				if tc.ctx.BlockContext == nil {
					tc.ctx.BlockContext = &types.BlockContext{
						ChainContext: &types.ChainContext{
							NetworkParameters: &types.NetworkParameters{
								DisabledGasCosts: false,
							},
						},
					}
				} else if tc.ctx.BlockContext.ChainContext == nil {
					tc.ctx.BlockContext.ChainContext = &types.ChainContext{
						NetworkParameters: &types.NetworkParameters{
							DisabledGasCosts: false,
						},
					}
				}

				res := app.Execute(tc.ctx, db, tx)
				if tc.err != nil {
					require.ErrorIs(t, tc.err, res.Error)
				} else {
					require.NoError(t, res.Error)
				}
			})
		})
	}
}

type mockAccount struct {
}

func (a *mockAccount) GetAccount(_ context.Context, _ sql.Executor, acctID []byte) (*types.Account, error) {
	return &types.Account{
		Identifier: acctID,
		Balance:    big.NewInt(0),
		Nonce:      0,
	}, nil
}

func (a *mockAccount) Spend(_ context.Context, _ sql.Executor, acctID []byte, amount *big.Int, nonce int64) error {
	return nil
}

func (a *mockAccount) Credit(_ context.Context, _ sql.Executor, acctID []byte, amount *big.Int) error {
	return nil
}

func (a *mockAccount) Transfer(_ context.Context, _ sql.Executor, from, to []byte, amount *big.Int) error {
	return nil
}

func (a *mockAccount) ApplySpend(_ context.Context, _ sql.Executor, acctID []byte, amount *big.Int, nonce int64) error {
	return nil
}
func (a *mockAccount) Commit() error {
	return nil
}

type mockValidator struct {
	getVoterFn getVoterPowerFunc
}

func (v *mockValidator) GetValidators() ([]*types.Validator, error) {
	return nil, nil
}

func (v *mockValidator) GetValidatorPower(_ context.Context, _ sql.Executor, pubKey []byte) (int64, error) {
	return v.getVoterFn()
}

func (v *mockValidator) SetValidatorPower(_ context.Context, _ sql.Executor, pubKey []byte, power int64) error {
	return nil
}

func (v *mockValidator) Commit() error {
	return nil
}

func getSigner(hexPrivKey string) *auth.Ed25519Signer {
	bts, err := hex.DecodeString(hexPrivKey)
	if err != nil {
		panic(err)
	}
	pk, err := crypto.UnmarshalEd25519PrivateKey(bts)
	if err != nil {
		panic(err)
	}

	return &auth.Ed25519Signer{
		Ed25519PrivateKey: *pk,
	}
}
