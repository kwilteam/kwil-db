package integration_test

import (
	"context"
	"flag"
	"strings"
	"testing"
	"time"

	"github.com/kwilteam/kwil-db/test/integration"
	"github.com/kwilteam/kwil-db/test/specifications"
)

var dev = flag.Bool("dev", false, "run for development purpose (no tests)")
var drivers = flag.String("drivers", "client,cli", "comma separated list of drivers to run")

var allServices = []string{integration.ExtContainer, "node0", "node1", "node2", "node3"}
var numServices = len(allServices)

func TestKwildDatabaseIntegration(t *testing.T) {
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
			helper.Setup(ctx, allServices)
			defer helper.Teardown()

			// running forever for local development
			if *dev {
				helper.WaitForSignals(t)
				return
			}

			node0Driver := helper.GetUserDriver(ctx, "node0", driverType)
			node1Driver := helper.GetUserDriver(ctx, "node1", driverType)
			node2Driver := helper.GetUserDriver(ctx, "node2", driverType)

			// Create a new database and verify that the database exists on other nodes
			specifications.DatabaseDeploySpecification(ctx, t, node0Driver)
			specifications.DatabaseVerifySpecification(ctx, t, node1Driver, true)
			specifications.DatabaseVerifySpecification(ctx, t, node2Driver, true)

			specifications.ExecuteDBInsertSpecification(ctx, t, node0Driver)
			specifications.ExecuteDBUpdateSpecification(ctx, t, node1Driver)
			specifications.ExecuteDBDeleteSpecification(ctx, t, node2Driver)

			// specifications.ExecutePermissionedActionSpecification(ctx, t, invalidUserDriver)

			specifications.DatabaseDropSpecification(ctx, t, node1Driver)
		})
	}
}

func TestKwildValidatorUpdatesIntegration(t *testing.T) {
	ctx := context.Background()

	const expiryBlocks = 15
	const blockInterval = time.Second
	const numVals, numNonVals = 3, 1
	opts := []integration.HelperOpt{
		integration.WithValidators(numVals),
		integration.WithNonValidators(numNonVals),
		integration.WithJoinExpiry(expiryBlocks),
		integration.WithBlockInterval(blockInterval),
	}

	expiryWait := 2 * expiryBlocks * blockInterval

	testDrivers := strings.Split(*drivers, ",")
	for _, driverType := range testDrivers {
		t.Run(driverType+"_driver", func(t *testing.T) {
			helper := integration.NewIntHelper(t, opts...)
			helper.Setup(ctx, allServices)
			defer helper.Teardown()

			// running forever for local development
			if *dev {
				helper.WaitForSignals(t)
				return
			}

			node0Driver := helper.GetOperatorDriver(ctx, "node0", driverType)
			node1Driver := helper.GetOperatorDriver(ctx, "node1", driverType)
			joinerDriver := helper.GetOperatorDriver(ctx, "node3", driverType)
			joinerPrivKey := helper.NodePrivateKey("node3")
			joinerPubKey := joinerPrivKey.PubKey().Bytes()

			// Start the network with 3 validators & 1 Non-validator
			specifications.CurrentValidatorsSpecification(ctx, t, node0Driver, 3)

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
			 - Requires at least 2 nodes to approve
			 - Consensus reached, Node3 is a Validator
			*/
			specifications.ValidatorNodeJoinSpecification(ctx, t, joinerDriver, joinerPubKey, 3)
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
			// Node 0, 1 approves
			specifications.ValidatorNodeApproveSpecification(ctx, t, node0Driver, joinerPubKey, 3, 3, false)
			specifications.ValidatorNodeApproveSpecification(ctx, t, node1Driver, joinerPubKey, 3, 4, true)
			specifications.CurrentValidatorsSpecification(ctx, t, node0Driver, 4)

			/* Remove node 2 (4 validators, nodes 0, 1, and 3 remove node 2)
			- node 0 votes to remove
			- node 2 is still a validator
			- node 1 votes to remove
			- node 2 is still a validator
			- node 3 votes to remove
			- node 2 is no longer a validator
			*/
			node2PubKey := helper.NodePrivateKey("node2").PubKey().Bytes()
			specifications.ValidatorNodeRemoveSpecificationV4R2(ctx, t, node0Driver, node1Driver, joinerDriver /* node3 */, node2PubKey) // joiner is a validator at node
		})
	}
}

func TestKwildNetworkSyncIntegration(t *testing.T) {
	ctx := context.Background()

	opts := []integration.HelperOpt{
		integration.WithValidators(4),
		integration.WithBlockInterval(time.Second),
	}

	testDrivers := strings.Split(*drivers, ",")
	for _, driverType := range testDrivers {
		t.Run(driverType+"_driver", func(t *testing.T) {
			helper := integration.NewIntHelper(t, opts...)
			// Bringup ext1, node 0,1,2 services but not node3
			helper.Setup(ctx, allServices[:numServices-1])
			defer helper.Teardown()

			// running forever for local development
			if *dev {
				helper.WaitForSignals(t)
				return
			}

			node0Driver := helper.GetUserDriver(ctx, "node0", driverType)
			node1Driver := helper.GetUserDriver(ctx, "node1", driverType)
			node2Driver := helper.GetUserDriver(ctx, "node2", driverType)

			// Create a new database and verify that the database exists on other nodes
			specifications.DatabaseDeploySpecification(ctx, t, node0Driver)
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
			helper.RunDockerComposeWithServices(ctx, allServices[numServices-1:])
			//node3Driver := helper.GetUserDriver(ctx, helper.ServiceContainer("node3"))
			node3Driver := helper.GetUserDriver(ctx, "node3", driverType)

			/*
			   1. This checks if the database exists on the new node
			   2. Verify if the user and posts are synced to the new node
			*/
			specifications.DatabaseVerifySpecification(ctx, t, node3Driver, true)

			expectPosts := 1
			specifications.ExecuteDBRecordsVerifySpecification(ctx, t, node3Driver, expectPosts)

			// NOTE: integration tests shows that we need somewhere to track the
			// test state, so we can verify across nodes
		})
	}
}
