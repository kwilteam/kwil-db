//go:build pglive

package validators

import (
	"bytes"
	"context"
	"testing"

	"github.com/kwilteam/kwil-db/core/log"
	"github.com/kwilteam/kwil-db/internal/sql/adapter"
	"github.com/kwilteam/kwil-db/internal/sql/pg"
)

// create user kwil_test_user with SUPERUSER replication;
// create database kwil_test_db owner kwil_test_user;
// create publication kwild_repl for all tables;

func Test_validatorStore(t *testing.T) {
	ctx := context.Background()

	cfg := &pg.PoolConfig{
		ConnConfig: pg.ConnConfig{
			Host:   "/var/run/postgresql",
			Port:   "",
			User:   "kwil_test_user",
			Pass:   "kwil", // would be ignored if pg_hba.conf set with trust
			DBName: "kwil_test_db",
		},
		MaxConns: 11,
	}
	db, err := pg.NewPool(ctx, cfg)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	defer db.Execute(ctx, `DROP SCHEMA IF EXISTS `+schemaName+` CASCADE`)

	ds := &adapter.DB{Datastore: db}

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

	votes, removals, err := vs.ActiveVotes(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(votes) != 0 {
		t.Fatalf("Starting votes not empty (%d)", len(votes))
	}
	if len(removals) != 0 {
		t.Fatalf("Starting removals not empty (%d)", len(removals))
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
	expiresAt := int64(3)
	err = vs.StartJoinRequest(ctx, joiner.PubKey, approvers, wantPower, expiresAt)
	if err != nil {
		t.Fatal(err)
	}

	// Expire the join request & delete it
	err = vs.DeleteJoinRequest(ctx, joiner.PubKey)
	if err != nil {
		t.Fatal(err)
	}

	// Add approval for expired join request
	err = vs.AddApproval(ctx, joiner.PubKey, v0Key)
	if err == nil {
		t.Fatalf("no error approving expired join requests")
	}

	joins, removals, err := vs.ActiveVotes(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(joins) != 0 {
		t.Fatalf("expected 0 active join requests, found %d", len(joins))
	}
	if len(removals) != 0 {
		t.Fatalf("expected 0 active removals, found %d", len(removals))
	}

	// Start a new join request
	err = vs.StartJoinRequest(ctx, joiner.PubKey, approvers, wantPower, expiresAt)
	if err != nil {
		t.Fatal(err)
	}

	joins, removals, err = vs.ActiveVotes(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(removals) != 0 {
		t.Fatalf("expected 0 active removals, found %d", len(removals))
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
	joins, removals, err = vs.ActiveVotes(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(removals) != 0 {
		t.Fatalf("expected 0 active removals, found %d", len(removals))
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
	joins, removals, err = vs.ActiveVotes(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(removals) != 0 {
		t.Fatalf("expected 0 active removals, found %d", len(removals))
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
