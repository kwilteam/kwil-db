package privval

/*
	Much of the code in this package is inspired or pulled directly from cometbft/privval,
	[https://github.com/cometbft/cometbft/blob/1fb31e05d304bcbdbf01aaeb073781aa9f441e34/privval/file.go#L100]

	Licensed under the Apache License, Version 2.0
*/

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	cometEd25519 "github.com/cometbft/cometbft/crypto/ed25519"
	cmtbytes "github.com/cometbft/cometbft/libs/bytes"
	"github.com/cometbft/cometbft/libs/protoio"
	tendermintTypes "github.com/cometbft/cometbft/proto/tendermint/types"
	cmttime "github.com/cometbft/cometbft/types/time"
	"github.com/cosmos/gogoproto/proto"

	"github.com/cometbft/cometbft/crypto"
	"github.com/cometbft/cometbft/types"
)

// NewValidatorSigner returns a new ValidatorSigner
// it takes in an ed25519 key, and a keyvalue store
// the key values store should NOT be atomically committed with other KV
// stores.  Instead, it should simply fsync after every write/commit
func NewValidatorSigner(privKey cometEd25519.PrivKey, storer AtomicReadWriter) (*ValidatorSigner, error) {
	if len(privKey.Bytes()) != cometEd25519.PrivateKeySize {
		return nil, fmt.Errorf("invalid private key size.  received: %d, expected: %d",
			len(privKey.Bytes()), cometEd25519.PrivateKeySize)
	}

	lss := &LastSignState{
		storer: storer,
	}
	err := lss.loadLatest()
	if err != nil {
		return nil, err
	}

	return &ValidatorSigner{
		privateKey:      privKey,
		lastSignedState: lss,
	}, nil
}

// ValidatorSigner implements CometBFT's cometTypes.PrivValidator
// It persists its most recent signature, which can be used during
// recovery to prevent double signing
type ValidatorSigner struct {
	// privateKey is the ed25519 private key used to sign messages
	privateKey crypto.PrivKey

	// lastSignedState is the most recent signature made by this validator
	lastSignedState *LastSignState
}

var _ types.PrivValidator = (*ValidatorSigner)(nil)

// GetPubKey returns the public key of the validator
// It is part of the cometTypes.PrivValidator interface
func (v *ValidatorSigner) GetPubKey() (crypto.PubKey, error) {
	return v.privateKey.PubKey(), nil

}

// SignProposal signs a proposal message
// It is part of the cometTypes.PrivValidator interface
func (v *ValidatorSigner) SignProposal(chainID string, proposal *tendermintTypes.Proposal) error {
	height, round, step := proposal.Height, proposal.Round, stepPropose

	sameHRS, err := v.lastSignedState.checkHRS(height, round, step)
	if err != nil {
		return err
	}

	signBytes := types.ProposalSignBytes(chainID, proposal)

	// We might crash before writing to the wal,
	// causing us to try to re-sign for the same HRS.
	// If signbytes are the same, use the last signature.
	// If they only differ by timestamp, use last timestamp and signature
	// Otherwise, return error
	if sameHRS {
		if bytes.Equal(signBytes, v.lastSignedState.SignBytes) {
			proposal.Signature = v.lastSignedState.Signature
		} else if timestamp, ok := checkProposalsOnlyDifferByTimestamp(v.lastSignedState.SignBytes, signBytes); ok {
			proposal.Signature = v.lastSignedState.Signature
			proposal.Timestamp = timestamp
		} else {
			err = fmt.Errorf("proposal sign bytes differ from last sign bytes")
		}
		return err
	}

	// Sign the proposal
	signature, err := v.signAndPersist(height, round, step, signBytes)
	if err != nil {
		return err
	}

	// Set the proposal signature
	proposal.Signature = signature

	return nil
}

// SignVote signs a vote message
// It is part of the cometTypes.PrivValidator interface
func (v *ValidatorSigner) SignVote(chainID string, vote *tendermintTypes.Vote) error {
	height, round, step := vote.Height, vote.Round, VoteToStep(vote)

	sameHRS, err := v.lastSignedState.checkHRS(height, round, step)
	if err != nil {
		return err
	}

	signBytes := types.VoteSignBytes(chainID, vote)

	// Vote extensions are non-deterministic, so it is possible that an
	// application may have created a different extension. We therefore always
	// re-sign the vote extensions of precommits. For prevotes and nil
	// precommits, the extension signature will always be empty.
	// Even if the signed over data is empty, we still add the signature.
	var extSig []byte
	if vote.Type == tendermintTypes.PrecommitType && !types.ProtoBlockIDIsNil(&vote.BlockID) {
		extSignBytes := types.VoteExtensionSignBytes(chainID, vote)
		extSig, err = v.privateKey.Sign(extSignBytes)
		if err != nil {
			return err
		}
	} else if len(vote.Extension) > 0 {
		return errors.New("unexpected vote extension - extensions are only allowed in non-nil precommits")
	}

	// We might crash before writing to the wal,
	// causing us to try to re-sign for the same HRS.
	// If signbytes are the same, use the last signature.
	// If they only differ by timestamp, use last timestamp and signature
	// Otherwise, return error
	if sameHRS {
		if bytes.Equal(signBytes, v.lastSignedState.SignBytes) {
			vote.Signature = v.lastSignedState.Signature
		} else if timestamp, ok := checkVotesOnlyDifferByTimestamp(v.lastSignedState.SignBytes, signBytes); ok {
			vote.Timestamp = timestamp
			vote.Signature = v.lastSignedState.Signature
		} else {
			err = fmt.Errorf("conflicting data")
		}

		vote.ExtensionSignature = extSig

		return err
	}

	// Sign the vote
	signature, err := v.signAndPersist(height, round, step, signBytes)
	if err != nil {
		return err
	}

	// Set the vote signature
	vote.Signature = signature
	vote.ExtensionSignature = extSig

	return nil
}

// signAndPersist signs a message and persists the signature
// it returns the signature after it has been persisted
func (v *ValidatorSigner) signAndPersist(height int64, round int32, step int8, signBytes []byte) ([]byte, error) {
	signature, err := v.privateKey.Sign(signBytes)
	if err != nil {
		return nil, err
	}

	v.lastSignedState.Height = height
	v.lastSignedState.Round = round
	v.lastSignedState.Step = step
	v.lastSignedState.SignBytes = signBytes
	v.lastSignedState.Signature = signature
	err = v.lastSignedState.store()
	if err != nil {
		return nil, err
	}

	return signature, nil
}

// LastSignState tracks the most recent signature
// made by this validator.  It is atomically committed to disk
// before it is used for anything else, and can be reloaded in case
// of a crash
type LastSignState struct {
	// Height is the height of the block that the message was signed for
	Height int64 `json:"height"`

	// Round is the consensus round that the message was signed for
	// CometBFT can have an arbitrary number of rounds per height
	Round int32 `json:"round"`

	// Step is the consensus step that the message was signed for
	// e.g. propose, prevote, precommit
	Step int8 `json:"step"`

	// Signature is the signature generated by the validator
	Signature []byte `json:"signature"`

	// SignBytes is the bytes that were signed by the validator
	SignBytes cmtbytes.HexBytes `json:"sign_bytes"`

	// storer is the store that this lastSignState is persisted to
	storer AtomicReadWriter
}

// store stores the lastSignState to the given KV store
// it is atomic, and will only commit if all writes succeed
func (l *LastSignState) store() error {
	bts, err := json.Marshal(l)
	if err != nil {
		return err
	}
	return l.storer.Write(bts)
}

// loadLatest loads the latest lastSignState from the given KV store
// if none exists, it sets all fields to zero values
func (l *LastSignState) loadLatest() (err error) {
	bts, err := l.storer.Read()
	if err != nil {
		return err
	}

	if bts == nil {
		l.setZero()
		return nil
	}

	return json.Unmarshal(bts, l)
}

// setZero sets all fields to zero values
func (l *LastSignState) setZero() {
	l.Height = 0
	l.Round = 0
	l.Step = 0
	l.Signature = nil
	l.SignBytes = nil
}

// checkHRS checks that the given height, round, and step match the lastSignState.
func (lss *LastSignState) checkHRS(height int64, round int32, step int8) (bool, error) {

	if lss.Height > height {
		return false, fmt.Errorf("%w: height regression. Got %v, last height %v", ErrHeightRegression, height, lss.Height)
	}

	if lss.Height == height {
		if lss.Round > round {
			return false, fmt.Errorf("%w: round regression at height %v. Got %v, last round %v", ErrRoundRegression, height, round, lss.Round)
		}

		if lss.Round == round {
			if lss.Step > step {
				return false, fmt.Errorf(
					"%w: step regression at height %v round %v. Got %v, last step %v",
					ErrStepRegression,
					height,
					round,
					step,
					lss.Step,
				)
			} else if lss.Step == step {
				if lss.SignBytes != nil {
					if lss.Signature == nil {
						return false, fmt.Errorf("%w: Signature is nil but SignBytes is not", ErrNilSignature)
					}
					return true, nil
				}
				return false, errors.New("no SignBytes found")
			}
		}
	}
	return false, nil
}

// Returns the timestamp from the lastSignBytes.
// Returns true if the only difference in the votes is their timestamp.
// Performs these checks on the canonical votes (excluding the vote extension
// and vote extension signatures).
func checkVotesOnlyDifferByTimestamp(lastSignBytes, newSignBytes []byte) (time.Time, bool) {
	var lastVote, newVote tendermintTypes.CanonicalVote
	if err := protoio.UnmarshalDelimited(lastSignBytes, &lastVote); err != nil {
		panic(fmt.Sprintf("LastSignBytes cannot be unmarshalled into vote: %v", err))
	}
	if err := protoio.UnmarshalDelimited(newSignBytes, &newVote); err != nil {
		panic(fmt.Sprintf("signBytes cannot be unmarshalled into vote: %v", err))
	}

	lastTime := lastVote.Timestamp
	// set the times to the same value and check equality
	now := cmttime.Now()
	lastVote.Timestamp = now
	newVote.Timestamp = now

	return lastTime, proto.Equal(&newVote, &lastVote)
}

// returns the timestamp from the lastSignBytes.
// returns true if the only difference in the proposals is their timestamp
func checkProposalsOnlyDifferByTimestamp(lastSignBytes, newSignBytes []byte) (time.Time, bool) {
	var lastProposal, newProposal tendermintTypes.CanonicalProposal
	if err := protoio.UnmarshalDelimited(lastSignBytes, &lastProposal); err != nil {
		panic(fmt.Sprintf("LastSignBytes cannot be unmarshalled into proposal: %v", err))
	}
	if err := protoio.UnmarshalDelimited(newSignBytes, &newProposal); err != nil {
		panic(fmt.Sprintf("signBytes cannot be unmarshalled into proposal: %v", err))
	}

	lastTime := lastProposal.Timestamp
	// set the times to the same value and check equality
	now := cmttime.Now()
	lastProposal.Timestamp = now
	newProposal.Timestamp = now

	return lastTime, proto.Equal(&newProposal, &lastProposal)
}

// this should be unexported, but is needed for testing
// A vote is either stepPrevote or stepPrecommit.
func VoteToStep(vote *tendermintTypes.Vote) int8 {
	switch vote.Type {
	case tendermintTypes.PrevoteType:
		return stepPrevote
	case tendermintTypes.PrecommitType:
		return stepPrecommit
	default:
		panic(fmt.Sprintf("Unknown vote type: %v", vote.Type))
	}
}

const (
	stepNone      int8 = 0 // Used to distinguish the initial state
	stepPropose   int8 = 1
	stepPrevote   int8 = 2
	stepPrecommit int8 = 3
)

// AtomicReadWriter is an interface for any store
// that can atomically read and write to a persistent store
type AtomicReadWriter interface {
	// Write should overwrite the current value with the given value
	Write([]byte) error
	// Read should return the current value
	// if the value is empty, it should return empty bytes and no error
	Read() ([]byte, error)
}

var (
	ErrHeightRegression = errors.New("height regression")
	ErrRoundRegression  = errors.New("round regression")
	ErrStepRegression   = errors.New("step regression")
	ErrNilSignature     = errors.New("signature is nil")
)
