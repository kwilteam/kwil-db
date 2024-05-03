package integration_test

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/kwilteam/kwil-db/test/integration"
	"github.com/kwilteam/kwil-db/test/specifications"
)

// We're testing gatewayclient and kwil-cli, so the tests are in kwil-db, not kgw.
// To test kgw, we need to run a kwil network with kgw services.
// Three ways to do this:
// 1. pull kgw repo, build kgw, run kgw services using `go`
// 2. pull kgw repo, build kgw, run kgw services using `docker-compose`
// 3. pull already built kgw image(built in kgw repo), run kgw services using `docker-compose`
// We choose 2, because it's the easiest, in CI.
//
// By default, we will skip kgw tests, so it's not a concern for local development.

// NOTE: KGW tests cannot be run in parallel, since the `domain` need to be
// preset, which require kgw always running(exposed) on `localhost:8090`.

func TestKGW(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping kgw test in short mode")
	}

	ctx := context.Background()

	opts := []integration.HelperOpt{
		integration.WithBlockInterval(time.Second),
		integration.WithValidators(4),
		integration.WithNonValidators(0),
	}

	testDrivers := strings.Split(*drivers, ",")
	for _, driverType := range testDrivers {
		if driverType == "http" { // http driver is not supported testing kgw
			continue
		}

		t.Run(driverType+"_driver", func(t *testing.T) {
			helper := integration.NewIntHelper(t, opts...)
			helper.Setup(ctx, basicServices)

			helper.RunDockerComposeWithServices(ctx, []string{"kgw"})
			// ensure kgw have upstream checked
			time.Sleep(time.Millisecond * 200)

			creatorDriver := helper.GetUserGatewayDriver(ctx, driverType, "creator")

			// When user deployed a database
			// invalid kuneiform
			specifications.DatabaseDeployInvalidSql1Specification(ctx, t, creatorDriver)
			specifications.DatabaseDeployInvalidExtensionSpecification(ctx, t, creatorDriver)
			// valid
			specifications.DatabaseDeploySpecification(ctx, t, creatorDriver)

			// Then user should be able to execute database actions
			visitorDriver := helper.GetUserGatewayDriver(ctx, driverType, "visitor")
			// creator should be able to execute all normal actions
			specifications.ExecuteOwnerActionSpecification(ctx, t, creatorDriver)
			specifications.ExecuteDBInsertSpecification(ctx, t, creatorDriver)
			specifications.ExecuteCallSpecification(ctx, t, creatorDriver, visitorDriver)
			specifications.ExecuteDBUpdateSpecification(ctx, t, creatorDriver)
			specifications.ExecuteDBDeleteSpecification(ctx, t, creatorDriver)
			// Also user should be able to call authn action after authentication
			db := specifications.SchemaLoader.Load(t, specifications.SchemaTestDB)
			dbid := creatorDriver.DBID(db.Name)
			// visitor can execute authn action
			// NOTE: due to the way we implemented client/cli for kgw authn, we cannot
			// test that authn action fails without authentication, e.g. the behavior
			// cannot be explicitly tested here
			//
			// successful call, bc schema on kgw is not synced yet, no authn rules are enforced
			specifications.ExecuteAuthnCallActionSpecification(ctx, t, visitorDriver, dbid)
			// sleep is necessary, longer than kgw schema sync interval
			time.Sleep(time.Second * 3)
			// successful call action, gatewayDriver will automatically authenticate if required
			specifications.ExecuteAuthnCallActionSpecification(ctx, t, visitorDriver, dbid)
			// successful call action, cookie from last call should be reused
			specifications.ExecuteAuthnCallActionSpecification(ctx, t, visitorDriver, dbid)
			// successful call procedure, cookie from last call should be reused
			specifications.ExecuteAuthnCallProcedureSpecification(ctx, t, visitorDriver, dbid)

			// test that the loaded extensions works
			specifications.ExecuteExtensionSpecification(ctx, t, creatorDriver)

			// and user should be able to drop database
			specifications.DatabaseDropSpecification(ctx, t, creatorDriver)
		})
	}
}
