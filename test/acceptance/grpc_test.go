package acceptance_test

import (
	"context"
	"crypto/ecdsa"
	"flag"
	"fmt"
	"github.com/stretchr/testify/require"
	"kwil/internal/app/kgw"
	"kwil/internal/app/kwild"
	"kwil/pkg/chain/types"
	"kwil/pkg/client"
	"kwil/pkg/databases"
	"kwil/pkg/log"
	"kwil/test/adapters"
	"kwil/test/specifications"
	"kwil/test/utils/deployer"
	eth_deployer "kwil/test/utils/deployer/eth-deployer"
	"math/big"
	"os"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/assert"
)

var remote = flag.Bool("remote", false, "run tests against remote environment")

func keepMiningBlocks(ctx context.Context, deployer deployer.Deployer, account string) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			time.Sleep(1 * time.Second)
			// to mine new blocks
			err := deployer.FundAccount(ctx, account, 1)
			if err != nil {
				fmt.Println("funded user account failed", err)
			}
		}
	}
}

type TestEnvCfg struct {
	UserPkStr         string
	DeployerPkStr     string
	DbSchemaPath      string
	NodeAddr          string // kwild address
	GatewayAddr       string // kgw address
	ChainRPCURL       string
	ChainSyncWaitTime time.Duration
	ChainCode         types.ChainCode
	FundAmount        int64
	Domination        *big.Int
	LogLevel          string

	// populated by init
	UserPK       *ecdsa.PrivateKey
	DeployerPK   *ecdsa.PrivateKey
	UserAddr     string
	DeployerAddr string
}

func NewTestEnv(userPkStr, deployerPkStr, dbSchemaPath, remoteKwildAddr, remoteGatewayAddr, remoteChainRPCURL string,
	chainSyncWaitTime time.Duration, chainCode types.ChainCode, fundAmount int64, domination *big.Int, logLevel string) TestEnvCfg {
	return TestEnvCfg{
		UserPkStr:         userPkStr,
		DeployerPkStr:     deployerPkStr,
		DbSchemaPath:      dbSchemaPath,
		NodeAddr:          remoteKwildAddr,
		GatewayAddr:       remoteGatewayAddr,
		ChainRPCURL:       remoteChainRPCURL,
		ChainSyncWaitTime: chainSyncWaitTime,
		ChainCode:         chainCode,
		FundAmount:        fundAmount,
		Domination:        domination,
		LogLevel:          logLevel,
	}
}

func (e *TestEnvCfg) init(t *testing.T) {
	var err error
	e.UserPK, err = crypto.HexToECDSA(e.UserPkStr)
	if err != nil {
		t.Fatal(fmt.Errorf("invalid user private key: %w", err))
	}
	e.DeployerPK, err = crypto.HexToECDSA(e.DeployerPkStr)
	if err != nil {
		t.Fatal(fmt.Errorf("invalid deployer private key: %w", err))
	}
	e.UserAddr = crypto.PubkeyToAddress(e.UserPK.PublicKey).Hex()
	e.DeployerAddr = crypto.PubkeyToAddress(e.DeployerPK.PublicKey).Hex()
}

func getTestEnvCfg(t *testing.T, remote bool) TestEnvCfg {
	var e TestEnvCfg
	if remote {
		e = NewTestEnv(
			os.Getenv("TEST_USER_PK"),
			os.Getenv("TEST_DEPLOYER_PK"),
			"./test-data/database_schema.json",
			os.Getenv("TEST_KWILD_ADDR"),
			os.Getenv("TEST_KGW_ADDR"),
			os.Getenv("TEST_PROVIDER"),
			15*time.Second,
			types.GOERLI,
			1,
			big.NewInt(10000),
			"debug")
	} else {
		e = NewTestEnv(
			adapters.UserAccountPK,
			adapters.DeployerAccountPK,
			"./test-data/database_schema.json",
			"",
			"",
			"",
			2*time.Second,
			types.GOERLI,
			100,
			big.NewInt(1000000000000000000),
			"debug")
	}

	e.init(t)
	return e
}

func setup(ctx context.Context, t *testing.T, cfg TestEnvCfg, logger log.Logger) (*kgw.KgwDriver, deployer.Deployer) {
	specifications.SetSchemaLoader(
		&specifications.FileDatabaseSchemaLoader{
			FilePath: cfg.DbSchemaPath,
			Modifier: func(db *databases.Database[[]byte]) {
				db.Owner = cfg.UserAddr
			}})

	if cfg.NodeAddr != "" {
		t.Logf("create kwild driver to %s", cfg.NodeAddr)
		kwilClt, err := client.New(ctx, cfg.NodeAddr)
		require.NoError(t, err, "failed to create kwil client")
		kwildDriver := kwild.NewKwildDriver(kwilClt, cfg.UserPK, logger)
		t.Logf("create kgw driver to %s", cfg.GatewayAddr)
		kgwDriver := kgw.NewKgwDriver(cfg.GatewayAddr, kwildDriver)
		return kgwDriver, nil
	}

	// ganache container
	ganacheC := adapters.StartGanacheDockerService(t, ctx, cfg.ChainCode.ToChainId().String())
	exposedChainRPC, err := ganacheC.ExposedEndpoint(ctx)
	require.NoError(t, err, "failed to get exposed endpoint")
	unexposedChainRPC, err := ganacheC.UnexposedEndpoint(ctx)
	require.NoError(t, err, "failed to get unexposed endpoint")
	// deploy token and escrow contract
	// @yaiba TODO: chain agnostic
	//t.Logf("create chain driver to %s", exposedChainRPC)
	//chainDriver := ethFund.New(exposedChainRPC, cfg.DeployerAddr, logger)
	t.Logf("create chain deployer to %s", exposedChainRPC)
	chainDeployer := eth_deployer.NewEthDeployer(exposedChainRPC, cfg.DeployerPkStr, cfg.Domination)
	tokenAddress, err := chainDeployer.DeployToken(ctx)
	require.NoError(t, err, "failed to deploy token")
	escrowAddress, err := chainDeployer.DeployEscrow(ctx, tokenAddress.String())
	require.NoError(t, err, "failed to deploy escrow")

	// postgres db container
	dbC := adapters.StartDBDockerService(t, ctx)

	// hasura container
	unexposedKwilPgURL := dbC.GetUnexposedDBUrl(ctx, adapters.KwildDatabase)
	unexposedHasuraPgURL := dbC.GetUnexposedDBUrl(ctx, "postgres")
	hasuraEnvs := map[string]string{
		"PG_DATABASE_URL":                      unexposedKwilPgURL,
		"HASURA_GRAPHQL_METADATA_DATABASE_URL": unexposedHasuraPgURL,
		"HASURA_GRAPHQL_ENABLE_CONSOLE":        "true",
		"HASURA_GRAPHQL_DEV_MODE":              "true",
		"HASURA_METADATA_DB":                   "postgres",
	}
	hasuraC := adapters.StartHasuraDockerService(ctx, t, hasuraEnvs)

	// kwild container
	unexposedHasuraEndpoint, err := hasuraC.UnexposedEndpoint(ctx)
	require.NoError(t, err)
	kwildEnv := map[string]string{
		"KWILD_FUND_RPC_URL":            unexposedChainRPC, // kwil will call using docker network
		"KWILD_FUND_POOL_ADDRESS":       escrowAddress.String(),
		"KWILD_FUND_WALLET":             cfg.DeployerPkStr,
		"KWILD_FUND_CHAIN_CODE":         fmt.Sprintf("%d", cfg.ChainCode),
		"KWILD_FUND_BLOCK_CONFIRMATION": "1",
		"KWILD_FUND_RECONNECT_INTERVAL": "30",
		"KWILD_GRAPHQL_ADDR":            unexposedHasuraEndpoint,
		// @yaiba can't get addr here, because the gw container is not ready yet
		// need a hacky way to get the addr
		"KWILD_GATEWAY_ADDR": "",
		"KWILD_DB_URL":       unexposedKwilPgURL,
		"KWILD_LOG_LEVEL":    cfg.LogLevel,
	}
	kwildC := adapters.StartKwildDockerService(t, ctx, kwildEnv)

	// kgw container
	exposedkwildEndpoint, err := kwildC.ExposedEndpoint(ctx)
	require.NoError(t, err)
	unexposedKwildEndpoint, err := kwildC.UnexposedEndpoint(ctx)
	require.NoError(t, err)
	kgwEnv := map[string]string{
		"KWILGW_KWILD_ADDR":         unexposedKwildEndpoint,
		"KWILGW_GRAPHQL_ADDR":       unexposedHasuraEndpoint,
		"KWILGW_LOG_LEVEL":          cfg.LogLevel,
		"KWILGW_SERVER_LISTEN_ADDR": ":8082",
	}
	kgwC := adapters.StartKgwDockerService(ctx, t, kgwEnv)

	//
	exposedKgwEndpoint, err := kgwC.ExposedEndpoint(ctx)
	require.NoError(t, err)

	cfg.NodeAddr = exposedkwildEndpoint
	cfg.GatewayAddr = exposedKgwEndpoint

	t.Logf("create kwild driver to %s", cfg.NodeAddr)
	kwilClt, err := client.New(ctx, cfg.NodeAddr,
		// TODO: to use returned chain rpc url
		client.WithChainRpcUrl(exposedChainRPC))

	require.NoError(t, err, "failed to create kwil client")
	kwildDriver := kwild.NewKwildDriver(kwilClt, cfg.UserPK, logger)
	t.Logf("create kgw driver to %s", cfg.GatewayAddr)
	kgwDriver := kgw.NewKgwDriver(cfg.GatewayAddr, kwildDriver)
	return kgwDriver, chainDeployer
}

func TestGrpcServerDatabaseService(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	tLogger := log.New(log.Config{
		Level:       "debug",
		OutputPaths: []string{"stdout"},
	})

	t.Run("should deposit fund", func(t *testing.T) {
		if *remote {
			t.Skip()
		}

		cfg := getTestEnvCfg(t, *remote)
		ctx := context.Background()
		// setup
		driver, chainDeployer := setup(ctx, t, cfg, tLogger)

		// Given user is funded with escrow token
		err := chainDeployer.FundAccount(ctx, cfg.UserAddr, cfg.FundAmount)
		assert.NoError(t, err, "failed to fund user config")

		// and user has approved funding_pool to spend his token
		specifications.ApproveTokenSpecification(ctx, t, driver)

		// should be able to deposit fund
		specifications.DepositFundSpecification(ctx, t, driver)
	})

	t.Run("should deploy and drop database", func(t *testing.T) {
		if *remote {
			t.Skip()
		}

		cfg := getTestEnvCfg(t, *remote)
		ctx := context.Background()
		// setup
		driver, chainDeployer := setup(ctx, t, cfg, tLogger)

		// Given user is funded with escrow token
		err := chainDeployer.FundAccount(ctx, cfg.UserAddr, cfg.FundAmount)
		assert.NoError(t, err, "failed to fund user config")
		go keepMiningBlocks(ctx, chainDeployer, cfg.UserAddr)
		// and user pledged fund to validator
		specifications.ApproveTokenSpecification(ctx, t, driver)
		specifications.DepositFundSpecification(ctx, t, driver)

		// chain sync, wait kwil to register user
		time.Sleep(cfg.ChainSyncWaitTime)

		// When user deployed database
		specifications.DatabaseDeploySpecification(ctx, t, driver)

		// Then user should be able to drop database
		specifications.DatabaseDropSpecification(ctx, t, driver)
	})

	t.Run("should execute database", func(t *testing.T) {
		cfg := getTestEnvCfg(t, *remote)
		ctx := context.Background()
		// setup
		driver, chainDeployer := setup(ctx, t, cfg, tLogger)

		// only local env test
		if !*remote {
			// Given user is funded
			err := chainDeployer.FundAccount(ctx, cfg.UserAddr, cfg.FundAmount)
			assert.NoError(t, err, "failed to fund user config")
			go keepMiningBlocks(ctx, chainDeployer, cfg.UserAddr)

			// and user pledged fund to validator
			specifications.ApproveTokenSpecification(ctx, t, driver)
			specifications.DepositFundSpecification(ctx, t, driver)
		}

		// chain sync, wait kwil to register user
		time.Sleep(cfg.ChainSyncWaitTime)

		// When user deployed database
		specifications.DatabaseDeploySpecification(ctx, t, driver)

		// Then user should be able to execute database
		specifications.ExecuteDBInsertSpecification(ctx, t, driver)
		specifications.ExecuteDBUpdateSpecification(ctx, t, driver)
		specifications.ExecuteDBDeleteSpecification(ctx, t, driver)

		// and user should be able to drop database
		specifications.DatabaseDropSpecification(ctx, t, driver)
	})
}
