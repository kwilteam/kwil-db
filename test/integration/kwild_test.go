package integration_test

import (
	"context"
	"flag"
	"fmt"
	"github.com/kwilteam/kwil-db/test/integration"
	"os"
	"os/signal"
	"syscall"
	"testing"
	"time"

	"github.com/kwilteam/kwil-db/test/specifications"
)

var remote = flag.Bool("remote", false, "run tests against remote environment")
var dev = flag.Bool("dev", false, "run for development purpose (no tests)")

func TestKwildDatabaseIntegration(t *testing.T) {
	ctx := context.Background()

	helper := integration.NewIntHelper(t)
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

	drivers := helper.GetDrivers(ctx)
	node0Driver := drivers[0]
	node1Driver := drivers[1]
	node2Driver := drivers[2]

	// Create a new database and verify that the database exists on other nodes
	fmt.Printf("Create database")
	specifications.DatabaseDeploySpecification(ctx, t, node0Driver)
	time.Sleep(30 * time.Second)
	specifications.DatabaseVerifySpecification(ctx, t, node1Driver, true)
	specifications.DatabaseVerifySpecification(ctx, t, node2Driver, true)

	specifications.ExecuteDBInsertSpecification(ctx, t, node0Driver)
	specifications.ExecuteDBUpdateSpecification(ctx, t, node1Driver)
	specifications.ExecuteDBDeleteSpecification(ctx, t, node2Driver)

	// specifications.ExecutePermissionedActionSpecification(ctx, t, invalidUserDriver)

	specifications.DatabaseDropSpecification(ctx, t, node1Driver)
}

//func TestKwildNetworkIntegration(t *testing.T) {
//
//
//	tLogger := log.New(log.Config{
//		Level:       "info",
//		OutputPaths: []string{"stdout"},
//	})
//
//	cfg := integration.GetTestEnvCfg(t, *remote)
//	ctx := context.Background()
//	// to stop mining blocks for current subtest
//	done := make(chan struct{})
//
//	// Bringup the KWIL DB cluster with 3 nodes
//	cfg.SchemaFile = "./test-data/test_db.kf"
//	//setupConfig()
//	fmt.Println("ChainRPCURL: ", cfg.ChainRPCURL)
//
//	cfg, kwildC := integration.SetupKwildCluster(ctx, t, cfg, path)
//
//	//time.Sleep(30 * time.Second)
//	// Create Kwil DB clients for each node
//	node0Driver := integration.SetupKwildDriver(ctx, t, cfg, kwildC[0], tLogger)
//	node1Driver := integration.SetupKwildDriver(ctx, t, cfg, kwildC[1], tLogger)
//	node2Driver := integration.SetupKwildDriver(ctx, t, cfg, kwildC[2], tLogger)
//
//	/* no more token
//	// Fund both the User accounts
//	err := chainDeployer.FundAccount(ctx, cfg.UserAddr, cfg.InitialFundAmount)
//	assert.NoError(t, err, "failed to fund user account")
//
//	err = chainDeployer.FundAccount(ctx, cfg.SecondUserAddr, cfg.InitialFundAmount)
//	assert.NoError(t, err, "failed to fund second user account")
//
//	go integration.KeepMiningBlocks(ctx, done, chainDeployer, cfg.UserAddr)
//
//	// and user pledged fund to validator
//
//	fmt.Println("Approve token1")
//	specifications.ApproveTokenSpecification(ctx, t, node0Driver)
//	fmt.Print("Deposit fund1")
//	time.Sleep(15 * time.Second)
//	specifications.DepositFundSpecification(ctx, t, node0Driver)
//
//	time.Sleep(cfg.ChainSyncWaitTime)
//	*/
//
//	// running forever for local development
//	if *dev {
//		integration.DumpEnv(&cfg)
//		<-done
//	}
//
//	node0PrivKey := "3za9smSSrMoaLUgzJcEncG79gn3dyeYxoYIielhvygIECZfoKhPmiR/RDtr79o/Jxe6jRUxJkVoZoeA/9NHZhQ=="
//	node1PubKey := "R0gA+mgclmqknbiTJrnVPfE0i9kCgSNoxJkHqpwh4f0="
//	node1PrivKey := "6uyWNA9LJNSBp0QNfQpDWZp+RxV+D8wFvll7duhudFhHSAD6aByWaqSduJMmudU98TSL2QKBI2jEmQeqnCHh/Q=="
//	node2PubKey := "9JL8gRIIvit2GgSPOnoCv1ZCTnTC33z9VjOdIi6iwgI="
//
//	// Create a new database and verify that the database exists on other nodes
//	fmt.Printf("Create database")
//
//	/*
//		Start with Genesis Node0
//		- Node1 requests to join
//		- Requires node0 to approve
//	*/
//	specifications.NetworkNodeValidatorSetSpecification(ctx, t, node0Driver, 1)
//	specifications.NetworkNodeJoinSpecification(ctx, t, node1Driver, node1PubKey)
//	specifications.NetworkNodeValidatorSetSpecification(ctx, t, node0Driver, 1)
//
//	specifications.NetworkNodeApproveSpecification(ctx, t, node0Driver, node1PubKey, node0PrivKey)
//	specifications.NetworkNodeDeploySpecification(ctx, t, node0Driver)
//	specifications.NetworkNodeValidatorSetSpecification(ctx, t, node2Driver, 2)
//
//	/* > 2/3rd majority approvals
//	1. Node 0 and Node1 are the current validators
//	2. Node2 requests to join
//	3. Node0 and Node1 need to approve for majority approval
//	*/
//	specifications.NetworkNodeJoinSpecification(ctx, t, node2Driver, node2PubKey)
//	specifications.NetworkNodeValidatorSetSpecification(ctx, t, node0Driver, 2)
//
//	specifications.NetworkNodeApproveSpecification(ctx, t, node0Driver, node2PubKey, node0PrivKey)
//	specifications.NetworkNodeDeploySpecification(ctx, t, node0Driver)
//	specifications.NetworkNodeValidatorSetSpecification(ctx, t, node1Driver, 2)
//
//	specifications.NetworkNodeApproveSpecification(ctx, t, node1Driver, node2PubKey, node1PrivKey)
//	specifications.NetworkNodeDeploySpecification(ctx, t, node1Driver)
//	specifications.NetworkNodeValidatorSetSpecification(ctx, t, node2Driver, 3)
//
//	specifications.NetworkNodeLeaveSpecification(ctx, t, node1Driver, node1PubKey)
//	specifications.NetworkNodeDeploySpecification(ctx, t, node1Driver)
//	specifications.NetworkNodeValidatorSetSpecification(ctx, t, node1Driver, 2)
//
//	close(done)
//}
