package validators

import (
	"bytes"
	"context"
	"errors"
	"testing"

	"github.com/kwilteam/kwil-db/pkg/log"
)

// These tests are internal, designed to isolate the core ValidatorMgr logic
// with a stub ValidatorStore (not just a testing sqlite DB). The black box
// tests for ValidatorMgr with the exported API are done in mgr_test.go.

// stubValStore is an in-memory store that implements the ValidatorStore interface.
type stubValStore struct {
	current []*Validator
	joins   []*JoinRequest
}

func (vs *stubValStore) Init(_ context.Context, vals []*Validator) error {
	vs.current = vals
	vs.joins = nil
	return nil
}

func (vs *stubValStore) CurrentValidators(context.Context) ([]*Validator, error) {
	return vs.current, nil
}

func (vs *stubValStore) RemoveValidator(_ context.Context, validator []byte) error {
	i := findValidator(validator, vs.current)
	if i == -1 {
		return errors.New("not present")
	}
	vs.current = append(vs.current[:i], vs.current[i+1:]...)
	return nil
}

func (vs *stubValStore) UpdateValidatorPower(_ context.Context, validator []byte, power int64) error {
	i := findValidator(validator, vs.current)
	if i == -1 {
		return errors.New("not present")
	}
	vs.current[i].Power = power
	return nil
}

func (vs *stubValStore) ActiveVotes(context.Context) ([]*JoinRequest, error) {
	return vs.joins, nil
}

func (vs *stubValStore) StartJoinRequest(_ context.Context, joiner []byte, approvers [][]byte, power int64, expiresAt int64) error {
	vs.joins = append(vs.joins, &JoinRequest{
		Candidate: joiner,
		Power:     power,
		Board:     approvers,
		ExpiresAt: expiresAt,
		Approved:  make([]bool, len(approvers)),
	})
	return nil
}

func (vs *stubValStore) DeleteJoinRequest(_ context.Context, joiner []byte) error {
	for i, ji := range vs.joins {
		if bytes.Equal(ji.Candidate, joiner) {
			vs.joins = append(vs.joins[:i], vs.joins[i+1:]...)
			return nil
		}
	}
	return errors.New("unknown candidate")
}

func (vs *stubValStore) AddApproval(_ context.Context, joiner []byte, approver []byte) error {
	for _, ji := range vs.joins {
		if bytes.Equal(ji.Candidate, joiner) {
			for i, ai := range ji.Board {
				if bytes.Equal(approver, ai) {
					ji.Approved[i] = true
					return nil
				}
			}
			return errors.New("not an approver for candidate")
		}
	}
	return errors.New("unknown candidate")
}

func (vs *stubValStore) AddValidator(_ context.Context, joiner []byte, power int64) error {
	vs.current = append(vs.current, &Validator{
		PubKey: joiner,
		Power:  power,
	})
	for i, ji := range vs.joins { // if there's a join request, clean it up
		if bytes.Equal(ji.Candidate, joiner) {
			vs.joins = append(vs.joins[:i], vs.joins[i+1:]...)
			break
		}
	}
	return nil
}

func newTestValidatorMgr(t *testing.T, store ValidatorStore) *ValidatorMgr {
	mgr := &ValidatorMgr{
		current:    make(map[string]struct{}),
		candidates: make(map[string]*joinReq),
		log:        log.NewStdOut(log.DebugLevel),
		db:         store,
		joinExpiry: 3,
	}
	if err := mgr.init(); err != nil {
		t.Fatal(err)
	}
	return mgr
}

func TestValidatorMgr_updates(t *testing.T) {
	ctx := context.Background()

	var numVals = 8
	resumeSet := make([]*Validator, numVals)
	for i := range resumeSet {
		resumeSet[i] = newValidator()
		resumeSet[i].Power = 1
	}
	store := &stubValStore{
		current: resumeSet,
	}

	// Build a ValidatorMgr with the in-memory store.
	mgr := newTestValidatorMgr(t, store)

	vals, err := mgr.CurrentSet(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(vals) != numVals {
		t.Errorf("expected %d in validator set, found %d", numVals, len(vals))
	}

	// Reinit to a genesis set containing a smaller set.
	numVals = 3
	genesisSet := make([]*Validator, numVals)
	copy(genesisSet, resumeSet)
	err = mgr.GenesisInit(ctx, genesisSet, 1)
	if err != nil {
		t.Fatal(err)
	}
	vals, err = mgr.CurrentSet(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(vals) != numVals {
		t.Errorf("expected %d in validator set, found %d", numVals, len(vals))
	}

	// Now we can test Join/Approve/Leave
	// thresh := threshold(numVals) // i.e. 2

	// existing, Join fail
	err = mgr.Join(ctx, genesisSet[0].PubKey, genesisSet[0].Power)
	if err == nil {
		t.Errorf("no error for exiting validator trying to join")
	}
	// new, Join success
	joiner := newValidator()
	joiner.Power = 1
	err = mgr.Join(ctx, joiner.PubKey, joiner.Power)
	if err != nil {
		t.Errorf("new validator failed to make a join request")
	}

	// Approve non-existent candidate, fail
	noone := newValidator()
	val0 := genesisSet[0].PubKey
	err = mgr.Approve(ctx, noone.PubKey, val0)
	if err == nil {
		t.Errorf("no error approving with non-existent join request")
	}

	// Approve existing candidate, invalid approver
	err = mgr.Approve(ctx, joiner.PubKey, noone.PubKey)
	if err == nil {
		t.Errorf("no error approval from non-validator")
	}

	// Approve existing candidate, self-approve
	err = mgr.Approve(ctx, joiner.PubKey, joiner.PubKey)
	if err == nil {
		t.Errorf("no error approval from non-validator")
	}

	// Approve existing candidate, self-approve
	err = mgr.Approve(ctx, joiner.PubKey, val0)
	if err != nil {
		t.Errorf("valid approval failed: %v", err)
	}

	// Approving twice, no error, but not counted
	err = mgr.Approve(ctx, joiner.PubKey, val0)
	if err != nil {
		t.Errorf("valid approval failed: %v", err)
	}

	// Should be no validator updates yet (subthresh at 1 of 2 required)
	updates := mgr.Finalize(ctx)
	if len(updates) != 0 {
		t.Fatalf("wanted no validator updates, got %d", len(updates))
	}

	// Second of two required approves
	val1 := genesisSet[1].PubKey
	err = mgr.Approve(ctx, joiner.PubKey, val1)
	if err != nil {
		t.Errorf("valid approval failed: %v", err)
	}
	updates = mgr.Finalize(ctx)
	if len(updates) != 1 {
		t.Fatalf("wanted a validator update, got %d", len(updates))
	}
	if !bytes.Equal(updates[0].PubKey, joiner.PubKey) {
		t.Errorf("added validator pubkey incorrect")
	}
	if updates[0].Power != joiner.Power {
		t.Errorf("validator added with power %d, wanted %d", updates[0].Power, joiner.Power)
	}

	// Finalize again (another block), should be empty updates
	updates = mgr.Finalize(ctx)
	if len(updates) != 0 {
		t.Fatalf("wanted no validator updates, got %d", len(updates))
	}
}
