package acceptance_test

import (
	"context"
	"crypto/ecdsa"
	"flag"
	"fmt"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/spf13/viper"
	"kwil/pkg/chain/client/dto"
	"kwil/pkg/chain/types"
	"kwil/pkg/databases"
	"kwil/pkg/fund"
	"kwil/pkg/grpc/client"
	kwil_client "kwil/pkg/kclient"
	"kwil/pkg/log"
	"kwil/test/adapters"
	"kwil/test/specifications"
	"kwil/test/utils/deployer"
	"math/big"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

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
				fmt.Println("funded user account failed", err)
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
		userAddr  string
		userPKStr string
		userPK    *ecdsa.PrivateKey
		// test deployer
		deployerPKStr string
		// database schema file
		dbSchemaPath string
		// remote kwil endpoint
		remoteKwildAddr string
		// remote blockchain endpoint
		remoteRpcUrl string
		// blochchain block produce interval
		chainSyncWaitTime  time.Duration
		fundingPoolAddress string
		_chainCode         types.ChainCode
		fundAmount         int64
		domination         *big.Int
		remoteGraphqlAddr  string
		remoteAPIKey       string
	)

	viper.SetDefault("log.level", "config")

	localEnv := func() {
		userPKStr = adapters.UserAccountPK
		deployerPKStr = adapters.DeployerAccountPK
		dbSchemaPath = "./test-data/database_schema.json"
		remoteKwildAddr = ""
		remoteRpcUrl = ""
		chainSyncWaitTime = 3 * time.Second
		_chainCode = types.GOERLI
		fundAmount = 100
		domination = big.NewInt(1000000000000000000)
	}

	remoteEnv := func() {
		// depends on the remote environment, change respectively
		userPKStr = os.Getenv("TEST_USER_PK")
		deployerPKStr = os.Getenv("TEST_DEPLOYER_PK")
		dbSchemaPath = "./test-data/database_schema.json"
		remoteKwildAddr = os.Getenv("TEST_KWILD_ADDR")
		remoteRpcUrl = os.Getenv("TEST_PROVIDER")
		chainSyncWaitTime = 15 * time.Second
		_chainCode = types.GOERLI
		fundAmount = 1
		fundingPoolAddress = os.Getenv("TEST_POOL_ADDRESS")
		remoteGraphqlAddr = os.Getenv("TEST_GRAPHQL_ADDR")
		domination = big.NewInt(10000)
		remoteAPIKey = os.Getenv("TEST_KGW_API_KEY")
	}

	tLogger := log.New(log.Config{
		Level:       "debug",
		OutputPaths: []string{"stdout"},
	})

	if *remote {
		remoteEnv()
	} else {
		localEnv()
	}

	userPK, err := crypto.HexToECDSA(userPKStr)
	if err != nil {
		t.Fatal(fmt.Errorf("invalid user private key: %w", err))
	}
	userAddr = crypto.PubkeyToAddress(userPK.PublicKey).Hex()

	fundCfg := fund.Config{
		Chain: dto.Config{
			ChainCode:         int64(_chainCode),
			RpcUrl:            remoteRpcUrl,
			BlockConfirmation: 10,
			ReconnectInterval: 30,
		},
		Wallet:      userPK,
		PoolAddress: fundingPoolAddress,
	}

	t.Run("should deposit fund", func(t *testing.T) {
		if *remote {
			t.Skip()
		}

		ctx := context.Background()
		// setup
		chainDriver, chainDeployer, _, _ := adapters.GetChainDriverAndDeployer(ctx, t, remoteRpcUrl, deployerPKStr, _chainCode, domination, &fundCfg, tLogger)

		// Given user is funded with escrow token
		err := chainDeployer.FundAccount(ctx, userAddr, fundAmount)
		assert.NoError(t, err, "failed to fund user config")

		// and user has approved funding_pool to spend his token
		specifications.ApproveTokenSpecification(ctx, t, chainDriver)

		// should be able to deposit fund
		specifications.DepositFundSpecification(ctx, t, chainDriver)
	})

	t.Run("should deploy and drop database", func(t *testing.T) {
		if *remote {
			t.Skip()
		}

		ctx := context.Background()
		// setup
		specifications.SetSchemaLoader(
			&specifications.FileDatabaseSchemaLoader{
				FilePath: dbSchemaPath,
				Modifier: func(db *databases.Database[[]byte]) {
					db.Owner = userAddr
				}})
		chainDriver, chainDeployer, userFundConfig, chainEnvs := adapters.GetChainDriverAndDeployer(ctx, t, remoteRpcUrl, deployerPKStr, _chainCode, domination, &fundCfg, tLogger)

		if !*remote {
			// Given user is funded
			err := chainDeployer.FundAccount(ctx, userAddr, fundAmount)
			assert.NoError(t, err, "failed to fund user config")
			go keepMiningBlocks(ctx, chainDeployer, userAddr)

			// and user pledged fund to validator
			specifications.ApproveTokenSpecification(ctx, t, chainDriver)
			specifications.DepositFundSpecification(ctx, t, chainDriver)
		}

		// When user deployed database
		cltConfig := &kwil_client.Config{
			Node: client.Config{
				Endpoint: remoteKwildAddr,
			},
			Fund: *userFundConfig,
			Log: log.Config{
				Level:       "info",
				OutputPaths: []string{"stdout"},
			},
		}
		grpcDriver := adapters.GetKwildDriver(ctx, t, remoteKwildAddr, cltConfig, chainEnvs)
		// chain sync, wait kwil to register user
		time.Sleep(chainSyncWaitTime)
		specifications.DatabaseDeploySpecification(ctx, t, grpcDriver)

		// Then user should be able to drop database
		specifications.DatabaseDropSpecification(ctx, t, grpcDriver)
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
		chainDriver, chainDeployer, userFundConfig, fundEnvs := adapters.GetChainDriverAndDeployer(ctx, t, remoteRpcUrl, deployerPKStr, _chainCode, domination, &fundCfg, tLogger)

		if !*remote {
			// Given user is funded
			err := chainDeployer.FundAccount(ctx, userAddr, fundAmount)
			assert.NoError(t, err, "failed to fund user config")
			go keepMiningBlocks(ctx, chainDeployer, userAddr)

			// and user pledged fund to validator
			specifications.ApproveTokenSpecification(ctx, t, chainDriver)
			specifications.DepositFundSpecification(ctx, t, chainDriver)
		}

		// When user deployed database
		cltConfig := &kwil_client.Config{
			Node: client.Config{
				Endpoint: remoteKwildAddr,
			},
			Fund: *userFundConfig,
		}
		grpcDriver := adapters.GetKgwDriver(ctx, t, remoteKwildAddr, remoteGraphqlAddr, remoteAPIKey, cltConfig, fundEnvs)
		// chain sync, wait kwil to register user
		time.Sleep(chainSyncWaitTime)
		specifications.DatabaseDeploySpecification(ctx, t, grpcDriver)

		// Then user should be able to execute database
		specifications.ExecuteDBInsertSpecification(ctx, t, grpcDriver)
		specifications.ExecuteDBUpdateSpecification(ctx, t, grpcDriver)
		specifications.ExecuteDBDeleteSpecification(ctx, t, grpcDriver)

		// and user should be able to drop database
		specifications.DatabaseDropSpecification(ctx, t, grpcDriver)
	})
}
