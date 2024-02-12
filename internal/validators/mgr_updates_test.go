package validators

import (
	"bytes"
	"context"
	"errors"
	"testing"

	"github.com/kwilteam/kwil-db/core/log"
	"github.com/kwilteam/kwil-db/internal/sql"
)

// These tests are internal, designed to isolate the core ValidatorMgr logic
// with a stub ValidatorStore (not just a testing sqlite DB). The black box
// tests for ValidatorMgr with the exported API are done in mgr_test.go.

// stubValStore is an in-memory store that implements the ValidatorStore interface.
type stubValStore struct {
	current []*Validator
	joins   []*JoinRequest
	removes []*ValidatorRemoveProposal
}

func (vs *stubValStore) Init(_ context.Context, _ sql.DB, vals []*Validator) error {
	vs.current = vals
	vs.joins = nil
	vs.removes = nil
	return nil
}

func (vs *stubValStore) CurrentValidators(context.Context, sql.DB) ([]*Validator, error) {
	return vs.current, nil
}

func (vs *stubValStore) RemoveValidator(_ context.Context, _ sql.DB, validator []byte) error {
	i := findValidator(validator, vs.current)
	if i == -1 {
		return errors.New("not present")
	}
	vs.current = append(vs.current[:i], vs.current[i+1:]...)
	return nil
}

func (vs *stubValStore) UpdateValidatorPower(_ context.Context, _ sql.DB, validator []byte, power int64) error {
	i := findValidator(validator, vs.current)
	if i == -1 {
		return errors.New("not present")
	}
	vs.current[i].Power = power
	return nil
}

func (vs *stubValStore) ActiveVotes(context.Context, sql.DB) ([]*JoinRequest, []*ValidatorRemoveProposal, error) {
	return vs.joins, nil, nil
}

func (vs *stubValStore) StartJoinRequest(_ context.Context, _ sql.DB, joiner []byte, approvers [][]byte, power int64, expiresAt int64) error {
	vs.joins = append(vs.joins, &JoinRequest{
		Candidate: joiner,
		Power:     power,
		Board:     approvers,
		ExpiresAt: expiresAt,
		Approved:  make([]bool, len(approvers)),
	})
	return nil
}

func (vs *stubValStore) DeleteJoinRequest(_ context.Context, _ sql.DB, joiner []byte) error {
	for i, ji := range vs.joins {
		if bytes.Equal(ji.Candidate, joiner) {
			vs.joins = append(vs.joins[:i], vs.joins[i+1:]...)
			return nil
		}
	}
	return errors.New("unknown candidate")
}

func (vs *stubValStore) AddRemoval(_ context.Context, _ sql.DB, target, validator []byte) error {
	return nil
}

func (vs *stubValStore) DeleteRemoval(ctx context.Context, _ sql.DB, target, validator []byte) error {
	return nil
}

func (vs *stubValStore) AddApproval(_ context.Context, _ sql.DB, joiner []byte, approver []byte) error {
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

func (vs *stubValStore) AddValidator(_ context.Context, _ sql.DB, joiner []byte, power int64) error {
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

func (vs *stubValStore) IsCurrent(_ context.Context, _ sql.DB, validator []byte) (bool, error) {
	return findValidator(validator, vs.current) != -1, nil
}

func newTestValidatorMgr(t *testing.T, store ValidatorStore, tx sql.DB) *ValidatorMgr {
	mgr := &ValidatorMgr{
		current:    make(map[string]struct{}),
		candidates: make(map[string]*joinReq),
		log:        log.NewStdOut(log.DebugLevel),
		db:         store,
		joinExpiry: 3,
	}
	if err := mgr.init(context.Background(), tx); err != nil {
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
	db := &mockDB{}

	// Build a ValidatorMgr with the in-memory store.
	mgr := newTestValidatorMgr(t, store, db)

	vals, err := mgr.CurrentSet(ctx, db)
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
	err = mgr.GenesisInit(ctx, db, genesisSet, 1)
	if err != nil {
		t.Fatal(err)
	}
	vals, err = mgr.CurrentSet(ctx, db)
	if err != nil {
		t.Fatal(err)
	}
	if len(vals) != numVals {
		t.Errorf("expected %d in validator set, found %d", numVals, len(vals))
	}

	// Now we can test Join/Approve/Leave
	// thresh := threshold(numVals) // i.e. 2

	// existing, Join fail
	err = mgr.Join(ctx, db, genesisSet[0].PubKey, genesisSet[0].Power)
	if err == nil {
		t.Errorf("no error for exiting validator trying to join")
	}
	// new, Join success
	joiner := newValidator()
	joiner.Power = 1
	err = mgr.Join(ctx, db, joiner.PubKey, joiner.Power)
	if err != nil {
		t.Errorf("new validator failed to make a join request")
	}

	// Approve non-existent candidate, fail
	noone := newValidator()
	val0 := genesisSet[0].PubKey
	err = mgr.Approve(ctx, db, noone.PubKey, val0)
	if err == nil {
		t.Errorf("no error approving with non-existent join request")
	}

	// Approve existing candidate, invalid approver
	err = mgr.Approve(ctx, db, joiner.PubKey, noone.PubKey)
	if err == nil {
		t.Errorf("no error approval from non-validator")
	}

	// Approve existing candidate, self-approve
	err = mgr.Approve(ctx, db, joiner.PubKey, joiner.PubKey)
	if err == nil {
		t.Errorf("no error approval from non-validator")
	}

	// Approve existing candidate, self-approve
	err = mgr.Approve(ctx, db, joiner.PubKey, val0)
	if err != nil {
		t.Errorf("valid approval failed: %v", err)
	}

	// Approving twice, no error, but not counted
	err = mgr.Approve(ctx, db, joiner.PubKey, val0)
	if err != nil {
		t.Errorf("valid approval failed: %v", err)
	}

	// Should be no validator updates yet (subthresh at 1 of 2 required)
	updates, _ := mgr.Finalize(ctx, db)
	if len(updates) != 0 {
		t.Fatalf("wanted no validator updates, got %d", len(updates))
	}

	// Second of two required approves
	val1 := genesisSet[1].PubKey
	err = mgr.Approve(ctx, db, joiner.PubKey, val1)
	if err != nil {
		t.Errorf("valid approval failed: %v", err)
	}
	updates, _ = mgr.Finalize(ctx, db)
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
	updates, _ = mgr.Finalize(ctx, db)
	if len(updates) != 0 {
		t.Fatalf("wanted no validator updates, got %d", len(updates))
	}
}

type mockDB struct{}

func (m *mockDB) AccessMode() sql.AccessMode {
	return sql.ReadWrite
}

func (m *mockDB) BeginTx(ctx context.Context) (sql.Tx, error) {
	return &mockTx{m}, nil
}

func (m *mockDB) Execute(ctx context.Context, stmt string, args ...any) (*sql.ResultSet, error) {
	return nil, nil
}

type mockTx struct {
	*mockDB
}

func (m *mockTx) Commit(ctx context.Context) error {
	return nil
}

func (m *mockTx) Rollback(ctx context.Context) error {
	return nil
}
