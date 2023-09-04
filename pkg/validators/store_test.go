package validators

import (
	"bytes"
	"context"
	"crypto/rand"
	"testing"

	"github.com/kwilteam/kwil-db/pkg/log"
	sqlTesting "github.com/kwilteam/kwil-db/pkg/sql/testing"
	"github.com/kwilteam/kwil-db/pkg/utils/random"
)

func Test_validatorStore(t *testing.T) {
	ds, td, err := sqlTesting.OpenTestDB("test_validator_store")
	if err != nil {
		t.Fatal(err)
	}
	defer td()

	ctx := context.Background()
	logger := log.NewStdOut(log.DebugLevel)
	vs, err := newValidatorStore(ctx, ds, logger)
	if err != nil {
		t.Fatal(err)
	}

	// This "test" steps through a positive use case while testing negative paths.

	// Ensure fresh store is usable
	vals, err := vs.CurrentValidators(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(vals) != 0 {
		t.Fatalf("Starting validator set not empty (%d)", len(vals))
	}

	votes, err := vs.ActiveVotes(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(votes) != 0 {
		t.Fatalf("Starting votes not empty (%d)", len(votes))
	}

	// Init for genesis validator set
	numValidators := 8 // 2/3 is 5.3333 => 6

	vals = make([]*Validator, numValidators)
	for i := range vals {
		vals[i] = newValidator()
	}

	err = vs.Init(ctx, vals)
	if err != nil {
		t.Fatal(err)
	}

	valsOut, err := vs.CurrentValidators(ctx)
	if err != nil {
		t.Fatal(err)
	}
	// Check that the slices are the same, disregarding order.
	if len(valsOut) != len(vals) {
		t.Fatalf("wanted %v validators, got %v", len(vals), len(valsOut))
	}
	for _, vi := range vals {
		i := findValidator(vi.PubKey, valsOut)
		if i == -1 {
			t.Fatalf("missing validator %v", vi.PubKey)
		}
		if valsOut[i].Power != vi.Power {
			t.Fatal("loaded validator power incorrect")
		}
	}

	// Update power
	const newPower = 8
	v0Key := vals[0].PubKey
	err = vs.UpdateValidatorPower(ctx, v0Key, newPower)
	if err != nil {
		t.Fatal(err)
	}
	valsOut, err = vs.CurrentValidators(ctx)
	if err != nil {
		t.Fatal(err)
	}
	i := findValidator(v0Key, valsOut)
	if i == -1 {
		t.Fatalf("missing validator %v", v0Key)
	}
	if valsOut[i].Power != newPower {
		t.Fatal("loaded validator power incorrect")
	}

	// Add a new validator (but no join or approves)
	vX := newValidator()
	numValidators++
	err = vs.AddValidator(ctx, vX.PubKey, vX.Power)
	if err != nil {
		t.Fatal(err)
	}
	valsOut, err = vs.CurrentValidators(ctx)
	if err != nil {
		t.Fatal(err)
	}
	i = findValidator(vX.PubKey, valsOut)
	if i == -1 {
		t.Fatalf("missing validator %v", vX.PubKey)
	}
	if numValidators != len(valsOut) {
		t.Fatalf("wanted %d validators, got %d", numValidators, len(valsOut))
	}
	err = vs.AddValidator(ctx, vX.PubKey, vX.Power)
	if err == nil {
		t.Fatal("expected an error re-adding an existing and empowered validator")
	}

	// Join requests
	joiner := newValidator()

	// Add approval for non-existent join request
	err = vs.AddApproval(ctx, joiner.PubKey, v0Key)
	if err == nil {
		t.Fatalf("no error approving non-existent join requests")
	}

	const wantPower = 11
	approvers := make([][]byte, len(valsOut))
	for i, vi := range valsOut {
		approvers[i] = vi.PubKey
	}
	err = vs.StartJoinRequest(ctx, joiner.PubKey, approvers, wantPower)
	if err != nil {
		t.Fatal(err)
	}

	joins, err := vs.ActiveVotes(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(joins) != 1 {
		t.Fatalf("expected 1 active join request, found %d", len(joins))
	}
	if !bytes.Equal(joins[0].Candidate, joiner.PubKey) {
		t.Fatalf("incorrect candidate pubkey in join request")
	}
	if joins[0].Power != wantPower {
		t.Errorf("wanted power %d in join request, got %d", wantPower, joins[0].Power)
	}
outer:
	for i, pki := range joins[0].Board {
		if joins[0].Approved[i] {
			t.Errorf("initial join request contained approval")
		}
		for _, pkj := range approvers {
			if bytes.Equal(pki, pkj) {
				continue outer
			}
		}
		t.Errorf("approver not found")
	}

	err = vs.AddApproval(ctx, joiner.PubKey, v0Key)
	if err != nil {
		t.Fatalf("unable to add approval")
	}
	joins, err = vs.ActiveVotes(ctx)
	if err != nil {
		t.Fatal(err)
	}
	var present bool
	for i, pki := range joins[0].Board {
		if bytes.Equal(pki, v0Key) {
			if !joins[0].Approved[i] {
				t.Error("approval not recorded")
			}
			present = true
			break
		}
	}
	if !present {
		t.Fatalf("approver missing")
	}

	// Let's say one vote is good enough.
	err = vs.AddValidator(ctx, joiner.PubKey, wantPower)
	if err != nil {
		t.Fatal(err)
	}
	numValidators++
	// the join request should be removed
	joins, err = vs.ActiveVotes(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(joins) != 0 {
		t.Error("inactive join request not removed on validator add")
	}

	valsOut, err = vs.CurrentValidators(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(valsOut) != numValidators {
		t.Errorf("expected %d validators, got %d", numValidators, len(valsOut))
	}
	if findValidator(joiner.PubKey, valsOut) == -1 {
		t.Errorf("new validator set did not include added validator")
	}
}

func findValidator(pubkey []byte, vals []*Validator) int {
	for i, v := range vals {
		if bytes.Equal(v.PubKey, pubkey) {
			return i
		}
	}
	return -1
}

var rng = random.New()

func randomBytes(l int) []byte {
	b := make([]byte, l)
	_, _ = rand.Read(b)
	return b
}

func newValidator() *Validator {
	return &Validator{
		PubKey: randomBytes(32),
		Power:  rng.Int63n(4) + 1, // in {1,2,3,4}
	}
}
