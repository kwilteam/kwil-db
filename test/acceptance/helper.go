package acceptance

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"os"

	//"kwil/cmd/kwil-cli/app"
	"kwil/internal/app/kwild"
	"kwil/pkg/chain/types"
	"kwil/pkg/client"
	"kwil/pkg/engine/models"
	"kwil/pkg/log"
	"kwil/test/acceptance/adapters"
	"kwil/test/acceptance/utils/deployer"
	eth_deployer "kwil/test/acceptance/utils/deployer/eth-deployer"
	"kwil/test/specifications"
	"math/big"
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
	UserPrivateKeyString             string
	SecondUserPrivateKeyString       string
	DatabaseDeployerPrivateKeyString string
	DBSchemaFilePath                 string
	NodeURL                          string // kwild address
	GatewayURL                       string // kgw address
	ChainRPCURL                      string
	ChainSyncWaitTime                time.Duration
	ChainCode                        types.ChainCode
	InitialFundAmount                int64
	denomination                     *big.Int
	LogLevel                         string

	// populated by init
	UserPrivateKey       *ecdsa.PrivateKey
	SecondUserPrivateKey *ecdsa.PrivateKey
	DeployerPrivateKey   *ecdsa.PrivateKey
	UserAddr             string
	SecondUserAddr       string
	DeployerAddr         string
}

func (e *TestEnvCfg) init(t *testing.T) {
	var err error
	e.UserPrivateKey, err = crypto.HexToECDSA(e.UserPrivateKeyString)
	if err != nil {
		t.Fatal(fmt.Errorf("invalid user private key: %w", err))
	}
	e.SecondUserPrivateKey, err = crypto.HexToECDSA(e.SecondUserPrivateKeyString)
	if err != nil {
		t.Fatal(fmt.Errorf("invalid second user private key: %w", err))
	}
	e.DeployerPrivateKey, err = crypto.HexToECDSA(e.DatabaseDeployerPrivateKeyString)
	if err != nil {
		t.Fatal(fmt.Errorf("invalid deployer private key: %w", err))
	}
	e.UserAddr = crypto.PubkeyToAddress(e.UserPrivateKey.PublicKey).Hex()
	e.SecondUserAddr = crypto.PubkeyToAddress(e.SecondUserPrivateKey.PublicKey).Hex()
	e.DeployerAddr = crypto.PubkeyToAddress(e.DeployerPrivateKey.PublicKey).Hex()
}

func GetTestEnvCfg(t *testing.T, remote bool) TestEnvCfg {
	var e TestEnvCfg

	if remote {
		e = TestEnvCfg{
			UserPrivateKeyString:             os.Getenv("TEST_USER_PK"),
			SecondUserPrivateKeyString:       os.Getenv("TEST_SECOND_USER_PK"),
			DatabaseDeployerPrivateKeyString: os.Getenv("TEST_DEPLOYER_PK"),
			DBSchemaFilePath:                 "./test-data/test_db.kf",
			NodeURL:                          os.Getenv("TEST_KWILD_ADDR"),
			GatewayURL:                       os.Getenv("TEST_KGW_ADDR"),
			ChainRPCURL:                      os.Getenv("TEST_PROVIDER"),
			ChainSyncWaitTime:                15 * time.Second,
			ChainCode:                        types.GOERLI,
			InitialFundAmount:                1,
			denomination:                     big.NewInt(10000),
			LogLevel:                         "debug",
		}
	} else {
		e = TestEnvCfg{
			UserPrivateKeyString:             adapters.UserAccountPrivateKey,
			SecondUserPrivateKeyString:       adapters.SecondUserPrivateKey,
			DatabaseDeployerPrivateKeyString: adapters.DeployerAccountPrivateKey,
			DBSchemaFilePath:                 "./test-data/test_db.kf",
			NodeURL:                          "",
			GatewayURL:                       "",
			ChainRPCURL:                      "",
			ChainSyncWaitTime:                2 * time.Second,
			ChainCode:                        types.GOERLI,
			InitialFundAmount:                100,
			denomination:                     big.NewInt(1000000000000000000),
			LogLevel:                         "debug",
		}
	}

	e.init(t)
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
	chainDeployer := GetDeployer("eth", exposedChainRPC, cfg.DatabaseDeployerPrivateKeyString, cfg.denomination)
	tokenAddress, err := chainDeployer.DeployToken(ctx)
	require.NoError(t, err, "failed to deploy token")
	escrowAddress, err := chainDeployer.DeployEscrow(ctx, tokenAddress.String())
	require.NoError(t, err, "failed to deploy escrow")

	kwildEnv := map[string]string{
		"KWILD_PRIVATE_KEY":                   cfg.DatabaseDeployerPrivateKeyString,
		"KWILD_DEPOSITS_BLOCK_CONFIRMATIONS":  "1",
		"KWILD_DEPOSITS_CHAIN_CODE":           fmt.Sprintf("%d", cfg.ChainCode),
		"KWILD_DEPOSITS_POOL_ADDRESS":         escrowAddress.String(),
		"KWILD_DEPOSITS_CLIENT_CHAIN_RPC_URL": unexposedChainRPC,
		"KWILD_LOG_LEVEL":                     cfg.LogLevel,
	}
	kwildC := adapters.StartKwildDockerService(t, ctx, kwildEnv)
	exposedKwildEndpoint, err := kwildC.ExposedEndpoint(ctx)
	require.NoError(t, err)
	exposedKgwEndpoint, err := kwildC.SecondExposedEndpoint(ctx)
	require.NoError(t, err)

	cfg.ChainRPCURL = exposedChainRPC
	cfg.NodeURL = exposedKwildEndpoint
	cfg.GatewayURL = exposedKgwEndpoint
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
			FilePath: cfg.DBSchemaFilePath,
			Modifier: func(db *models.Dataset) {
				db.Owner = cfg.UserAddr
				// NOTE: this is a hack to make sure the db name is temporary unique
				db.Name = fmt.Sprintf("%s_%s", db.Name, time.Now().Format("20160102"))
			}})
}

/*
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
}*/

func setupGrpcDriver(ctx context.Context, t *testing.T, cfg TestEnvCfg, logger log.Logger) (KwilAcceptanceDriver, deployer.Deployer, TestEnvCfg) {
	setSchemaLoader(cfg)

	if cfg.NodeURL != "" {
		t.Logf("create kwild driver to %s, (gateway: %s)", cfg.NodeURL, cfg.GatewayURL)
		kwilClt, err := client.New(ctx, cfg.NodeURL)
		require.NoError(t, err, "failed to create kwil client")

		kwildDriver := kwild.NewKwildDriver(kwilClt, cfg.UserPrivateKey, cfg.GatewayURL, logger)
		return kwildDriver, nil, cfg
	}

	updatedCfg, chainDeployer := setupCommon(ctx, t, cfg)

	t.Logf("create kwild driver to %s, (gateway: %s)", updatedCfg.NodeURL, updatedCfg.GatewayURL)

	kwilClt, err := client.New(ctx, updatedCfg.NodeURL,
		client.WithChainRpcUrl(updatedCfg.ChainRPCURL),
		client.WithPrivateKey(updatedCfg.UserPrivateKey),
	)
	require.NoError(t, err, "failed to create kwil client")

	kwildDriver := kwild.NewKwildDriver(kwilClt, updatedCfg.UserPrivateKey, updatedCfg.GatewayURL, logger)
	return kwildDriver, chainDeployer, updatedCfg
}

// NewClient creates a new client that is a KwilAcceptanceDriver
// this can be used to simulate several "wallets" in the same test
func newGRPCClient(ctx context.Context, t *testing.T, cfg *TestEnvCfg, logger log.Logger) KwilAcceptanceDriver {
	kwilClt, err := client.New(ctx, cfg.NodeURL,
		client.WithChainRpcUrl(cfg.ChainRPCURL),
		client.WithPrivateKey(cfg.UserPrivateKey),
	)
	require.NoError(t, err, "failed to create kwil client")

	kwildDriver := kwild.NewKwildDriver(kwilClt, cfg.UserPrivateKey, cfg.GatewayURL, logger)
	return kwildDriver
}

func GetDriver(ctx context.Context, t *testing.T, driverType string, cfg TestEnvCfg, logger log.Logger) (KwilAcceptanceDriver, deployer.Deployer, TestEnvCfg) {
	switch driverType {
	//case "cli":
	//	return setupCliDriver(ctx, t, cfg, logger)
	case "grpc":
		return setupGrpcDriver(ctx, t, cfg, logger)
	default:
		panic("unknown driver type")
	}
}

func NewClient(ctx context.Context, t *testing.T, driverType string, cfg TestEnvCfg, logger log.Logger) KwilAcceptanceDriver {
	// sort of hacky, but we want to use the second user's private key
	cfg.UserPrivateKeyString = cfg.SecondUserPrivateKeyString
	cfg.UserPrivateKey = cfg.SecondUserPrivateKey
	switch driverType {
	//case "cli":
	//	return setupCliDriver(ctx, t, cfg, logger)
	case "grpc":
		return newGRPCClient(ctx, t, &cfg, logger)
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
