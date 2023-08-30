package acceptance_test

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"syscall"
	"testing"

	"github.com/kwilteam/kwil-db/test/acceptance"
	"github.com/kwilteam/kwil-db/test/specifications"
)

var dev = flag.Bool("dev", false, "run for development purpose (no tests)")
var remote = flag.Bool("remote", false, "test against remote node")

func TestKwildGrpcAcceptance(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	ctx := context.Background()

	helper := acceptance.NewActHelper(t)
	helper.LoadConfig()

	if !*remote {
		helper.Setup(ctx)
		defer helper.Teardown()
	}

	// running forever for local development
	if *dev {
		done := make(chan os.Signal, 1)
		signal.Notify(done, syscall.SIGINT, syscall.SIGTERM)

		// block waiting for a signal
		s := <-done
		t.Logf("Got signal: %v\n", s)
		helper.Teardown()
		t.Logf("Teardown done\n")
		return
	}

	aliceDriver := helper.GetAliceDriver(ctx)
	bobDriver := helper.GetBobDriver(ctx)

	// When user deployed database
	//specifications.DatabaseDeployInvalidSqlSpecification(ctx, t, driver)
	//specifications.DatabaseDeployInvalidExtensionSpecification(ctx, t, driver)
	specifications.DatabaseDeploySpecification(ctx, t, aliceDriver)

	//// Then user should be able to execute database
	specifications.ExecuteOwnerActionSpecification(ctx, t, aliceDriver)

	specifications.ExecuteOwnerActionFailSpecification(ctx, t, bobDriver)
	specifications.ExecuteDBInsertSpecification(ctx, t, aliceDriver)
	specifications.ExecuteCallSpecification(ctx, t, aliceDriver)

	specifications.ExecuteDBUpdateSpecification(ctx, t, aliceDriver)
	specifications.ExecuteDBDeleteSpecification(ctx, t, aliceDriver)

	// test that the loaded extensions works
	specifications.ExecuteExtensionSpecification(ctx, t, aliceDriver)

	// and user should be able to drop database
	specifications.DatabaseDropSpecification(ctx, t, aliceDriver)
}
