package integration_test

import (
	"context"
	"encoding/hex"
	"flag"
	"fmt"
	"math/big"
	"strings"
	"testing"
	"time"

	"github.com/kwilteam/kwil-db/core/client"
	"github.com/kwilteam/kwil-db/core/types"
	"github.com/kwilteam/kwil-db/core/types/transactions"
	"github.com/kwilteam/kwil-db/test/driver"
	"github.com/kwilteam/kwil-db/test/integration"
	"github.com/kwilteam/kwil-db/test/specifications"
	"github.com/stretchr/testify/require"
)

var dev = flag.Bool("dev", false, "run for development purpose (no tests)")

var spamTest = flag.Bool("spam", false, "run the spam test that requires a special docker image to be built")

var forkTest = flag.Bool("fork", false, "run the fork test that requires a special docker image to be built")

var drivers = flag.String("drivers", "jsonrpc,cli", "comma separated list of drivers to run")

// NOTE: `-parallel` is a flag that is already used by `go test`
var parallelMode = flag.Bool("parallel-mode", false, "run tests in parallel mode")

// Here we make clear the services will be used in each stage
var basicServices = []string{integration.ExtContainer, "pg0", "pg1", "pg2", "node0", "node1", "node2"}
var newServices = []string{integration.Ext3Container, "pg3", "node3"}

// NOTE: allServices will be sorted by docker-compose(in setup), so the order is not reliable
var allServices = []string{integration.ExtContainer, integration.Ext3Container,
	"pg0", "pg1", "pg2", "pg3", "node0", "node1", "node2", "node3",
}

var migrationServices = []string{"new-ext1", "new-pg0", "new-pg1", "new-pg2", "new-node0", "new-node1", "new-node2"}

var migrationServices2 = []string{"new-ext3", "new-pg3", "new-node3"}

var singleNodeServices = []string{integration.ExtContainer, "pg0", "node0"}

var byzAllServices = []string{integration.ExtContainer, integration.Ext3Container, "pg0", "pg1", "pg2", "pg3", "pg4", "pg5", "node0", "node1", "node2", "node3", "node4", "node5"}

func TestLocalDevSetup(t *testing.T) {
	if !*dev {
		t.Skip("skipping local dev setup")
	}

	// running forever for local development

	ctx := context.Background()

	opts := []integration.HelperOpt{
		integration.WithBlockInterval(time.Second),
		integration.WithValidators(4),
		integration.WithNonValidators(0),
		integration.WithExposedRPCPorts(),
	}

	helper := integration.NewIntHelper(t, opts...)
	helper.Setup(ctx, allServices)

	helper.WaitForSignals(t)
}

func TestKwildDatabaseIntegration(t *testing.T) {
	if *parallelMode {
		t.Parallel()
	}

	ctx := context.Background()

	opts := []integration.HelperOpt{
		integration.WithBlockInterval(time.Second),
		integration.WithValidators(4),
		integration.WithNonValidators(0),
	}

	testDrivers := strings.Split(*drivers, ",")
	for _, driverType := range testDrivers {
		t.Run(driverType+"_driver", func(t *testing.T) {
			helper := integration.NewIntHelper(t, opts...)
			helper.Setup(ctx, basicServices)

			node0Driver := helper.GetUserDriver(ctx, "node0", driverType, nil)
			node1Driver := helper.GetUserDriver(ctx, "node1", driverType, nil)
			node2Driver := helper.GetUserDriver(ctx, "node2", driverType, nil)

			// Create a new database and verify that the database exists on other nodes
			specifications.DatabaseDeploySpecification(ctx, t, node0Driver)
			// TODO: wait for node 1 and 2 to hit whatever height 0 is at
			time.Sleep(2 * time.Second)
			specifications.DatabaseVerifySpecification(ctx, t, node1Driver, true)
			specifications.DatabaseVerifySpecification(ctx, t, node2Driver, true)

			specifications.ExecuteDBInsertSpecification(ctx, t, node0Driver)

			// restart node1 and ensure that the app state is synced
			helper.RestartNode(ctx, "node1", 15*time.Second)
			node1Driver = helper.GetUserDriver(ctx, "node1", driverType, nil)

			specifications.ExecuteDBUpdateSpecification(ctx, t, node1Driver)
			specifications.ExecuteDBDeleteSpecification(ctx, t, node2Driver)

			// specifications.ExecutePermissionedActionSpecification(ctx, t, invalidUserDriver)

			specifications.DatabaseDropSpecification(ctx, t, node1Driver)

		})
	}
}

func TestKwildValidatorRemoval(t *testing.T) {
	if *parallelMode {
		t.Parallel()
	}

	ctx := context.Background()

	// In this test, we will have a set of 4 validators, where 3 of the
	// validators are required to remove one.
	const numVals, numNonVals = 4, 0
	const blockInterval = time.Second

	opts := []integration.HelperOpt{
		integration.WithValidators(numVals),
		integration.WithNonValidators(numNonVals),
		integration.WithBlockInterval(blockInterval),
	}

	testDrivers := strings.Split(*drivers, ",")
	for _, driverType := range testDrivers {
		if driverType != "cli" {
			continue // admin service is cli->jsonrpc only
		}

		t.Run(driverType+"_driver", func(t *testing.T) {
			helper := integration.NewIntHelper(t, opts...)
			helper.Setup(ctx, allServices)

			// Wait for the network to produce atleast 1 block for the genesis validators to get committed and synced.
			time.Sleep(2 * time.Second)

			node0Driver := helper.GetOperatorDriver(ctx, "node0", driverType)
			node1Driver := helper.GetOperatorDriver(ctx, "node1", driverType)
			node2Driver := helper.GetOperatorDriver(ctx, "node2", driverType)
			targetPubKey := helper.NodePrivateKey("node3").PubKey().Bytes()

			/* Remove node 3 (4 validators, nodes 0, 1, and 2 remove node 3)
			- node 0 votes to remove
			- node 3 is still a validator
			- node 1 votes to remove
			- node 3 is still a validator
			- node 2 votes to remove
			- node 3 is no longer a validator
			*/
			specifications.ValidatorNodeRemoveSpecificationV4R1(ctx, t, node0Driver, node1Driver, node2Driver, targetPubKey) // joiner is a validator at node

			//
			// Node 3 is no longer a validator, removal should fail
			specifications.RemoveNonValidatorSpecification(ctx, t, node2Driver, targetPubKey)
		})
	}
}

func TestKwildValidatorUpdatesIntegration(t *testing.T) {
	if *parallelMode {
		t.Parallel()
	}

	ctx := context.Background()

	const expiryBlocks = 20
	const blockInterval = time.Second
	const numVals, numNonVals = 3, 1
	opts := []integration.HelperOpt{
		integration.WithValidators(numVals),
		integration.WithNonValidators(numNonVals),
		integration.WithJoinExpiry(expiryBlocks),
		integration.WithBlockInterval(blockInterval),
		integration.WithGas(), // must give the joining node some gas too
	}

	const expiryWait = 3 * expiryBlocks * blockInterval / 2

	testDrivers := strings.Split(*drivers, ",")
	for _, driverType := range testDrivers {
		if driverType != "cli" {
			continue // admin service is cli->jsonrpc only still
		}

		t.Run(driverType+"_driver", func(t *testing.T) {
			helper := integration.NewIntHelper(t, opts...)
			helper.Setup(ctx, allServices)

			node0Driver := helper.GetOperatorDriver(ctx, "node0", driverType)
			node0PubKey := helper.NodePrivateKey("node0").PubKey().Bytes()
			node1Driver := helper.GetOperatorDriver(ctx, "node1", driverType)
			joinerDriver := helper.GetOperatorDriver(ctx, "node3", driverType)
			joinerPrivKey := helper.NodePrivateKey("node3")
			joinerPubKey := joinerPrivKey.PubKey().Bytes()

			// Wait for the network to produce atleast 1 block for the genesis validators to get committed and synced.
			time.Sleep(2 * time.Second)

			// Start the network with 3 validators & 1 Non-validator
			specifications.CurrentValidatorsSpecification(ctx, t, node0Driver, 3)

			// Reject Joins from existing validators
			specifications.JoinExistingValidatorSpecification(ctx, t, node0Driver, node0PubKey)

			// Reject leaves from non-validators
			specifications.NonValidatorLeaveSpecification(ctx, t, joinerDriver, joinerPubKey)

			/*
				Join Expiry:
				- Node3 requests to join
				- No approval from other nodes
				- Join request should expire after 15 blocks
			*/
			specifications.ValidatorJoinExpirySpecification(ctx, t, joinerDriver, joinerPubKey, expiryWait)

			/*
				Join Process:
				- Node3 requests to join
				- Requires at least 2 nodes to approvee
				- Consensus reached, Node3 is a Validator
			*/
			specifications.ValidatorNodeJoinSpecification(ctx, t, joinerDriver, joinerPubKey, 3)
			time.Sleep(2 * time.Second)

			// Node3 cant self approve
			specifications.ValidatorNodeSelfApproveSpecification(ctx, t, joinerDriver, joinerPubKey)

			// Node 0,1 approves
			specifications.ValidatorNodeApproveSpecification(ctx, t, node0Driver, joinerPubKey, 3, 3, false)
			specifications.ValidatorNodeApproveSpecification(ctx, t, node1Driver, joinerPubKey, 3, 4, true)
			specifications.CurrentValidatorsSpecification(ctx, t, node0Driver, 4)

			/*
				Leave Process:
				- node3 issues a leave request -> removes it from the validator list
				- Validator set count should be reduced by 1
			*/
			specifications.ValidatorNodeLeaveSpecification(ctx, t, joinerDriver)

			/*
				Rejoin: (same as join process)
			*/
			specifications.ValidatorNodeJoinSpecification(ctx, t, joinerDriver, joinerPubKey, 3)
			time.Sleep(2 * time.Second)
			// Node 0, 1 approves
			specifications.ValidatorNodeApproveSpecification(ctx, t, node0Driver, joinerPubKey, 3, 3, false)
			specifications.ValidatorNodeApproveSpecification(ctx, t, node1Driver, joinerPubKey, 3, 4, true)
			specifications.CurrentValidatorsSpecification(ctx, t, node0Driver, 4)
		})
	}
}

func TestKwildNetworkSyncIntegration(t *testing.T) {
	if *parallelMode {
		t.Parallel()
	}

	ctx := context.Background()

	opts := []integration.HelperOpt{
		integration.WithValidators(4),
		integration.WithBlockInterval(time.Second),
	}

	testDrivers := strings.Split(*drivers, ",")
	for _, driverType := range testDrivers {
		t.Run(driverType+"_driver", func(t *testing.T) {
			helper := integration.NewIntHelper(t, opts...)
			helper.Setup(ctx, basicServices)

			node0Driver := helper.GetUserDriver(ctx, "node0", driverType, nil)
			node1Driver := helper.GetUserDriver(ctx, "node1", driverType, nil)
			node2Driver := helper.GetUserDriver(ctx, "node2", driverType, nil)

			// Create a new database and verify that the database exists on other nodes
			specifications.DatabaseDeploySpecification(ctx, t, node0Driver)
			time.Sleep(time.Second * 2) // need time to sync
			specifications.DatabaseVerifySpecification(ctx, t, node1Driver, true)
			specifications.DatabaseVerifySpecification(ctx, t, node2Driver, true)

			// Insert 1 User and 1or2 Posts
			specifications.ExecuteDBInsertSpecification(ctx, t, node0Driver)

			// Spin up node 4 and ensure that the database is synced to node4
			/*
				1. Generate config for node 4: place it in the homedir/newNode
				2. Run docker compose up on the new node and get the container
				3. Get the node driver
				4. Verify that the database exists on the new node
			*/
			helper.RunDockerComposeWithServices(ctx, newServices)
			node3Driver := helper.GetUserDriver(ctx, "node3", driverType, nil)

			/*
				1. This checks if the database exists on the new node
				2. Verify if the user and posts are synced to the new node
			*/
			time.Sleep(time.Second * 4) // need time to catch up
			specifications.DatabaseVerifySpecification(ctx, t, node3Driver, true)

			expectPosts := 1
			specifications.ExecuteDBRecordsVerifySpecification(ctx, t, node3Driver, expectPosts)

			// NOTE: integration tests shows that we need somewhere to track the
			// test state, so we can verify across nodes
		})
	}
}

// TestKwildNetworkHardfork checks that the basic height based rule changes with
// hard forks are working.
//
// This test completely breaks the mould used in the other integration tests
// since the "gremlin" test hard fork does not and should not corresponding
// modification to client tooling. We're just type asserting our way into the
// underlying client types to do the checks needed.
func TestKwildNetworkHardfork(t *testing.T) {
	if !*forkTest {
		t.Skip("Fork test requires -fork flag (and special docker image).")
	}
	if *parallelMode {
		t.Parallel()
	}

	ctx := context.Background()

	// At block 8, enable the "gremlin" hardfork rule changes.
	// We should check:
	// - all of network follows new rules (no liveness issues)
	// - a "noop" transaction works after activation
	// - the dc18 account credited in the state mode gets it's 42000 atoms
	// - the consensus params has app version set to 1 (can we even check that?)
	gremlinHeight := uint64(8)
	blockInterval := time.Second

	opts := []integration.HelperOpt{
		integration.WithForkNode(),
		integration.WithValidators(4),
		integration.WithBlockInterval(blockInterval),
		integration.WithForks(map[string]*uint64{
			"gremlin": &gremlinHeight,
		}),
	}

	driverType := "jsonrpc"

	t.Run(driverType+"_driver", func(t *testing.T) {
		helper := integration.NewIntHelper(t, opts...)
		helper.Setup(ctx, basicServices)

		// Wait for the network to produce at least 1 block for the genesis
		// validators to get committed and synced.
		time.Sleep(2 * blockInterval)

		node0Driver := helper.GetUserDriver(ctx, "node0", driverType, nil)
		node1Driver := helper.GetUserDriver(ctx, "node1", driverType, nil)
		node2Driver := helper.GetUserDriver(ctx, "node2", driverType, nil)
		// node3Driver := helper.GetUserDriver(ctx, "node3", driverType)
		// targetPubKey := helper.NodePrivateKey("node3").PubKey().Bytes()

		// Create a new database and verify that the database exists on other nodes
		specifications.DatabaseDeploySpecification(ctx, t, node0Driver)
		time.Sleep(time.Second * 2) // need time to sync
		specifications.DatabaseVerifySpecification(ctx, t, node1Driver, true)
		specifications.DatabaseVerifySpecification(ctx, t, node2Driver, true)

		// ample time to reach activation height, when network failure could occur
		time.Sleep(time.Duration(gremlinHeight+6 /* 6? use chain_info in loop instead */) * blockInterval)

		// Now we are going to test that it was activated:
		// - make a "noop" transaction
		// - check the dc18 account balance

		cl := node0Driver.(*driver.KwildClientDriver).Client().(*client.Client)
		n0AcctStatus, err := cl.GetAccount(ctx, cl.Signer.Identity(), types.AccountStatusLatest)
		if err != nil {
			t.Fatal(err)
		}
		noopTx, err := transactions.CreateTransaction(&noopPayload{}, helper.ChainID(), uint64(n0AcctStatus.Nonce)+1)
		if err != nil {
			t.Fatal(err)
		}
		noopTx.Body.Fee = big.NewInt(42000)
		if err = noopTx.Sign(cl.Signer); err != nil {
			t.Fatal(err)
		}
		noopTxHash, err := cl.SvcClient().Broadcast(ctx, noopTx, 2)
		if err != nil {
			t.Fatal(err)
		}
		specifications.ExpectTxSuccess(t, node0Driver, ctx, noopTxHash)

		dc18, _ := hex.DecodeString("dc18f4993e93b50486e3e54e27d91d57cee1da07")
		dc18Balance, err := node0Driver.AccountBalance(ctx, dc18)
		if err != nil {
			t.Fatal(err)
		}
		if dc18Balance.Cmp(big.NewInt(42000)) != 0 {
			t.Errorf("expected dc18 acct balance %v, got %v", 42000, dc18Balance)
		}

		// Insert 1 User and 1or2 Posts
		specifications.ExecuteDBInsertSpecification(ctx, t, node0Driver)

		// Start 4th node and ensure that the database is caught up, cleanly
		// applying the hardfork changes in catchup.
		helper.RunDockerComposeWithServices(ctx, newServices)
		node3Driver := helper.GetUserDriver(ctx, "node3", driverType, nil)

		// Checks if the database exists on the new node and that the user and
		// posts are synced.
		time.Sleep(time.Second * 8) // need time to catch up
		specifications.DatabaseVerifySpecification(ctx, t, node3Driver, true)

		const expectPosts = 1
		specifications.ExecuteDBRecordsVerifySpecification(ctx, t, node3Driver, expectPosts)
	})
}

type noopPayload struct{}

func (a *noopPayload) MarshalBinary() ([]byte, error) {
	return []byte{0x42}, nil
}

func (a *noopPayload) UnmarshalBinary(b []byte) error { // unused
	return nil
}

func (a *noopPayload) Type() transactions.PayloadType {
	return "noop"
}

func TestKwildEthDepositOracleIntegration(t *testing.T) {
	if *parallelMode {
		t.Parallel()
	}

	type testcase struct {
		name          string
		numValidators int
		serviceNames  []string
	}

	testcases := []testcase{
		{"single-node", 1, singleNodeServices},
		{"multi-node", 4, allServices},
	}

	testDrivers := strings.Split(*drivers, ",")
	for _, driverType := range testDrivers {
		for _, tc := range testcases {
			t.Run(tc.name+"_"+driverType+"_driver", func(t *testing.T) {
				ctx := context.Background()
				opts := []integration.HelperOpt{
					integration.WithBlockInterval(time.Second),
					integration.WithValidators(tc.numValidators),
					integration.WithNonValidators(0),
					integration.WithETHDevNet(),
					integration.WithGas(),
					integration.WithEthDepositOracle(true),
				}
				helper := integration.NewIntHelper(t, opts...)
				helper.Setup(ctx, tc.serviceNames)

				// get deployer
				ctxMiner, cancel := context.WithCancel(ctx)
				defer cancel()
				deployer := helper.EthDeployer(false)
				err := deployer.KeepMining(ctxMiner)
				require.NoError(t, err)

				// Get the user driver
				userDriver := helper.GetUserDriver(ctx, "node0", driverType, deployer)

				// Deposit the amount to the escrow
				amount := big.NewInt(10)
				specifications.DepositSuccessSpecification(ctx, t, userDriver, amount)

				// Deploy DB without enough funds
				specifications.DeployDbInsufficientFundsSpecification(ctx, t, userDriver)

				// Deploy DB with enough funds
				specifications.DeployDbSuccessSpecification(ctx, t, userDriver)
			})
		}
	}
}

func TestKwildEthDepositOracleExpiryIntegration(t *testing.T) {
	if *parallelMode {
		t.Parallel()
	}

	t.Skip("Skipping test as currently there is no way to update the resolution expiry")

	opts := []integration.HelperOpt{
		integration.WithBlockInterval(time.Second),
		integration.WithValidators(5),
		integration.WithNonValidators(0),
		integration.WithGas(),
		integration.WithETHDevNet(),
		integration.WithEthDepositOracle(true),
		integration.WithNumByzantineExpiryNodes(1), // 1 node listens on a different escrow contract and submits votes for events on the byz contract which never gets approved
		integration.WithVoteExpiry(4),
	}

	testDrivers := strings.Split(*drivers, ",")
	for _, driverType := range testDrivers {
		t.Run(driverType+"_driver", func(t *testing.T) {
			ctx := context.Background()

			helper := integration.NewIntHelper(t, opts...)
			helper.Setup(ctx, append(allServices, "pg4", "node4"))

			ctxMiner, cancel := context.WithCancel(ctx)
			defer cancel()
			byzDeployer := helper.EthDeployer(true)
			err := byzDeployer.KeepMining(ctxMiner)
			require.NoError(t, err)

			// Get the user driver
			node0Driver := helper.GetUserDriver(ctx, "node0", driverType, byzDeployer)

			// Nodes: 5
			// Threshold approvals: 5 * 2/3 = 4
			// Expiry refund threshold: 5 * 1/3 = 2
			// Now that we have only 1 byz node, tx fees are never refunded for votes submitted.
			specifications.DepositResolutionExpirySpecification(ctx, t, node0Driver, helper.NodeKeys())

		})
	}
}

func TestKwildEthDepositOracleExpiryRefundIntegration(t *testing.T) {
	if *parallelMode {
		t.Parallel()
	}

	t.Skip("Skipping test as currently there is no way to update the resolution expiry")

	opts := []integration.HelperOpt{
		integration.WithBlockInterval(time.Second),
		integration.WithValidators(5),
		integration.WithNonValidators(0),
		integration.WithGas(),
		integration.WithETHDevNet(),
		integration.WithEthDepositOracle(true),
		integration.WithNumByzantineExpiryNodes(2), // 2 nodes listen on different escrow contracts and submits votes for events on the byz contract which never gets approved.
		integration.WithVoteExpiry(4),
	}

	testDrivers := strings.Split(*drivers, ",")
	for _, driverType := range testDrivers {
		t.Run(driverType+"_driver", func(t *testing.T) {
			ctx := context.Background()

			helper := integration.NewIntHelper(t, opts...)
			helper.Setup(ctx, append(allServices, "pg4", "node4"))

			ctxMiner, cancel := context.WithCancel(ctx)
			defer cancel()
			byzDeployer := helper.EthDeployer(true)
			err := byzDeployer.KeepMining(ctxMiner)
			require.NoError(t, err)

			// Get the user driver
			node0Driver := helper.GetUserDriver(ctx, "node0", driverType, byzDeployer)

			// Nodes: 5
			// Threshold approvals: 5 * 2/3 = 4
			// Expiry refund threshold: 5 * 1/3 = 2
			// As minimun threshold is met for the expiry refund, the tx fees for the byz nodes is refunded.
			specifications.DepositResolutionExpiryRefundSpecification(ctx, t, node0Driver, helper.NodeKeys())

		})
	}
}

func TestKwildEthDepositOracleValidatorUpdates(t *testing.T) {
	if *parallelMode {
		t.Parallel()
	}

	opts := []integration.HelperOpt{
		integration.WithBlockInterval(time.Second),
		integration.WithValidators(6),
		integration.WithNonValidators(0),
		integration.WithGas(),
		integration.WithETHDevNet(),
		integration.WithEthDepositOracle(true),
		integration.WithNumByzantineExpiryNodes(2), // 2 nodes listen on different escrow contracts and submits votes for events on the byz contract which never gets approved.
	}

	testDrivers := strings.Split(*drivers, ",")
	for _, driverType := range testDrivers {
		if driverType != "cli" {
			continue // admin service is cli->jsonrpc only still
		}

		t.Run(driverType+"_driver", func(t *testing.T) {
			ctx := context.Background()

			helper := integration.NewIntHelper(t, opts...)
			helper.Setup(ctx, byzAllServices)

			// get deployer
			ctx2, cancel := context.WithCancel(ctx)
			defer cancel()
			deployer := helper.EthDeployer(false)
			err := deployer.KeepMining(ctx2)
			require.NoError(t, err)

			// Get node drivers
			nodeDrivers := make(map[string]specifications.ValidatorOpsDsl, 6)
			for i := 0; i <= 5; i++ {
				node := fmt.Sprintf("node%d", i)
				nodeDrivers[node] = helper.GetOperatorDriver(ctx, node, driverType)
			}

			user0Driver := helper.GetUserDriver(ctx, "node1", driverType, deployer)

			specifications.EthDepositValidatorUpdatesSpecification(ctx, t, nodeDrivers, user0Driver, helper.NodeKeys())

		})
	}
}

func TestKwildEthDepositFundTransfer(t *testing.T) {
	if *parallelMode {
		t.Parallel()
	}

	opts := []integration.HelperOpt{
		integration.WithBlockInterval(time.Second),
		integration.WithValidators(4),
		integration.WithNonValidators(0),
		integration.WithGas(),
		integration.WithETHDevNet(),
		integration.WithEthDepositOracle(true),
	}

	testDrivers := strings.Split(*drivers, ",")
	for _, driverType := range testDrivers {
		t.Run(driverType+"_driver", func(t *testing.T) {
			ctx := context.Background()

			helper := integration.NewIntHelper(t, opts...)
			helper.Setup(ctx, allServices)
			// defer helper.Teardown()

			// This tests out the ways in which the validator accounts can be funded
			// One way is during network bootstrapping using allocs in the genesis file
			// Other, is through transfer from one kwil account to another

			// Get deposits credited into user account from escrow (or) using alloc in the genesis file
			// Transfer from user account to validator account
			// validate that the validator account has the funds

			ctx, cancel := context.WithCancel(ctx)
			defer cancel()
			ethdeployer := helper.EthDeployer(false)
			senderDriver := helper.GetUserDriver(ctx, "node0", driverType, ethdeployer)

			// node0 key
			valIdentity := helper.NodePrivateKey("node0").PubKey().Bytes()

			specifications.FundValidatorSpecification(ctx, t, senderDriver, valIdentity)
		})
	}
}

func TestSpamListener(t *testing.T) {
	if !*spamTest {
		t.Skip("Spam test requires -spam flag.")
	}
	if *parallelMode {
		t.Parallel()
	}

	opts := []integration.HelperOpt{
		integration.WithValidators(4),
		integration.WithNonValidators(0),
		integration.WithGas(),
		integration.WithSpamOracle(),
		integration.WithSnapshots(),
		integration.WithRecurringHeight(10),
	}

	ctx := context.Background()

	helper := integration.NewIntHelper(t, opts...)
	helper.Setup(ctx, allServices)

	// Wait for the network to produce atleast 1 block for the genesis validators to get committed and synced.
	time.Sleep(2 * time.Second)
	node0Driver := helper.GetUserDriver(ctx, "node0", "jsonrpc", nil)

	// Verify that the spam listener is running and does not overwhelm the network
	// Also keep issuing some transactions to ensure that the network is up and not saturated
	for i := 0; i < 10; i++ {
		specifications.DatabaseDeploySpecification(ctx, t, node0Driver)

		specifications.DatabaseDropSpecification(ctx, t, node0Driver)
		time.Sleep(4 * time.Second)
	}
	fmt.Println("Spam listener test completed")
}

func TestKwildPrivateNetworks(t *testing.T) {
	if *parallelMode {
		t.Parallel()
	}

	blockInterval := time.Second
	opts := []integration.HelperOpt{
		integration.WithValidators(1),
		integration.WithNonValidators(2),
		integration.WithBlockInterval(blockInterval),
		integration.WithJoinExpiry(15),
		// integration.WithAdminRPC("0.0.0.0:8485"),
		integration.PopulatePersistentPeers(true),
		integration.WithPrivateMode(), // Enables private mode
	}

	ctx := context.Background()

	driverType := "cli"
	t.Run("cli_driver", func(t *testing.T) {
		helper := integration.NewIntHelper(t, opts...)
		helper.Setup(ctx, basicServices) // 3 nodes, all connected to each other

		nodeIDs := helper.NodeIDs()
		node0 := nodeIDs["node0"]
		node1 := nodeIDs["node1"]
		node2 := nodeIDs["node2"]

		userNode0Driver := helper.GetUserDriver(ctx, "node0", driverType, nil)
		userNode1Driver := helper.GetUserDriver(ctx, "node1", driverType, nil)
		userNode2Driver := helper.GetUserDriver(ctx, "node2", driverType, nil)

		node0Driver := helper.GetOperatorDriver(ctx, "node0", driverType)
		node1Driver := helper.GetOperatorDriver(ctx, "node1", driverType)
		node2Driver := helper.GetOperatorDriver(ctx, "node2", driverType)

		// Isolate all the nodes from each other
		// n0 n1 n2
		specifications.RemovePeerSpecification(ctx, t, node0Driver, node1)
		specifications.RemovePeerSpecification(ctx, t, node1Driver, node0)
		specifications.RemovePeerSpecification(ctx, t, node0Driver, node2)
		specifications.RemovePeerSpecification(ctx, t, node2Driver, node0)
		specifications.RemovePeerSpecification(ctx, t, node1Driver, node2)
		specifications.RemovePeerSpecification(ctx, t, node2Driver, node1)

		// List whitelisted peers
		specifications.ListPeersSpecification(ctx, t, node0Driver, []string{})
		specifications.ListPeersSpecification(ctx, t, node1Driver, []string{})
		specifications.ListPeersSpecification(ctx, t, node2Driver, []string{})

		// Deploy a database on node0 and verify that the database does not exist on node1 and node2
		specifications.DatabaseDeploySpecification(ctx, t, userNode0Driver)
		time.Sleep(2 * time.Second)
		specifications.DatabaseVerifySpecification(ctx, t, userNode1Driver, false)
		specifications.DatabaseVerifySpecification(ctx, t, userNode2Driver, false)

		// Allow node0 to accept connections from node1
		// n0: [n0, n1]
		// n1: [n1, n0]
		specifications.AddPeerSpecification(ctx, t, node0Driver, node1)
		specifications.AddPeerSpecification(ctx, t, node1Driver, node0)

		specifications.ListPeersSpecification(ctx, t, node0Driver, []string{node1})
		specifications.ListPeersSpecification(ctx, t, node1Driver, []string{node0})
		specifications.ListPeersSpecification(ctx, t, node2Driver, []string{})

		// Verify that the database exists on node1 but not on node2
		time.Sleep(30 * time.Second) // TODO: node eventually connects and syncs // figure out the time interval
		specifications.DatabaseVerifySpecification(ctx, t, userNode1Driver, true)
		specifications.DatabaseVerifySpecification(ctx, t, userNode2Driver, false)

		// Allow node0 to accept connections from node2
		// n0: [n0, n1, n2]
		// n1: [n1, n0]
		// n2: [n2, n0]
		specifications.AddPeerSpecification(ctx, t, node0Driver, node2)
		specifications.AddPeerSpecification(ctx, t, node2Driver, node0)

		specifications.ListPeersSpecification(ctx, t, node0Driver, []string{node1, node2})
		specifications.ListPeersSpecification(ctx, t, node1Driver, []string{node0})
		specifications.ListPeersSpecification(ctx, t, node2Driver, []string{node0})

		// ensure that the database exists on all nodes
		time.Sleep(30 * time.Second)
		specifications.DatabaseVerifySpecification(ctx, t, userNode2Driver, true)

		// ensure that node2 is not connected to node1
		specifications.PeerConnectivitySpecification(ctx, t, node2Driver, node1, false)
		specifications.PeerConnectivitySpecification(ctx, t, node1Driver, node2, false)

		// allow node1 to accept connections from node2, but not the other way around
		// n0: [n0, n1, n2]
		// n1: [n1, n0, n2]
		// n2: [n2, n0]
		specifications.AddPeerSpecification(ctx, t, node1Driver, node2)
		specifications.ListPeersSpecification(ctx, t, node1Driver, []string{node0, node2})
		specifications.ListPeersSpecification(ctx, t, node2Driver, []string{node0})

		// so both the nodes are still not connected
		specifications.PeerConnectivitySpecification(ctx, t, node1Driver, nodeIDs["node2"], false)
		specifications.PeerConnectivitySpecification(ctx, t, node2Driver, nodeIDs["node1"], false)

		// now make node1 a validator
		node1PrivKey := helper.NodePrivateKey("node1")
		node1PubKey := node1PrivKey.PubKey().Bytes()
		specifications.ValidatorNodeJoinSpecification(ctx, t, node1Driver, node1PubKey, 1)
		specifications.ValidatorNodeApproveSpecification(ctx, t, node0Driver, node1PubKey, 1, 2, true)

		time.Sleep(30 * time.Second)
		specifications.CurrentValidatorsSpecification(ctx, t, node0Driver, 2)
		specifications.CurrentValidatorsSpecification(ctx, t, node2Driver, 2)

		// node2 automatocally adds node1 as a peer as it is a validator
		// time.Sleep(30 * time.Second)
		// n0, n1, n3: [n0, n1, n2]
		specifications.ListPeersSpecification(ctx, t, node2Driver, []string{node1, node0})
		specifications.ListPeersSpecification(ctx, t, node1Driver, []string{node2, node0})

		// node1 removes node2 from its peer list
		specifications.RemovePeerSpecification(ctx, t, node1Driver, node2)

		// Node2 sends a validator join request
		// node1 approves the request -> node1 trusts node2 -> adds to its peer list
		// validator request expires -> node1 removes node2 from its peer list
		const expiryWait = 25 * time.Second
		node2PrivKey := helper.NodePrivateKey("node2")
		node2PubKey := node2PrivKey.PubKey().Bytes()
		specifications.ValidatorNodeJoinSpecification(ctx, t, node2Driver, node2PubKey, 2)
		specifications.ListPeersSpecification(ctx, t, node1Driver, []string{node0})
		specifications.ValidatorNodeApproveSpecification(ctx, t, node1Driver, node2PubKey, 2, 2, false)
		specifications.ListPeersSpecification(ctx, t, node1Driver, []string{node0, node2})

		time.Sleep(expiryWait)
		// as join request expires, node1 removes node2 from its peer list
		specifications.ListPeersSpecification(ctx, t, node1Driver, []string{node0})
	})
}

func TestStatesync(t *testing.T) {
	if *parallelMode {
		t.Parallel()
	}

	ctx := context.Background()

	opts := []integration.HelperOpt{
		integration.WithValidators(4),
		integration.WithBlockInterval(time.Second),
		integration.WithSnapshots(),
		integration.WithRecurringHeight(5),
	}

	/*
		Node 1, 2, 3 has snapshots enabled
		Node 4 tries to sync with the network, with statesync enabled.
		Node4 should be able to sync with the network and catch up with the latest state (maybe check for the database existence)
	*/
	testDrivers := strings.Split(*drivers, ",")
	for _, driverType := range testDrivers {
		t.Run(driverType+"_driver", func(t *testing.T) {
			helper := integration.NewIntHelper(t, opts...)
			helper.Setup(ctx, basicServices)

			node0Driver := helper.GetUserDriver(ctx, "node0", driverType, nil)
			node1Driver := helper.GetUserDriver(ctx, "node1", driverType, nil)
			node2Driver := helper.GetUserDriver(ctx, "node2", driverType, nil)

			// Create a new database and verify that the database exists on other nodes
			specifications.DatabaseDeploySpecification(ctx, t, node0Driver)
			time.Sleep(time.Second * 2) // need time to sync
			specifications.DatabaseVerifySpecification(ctx, t, node1Driver, true)
			specifications.DatabaseVerifySpecification(ctx, t, node2Driver, true)

			// Insert 1 User and 1or2 Posts

			// Wait for atleast 5 secs for the snapshots to be created
			time.Sleep(7 * time.Second)

			// Spin up node 4 and ensure that the database is synced to node4
			/*
				1. Generate config for node 4: place it in the homedir/newNode
				2. Run docker compose up on the new node and get the container
				3. Get the node driver
				4. Verify that the database exists on the new node
			*/

			rootDir, err := helper.TestnetDir()
			require.NoError(t, err)

			helper.EnableStatesync(ctx, rootDir, "node3", []string{"node0", "node1", "node2"})
			helper.RunDockerComposeWithServices(ctx, newServices)

			// Let the node catch up with the network
			time.Sleep(10 * time.Second)

			node3Driver := helper.GetUserDriver(ctx, "node3", driverType, nil)

			/*
				1. This checks if the database exists on the new node
				2. Verify if the user and posts are synced to the new node
			*/
			time.Sleep(time.Second * 4) // need time to catch up
			specifications.DatabaseVerifySpecification(ctx, t, node3Driver, true)

			specifications.ExecuteDBInsertSpecification(ctx, t, node3Driver)

			expectPosts := 1
			specifications.ExecuteDBRecordsVerifySpecification(ctx, t, node3Driver, expectPosts)
			specifications.ExecuteDBRecordsVerifySpecification(ctx, t, node0Driver, expectPosts)
		})
	}
}

func TestLongRunningNetworkMigrations(t *testing.T) {

	if *parallelMode {
		t.Parallel()
	}

	ctx := context.Background()

	opts := []integration.HelperOpt{
		integration.WithValidators(4),
		integration.WithBlockInterval(time.Second),
		integration.WithAdminRPC("0.0.0.0:8485"),
		integration.WithSnapshots(),
		integration.WithRecurringHeight(5),
	}

	t.Run("jsonrpc_driver", func(t *testing.T) {
		helper := integration.NewIntHelper(t, opts...)
		helper.Setup(ctx, allServices)

		// Prepare the network for migration
		var addresses []string
		for i := 0; i < 4; i++ {
			_, addr, err := helper.JSONRPCListenAddress(ctx, fmt.Sprintf("node%d", i))
			require.NoError(t, err)
			addresses = append(addresses, addr)
		}

		user0Driver := helper.GetUserDriver(ctx, "node0", "jsonrpc", nil)
		user1Driver := helper.GetUserDriver(ctx, "node1", "jsonrpc", nil)
		user2Driver := helper.GetUserDriver(ctx, "node2", "jsonrpc", nil)

		node0Driver := helper.GetOperatorDriver(ctx, "node0", "jsonrpc")
		node1Driver := helper.GetOperatorDriver(ctx, "node1", "jsonrpc")
		node2Driver := helper.GetOperatorDriver(ctx, "node2", "jsonrpc")

		// Create a new database and verify that the database exists on other nodes
		specifications.DatabaseDeploySpecification(ctx, t, user0Driver)
		time.Sleep(time.Second * 5) // need time to syncx
		specifications.DatabaseVerifySpecification(ctx, t, user1Driver, true)
		specifications.DatabaseVerifySpecification(ctx, t, user2Driver, true)

		newDir := helper.MigrationSetup(ctx)

		// Trigger a network migration request
		specifications.SubmitMigrationProposal(ctx, t, node0Driver, integration.MigrationChainID)

		// node1 approves the migration and verifies that the migration is still pending
		specifications.ApproveMigration(ctx, t, node1Driver, true)

		// node1 approves again and verifies that the migration is still pending as duplicate approvals are not allowed
		specifications.ApproveMigration(ctx, t, node1Driver, true)

		// node2 approves the migration and verifies that the migration is approved
		specifications.ApproveMigration(ctx, t, node2Driver, false)

		// Wait for the migration to start
		time.Sleep(10 * time.Second)

		// Retrieve the genesis state and update the config for the new network
		specifications.InstallGenesisState(ctx, t, node0Driver, newDir, 4, addresses)

		// Bring up the new network using the genesis state and the genesis file installed from the old network
		helper.RunDockerComposeWithServices(ctx, migrationServices)

		// Insert 1 User and 1or2 Posts
		specifications.ExecuteDBInsertSpecification(ctx, t, user0Driver)

		/*
			1. This checks if the database exists on the new node
			2. Verify if the user and posts are synced to the new node
		*/
		time.Sleep(time.Second * 30) // This is for the changesets to be synced and voted and resolved in the new network

		newNodeDriver := helper.GetMigrationUserDriver(ctx, "new-node0", "jsonrpc", nil)
		specifications.DatabaseVerifySpecification(ctx, t, newNodeDriver, true)

		// The user and posts should be synced to the new node through the changeset_listener extension
		// Below specification checks if the user and posts are synced to the new node correctly
		expectPosts := 1
		specifications.ExecuteDBRecordsVerifySpecificationEventually(ctx, t, newNodeDriver, expectPosts)

		// Bring up node3-1 using statesync
		helper.EnableStatesync(ctx, newDir, "new-node3", []string{"new-node0", "new-node1", "new-node2"})
		helper.RunDockerComposeWithServices(ctx, migrationServices2)

		newNode3Driver := helper.GetMigrationUserDriver(ctx, "new-node3", "jsonrpc", nil)

		specifications.DatabaseVerifySpecificationEventually(ctx, t, newNode3Driver)

		specifications.ExecuteDBRecordsVerifySpecificationEventually(ctx, t, newNode3Driver, expectPosts)

		t.Logf("Long Running Network Migration Test: Completed Successfully")
	})
}
