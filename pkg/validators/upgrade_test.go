package validators

import (
	"context"
	"os"
	"testing"

	"github.com/kwilteam/kwil-db/pkg/log"
	sqlTesting "github.com/kwilteam/kwil-db/pkg/sql/testing"
)

func setup(srcfile string) {
	os.MkdirAll("tmp", os.ModePerm)
	bts, err := os.ReadFile(srcfile)
	if err != nil {
		panic(err)
	}

	err = os.WriteFile("./tmp/validator_db.sqlite", bts, os.ModePerm)
	if err != nil {
		panic(err)
	}
}
func TestValidatorStoreUpgradeLegacyToV1(t *testing.T) {
	setup("./test_data/version0.sqlite")
	ds, td, err := sqlTesting.OpenTestDB("validator_db")
	if err != nil {
		t.Fatal(err)
	}
	defer td()
	ctx := context.Background()
	logger := log.NewStdOut(log.DebugLevel)

	// vs
	vs := &validatorStore{
		db:  ds,
		log: logger,
	}

	results, err := vs.db.Query(ctx, "SELECT COUNT(*) FROM validators", nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 1 {
		t.Fatalf("Expected 1 result, got %d", len(results))
	}

	// CheckVersion
	version, action, err := vs.CheckVersion(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if version != 0 {
		t.Fatalf("Expected version 0, got %d", version)
	}
	if action != upgradeActionLegacy {
		t.Fatalf("Expected action %d, got %d", upgradeActionLegacy, action)
	}

	// Get JoinRequest entries
	_, err = vs.ActiveVotes(ctx)
	if err == nil {
		t.Fatal(err)
	}

	// Upgrade
	err = vs.databaseUpgrade(ctx)
	if err != nil {
		t.Fatal(err)
	}

	// Check Version Table
	version, err = vs.currentVersion(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if version != valStoreVersion {
		t.Fatalf("Expected version %d, got %d", valStoreVersion, version)
	}

	// Get JoinRequest entries
	votes, err := vs.ActiveVotes(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(votes) != 3 {
		t.Fatalf("Starting votes not empty (%d)", len(votes))
	}

}

func TestValidatorStoreUpgradeV1(t *testing.T) {
	setup("./test_data/version1.sqlite")
	ds, td, err := sqlTesting.OpenTestDB("validator_db")
	if err != nil {
		t.Fatal(err)
	}
	defer td()
	ctx := context.Background()
	logger := log.NewStdOut(log.DebugLevel)

	// vs
	vs := &validatorStore{
		db:  ds,
		log: logger,
	}

	results, err := vs.db.Query(ctx, "SELECT COUNT(*) FROM validators", nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 1 {
		t.Fatalf("Expected 1 result, got %d", len(results))
	}

	// CheckVersion
	version, action, err := vs.CheckVersion(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if version != 1 {
		t.Fatalf("Expected version 0, got %d", version)
	}
	if action != upgradeActionNone {
		t.Fatalf("Expected action %d, got %d", upgradeActionLegacy, action)
	}

	// Get JoinRequest entries
	votes, err := vs.ActiveVotes(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(votes) != 3 {
		t.Fatalf("Starting votes not empty (%d)", len(votes))
	}

	// Upgrade
	err = vs.databaseUpgrade(ctx)
	if err != nil {
		t.Fatal(err)
	}

	// Check Version Table
	version, err = vs.currentVersion(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if version != valStoreVersion {
		t.Fatalf("Expected version %d, got %d", valStoreVersion, version)
	}

	// Get JoinRequest entries
	votes, err = vs.ActiveVotes(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(votes) != 3 {
		t.Fatalf("Starting votes not empty (%d)", len(votes))
	}
}
