package acceptance_test

import (
	"context"
	"flag"
	"math/big"
	"strings"
	"testing"
	"time"

	"github.com/kwilteam/kwil-db/test/acceptance"
	"github.com/kwilteam/kwil-db/test/specifications"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var dev = flag.Bool("dev", false, "run for development purpose (no tests)")
var remote = flag.Bool("remote", false, "test against remote node")

// NOTE: `-parallel` is a flag that is already used by `go test`
var parallelMode = flag.Bool("parallel-mode", false, "run tests in parallelMode mode")
var drivers = flag.String("drivers", "jsonrpc,cli", "comma separated list of drivers to run")

func TestLocalDevSetup(t *testing.T) {
	if !*dev {
		t.Skip("skipping local dev setup")
	}

	// running forever for local development

	ctx := context.Background()
	helper := acceptance.NewActHelper(t)
	cfg := helper.LoadConfig()
	cfg.DockerComposeFile = "./docker-compose-dev.yml" // use the dev compose file

	helper.Setup(ctx)
	helper.WaitUntilInterrupt()
}

func TestKwildTransferAcceptance(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	if *parallelMode {
		t.Parallel()
	}

	ctx := context.Background()
	testDrivers := strings.Split(*drivers, ",")
	for _, driverType := range testDrivers {
		// NOTE: those tests should not be run concurrently
		t.Run(driverType+"_driver", func(t *testing.T) {
			// setup for each driver
			helper := acceptance.NewActHelper(t)
			cfg := helper.LoadConfig()
			cfg.GasEnabled = true
			if !*remote {
				helper.Setup(ctx)
			}
			// Ensure that the fee for a transfer transaction is as expected.
			var transferPrice = big.NewInt(210_000)
			senderIdentity := helper.GetConfig().CreatorIdent()
			receiverIdentity := helper.GetConfig().VisitorIdent()
			t.Log("creator private key: ", helper.GetConfig().CreatorRawPk)

			// =================

			// Wait for Genesis allocs to get credited at the end of 1st block, before issuing any transactions.
			time.Sleep(2 * time.Second)

			senderDriver := helper.GetDriver(driverType, "creator")
			sender := specifications.TransferAmountDsl(senderDriver)

			receiverDriver := helper.GetDriver(driverType, "visitor")
			receiver := specifications.TransferAmountDsl(receiverDriver)

			bal0Sender, err := sender.AccountBalance(ctx, senderIdentity)
			assert.NoError(t, err)
			bal0Receiver, err := sender.AccountBalance(ctx, receiverIdentity)
			assert.NoError(t, err)

			// An unfunded account can't send (should check balance first)
			amt := big.NewInt(0).Add(transferPrice, transferPrice) // 2 x fee -- enough to ensure they can send back
			_, err = receiver.TransferAmt(ctx, senderIdentity, amt)
			require.Error(t, err, "should have failed to send")

			// When I transfer to an account
			txHash, err := sender.TransferAmt(ctx, receiverIdentity, amt)
			require.NoError(t, err, "failed to send transfer tx")

			// Then I expect success
			specifications.ExpectTxSuccess(t, sender, ctx, txHash)

			gotBal, err := sender.AccountBalance(ctx, senderIdentity)
			assert.NoError(t, err)

			// Sender balance should be reduced by amt+fees
			expectSpent := big.NewInt(0).Add(amt, transferPrice)
			expectBal := big.NewInt(0).Sub(bal0Sender, expectSpent)
			assert.EqualValues(t, expectBal, gotBal)

			// The receiver balance should be increased by the amount sent
			gotBal, err = sender.AccountBalance(ctx, receiverIdentity)
			assert.NoError(t, err)
			expectBal = big.NewInt(0).Add(bal0Receiver, amt)
			assert.EqualValues(t, expectBal, gotBal)

			// Receiver should be able to send back
			amt = big.NewInt(0).Set(transferPrice) // should leave us at exactly zero
			txHash, err = receiver.TransferAmt(ctx, senderIdentity, amt)
			require.NoError(t, err, "failed to send transfer tx")
			specifications.ExpectTxSuccess(t, sender, ctx, txHash)
			gotBal, err = receiver.AccountBalance(ctx, receiverIdentity)
			assert.NoError(t, err)
			expectBal = big.NewInt(0)
			assert.EqualValues(t, expectBal, gotBal)
		})
	}
}

// TestKwildProcedures runs acceptance tests against a single kwild node,
// testing Kuneiforms procedural language.
func TestKwildProcedures(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	if *parallelMode {
		t.Parallel()
	}

	ctx := context.Background()
	testDrivers := strings.Split(*drivers, ",")
	for _, driverType := range testDrivers {
		t.Run(driverType+"_driver", func(t *testing.T) {
			// setup for each driver
			helper := acceptance.NewActHelper(t)
			helper.LoadConfig()
			if !*remote {
				helper.Setup(ctx)
			}
			creatorDriver := helper.GetDriver(driverType, "creator")

			specifications.ExecuteProcedureSpecification(ctx, t, creatorDriver)
		})
	}
}

// TestKwildAcceptance runs acceptance tests again a single kwild node(and
// are not concurrent), using different drivers: clientDriver, cliDriver.
// The tests here are not exhaustive, and are meant to only test happy paths.
func TestKwildAcceptance(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	if *parallelMode {
		t.Parallel()
	}

	ctx := context.Background()
	testDrivers := strings.Split(*drivers, ",")
	for _, driverType := range testDrivers {
		t.Run(driverType+"_driver", func(t *testing.T) {
			// setup for each driver
			helper := acceptance.NewActHelper(t)
			helper.LoadConfig()
			if !*remote {
				helper.Setup(ctx)
			}
			creatorDriver := helper.GetDriver(driverType, "creator")

			// ================
			// When user deployed database
			specifications.DatabaseDeployInvalidSql1Specification(ctx, t, creatorDriver)
			specifications.DatabaseDeployInvalidExtensionSpecification(ctx, t, creatorDriver)
			specifications.DatabaseDeploySpecification(ctx, t, creatorDriver)

			// Then user should be able to execute database
			specifications.ExecuteOwnerActionSpecification(ctx, t, creatorDriver)

			// TODO: This test doesn't looks good, the spec suppose to expect
			// only one parameter, the driver.
			// Read `test/specifications/README.md` for more information.
			db := specifications.SchemaLoader.Load(t, specifications.SchemaTestDB)
			dbid := creatorDriver.DBID(db.Name)
			visitorDriver := helper.GetDriver(driverType, "visitor")
			specifications.ExecuteOwnerActionFailSpecification(ctx, t, visitorDriver, dbid)

			specifications.ExecuteDBInsertSpecification(ctx, t, creatorDriver)
			specifications.ExecuteCallSpecification(ctx, t, creatorDriver, visitorDriver)

			specifications.ExecuteDBUpdateSpecification(ctx, t, creatorDriver)
			specifications.ExecuteDBDeleteSpecification(ctx, t, creatorDriver)

			// test that the loaded extensions works
			specifications.ExecuteExtensionSpecification(ctx, t, creatorDriver)

			specifications.ExecutePrivateActionSpecification(ctx, t, creatorDriver)

			// and user should be able to drop database
			specifications.DatabaseDropSpecification(ctx, t, creatorDriver)

			specifications.ExecuteChainInfoSpecification(ctx, t, creatorDriver, acceptance.TestChainID)
			// there's one node in the network and we're the validator
			// @brennan I am commenting this out temporarily, but it seems to be _mostly_ useless
			// all it does is check that the node is a validator, which is not really a useful test,
			// and couples the rest of acceptance to the validator rpcs, which should probably
			// be a standalone set of tests anyways
			//specifications.CurrentValidatorsSpecification(ctx, t, creatorDriver, 1)

			// The other network/validator specs require multiple nodes in a network

			// TODO: test inputting invalid utf-8 into action that needs string (should fail)
			// this previously crashed the node

			// Test notices
			specifications.ExecuteNoticeSpecification(ctx, t, creatorDriver)
		})
	}
}

// TestTypes checks that type serialization works correctly over RLP, JSON,
// and Postgres.
func TestTypes(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	if *parallelMode {
		t.Parallel()
	}

	ctx := context.Background()
	testDrivers := strings.Split(*drivers, ",")
	for _, driverType := range testDrivers {
		t.Run(driverType+"_driver", func(t *testing.T) {
			// setup for each driver
			helper := acceptance.NewActHelper(t)
			helper.LoadConfig()
			if !*remote {
				helper.Setup(ctx)
			}
			creatorDriver := helper.GetDriver(driverType, "creator")

			// we only test nils if using json rpc driver, becase the cli driver cannot support nil
			// args to actions/procedures
			specifications.ExecuteTypesSpecification(ctx, t, creatorDriver, driverType == "jsonrpc")

			// test contextual vars
			specifications.ExecuteContextualVarsSpecification(ctx, t, creatorDriver)
		})
	}
}

func TestDataPrivateMode(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	if *parallelMode {
		t.Parallel()
	}

	ctx := context.Background()
	testDrivers := strings.Split(*drivers, ",")
	for _, driverType := range testDrivers {
		t.Run(driverType+"_driver", func(t *testing.T) {
			// setup for each driver
			helper := acceptance.NewActHelper(t)
			cfg := helper.LoadConfig()
			cfg.PrivateRPC = true
			if !*remote {
				helper.Setup(ctx)
			}

			noAuthDriver := helper.GetDriver(driverType, "")
			authDriver := helper.GetDriver(driverType, "creator")

			specifications.DatabaseDeploySpecification(ctx, t, authDriver)

			specifications.ExecuteCallPrivateModeSpecification(ctx, t, authDriver, noAuthDriver)
		})
	}
}
