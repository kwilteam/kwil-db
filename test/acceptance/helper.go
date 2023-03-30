package acceptance

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"kwil/cmd/kwil-cli/app"
	"kwil/internal/app/kwild"
	"kwil/pkg/chain/types"
	client "kwil/pkg/client2"
	"kwil/pkg/databases"
	"kwil/pkg/log"
	"kwil/test/acceptance/adapters"
	"kwil/test/acceptance/utils/deployer"
	eth_deployer "kwil/test/acceptance/utils/deployer/eth-deployer"
	"kwil/test/specifications"
	"math/big"
	"os"
	"path"
	"runtime"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/require"
	"github.com/tonistiigi/go-rosetta"
)

// KeepMiningBlocks is a helper function to keep mining blocks
// since kwild need to mine blocks to process transactions, and ganache is configured to mine a block for every tx
// so we need to keep produce txs to keep mining blocks
func KeepMiningBlocks(ctx context.Context, done chan struct{}, deployer deployer.Deployer, account string) {
	for {
		select {
		case <-ctx.Done():
			return
		case <-done:
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
	DBSchemaPath      string
	NodeAddr          string // kwild address
	GatewayAddr       string // kgw address
	ChainRPCURL       string
	ChainSyncWaitTime time.Duration
	ChainCode         types.ChainCode
	FundAmount        int64
	denomination      *big.Int
	LogLevel          string

	// populated by init
	UserPK       *ecdsa.PrivateKey
	DeployerPK   *ecdsa.PrivateKey
	UserAddr     string
	DeployerAddr string
}

func NewTestEnv(userPkStr, deployerPkStr, dbSchemaPath, remoteKwildAddr, remoteGatewayAddr, remoteChainRPCURL string,
	chainSyncWaitTime time.Duration, chainCode types.ChainCode, fundAmount int64, denomination *big.Int, logLevel string) TestEnvCfg {
	return TestEnvCfg{
		UserPkStr:         userPkStr,
		DeployerPkStr:     deployerPkStr,
		DBSchemaPath:      dbSchemaPath,
		NodeAddr:          remoteKwildAddr,
		GatewayAddr:       remoteGatewayAddr,
		ChainRPCURL:       remoteChainRPCURL,
		ChainSyncWaitTime: chainSyncWaitTime,
		ChainCode:         chainCode,
		FundAmount:        fundAmount,
		denomination:      denomination,
		LogLevel:          logLevel,
	}
}

func (e *TestEnvCfg) Init(t *testing.T) {
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

func GetTestEnvCfg(t *testing.T, remote bool) TestEnvCfg {
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

	e.Init(t)
	return e
}

func setupCommon(ctx context.Context, t *testing.T, cfg TestEnvCfg) (TestEnvCfg, deployer.Deployer) {
	// ganache container
	ganacheC := adapters.StartGanacheDockerService(t, ctx, cfg.ChainCode.ToChainId().String())
	exposedChainRPC, err := ganacheC.ExposedEndpoint(ctx)
	require.NoError(t, err, "failed to get exposed endpoint")
	unexposedChainRPC, err := ganacheC.UnexposedEndpoint(ctx)
	require.NoError(t, err, "failed to get unexposed endpoint")

	// deploy token and escrow contract
	t.Logf("create chain deployer to %s", exposedChainRPC)
	chainDeployer := GetDeployer("eth", exposedChainRPC, cfg.DeployerPkStr, cfg.denomination)
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
		"KWILD_FUND_PUBLIC_RPC_URL":     exposedChainRPC,   // user will call using host network
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
	exposedKwildEndpoint, err := kwildC.ExposedEndpoint(ctx)
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

	cfg.ChainRPCURL = exposedChainRPC
	cfg.NodeAddr = exposedKwildEndpoint
	cfg.GatewayAddr = exposedKgwEndpoint
	return cfg, chainDeployer
}

func arch() string {
	arch := runtime.GOARCH
	if rosetta.Enabled() {
		arch += " (rosetta)"
	}
	return arch
}

func setSchemaLoader(cfg TestEnvCfg) {
	specifications.SetSchemaLoader(
		&specifications.FileDatabaseSchemaLoader{
			FilePath: cfg.DBSchemaPath,
			Modifier: func(db *databases.Database[[]byte]) {
				db.Owner = cfg.UserAddr
				// NOTE: this is a hack to make sure the db name is temporary unique
				db.Name = fmt.Sprintf("%s_%s", db.Name, time.Now().Format("20160102"))
			}})
}

func setupCliDriver(ctx context.Context, t *testing.T, cfg TestEnvCfg, logger log.Logger) (KwilACTDriver, deployer.Deployer) {
	setSchemaLoader(cfg)

	_, currentFilePath, _, _ := runtime.Caller(1)
	binPath := path.Join(path.Dir(currentFilePath), fmt.Sprintf("../../.build/kwil-cli-%s-%s", runtime.GOOS, arch()))
	if cfg.NodeAddr != "" {
		t.Logf("create cli driver to %s", cfg.NodeAddr)
		cliDriver := app.NewKwilCliDriver(binPath, cfg.NodeAddr, cfg.GatewayAddr, "", cfg.UserPkStr, cfg.UserAddr, logger)
		return cliDriver, nil
	}

	updatedCfg, chainDeployer := setupCommon(ctx, t, cfg)

	t.Logf("create cli driver to %s", updatedCfg.NodeAddr)
	cliDriver := app.NewKwilCliDriver(binPath, updatedCfg.NodeAddr, updatedCfg.GatewayAddr, updatedCfg.ChainRPCURL, updatedCfg.UserPkStr, cfg.UserAddr, logger)
	return cliDriver, chainDeployer
}

func setupGrpcDriver(ctx context.Context, t *testing.T, cfg TestEnvCfg, logger log.Logger) (KwilACTDriver, deployer.Deployer) {
	setSchemaLoader(cfg)

	if cfg.NodeAddr != "" {
		t.Logf("create kwild driver to %s, (gateway: %s)", cfg.NodeAddr, cfg.GatewayAddr)
		kwilClt, err := client.New(ctx, cfg.NodeAddr)
		require.NoError(t, err, "failed to create kwil client")
		kwildDriver := kwild.NewKwildDriver(kwilClt, cfg.UserPK, cfg.GatewayAddr, logger)
		return kwildDriver, nil
	}

	updatedCfg, chainDeployer := setupCommon(ctx, t, cfg)

	t.Logf("create kwild driver to %s, (gateway: %s)", updatedCfg.NodeAddr, updatedCfg.GatewayAddr)
	kwilClt, err := client.New(ctx, updatedCfg.NodeAddr)
	require.NoError(t, err, "failed to create kwil client")
	kwildDriver := kwild.NewKwildDriver(kwilClt, updatedCfg.UserPK, updatedCfg.GatewayAddr, logger)
	return kwildDriver, chainDeployer
}

func GetDriver(ctx context.Context, t *testing.T, driverType string, cfg TestEnvCfg, logger log.Logger) (KwilACTDriver, deployer.Deployer) {
	switch driverType {
	case "cli":
		return setupCliDriver(ctx, t, cfg, logger)
	case "grpc":
		return setupGrpcDriver(ctx, t, cfg, logger)
	default:
		panic("unknown driver type")
	}
}

func GetDeployer(deployerType string, rpcURL string, privateKeyStr string, domination *big.Int) deployer.Deployer {
	switch deployerType {
	case "eth":
		return eth_deployer.NewEthDeployer(rpcURL, privateKeyStr, eth_deployer.WithDomination(domination))
	default:
		panic("unknown deployer type")
	}
}
