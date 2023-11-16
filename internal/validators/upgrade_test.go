package validators

import (
	"context"
	"os"
	"testing"

	"github.com/kwilteam/kwil-db/core/log"
	"github.com/kwilteam/kwil-db/internal/sql/adapter"
	"github.com/kwilteam/kwil-db/internal/sql/sqlite"
)

func setup(srcFile string) {
	// Copies the db file to tmp
	os.MkdirAll("tmp", os.ModePerm)
	bts, err := os.ReadFile(srcFile)
	if err != nil {
		panic(err)
	}

	err = os.WriteFile("./tmp/validator_db.sqlite", bts, os.ModePerm)
	if err != nil {
		panic(err)
	}
}

func TestValidatorStoreUpgradeLegacy(t *testing.T) {
	setup("./test_data/version0.sqlite")

	defer deleteTempDir()

	ctx := context.Background()

	pool, err := sqlite.NewPool(ctx, "./tmp/validator_db.sqlite", 1, 1, true)
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()

	ds := &adapter.PoolAdapater{Pool: pool}

	logger := log.NewStdOut(log.DebugLevel)

	// validator store
	vs := &validatorStore{
		db:  ds,
		log: logger,
	}

	// Verify validator count is 1
	results, err := vs.db.Query(ctx, "SELECT COUNT(*) FROM validators", nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 1 {
		t.Fatalf("Expected 1 result, got %d", len(results))
	}

	// Expected values: version 0, action upgradeActionLegacy
	version, action, err := vs.checkVersion(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if version != 0 {
		t.Fatalf("Expected version 0, got %d", version)
	}
	if action != upgradeActionRunMigrations {
		t.Fatalf("Expected action %s, got %s",
			upgradeActionString(upgradeActionRunMigrations),
			upgradeActionString(action))
	}

	//	Expect failure as expiresAt column doesn't exist in legacy code
	_, _, err = vs.ActiveVotes(ctx)
	if err == nil {
		t.Fatal(err)
	}

	// Upgrade DB to latest
	err = vs.initOrUpgradeDatabase(ctx)
	if err != nil {
		t.Fatal(err)
	}

	// Check Version Table to ensure version is latest
	version, err = vs.currentVersion(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if version != valStoreVersion {
		t.Fatalf("Expected version %d, got %d", valStoreVersion, version)
	}
}

func TestValidatorStoreUpgradeV1(t *testing.T) {
	setup("./test_data/version1.sqlite")

	defer deleteTempDir()

	ctx := context.Background()

	pool, err := sqlite.NewPool(ctx, "./tmp/validator_db.sqlite", 1, 1, true)
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()

	ds := &adapter.PoolAdapater{Pool: pool}

	logger := log.NewStdOut(log.DebugLevel)

	realValStoreVersion := valStoreVersion
	defer func() {
		valStoreVersion = realValStoreVersion
	}()

	valStoreVersion = 1

	// validator store
	vs := &validatorStore{
		db:  ds,
		log: logger,
	}

	// Verify validator count is 1
	results, err := vs.db.Query(ctx, "SELECT COUNT(*) FROM validators", nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 1 {
		t.Fatalf("Expected 1 result, got %d", len(results))
	}

	// Expected values:  version 1, action upgradeActionNone
	versionPre, action, err := vs.checkVersion(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if versionPre != 1 {
		t.Fatalf("Expected version 0, got %d", versionPre)
	}
	if action != upgradeActionNone {
		t.Fatalf("Expected action %s, got %s",
			upgradeActionString(upgradeActionNone),
			upgradeActionString(action))
	}

	// Three entries in join_reqs table with expiresAt column
	votes, err := vs.allActiveJoinReqs(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(votes) != 3 {
		t.Fatalf("Starting votes not empty (%d)", len(votes))
	}

	// Upgrade
	err = vs.initOrUpgradeDatabase(ctx)
	if err != nil {
		t.Fatal(err)
	}

	// Check Version Table
	versionPost, err := vs.currentVersion(ctx)
	if err != nil {
		t.Fatal(err)
	}
	// Version should be 1, no upgrade required
	if versionPost != versionPre {
		t.Fatalf("Expected version %d, got %d", versionPre, versionPost)
	}

	valStoreVersion = 2

	// Now it should upgrade from 1->2

	// Upgrade
	err = vs.initOrUpgradeDatabase(ctx)
	if err != nil {
		t.Fatal(err)
	}

	// Check Version Table
	versionPost, err = vs.currentVersion(ctx)
	if err != nil {
		t.Fatal(err)
	}
	// Version should be 1, no upgrade required
	if versionPost != valStoreVersion {
		t.Fatalf("Expected version %d, got %d", versionPre, valStoreVersion)
	}
}

func TestValidatorStoreUpgradeFuture(t *testing.T) {
	setup("./test_data/version9999.sqlite")

	defer deleteTempDir()

	ctx := context.Background()

	pool, err := sqlite.NewPool(ctx, "./tmp/validator_db.sqlite", 1, 1, true)
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()

	ds := &adapter.PoolAdapater{Pool: pool}

	logger := log.NewStdOut(log.DebugLevel)

	// validator store
	vs := &validatorStore{
		db:  ds,
		log: logger,
	}

	// invalid version
	_, _, err = vs.checkVersion(ctx)
	if err == nil {
		t.Fatal("expected an error for an unknown version, but got none")
	}

	// Upgrade should fail as version is invalid
	err = vs.initOrUpgradeDatabase(ctx)
	if err == nil {
		t.Fatal(err)
	}
}
