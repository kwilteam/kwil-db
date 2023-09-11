package integration_test

import (
	"context"
	"flag"
	"testing"

	"github.com/kwilteam/kwil-db/test/integration"
	"github.com/kwilteam/kwil-db/test/specifications"
)

var dev = flag.Bool("dev", false, "run for development purpose (no tests)")

var allServices = []string{integration.ExtContainer, "node0", "node1", "node2", "node3"}
var numServices = len(allServices)

func TestKwildDatabaseIntegration(t *testing.T) {
	ctx := context.Background()

	opts := []integration.HelperOpt{
		integration.WithValidators(4),
		integration.WithNonValidators(0),
	}

	helper := integration.NewIntHelper(t, opts...)
	helper.LoadConfig()
	helper.Setup(ctx, allServices)
	defer helper.Teardown()

	// running forever for local development
	if *dev {
		helper.WaitForSignals(t)
		return
	}

	drivers := helper.GetDrivers(ctx)
	node0Driver := drivers[0]
	node1Driver := drivers[1]
	node2Driver := drivers[2]

	// Create a new database and verify that the database exists on other nodes
	specifications.DatabaseDeploySpecification(ctx, t, node0Driver)
	specifications.DatabaseVerifySpecification(ctx, t, node1Driver, true)
	specifications.DatabaseVerifySpecification(ctx, t, node2Driver, true)

	specifications.ExecuteDBInsertSpecification(ctx, t, node0Driver)
	specifications.ExecuteDBUpdateSpecification(ctx, t, node1Driver)
	specifications.ExecuteDBDeleteSpecification(ctx, t, node2Driver)

	// specifications.ExecutePermissionedActionSpecification(ctx, t, invalidUserDriver)

	specifications.DatabaseDropSpecification(ctx, t, node1Driver)
}

func TestKwildValidatorUpdatesIntegration(t *testing.T) {
	ctx := context.Background()

	opts := []integration.HelperOpt{
		integration.WithValidators(3),
		integration.WithNonValidators(1),
	}

	helper := integration.NewIntHelper(t, opts...)
	helper.LoadConfig()

	helper.Setup(ctx, allServices)
	defer helper.Teardown()

	// running forever for local development
	if *dev {
		helper.WaitForSignals(t)
		return
	}

	node0Driver := helper.GetNodeDriver(ctx, "node0")
	node1Driver := helper.GetNodeDriver(ctx, "node1")
	joinerDriver := helper.GetNodeDriver(ctx, "node3")
	joinerPkey := helper.NodePrivateKey("node3")
	joinerPubKey := joinerPkey.PubKey().Bytes()

	// Start the network with 3 validators & 1 Non-validator
	specifications.NetworkNodeValidatorSetSpecification(ctx, t, node0Driver, 3)

	/*
	 Join Process:
	 - Node3 requests to join
	 - Requires atleast 2 nodes to approve
	 - Consensus reached, Node3 is a Validator
	*/
	specifications.NetworkNodeJoinSpecification(ctx, t, joinerDriver, joinerPubKey, 3)
	// Node 0,1 approves
	specifications.NetworkNodeApproveSpecification(ctx, t, node0Driver, joinerPubKey, 3, 3, false)
	specifications.NetworkNodeApproveSpecification(ctx, t, node1Driver, joinerPubKey, 3, 4, true)
	specifications.NetworkNodeValidatorSetSpecification(ctx, t, node0Driver, 4)

	/*
	 Leave Process:
	 - node3 issues a leave request -> removes it from the validator list
	 - Validatorset count should be reduced by 1
	*/
	specifications.NetworkNodeLeaveSpecification(ctx, t, joinerDriver)

	/*
	 Rejoin: (same as join process)
	*/
	specifications.NetworkNodeJoinSpecification(ctx, t, joinerDriver, joinerPubKey, 3)
	// Node 0, 1 approves
	specifications.NetworkNodeApproveSpecification(ctx, t, node0Driver, joinerPubKey, 3, 3, false)
	specifications.NetworkNodeApproveSpecification(ctx, t, node1Driver, joinerPubKey, 3, 4, true)
	specifications.NetworkNodeValidatorSetSpecification(ctx, t, node0Driver, 4)
}

func TestKwildNetworkSyncIntegration(t *testing.T) {
	ctx := context.Background()

	opts := []integration.HelperOpt{
		integration.WithValidators(4),
	}

	helper := integration.NewIntHelper(t, opts...)
	helper.LoadConfig()
	// Bringup ext1, node 0,1,2 services but not node3
	helper.Setup(ctx, allServices[:numServices-1])
	defer helper.Teardown()

	// running forever for local development
	if *dev {
		helper.WaitForSignals(t)
		return
	}

	drivers := helper.GetDrivers(ctx)
	node0Driver := drivers[0]
	node1Driver := drivers[1]
	node2Driver := drivers[2]

	// Create a new database and verify that the database exists on other nodes
	specifications.DatabaseDeploySpecification(ctx, t, node0Driver)
	specifications.DatabaseVerifySpecification(ctx, t, node1Driver, true)
	specifications.DatabaseVerifySpecification(ctx, t, node2Driver, true)

	// Insert 1 User 4 Posts
	specifications.ExecuteDBInsertSpecification(ctx, t, node0Driver)

	// Spin up node 4 and ensure that the database is synced to node4
	/*
		1. Generate config for node 4: place it in the homedir/newNode
		2. Run docker compose up on the new node and get the container
		3. Get the node driver
		4. Verify that the database exists on the new node
	*/
	helper.RunDockerComposeWithServices(ctx, allServices[numServices-1:])
	node3Driver := helper.GetDriver(ctx, helper.ServiceContainer("node3"))

	/*
	   1. This checks if the database exists on the new node
	   2. Verify if the user and posts are synced to the new node
	*/
	specifications.DatabaseVerifySpecification(ctx, t, node3Driver, true)
	specifications.ExecuteDBRecordsVerifySpecification(ctx, t, node3Driver, 4)
}
