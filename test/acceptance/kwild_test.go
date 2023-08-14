package acceptance_test

import (
	"context"
	"flag"
	"github.com/kwilteam/kwil-db/test/acceptance"
	"github.com/kwilteam/kwil-db/test/specifications"
	"os"
	"os/signal"
	"syscall"
	"testing"
)

var remote = flag.Bool("remote", false, "run tests against remote environment")
var dev = flag.Bool("dev", false, "run for development purpose (no tests)")

func TestKwildGrpcAcceptance(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	ctx := context.Background()

	helper := acceptance.NewActHelper(t)
	helper.LoadConfig()
	helper.Setup(ctx)
	defer helper.Teardown()

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

	driver := helper.GetAliceDriver(ctx)
	secondDriver := helper.GetBobDriver(ctx)

	// When user deployed database
	specifications.DatabaseDeployInvalidSqlSpecification(ctx, t, driver)
	specifications.DatabaseDeployInvalidExtensionSpecification(ctx, t, driver)
	specifications.DatabaseDeploySpecification(ctx, t, driver)

	//// Then user should be able to execute database
	specifications.ExecuteOwnerActionSpecification(ctx, t, driver)

	specifications.ExecuteOwnerActionFailSpecification(ctx, t, secondDriver)
	specifications.ExecuteDBInsertSpecification(ctx, t, driver)
	specifications.ExecuteCallSpecification(ctx, t, driver)

	specifications.ExecuteDBUpdateSpecification(ctx, t, driver)
	specifications.ExecuteDBDeleteSpecification(ctx, t, driver)

	// test that the loaded extensions works
	specifications.ExecuteExtensionSpecification(ctx, t, driver)

	// and user should be able to drop database
	specifications.DatabaseDropSpecification(ctx, t, driver)
}
