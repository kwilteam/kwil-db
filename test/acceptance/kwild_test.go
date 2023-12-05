package acceptance_test

import (
	"context"
	"flag"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/kwilteam/kwil-db/test/acceptance"
	"github.com/kwilteam/kwil-db/test/specifications"
)

var dev = flag.Bool("dev", false, "run for development purpose (no tests)")
var remote = flag.Bool("remote", false, "test against remote node")

var drivers = flag.String("drivers", "grpc,cli", "comma separated list of drivers to run")

// TestKwildAcceptance runs acceptance tests again a single kwild node(and
// are not concurrent), using different drivers: clientDriver, cliDriver.
// The tests here are not exhaustive, and are meant to only test happy paths.
func TestKwildAcceptance(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	ctx := context.Background()

	helper := acceptance.NewActHelper(t)
	helper.LoadConfig()

	if !*remote {
		helper.Setup(ctx)
	}

	// running forever for local development
	if *dev {
		helper.WaitUntilInterrupt()
	}

	testDrivers := strings.Split(*drivers, ",")
	for _, driverType := range testDrivers {
		// NOTE: those tests should not be run concurrently

		t.Run(driverType+"_driver", func(t *testing.T) {
			creatorDriver := helper.GetDriver(driverType, "creator")

			// When user deployed database
			//specifications.DatabaseDeployInvalidSql1Specification(ctx, t, creatorDriver)
			start := time.Now()
			specifications.DatabaseDeployInvalidExtensionSpecification(ctx, t, creatorDriver)
			fmt.Println("DatabaseDeployInvalidExtensionSpecification took", time.Since(start))
			specifications.DatabaseDeploySpecification(ctx, t, creatorDriver)
			fmt.Println("DatabaseDeploySpecification took", time.Since(start))

			//Then user should be able to execute database
			specifications.ExecuteOwnerActionSpecification(ctx, t, creatorDriver)
			fmt.Println("ExecuteOwnerActionSpecification took", time.Since(start))

			// TODO: This test doesn't looks good, the spec suppose to expect
			// only one parameter, the driver.
			// Read `test/specifications/README.md` for more information.
			db := specifications.SchemaLoader.Load(t, specifications.SchemaTestDB)
			dbid := creatorDriver.DBID(db.Name)
			visitorDriver := helper.GetDriver(driverType, "visitor")
			specifications.ExecuteOwnerActionFailSpecification(ctx, t, visitorDriver, dbid)
			fmt.Println("ExecuteOwnerActionFailSpecification took", time.Since(start))

			specifications.ExecuteDBInsertSpecification(ctx, t, creatorDriver)
			fmt.Println("ExecuteDBInsertSpecification took", time.Since(start))
			specifications.ExecuteCallSpecification(ctx, t, creatorDriver, visitorDriver)
			fmt.Println("ExecuteCallSpecification took", time.Since(start))

			specifications.ExecuteDBUpdateSpecification(ctx, t, creatorDriver)
			fmt.Println("ExecuteDBUpdateSpecification took", time.Since(start))
			specifications.ExecuteDBDeleteSpecification(ctx, t, creatorDriver)
			fmt.Println("ExecuteDBDeleteSpecification took", time.Since(start))

			// test that the loaded extensions works
			specifications.ExecuteExtensionSpecification(ctx, t, creatorDriver)
			fmt.Println("ExecuteExtensionSpecification took", time.Since(start))

			// and user should be able to drop database
			specifications.DatabaseDropSpecification(ctx, t, creatorDriver)
			fmt.Println("DatabaseDropSpecification took", time.Since(start))

			specifications.ExecuteChainInfoSpecification(ctx, t, creatorDriver, acceptance.TestChainID)
			fmt.Println("ExecuteChainInfoSpecification took", time.Since(start))
			// there's one node in the network and we're the validator
			// @brennan I am commenting this out temporarily, but it seems to be _mostly_ useless
			// all it does is check that the node is a validator, which is not really a useful test,
			// and couples the rest of acceptance to the validator rpcs, which should probably
			// be a standalone set of tests anyways
			//specifications.CurrentValidatorsSpecification(ctx, t, creatorDriver, 1)

			// The other network/validator specs require multiple nodes in a network
		})
	}

	t.FailNow()
}
