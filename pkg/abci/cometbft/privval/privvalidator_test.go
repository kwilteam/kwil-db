package privval_test

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	cometEd25519 "github.com/cometbft/cometbft/crypto/ed25519"
	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	"github.com/cometbft/cometbft/types"
	"github.com/kwilteam/kwil-db/pkg/abci/cometbft/privval"
	"github.com/stretchr/testify/assert"
)

const defaultChainID = "test-chain"
const defaultPrivateKey = "7c67e60fce0c403ff40193a3128e5f3d8c2139aed36d76d7b5f1e70ec19c43f00aa611bf555596912bc6f9a9f169f8785918e7bab9924001895798ff13f05842"

func Test_C(t *testing.T) {
	pk := cometEd25519.PrivKey(defaultPrivateKey)
	fmt.Println(base64.StdEncoding.EncodeToString(pk.Bytes()))
	fmt.Println(base64.StdEncoding.EncodeToString(pk.PubKey().Bytes()))
	panic("")
}

func Test_D(t *testing.T) {
	pk := cometEd25519.GenPrivKey()
	fmt.Println(hex.DecodeString("7c67e60fce0c403ff40193a3128e5f3d8c2139aed36d76d7b5f1e70ec19c43f00aa611bf555596912bc6f9a9f169f8785918e7bab9924001895798ff13f05842"))
	fmt.Println(hex.EncodeToString(pk.Bytes()))
	fmt.Println(base64.StdEncoding.EncodeToString(pk.PubKey().Bytes()))
	fmt.Println(base64.StdEncoding.EncodeToString(pk.PubKey().Address()))
	panic("")
}

func Test_PrivValidatorVote(t *testing.T) {
	type testCase struct {
		// name is the name of the test case.
		name string
		// lastSigned is the last signed vote.
		// it can be nil.
		lastSigned *cmtproto.Vote
		// vote is the vote to sign.
		vote *cmtproto.Vote
		// secondVote is the second vote to sign.
		// it can be nil.
		secondVote *cmtproto.Vote

		// chainid is the default chain ID to use.
		chainID string

		// privKey is the private key to use.
		privKey string

		// err is the expected error.
		// if nil, no error is expected.
		err error

		// after is a function to run after the test case.
		// it can be nil.
		after func(t *testing.T, tc *testCase)
	}

	tests := []testCase{
		// {
		// 	name:    "signing a vote with no other votes signed",
		// 	vote:    testVote(),
		// 	chainID: defaultChainID,
		// 	privKey: defaultPrivateKey,
		// },
		// {
		// 	name:       "signing two separate votes, validly",
		// 	vote:       testVote(height(1)),
		// 	secondVote: testVote(height(2)),
		// 	chainID:    defaultChainID,
		// 	privKey:    defaultPrivateKey,
		// },
		// {
		// 	name:       "signing a vote with a different previous vote signed",
		// 	lastSigned: testVote(height(1)),
		// 	vote:       testVote(height(2)),
		// 	chainID:    defaultChainID,
		// 	privKey:    defaultPrivateKey,
		// },
		// {
		// 	name:       "signing the same vote despite it being signed already, first vote is last signed",
		// 	lastSigned: testVote(signed("sig")),
		// 	vote:       testVote(),
		// 	chainID:    defaultChainID,
		// 	privKey:    defaultPrivateKey,
		// 	after: func(t *testing.T, tc *testCase) {
		// 		// it should have the same signature as the last signed vote.
		// 		assert.Equal(t, tc.lastSigned.Signature, tc.vote.Signature)
		// 	},
		// },
		{
			name:       "signing same vote twice, with different timestamps",
			lastSigned: testVote(signed("sig"), timestamped(100)),
			vote:       testVote(timestamped(200)),
			chainID:    defaultChainID,
			privKey:    defaultPrivateKey,
			after: func(t *testing.T, tc *testCase) {
				// it should have the same signature as the last signed vote.
				assert.Equal(t, tc.lastSigned.Signature, tc.vote.Signature)
				assert.Equal(t, tc.lastSigned.Timestamp.UTC(), tc.vote.Timestamp.UTC())
			},
		},
		{
			name:       "test height regression",
			lastSigned: testVote(height(100)),
			vote:       testVote(height(99)),
			chainID:    defaultChainID,
			privKey:    defaultPrivateKey,
			err:        privval.ErrHeightRegression,
		},
		{
			name:       "test round regression",
			lastSigned: testVote(round(100)),
			vote:       testVote(round(99)),
			chainID:    defaultChainID,
			privKey:    defaultPrivateKey,
			err:        privval.ErrRoundRegression,
		},
		{
			name:       "test step regression",
			lastSigned: testVote(step(cmtproto.PrecommitType)),
			vote:       testVote(step(cmtproto.PrevoteType)),
			chainID:    defaultChainID,
			privKey:    defaultPrivateKey,
			err:        privval.ErrStepRegression,
		},
	}

	// test cases for signing a vote.
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// outer function to catch any returns errors
			err := func() error {
				privKeyBts, err := hex.DecodeString(tc.privKey)
				if err != nil {
					return err
				}

				store := newMockStore()
				if tc.lastSigned != nil {
					if err := setKeys(store, tc.lastSigned.Height, tc.lastSigned.Round, privval.VoteToStep(tc.lastSigned), tc.lastSigned.Signature, types.VoteSignBytes(tc.chainID, tc.lastSigned)); err != nil {
						return err
					}
				}

				privVal, err := privval.NewValidatorSigner(privKeyBts, store)
				if err != nil {
					return err
				}

				// sign the vote
				err = privVal.SignVote(tc.chainID, tc.vote)
				if err != nil {
					return err
				}

				assert.NotNil(t, tc.vote.Signature)

				if tc.secondVote != nil {
					// sign the second vote
					err = privVal.SignVote(tc.chainID, tc.secondVote)
					if err != nil {
						return err
					}

					assert.Equal(t, tc.vote.Timestamp, tc.secondVote.Timestamp)
				}

				return nil
			}()
			if err != nil {
				if tc.err == nil {
					t.Fatalf("unexpected error: %v", err)
				}

				assert.ErrorIs(t, err, tc.err)
				return
			}
			if tc.err != nil {
				t.Fatalf("expected error: %v", tc.err)
			}

			if tc.after != nil {
				tc.after(t, &tc)
			}
		})
	}
}

func Test_Proposals(t *testing.T) {
	type testCase struct {
		// name is the name of the test case.
		name string
		// lastSigned is the last signed vote.
		// it can be nil.
		lastSigned *cmtproto.Proposal
		// vote is the vote to sign.
		vote *cmtproto.Proposal
		// secondVote is the second vote to sign.
		// it can be nil.
		secondVote *cmtproto.Proposal

		// err is the expected error.
		// if nil, no error is expected.
		err error

		// after is a function to run after the test case.
		// it can be nil.
		after func(t *testing.T, tc *testCase)
	}

	tests := []testCase{
		{
			name: "signing a vote with no other votes signed",
			vote: testProposal(),
		},
		{
			name:       "signing two separate votes, validly",
			vote:       testProposal(height(1)),
			secondVote: testProposal(height(2)),
		},
		{
			name:       "signing a vote with a different previous vote signed",
			lastSigned: testProposal(height(1)),
			vote:       testProposal(height(2)),
		},
		{
			name:       "signing the same vote despite it being signed already, first vote is last signed",
			lastSigned: testProposal(signed("sig")),
			vote:       testProposal(),

			after: func(t *testing.T, tc *testCase) {
				// it should have the same signature as the last signed vote.
				assert.Equal(t, tc.lastSigned.Signature, tc.vote.Signature)
			},
		},
		{
			name:       "signing same vote twice, with different timestamps",
			lastSigned: testProposal(signed("sig"), timestamped(100)),
			vote:       testProposal(timestamped(200)),

			after: func(t *testing.T, tc *testCase) {
				// it should have the same signature as the last signed vote.
				assert.Equal(t, tc.lastSigned.Signature, tc.vote.Signature)
				assert.Equal(t, tc.lastSigned.Timestamp.UTC(), tc.vote.Timestamp.UTC())
			},
		},
		{
			name:       "test height regression",
			lastSigned: testProposal(height(100)),
			vote:       testProposal(height(99)),

			err: privval.ErrHeightRegression,
		},
		{
			name:       "test round regression",
			lastSigned: testProposal(round(100)),
			vote:       testProposal(round(99)),

			err: privval.ErrRoundRegression,
		},
	}

	// test cases for signing a vote.
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// outer function to catch any returns errors
			err := func() error {
				privKeyBts, err := hex.DecodeString(defaultPrivateKey)
				if err != nil {
					return err
				}

				store := newMockStore()
				if tc.lastSigned != nil {
					if err := setKeys(store, tc.lastSigned.Height, tc.lastSigned.Round, 1, tc.lastSigned.Signature, types.ProposalSignBytes(defaultChainID, tc.lastSigned)); err != nil {
						return err
					}
				}

				privVal, err := privval.NewValidatorSigner(privKeyBts, store)
				if err != nil {
					return err
				}

				// sign the vote
				err = privVal.SignProposal(defaultChainID, tc.vote)
				if err != nil {
					return err
				}

				assert.NotNil(t, tc.vote.Signature)

				if tc.secondVote != nil {
					// sign the second vote
					err = privVal.SignProposal(defaultChainID, tc.secondVote)
					if err != nil {
						return err
					}

					assert.Equal(t, tc.vote.Timestamp, tc.secondVote.Timestamp)
				}

				return nil
			}()
			if err != nil {
				if tc.err == nil {
					t.Fatalf("unexpected error: %v", err)
				}

				assert.ErrorIs(t, err, tc.err)
				return
			}
			if tc.err != nil {
				t.Fatalf("expected error: %v", tc.err)
			}

			if tc.after != nil {
				tc.after(t, &tc)
			}
		})
	}
}

// setKeys from vote sets the keys in the AtomicKV from the vote.
// this is useful for testing starting up the atomic KV with an existing vote.
func setKeys(store privval.AtomicReadWriter, ht int64, rnd int32, stp int8, signature []byte, signBytes []byte) error {
	latest := privval.LastSignState{
		Height:    ht,
		Round:     rnd,
		Step:      stp,
		Signature: signature,
		SignBytes: signBytes,
	}

	latestBts, err := json.Marshal(latest)
	if err != nil {
		return err
	}

	return store.Write(latestBts)
}

// mockStore implements AtomicReadWriter
type mockStore struct {
	latest []byte
}

func newMockStore() *mockStore {
	return &mockStore{
		latest: nil,
	}
}

// testVote is a valid vote for height 1
func testVote(opts ...testVotOpt) *cmtproto.Vote {
	options := defaultOptions()
	for _, opt := range opts {
		opt(options)
	}

	return &cmtproto.Vote{
		Type:   options.step,
		Height: options.height,
		Round:  options.round,
		BlockID: cmtproto.BlockID{
			Hash: hash("hash1"),
			PartSetHeader: cmtproto.PartSetHeader{
				Total: 1,
				Hash:  hash("hash12"),
			},
		},
		Timestamp:        time.Unix(options.timestamp, 0),
		ValidatorAddress: []byte("validator1"),
		ValidatorIndex:   1,
		Signature:        options.signature,
	}
}

func testProposal(opts ...testVotOpt) *cmtproto.Proposal {
	options := defaultOptions()
	for _, opt := range opts {
		opt(options)
	}
	if options.step != 1 {
		panic("cannot create proposal with step != 1")
	}

	return &cmtproto.Proposal{
		Type:     1,
		Height:   options.height,
		Round:    options.round,
		PolRound: 0,
		BlockID: cmtproto.BlockID{
			Hash: hash("hash1"),
			PartSetHeader: cmtproto.PartSetHeader{
				Total: 1,
				Hash:  hash("hash12"),
			},
		},
		Timestamp: time.Unix(options.timestamp, 0),
		Signature: options.signature,
	}
}

func defaultOptions() *testVoteOptions {
	return &testVoteOptions{
		timestamp: 500,
		height:    10,
		round:     0,
		step:      cmtproto.PrevoteType,
	}
}

type testVoteOptions struct {
	timestamp int64
	signature []byte
	height    int64
	round     int32
	step      cmtproto.SignedMsgType
}

type testVotOpt func(*testVoteOptions)

func timestamped(ts int64) testVotOpt {
	return func(opts *testVoteOptions) {
		opts.timestamp = ts
	}
}

func signed(sig string) testVotOpt {
	return func(opts *testVoteOptions) {
		opts.signature = []byte(sig)
	}
}

func height(h int64) testVotOpt {
	return func(opts *testVoteOptions) {
		opts.height = h
	}
}

func round(r int32) testVotOpt {
	return func(opts *testVoteOptions) {
		opts.round = r
	}
}

func step(s cmtproto.SignedMsgType) testVotOpt {
	return func(opts *testVoteOptions) {
		opts.step = s
	}
}

func (m *mockStore) Read() ([]byte, error) {
	return m.latest, nil
}

func (m *mockStore) Write(p0 []byte) error {
	m.latest = p0
	return nil
}

func hash(s string) []byte {
	hasher := sha256.New()
	hasher.Write([]byte(s))
	return hasher.Sum(nil)
}
