package integration_tests

import (
	"context"
	"flag"
	"fmt"
	"os/exec"
	"testing"
	"time"

	"github.com/kwilteam/kwil-db/pkg/log"
	"github.com/kwilteam/kwil-db/test/acceptance"
	"github.com/kwilteam/kwil-db/test/specifications"
	"github.com/stretchr/testify/assert"
)

var remote = flag.Bool("remote", false, "run tests against remote environment")
var dev = flag.Bool("dev", false, "run for development purpose (no tests)")

func setupConfig() {
	cmd := exec.Command("cp", "--recursive", "./cluster_data/kwil/k1/node0-cpy", "./cluster_data/kwil/k1/node0")
	cmd.Run()
	cmd = exec.Command("cp", "--recursive", "./cluster_data/kwil/k2/node1-cpy", "./cluster_data/kwil/k2/node1")
	cmd.Run()
	cmd = exec.Command("cp", "--recursive", "./cluster_data/kwil/k3/node2-cpy", "./cluster_data/kwil/k3/node2")
	cmd.Run()
}

func teardownConfig() {
	cmd := exec.Command("rm", "-rf", "./cluster_data/kwil/k1/node0")
	cmd.Run()
	cmd = exec.Command("cp", "-rf", "./cluster_data/kwil/k2/node1")
	cmd.Run()
	cmd = exec.Command("cp", "-rf", "./cluster_data/kwil/k3/node2")
	cmd.Run()
}

func TestKwildIntegration(t *testing.T) {
	tLogger := log.New(log.Config{
		Level:       "info",
		OutputPaths: []string{"stdout"},
	})

	cfg := acceptance.GetTestEnvCfg(t, *remote)
	ctx := context.Background()
	// to stop mining blocks for current subtest
	done := make(chan struct{})

	// Bringup the KWIL DB cluster with 3 nodes
	cfg.DBSchemaFilePath = "./test-data/test_db.kf"
	setupConfig()
	fmt.Println("ChainRPCURL: ", cfg.ChainRPCURL)
	cfg, kwildC, chainDeployer := acceptance.SetupKwildCluster(ctx, t, cfg)

	//time.Sleep(30 * time.Second)
	// Create Kwil DB clients for each node
	node1Driver := acceptance.SetupKwildDriver(ctx, t, cfg, kwildC[0], tLogger)
	node2Driver := acceptance.SetupKwildDriver(ctx, t, cfg, kwildC[1], tLogger)
	node3Driver := acceptance.SetupKwildDriver(ctx, t, cfg, kwildC[2], tLogger)

	correctPrivateKey := cfg.UserPrivateKey
	correctPrivateKeyString := cfg.UserPrivateKeyString
	cfg.UserPrivateKey = cfg.SecondUserPrivateKey
	cfg.UserPrivateKeyString = cfg.SecondUserPrivateKeyString

	// Create invalid user driver
	invalidUserDriver := acceptance.SetupKwildDriver(ctx, t, cfg, kwildC[0], tLogger)
	cfg.UserPrivateKey = correctPrivateKey
	cfg.UserPrivateKeyString = correctPrivateKeyString

	// Fund both the User accounts
	err := chainDeployer.FundAccount(ctx, cfg.UserAddr, cfg.InitialFundAmount)
	assert.NoError(t, err, "failed to fund user account")

	err = chainDeployer.FundAccount(ctx, cfg.SecondUserAddr, cfg.InitialFundAmount)
	assert.NoError(t, err, "failed to fund second user account")

	go acceptance.KeepMiningBlocks(ctx, done, chainDeployer, cfg.UserAddr)

	// and user pledged fund to validator
	fmt.Println("Approve token1")
	specifications.ApproveTokenSpecification(ctx, t, node1Driver)
	fmt.Print("Deposit fund1")
	time.Sleep(15 * time.Second)
	specifications.DepositFundSpecification(ctx, t, node1Driver)

	// second user
	fmt.Println("Approve token2")
	specifications.ApproveTokenSpecification(ctx, t, invalidUserDriver)
	time.Sleep(15 * time.Second)
	fmt.Print("Deposit fund2")
	specifications.DepositFundSpecification(ctx, t, invalidUserDriver)
	time.Sleep(cfg.ChainSyncWaitTime)
	// running forever for local development
	if *dev {
		acceptance.DumpEnv(&cfg)
		<-done
	}

	// Create a new database and verify that the database exists on other nodes
	fmt.Printf("Create database")
	specifications.DatabaseDeploySpecification(ctx, t, node1Driver)
	time.Sleep(30 * time.Second)
	specifications.DatabaseVerifySpecification(ctx, t, node2Driver, true)
	specifications.DatabaseVerifySpecification(ctx, t, node3Driver, true)

	// // Execute actions on the database
	specifications.ExecuteDBDeleteSpecification(ctx, t, node1Driver)
	specifications.ExecuteDBUpdateSpecification(ctx, t, node2Driver)
	specifications.ExecuteDBDeleteSpecification(ctx, t, node3Driver)

	// // Test permissioned actions
	// specifications.ExecutePermissionedActionSpecification(ctx, t, invalidUserDriver)

	// // Drop the database and verify that the database does not exist on other nodes
	// specifications.DatabaseDropSpecification(ctx, t, node1Driver)
	// specifications.DatabaseVerifySpecification(ctx, t, node2Driver, false)
	// specifications.DatabaseVerifySpecification(ctx, t, node3Driver, false)
	// Teardown
	teardownConfig()
}
