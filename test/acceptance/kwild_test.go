package acceptance_test

import (
	"context"
	"flag"
	"strings"
	"testing"

	"github.com/kwilteam/kwil-db/test/acceptance"
	"github.com/kwilteam/kwil-db/test/specifications"
)

var dev = flag.Bool("dev", false, "run for development purpose (no tests)")
var remote = flag.Bool("remote", false, "test against remote node")

var drivers = flag.String("drivers", "http,cli", "comma separated list of drivers to run")

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
			specifications.DatabaseDeployInvalidExtensionSpecification(ctx, t, creatorDriver)
			specifications.DatabaseDeploySpecification(ctx, t, creatorDriver)

			//Then user should be able to execute database
			specifications.ExecuteOwnerActionSpecification(ctx, t, creatorDriver)

			// TODO: This test doesn't looks good, the spec suppose to expect
			// only one parameter, the driver.
			// Read `test/specifications/README.md` for more information.
			db := specifications.SchemaLoader.Load(t, specifications.SchemaTestDB)
			dbid := creatorDriver.DBID(db.Name)
			visitorDriver := helper.GetDriver(driverType, "visitor")
			specifications.ExecuteOwnerActionFailSpecification(ctx, t, visitorDriver, dbid)

			specifications.ExecuteDBInsertSpecification(ctx, t, creatorDriver)
			specifications.ExecuteCallSpecification(ctx, t, creatorDriver, visitorDriver)

			specifications.ExecuteDBUpdateSpecification(ctx, t, creatorDriver)
			specifications.ExecuteDBDeleteSpecification(ctx, t, creatorDriver)

			// test that the loaded extensions works
			specifications.ExecuteExtensionSpecification(ctx, t, creatorDriver)

			// and user should be able to drop database
			specifications.DatabaseDropSpecification(ctx, t, creatorDriver)

			// there's one node in the network and we're the validator
			specifications.CurrentValidatorsSpecification(ctx, t, creatorDriver, 1)

			// The other network/validator specs require multiple nodes in a network
		})
	}
}
