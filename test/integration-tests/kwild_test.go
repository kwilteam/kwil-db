package integration_tests

import (
	"context"
	"flag"
	"fmt"
	"path/filepath"
	"testing"
	"time"

	"github.com/kwilteam/kwil-db/pkg/log"
	"github.com/kwilteam/kwil-db/pkg/utils"
	"github.com/kwilteam/kwil-db/test/acceptance"
	"github.com/kwilteam/kwil-db/test/specifications"
)

var remote = flag.Bool("remote", false, "run tests against remote environment")
var dev = flag.Bool("dev", false, "run for development purpose (no tests)")

func teardownConfig(path string) {
	n0_path := filepath.Join(path, "kwil/node0")
	fmt.Println("Path: ", n0_path)

	utils.ResetAll(filepath.Join(n0_path, "data"), filepath.Join(n0_path, "config/addrbook.json"),
		filepath.Join(n0_path, "config/priv_validator_key.json"),
		filepath.Join(n0_path, "data/priv_validator_state.json"))

	n1_path := filepath.Join(path, "kwil/node1")
	utils.ResetAll(filepath.Join(n1_path, "data"), filepath.Join(n1_path, "config/addrbook.json"),
		filepath.Join(n1_path, "config/priv_validator_key.json"),
		filepath.Join(n1_path, "data/priv_validator_state.json"))

	n2_path := filepath.Join(path, "kwil/node2")
	utils.ResetAll(filepath.Join(n2_path, "data"), filepath.Join(n2_path, "config/addrbook.json"),
		filepath.Join(n2_path, "config/priv_validator_key.json"),
		filepath.Join(n2_path, "data/priv_validator_state.json"))
}

func TestKwildDatabaseIntegration(t *testing.T) {
	path := "./cluster_data/database/"
	teardownConfig(path)
	defer teardownConfig(path)

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
	//setupConfig()
	fmt.Println("ChainRPCURL: ", cfg.ChainRPCURL)

	cfg, kwildC := acceptance.SetupKwildCluster(ctx, t, cfg, path)

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
	// invalidUserDriver := acceptance.SetupKwildDriver(ctx, t, cfg, kwildC[0], tLogger)
	cfg.UserPrivateKey = correctPrivateKey
	cfg.UserPrivateKeyString = correctPrivateKeyString

	// Fund both the User accounts
	/* xxx
	err := chainDeployer.FundAccount(ctx, cfg.UserAddr, cfg.InitialFundAmount)
	assert.NoError(t, err, "failed to fund user account")

	err = chainDeployer.FundAccount(ctx, cfg.SecondUserAddr, cfg.InitialFundAmount)
	assert.NoError(t, err, "failed to fund second user account")

	go acceptance.KeepMiningBlocks(ctx, done, chainDeployer, cfg.UserAddr)
	*/

	// and user pledged fund to validator
	/* no more token
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

	time.Sleep(cfg.ChainSyncWaitTime)
	*/

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

	specifications.ExecuteDBInsertSpecification(ctx, t, node1Driver)
	specifications.ExecuteDBUpdateSpecification(ctx, t, node2Driver)
	specifications.ExecuteDBDeleteSpecification(ctx, t, node3Driver)

	// specifications.ExecutePermissionedActionSpecification(ctx, t, invalidUserDriver)

	specifications.DatabaseDropSpecification(ctx, t, node1Driver)
	close(done)
}

func TestKwildNetworkIntegration(t *testing.T) {
	path := "./cluster_data/network/"
	teardownConfig(path)
	defer teardownConfig(path)

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
	//setupConfig()
	fmt.Println("ChainRPCURL: ", cfg.ChainRPCURL)

	cfg, kwildC := acceptance.SetupKwildCluster(ctx, t, cfg, path)

	//time.Sleep(30 * time.Second)
	// Create Kwil DB clients for each node
	node0Driver := acceptance.SetupKwildDriver(ctx, t, cfg, kwildC[0], tLogger)
	node1Driver := acceptance.SetupKwildDriver(ctx, t, cfg, kwildC[1], tLogger)
	node2Driver := acceptance.SetupKwildDriver(ctx, t, cfg, kwildC[2], tLogger)

	/* no more token
	// Fund both the User accounts
	err := chainDeployer.FundAccount(ctx, cfg.UserAddr, cfg.InitialFundAmount)
	assert.NoError(t, err, "failed to fund user account")

	err = chainDeployer.FundAccount(ctx, cfg.SecondUserAddr, cfg.InitialFundAmount)
	assert.NoError(t, err, "failed to fund second user account")

	go acceptance.KeepMiningBlocks(ctx, done, chainDeployer, cfg.UserAddr)

	// and user pledged fund to validator

	fmt.Println("Approve token1")
	specifications.ApproveTokenSpecification(ctx, t, node0Driver)
	fmt.Print("Deposit fund1")
	time.Sleep(15 * time.Second)
	specifications.DepositFundSpecification(ctx, t, node0Driver)

	time.Sleep(cfg.ChainSyncWaitTime)
	*/

	// running forever for local development
	if *dev {
		acceptance.DumpEnv(&cfg)
		<-done
	}

	node0PrivKey := "3za9smSSrMoaLUgzJcEncG79gn3dyeYxoYIielhvygIECZfoKhPmiR/RDtr79o/Jxe6jRUxJkVoZoeA/9NHZhQ=="
	node1PubKey := "R0gA+mgclmqknbiTJrnVPfE0i9kCgSNoxJkHqpwh4f0="
	node1PrivKey := "6uyWNA9LJNSBp0QNfQpDWZp+RxV+D8wFvll7duhudFhHSAD6aByWaqSduJMmudU98TSL2QKBI2jEmQeqnCHh/Q=="
	node2PubKey := "9JL8gRIIvit2GgSPOnoCv1ZCTnTC33z9VjOdIi6iwgI="

	// Create a new database and verify that the database exists on other nodes
	fmt.Printf("Create database")

	/*
		Start with Genesis Node0
		- Node1 requests to join
		- Requires node0 to approve
	*/
	specifications.NetworkNodeValidatorSetSpecification(ctx, t, node0Driver, 1)
	specifications.NetworkNodeJoinSpecification(ctx, t, node1Driver, node1PubKey)
	specifications.NetworkNodeValidatorSetSpecification(ctx, t, node0Driver, 1)

	specifications.NetworkNodeApproveSpecification(ctx, t, node0Driver, node1PubKey, node0PrivKey)
	specifications.NetworkNodeDeploySpecification(ctx, t, node0Driver)
	specifications.NetworkNodeValidatorSetSpecification(ctx, t, node2Driver, 2)

	/* > 2/3rd majority approvals
	1. Node 0 and Node1 are the current validators
	2. Node2 requests to join
	3. Node0 and Node1 need to approve for majority approval
	*/
	specifications.NetworkNodeJoinSpecification(ctx, t, node2Driver, node2PubKey)
	specifications.NetworkNodeValidatorSetSpecification(ctx, t, node0Driver, 2)

	specifications.NetworkNodeApproveSpecification(ctx, t, node0Driver, node2PubKey, node0PrivKey)
	specifications.NetworkNodeDeploySpecification(ctx, t, node0Driver)
	specifications.NetworkNodeValidatorSetSpecification(ctx, t, node1Driver, 2)

	specifications.NetworkNodeApproveSpecification(ctx, t, node1Driver, node2PubKey, node1PrivKey)
	specifications.NetworkNodeDeploySpecification(ctx, t, node1Driver)
	specifications.NetworkNodeValidatorSetSpecification(ctx, t, node2Driver, 3)

	specifications.NetworkNodeLeaveSpecification(ctx, t, node1Driver, node1PubKey)
	specifications.NetworkNodeDeploySpecification(ctx, t, node1Driver)
	specifications.NetworkNodeValidatorSetSpecification(ctx, t, node1Driver, 2)

	close(done)
}
