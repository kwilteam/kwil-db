package acceptance_test

import (
	"context"
	"flag"
	"fmt"
	"github.com/ethereum/go-ethereum/crypto"
	"kwil/pkg/chain/types"
	"kwil/test/adapters"
	"kwil/test/specifications"
	"kwil/test/utils/deployer"
	"kwil/x/types/databases"
	"math/big"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

var buildKwilOnce sync.Once

var remote = flag.Bool("remote", false, "run tests against remote environment")

func keepMiningBlocks(ctx context.Context, deployer deployer.Deployer, account string) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			time.Sleep(3 * time.Second)
			// to mine new blocks
			err := deployer.FundAccount(ctx, account, 1)
			if err != nil {
				fmt.Println("funded user info failed", err)
			}
		}
	}
}

func TestGrpcServerDatabaseService(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	var (
		// test user
		userAddr string
		userPK   string
		// test deployer
		deployerPK string
		// database schema file
		dbSchemaPath string
		// remote kwild endpoint
		remoteKwildAddr string
		// remote blockchain endpoint
		providerEndpoint string
		// blochchain block produce interval
		chainSyncWaitTime  time.Duration
		fundingPoolAddress string
		_chainCode         types.ChainCode
		fundAmount         int64
		domination         *big.Int
		remoteDBUrl        string
	)

	localEnv := func() {
		userPK = adapters.UserAccountPK
		deployerPK = adapters.DeployerAccountPK
		dbSchemaPath = "./test-data/database_schema.json"
		remoteKwildAddr = ""
		providerEndpoint = ""
		chainSyncWaitTime = 3 * time.Second
		_chainCode = types.GOERLI
		fundAmount = 100
		domination = big.NewInt(1000000000000000000)
	}

	remoteEnv := func() {
		// depends on the remote environment, change respectively
		userPK = os.Getenv("TEST_USER_PK")
		deployerPK = os.Getenv("TEST_DEPLOYER_PK")
		dbSchemaPath = "./test-data/database_schema.json"
		remoteKwildAddr = os.Getenv("TEST_KWILD_ADDR")
		providerEndpoint = os.Getenv("TEST_PROVIDER")
		chainSyncWaitTime = 15 * time.Second
		_chainCode = types.GOERLI
		fundAmount = 1
		fundingPoolAddress = os.Getenv("TEST_POOL_ADDRESS")
		remoteDBUrl = os.Getenv("TEST_DB_URL")
		domination = big.NewInt(10000)
	}

	if *remote {
		remoteEnv()
	} else {
		localEnv()
	}

	userPrivateKey, err := crypto.HexToECDSA(userPK)
	if err != nil {
		t.Fatal(fmt.Errorf("invalid user private key: %s", err))
	}
	userAddr = crypto.PubkeyToAddress(userPrivateKey.PublicKey).Hex()

	t.Run("should deposit fund", func(t *testing.T) {
		if *remote {
			t.Skip()
		}

		ctx := context.Background()
		// setup
		chainDriver, chainDeployer, _, _ := adapters.GetChainDriverAndDeployer(
			t, ctx, providerEndpoint, deployerPK, _chainCode, userPK, fundingPoolAddress, domination)

		// Given user is funded with escrow token
		err := chainDeployer.FundAccount(ctx, userAddr, fundAmount)
		assert.NoError(t, err, "failed to fund user info")

		// and user has approved funding_pool to spend his token
		specifications.ApproveTokenSpecification(t, ctx, chainDriver)

		// should be able to deposit fund
		specifications.DepositFundSpecification(t, ctx, chainDriver)
	})

	t.Run("should deploy and drop database", func(t *testing.T) {
		ctx := context.Background()
		// setup
		specifications.SetSchemaLoader(
			&specifications.FileDatabaseSchemaLoader{
				FilePath: dbSchemaPath,
				Modifier: func(db *databases.Database[[]byte]) {
					db.Owner = userAddr
				}})
		chainDriver, chainDeployer, userFundConfig, chainEnvs := adapters.GetChainDriverAndDeployer(
			t, ctx, providerEndpoint, deployerPK, _chainCode, userPK, fundingPoolAddress, domination)

		if !*remote {
			// Given user is funded
			err := chainDeployer.FundAccount(ctx, userAddr, fundAmount)
			assert.NoError(t, err, "failed to fund user info")
			go keepMiningBlocks(ctx, chainDeployer, userAddr)

			// and user pledged fund to validator
			specifications.ApproveTokenSpecification(t, ctx, chainDriver)
			specifications.DepositFundSpecification(t, ctx, chainDriver)
		}

		// When user deployed database
		grpcDriver := adapters.GetGrpcDriver(t, ctx, remoteKwildAddr, chainEnvs, remoteDBUrl)
		grpcDriver.SetFundConfig(userFundConfig)
		// chain sync, wait kwild to register user
		time.Sleep(chainSyncWaitTime)
		specifications.DatabaseDeploySpecification(t, ctx, grpcDriver)

		// Then user should be able to drop database
		specifications.DatabaseDropSpecification(t, ctx, grpcDriver)
	})

	t.Run("should execute database", func(t *testing.T) {
		ctx := context.Background()
		// setup
		specifications.SetSchemaLoader(
			&specifications.FileDatabaseSchemaLoader{
				FilePath: dbSchemaPath,
				Modifier: func(db *databases.Database[[]byte]) {
					db.Owner = userAddr
				}})
		chainDriver, chainDeployer, userFundConfig, chainEnvs := adapters.GetChainDriverAndDeployer(
			t, ctx, providerEndpoint, deployerPK, _chainCode, userPK, fundingPoolAddress, domination)

		if !*remote {
			// Given user is funded
			err := chainDeployer.FundAccount(ctx, userAddr, fundAmount)
			assert.NoError(t, err, "failed to fund user info")
			go keepMiningBlocks(ctx, chainDeployer, userAddr)

			// and user pledged fund to validator
			specifications.ApproveTokenSpecification(t, ctx, chainDriver)
			specifications.DepositFundSpecification(t, ctx, chainDriver)
		}

		// When user deployed database
		grpcDriver := adapters.GetGrpcDriver(t, ctx, remoteKwildAddr, chainEnvs, remoteDBUrl)
		grpcDriver.SetFundConfig(userFundConfig)
		// chain sync, wait kwild to register user
		time.Sleep(chainSyncWaitTime)
		specifications.DatabaseDeploySpecification(t, ctx, grpcDriver)

		// Then user should be able to execute database
		specifications.ExecuteDBInsertSpecification(t, ctx, grpcDriver)
		specifications.ExecuteDBUpdateSpecification(t, ctx, grpcDriver)
		specifications.ExecuteDBDeleteSpecification(t, ctx, grpcDriver)

		// and user should be able to drop database
		specifications.DatabaseDropSpecification(t, ctx, grpcDriver)
	})
}
