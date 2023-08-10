package acceptance_test

import (
	"context"
	"flag"
	"fmt"
	"path/filepath"
	"testing"

	"github.com/kwilteam/kwil-db/pkg/log"
	"github.com/kwilteam/kwil-db/pkg/utils"
	"github.com/kwilteam/kwil-db/test/acceptance"
	"github.com/kwilteam/kwil-db/test/specifications"
)

var remote = flag.Bool("remote", false, "run tests against remote environment")
var dev = flag.Bool("dev", false, "run for development purpose (no tests)")

func teardownConfig(path string) {
	utils.ResetAll(filepath.Join(path, "data"), filepath.Join(path, "config/addrbook.json"),
		filepath.Join(path, "config/priv_validator_key.json"),
		filepath.Join(path, "data/priv_validator_state.json"))
}

func TestKwildAcceptance(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	path := "./test-data/"
	teardownConfig(filepath.Join(path, "kwil/node0"))
	defer teardownConfig(filepath.Join(path, "kwil/node0"))

	tLogger := log.New(log.Config{
		Level:       "info",
		OutputPaths: []string{"stdout"},
	})

	cases := []struct {
		name       string
		driverType string
	}{
		{"grpc driver", "grpc"},
		//{"cli driver", "cli"},
	}

	for _, c := range cases {
		t.Run(fmt.Sprintf("should execute database using %s", c.name), func(t *testing.T) {
			cfg := acceptance.GetTestEnvCfg(t, *remote)
			ctx := context.Background()
			// to stop mining blocks for current subtest
			done := make(chan struct{})

			// setup

			driver, runningCfg := acceptance.GetDriver(ctx, t, c.driverType, cfg, tLogger, path)
			secondDriver := acceptance.NewClient(ctx, t, c.driverType, runningCfg, tLogger)

			/* xxx

			// NOTE: only local env test, public network test takes too long
			// thus here test assume user is funded
			if !*remote {
				// Given user is funded
				err := chainDeployer.FundAccount(ctx, cfg.UserAddr, cfg.InitialFundAmount)
				assert.NoError(t, err, "failed to fund user account")

				err = chainDeployer.FundAccount(ctx, cfg.SecondUserAddr, cfg.InitialFundAmount)
				assert.NoError(t, err, "failed to fund second user account")

				go acceptance.KeepMiningBlocks(ctx, done, chainDeployer, cfg.UserAddr)

				// and user pledged fund to validator
				specifications.ApproveTokenSpecification(ctx, t, driver)
				specifications.DepositFundSpecification(ctx, t, driver)

				// second user
				specifications.ApproveTokenSpecification(ctx, t, secondDriver)
				specifications.DepositFundSpecification(ctx, t, secondDriver)
			}

			// chain sync, wait kwil to register user
			time.Sleep(cfg.ChainSyncWaitTime)

			*/

			// running forever for local development
			if *dev {
				acceptance.DumpEnv(&runningCfg)
				<-done
			}

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

			close(done)
		})
	}
}
