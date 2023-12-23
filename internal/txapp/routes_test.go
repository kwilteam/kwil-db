package txapp

import (
	"testing"

	"context"

	"math/big"

	"github.com/kwilteam/kwil-db/core/crypto"
	"github.com/kwilteam/kwil-db/core/crypto/auth"
	"github.com/kwilteam/kwil-db/core/log"
	"github.com/kwilteam/kwil-db/core/types"
	"github.com/kwilteam/kwil-db/core/types/transactions"
	"github.com/kwilteam/kwil-db/internal/accounts"
	"github.com/kwilteam/kwil-db/internal/validators"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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
		fee     string                                    // optional, if nil, will automatically use 0
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
			fn: func(t *testing.T, callback func(*TxApp)) {
				approveCount := 0
				deleteCount := 0

				callback(&TxApp{
					VoteStore: &mockVoteStore{
						approve: func(ctx context.Context, resolutionID types.UUID, expiration int64, from []byte) error {
							approveCount++

							return nil
						},
						containsBodyOrFinished: func(ctx context.Context, resolutionID types.UUID) (bool, error) {
							return true, nil
						},
					},
					EventStore: &mockEventStore{
						deleteEvent: func(ctx context.Context, id types.UUID) error {
							deleteCount++

							return nil
						},
					},
					Validators: &mockValidatorStore{
						isCurrent: func(ctx context.Context, validator []byte) (bool, error) {
							return true, nil
						},
					},
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
			fn: func(t *testing.T, callback func(*TxApp)) {
				approveCount := 0
				deleteCount := 0

				callback(&TxApp{
					VoteStore: &mockVoteStore{
						approve: func(ctx context.Context, resolutionID types.UUID, expiration int64, from []byte) error {
							approveCount++

							return nil
						},
						containsBodyOrFinished: func(ctx context.Context, resolutionID types.UUID) (bool, error) {
							return true, nil
						},
					},
					EventStore: &mockEventStore{
						deleteEvent: func(ctx context.Context, id types.UUID) error {
							deleteCount++

							return nil
						},
					},
					Validators: &mockValidatorStore{
						isCurrent: func(ctx context.Context, validator []byte) (bool, error) {
							return true, nil
						},
					},
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
			fn: func(t *testing.T, callback func(*TxApp)) {
				callback(&TxApp{
					Validators: &mockValidatorStore{
						isCurrent: func(ctx context.Context, validator []byte) (bool, error) {
							return false, nil
						},
					},
				})
			},
			payload: &transactions.ValidatorVoteIDs{
				ResolutionIDs: []types.UUID{
					types.NewUUIDV5([]byte("test")),
				},
			},
			err: ErrCallerNotValidator,
		},
		// TODO: we should test that we properly hit eventstore.MarkReceived
		{
			// testing validator_vote_bodies, as the proposer
			name: "validator_vote_bodies, as proposer",
			fn: func(t *testing.T, callback func(*TxApp)) {
				deleteCount := 0

				callback(&TxApp{
					VoteStore: &mockVoteStore{
						hasVoted: func(ctx context.Context, resolutionID types.UUID, voter []byte) (bool, error) {
							return true, nil
						},
					},
					EventStore: &mockEventStore{
						deleteEvent: func(ctx context.Context, id types.UUID) error {
							deleteCount++

							return nil
						},
					},
					Validators: &mockValidatorStore{
						isCurrent: func(ctx context.Context, validator []byte) (bool, error) {
							return true, nil
						},
					},
				})
				assert.Equal(t, 1, deleteCount)
			},
			payload: &transactions.ValidatorVoteBodies{
				Events: []*types.VotableEvent{
					{
						Type: "asdfadsf",
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
			fn: func(t *testing.T, callback func(*TxApp)) {
				deleteCount := 0

				callback(&TxApp{
					VoteStore: &mockVoteStore{
						hasVoted: func(ctx context.Context, resolutionID types.UUID, voter []byte) (bool, error) {
							return true, nil
						},
					},
					EventStore: &mockEventStore{
						deleteEvent: func(ctx context.Context, id types.UUID) error {
							deleteCount++

							return nil
						},
					},
					Validators: &mockValidatorStore{
						isCurrent: func(ctx context.Context, validator []byte) (bool, error) {
							return true, nil
						},
					},
				})
				assert.Equal(t, 0, deleteCount) // 0, since this does not go through
			},
			payload: &transactions.ValidatorVoteBodies{
				Events: []*types.VotableEvent{
					{
						Type: "asdfadsf",
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

			// build tx
			tx, err := transactions.CreateTransaction(tc.payload, "chainid", 1)
			require.NoError(t, err)

			if tc.fee == "" {
				tx.Body.Fee = big.NewInt(0)
			} else {
				bigFee, ok := new(big.Int).SetString(tc.fee, 10)
				if !ok {
					t.Fatal("invalid fee")
				}

				tx.Body.Fee = bigFee
			}

			err = tx.Sign(tc.from)
			require.NoError(t, err)

			if tc.fn == nil {
				require.Fail(t, "no callback provided")
			}

			tc.fn(t, func(app *TxApp) {
				// since every test case needs an account store, we'll just create a mock one here
				// if one isn't provided
				if app.Accounts == nil {
					app.Accounts = &mockAccountStore{}
				}
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

type mockAccountStore struct {
	getAccount func(ctx context.Context, acctID []byte) (*accounts.Account, error)
	credit     func(ctx context.Context, acctID []byte, amt *big.Int) error
	spend      func(ctx context.Context, spend *accounts.Spend) error
	transfer   func(ctx context.Context, to []byte, from []byte, amt *big.Int) error
}

func (m *mockAccountStore) GetAccount(ctx context.Context, acctID []byte) (*accounts.Account, error) {
	if m.getAccount != nil {
		return m.getAccount(ctx, acctID)
	}

	return &accounts.Account{
		Identifier: acctID,
		Balance:    big.NewInt(0),
		Nonce:      0,
	}, nil
}

func (m *mockAccountStore) Credit(ctx context.Context, acctID []byte, amt *big.Int) error {
	if m.credit != nil {
		return m.credit(ctx, acctID, amt)
	}
	return nil
}

func (m *mockAccountStore) Spend(ctx context.Context, spend *accounts.Spend) error {
	if m.spend != nil {
		return m.spend(ctx, spend)
	}
	return nil
}

func (m *mockAccountStore) Transfer(ctx context.Context, to []byte, from []byte, amt *big.Int) error {
	if m.transfer != nil {
		return m.transfer(ctx, to, from, amt)
	}
	return nil
}

type mockVoteStore struct {
	alreadyProcessed            func(ctx context.Context, resolutionID types.UUID) (bool, error)
	approve                     func(ctx context.Context, resolutionID types.UUID, expiration int64, from []byte) error
	containsBodyOrFinished      func(ctx context.Context, resolutionID types.UUID) (bool, error)
	createResolution            func(ctx context.Context, event *types.VotableEvent, expiration int64) error
	expire                      func(ctx context.Context, blockheight int64) error
	hasVoted                    func(ctx context.Context, resolutionID types.UUID, voter []byte) (bool, error)
	processConfirmedResolutions func(ctx context.Context) ([]types.UUID, error)
	updateVoter                 func(ctx context.Context, identifier []byte, power int64) error
}

func (m *mockVoteStore) AlreadyProcessed(ctx context.Context, resolutionID types.UUID) (bool, error) {
	if m.alreadyProcessed != nil {
		return m.alreadyProcessed(ctx, resolutionID)
	}

	return false, nil
}

func (m *mockVoteStore) Approve(ctx context.Context, resolutionID types.UUID, expiration int64, from []byte) error {
	if m.approve != nil {
		return m.approve(ctx, resolutionID, expiration, from)
	}

	return nil
}

func (m *mockVoteStore) ContainsBodyOrFinished(ctx context.Context, resolutionID types.UUID) (bool, error) {
	if m.containsBodyOrFinished != nil {
		return m.containsBodyOrFinished(ctx, resolutionID)
	}

	return false, nil
}

func (m *mockVoteStore) CreateResolution(ctx context.Context, event *types.VotableEvent, expiration int64) error {
	if m.createResolution != nil {
		return m.createResolution(ctx, event, expiration)
	}

	return nil
}

func (m *mockVoteStore) Expire(ctx context.Context, blockheight int64) error {
	if m.expire != nil {
		return m.expire(ctx, blockheight)
	}

	return nil
}

func (m *mockVoteStore) HasVoted(ctx context.Context, resolutionID types.UUID, voter []byte) (bool, error) {
	if m.hasVoted != nil {
		return m.hasVoted(ctx, resolutionID, voter)
	}

	return false, nil
}

func (m *mockVoteStore) ProcessConfirmedResolutions(ctx context.Context) ([]types.UUID, error) {
	if m.processConfirmedResolutions != nil {
		return m.processConfirmedResolutions(ctx)
	}

	return nil, nil
}

func (m *mockVoteStore) UpdateVoter(ctx context.Context, identifier []byte, power int64) error {
	if m.updateVoter != nil {
		return m.updateVoter(ctx, identifier, power)
	}

	return nil
}

type mockEventStore struct {
	deleteEvent  func(ctx context.Context, id types.UUID) error
	getEvents    func(ctx context.Context) ([]*types.VotableEvent, error)
	markReceived func(ctx context.Context, id types.UUID) error
}

func (m *mockEventStore) DeleteEvent(ctx context.Context, id types.UUID) error {
	if m.deleteEvent != nil {
		return m.deleteEvent(ctx, id)
	}

	return nil
}

func (m *mockEventStore) GetEvents(ctx context.Context) ([]*types.VotableEvent, error) {
	if m.getEvents != nil {
		return m.getEvents(ctx)
	}

	return nil, nil
}

func (m *mockEventStore) MarkReceived(ctx context.Context, id types.UUID) error {
	if m.markReceived != nil {
		return m.markReceived(ctx, id)
	}

	return nil
}

type mockValidatorStore struct {
	approve           func(ctx context.Context, joiner []byte, approver []byte) error
	finalize          func(ctx context.Context) ([]*validators.Validator, error)
	isCurrent         func(ctx context.Context, validator []byte) (bool, error)
	join              func(ctx context.Context, joiner []byte, power int64) error
	leave             func(ctx context.Context, joiner []byte) error
	remove            func(ctx context.Context, target []byte, validator []byte) error
	updateBlockHeight func(blockHeight int64)
}

func (m *mockValidatorStore) Approve(ctx context.Context, joiner []byte, approver []byte) error {
	if m.approve != nil {
		return m.approve(ctx, joiner, approver)
	}

	return nil
}

func (m *mockValidatorStore) Finalize(ctx context.Context) ([]*validators.Validator, error) {
	if m.finalize != nil {
		return m.finalize(ctx)
	}

	return nil, nil
}

func (m *mockValidatorStore) IsCurrent(ctx context.Context, validator []byte) (bool, error) {
	if m.isCurrent != nil {
		return m.isCurrent(ctx, validator)
	}

	return false, nil
}

func (m *mockValidatorStore) Join(ctx context.Context, joiner []byte, power int64) error {
	if m.join != nil {
		return m.join(ctx, joiner, power)
	}

	return nil
}

func (m *mockValidatorStore) Leave(ctx context.Context, joiner []byte) error {
	if m.leave != nil {
		return m.leave(ctx, joiner)
	}

	return nil
}

func (m *mockValidatorStore) Remove(ctx context.Context, target []byte, validator []byte) error {
	if m.remove != nil {
		return m.remove(ctx, target, validator)
	}

	return nil
}

func (m *mockValidatorStore) UpdateBlockHeight(blockHeight int64) {
	if m.updateBlockHeight != nil {
		m.updateBlockHeight(blockHeight)
	}
}
