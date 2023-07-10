package acceptance

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"os"
	"path/filepath"

	"github.com/kwilteam/kwil-db/pkg/kuneiform/schema"

	//"github.com/kwilteam/kwil-db/cmd/kwil-cli/app"
	"math/big"
	"runtime"
	"testing"
	"time"

	"github.com/kwilteam/kwil-db/internal/app/kwild"
	"github.com/kwilteam/kwil-db/pkg/chain/types"
	"github.com/kwilteam/kwil-db/pkg/client"
	"github.com/kwilteam/kwil-db/pkg/log"
	"github.com/kwilteam/kwil-db/test/acceptance/adapters"
	"github.com/kwilteam/kwil-db/test/acceptance/utils/deployer"
	eth_deployer "github.com/kwilteam/kwil-db/test/acceptance/utils/deployer/eth-deployer"
	"github.com/kwilteam/kwil-db/test/specifications"

	rpchttp "github.com/cometbft/cometbft/rpc/client/http"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/compose"
	"github.com/testcontainers/testcontainers-go/wait"
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

func SetupKwildCluster(ctx context.Context, t *testing.T, cfg TestEnvCfg, path string) (TestEnvCfg, []*testcontainers.DockerContainer, deployer.Deployer) {
	// Create Ganache container
	fmt.Println("ChainRPCURL: ", cfg.ChainRPCURL)
	t.Logf("Create ganache container:  %s\n", cfg.ChainCode.ToChainId().String())
	dockerComposeId := fmt.Sprintf("%d", time.Now().Unix())
	t.Log("dockerComposeId", dockerComposeId)
	pathG := filepath.Join(path, "/ganache/docker-compose.yml")
	composeG, err := compose.NewDockerCompose(pathG)
	require.NoError(t, err, "failed to create ganache docker compose")
	err = composeG.
		WithEnv(map[string]string{
			"uid": dockerComposeId,
		}).
		WaitForService("ganache", wait.NewLogStrategy("RPC Listening on 0.0.0.0:8545")).
		Up(ctx)
	require.NoError(t, err, "failed to start ganache container")
	t.Log("Ganache container is up")
	t.Cleanup(func() {
		assert.NoError(t, composeG.Down(context.Background(), compose.RemoveOrphans(true), compose.RemoveImagesLocal), "compose.Down()")
	})

	serviceG := composeG.Services()
	assert.Contains(t, serviceG, "ganache")

	ganacheC, err := composeG.ServiceContainer(ctx, "ganache")
	require.NoError(t, err, "failed to get ganache container")

	exposedChainRPC, err := ganacheC.PortEndpoint(ctx, "8545", "ws")
	t.Log("exposedChainRPC", exposedChainRPC)
	require.NoError(t, err, "failed to get exposed endpoint")
	ganacheIp, err := ganacheC.ContainerIP(ctx)
	require.NoError(t, err, "failed to get ganache container ip")
	unexposedChainRPC := fmt.Sprintf("ws://%s:8545", ganacheIp)
	t.Log("unexposedChainRPC", unexposedChainRPC)

	// Deploy contracts
	chainDeployer := GetDeployer("eth", exposedChainRPC, cfg.DatabaseDeployerPrivateKeyString, cfg.denomination)
	tokenAddress, err := chainDeployer.DeployToken(ctx)
	require.NoError(t, err, "failed to deploy token")
	t.Log("Token address: ", tokenAddress.String())
	escrowAddress, err := chainDeployer.DeployEscrow(ctx, tokenAddress.String())
	require.NoError(t, err, "failed to deploy contract")
	t.Logf("Escrow address: %s\n", escrowAddress.String())
	cfg.ChainRPCURL = exposedChainRPC
	t.Log("create Kwil cluster container")
	fmt.Println("ChainRPCURL: ", cfg.ChainRPCURL)
	time.Sleep(20 * time.Second)

	// Create Kwil cluster container
	pathK := filepath.Join(path, "/kwil/docker-compose.yml")
	composeKwild, err := compose.NewDockerCompose(pathK)
	require.NoError(t, err, "failed to create docker compose object for kwild cluster")
	fmt.Println("Unexposed chain rpc: ", unexposedChainRPC)
	err = composeKwild.
		WithEnv(map[string]string{
			"uid":                                 dockerComposeId,
			"KWILD_DEPOSITS_CLIENT_CHAIN_RPC_URL": unexposedChainRPC,
			// "KWILD_DEPOSITS_POOL_ADDRESS":         escrowAddress.String(),
			// "KWILD_PRIVATE_KEY":                   "b08786f38934aac966d10f0bc79a72f15067896d3b3beba721b5c235ffc5cc5f",
			// "KWILD_DEPOSITS_BLOCK_CONFIRMATIONS":  "1",
			// "KWILD_DEPOSITS_CHAIN_CODE":           "2",
			// "KWILD_LOG_LEVEL":                     "debug",
			// "COMET_BFT_HOME":                      "/apt/comet-bft",
		}).
		WaitForService("k1", wait.NewLogStrategy("grpc server started")).
		WaitForService("k2", wait.NewLogStrategy("grpc server started")).
		WaitForService("k3", wait.NewLogStrategy("grpc server started")).
		Up(ctx)
	require.NoError(t, err, "failed to start kwild cluster container")
	t.Cleanup(func() {
		assert.NoError(t, composeKwild.Down(context.Background(), compose.RemoveOrphans(true), compose.RemoveImagesLocal), "compose.Down()")
	})

	kwildserviceNames := composeKwild.Services()
	t.Log("serviceNames", kwildserviceNames)
	var kwildC []*testcontainers.DockerContainer
	for _, name := range kwildserviceNames {
		container, err := composeKwild.ServiceContainer(ctx, name)
		require.NoError(t, err, "failed to get container for service %s", name)
		kwildC = append(kwildC, container)
		/* ports, err := container.Ports(ctx)
		require.NoError(t, err, "failed to get ports for service %s", name)
		t.Logf("ports: %v for container name: %s", ports, name)

		nodeURL, err := container.PortEndpoint(ctx, "50051", "")
		require.NoError(t, err, "failed to get node url for service %s", name)
		t.Logf("nodeURL: %s for container name: %s", nodeURL, name)
		gatewayURL, err := container.PortEndpoint(ctx, "8080", "")
		require.NoError(t, err, "failed to get gateway url for service %s", name)
		t.Logf("gatewayURL: %s for container name: %s", gatewayURL, name) */
	}
	return cfg, kwildC, chainDeployer
}

func SetupKwildDriver(ctx context.Context, t *testing.T, cfg TestEnvCfg, kwildC *testcontainers.DockerContainer, logger log.Logger) KwilAcceptanceDriver {
	setSchemaLoader(cfg)

	nodeURL, err := kwildC.PortEndpoint(ctx, "50051", "")
	require.NoError(t, err)
	gatewayURL, err := kwildC.PortEndpoint(ctx, "8080", "")
	require.NoError(t, err)
	cometBftURL, err := kwildC.PortEndpoint(ctx, "26657", "tcp")
	require.NoError(t, err)

	name, err := kwildC.Name(ctx)
	require.NoError(t, err)

	t.Logf("nodeURL: %s gatewayURL: %s for container name: %s", nodeURL, gatewayURL, name)
	kwilClt, err := client.New(ctx, nodeURL,
		client.WithChainRpcUrl(cfg.ChainRPCURL),
		client.WithPrivateKey(cfg.UserPrivateKey),
	)
	require.NoError(t, err, "failed to create kwil client")

	bcClt, err := rpchttp.New(cometBftURL, "")
	require.NoError(t, err, "failed to create comet bft client")

	kwildDriver := kwild.NewKwildDriver(kwilClt, bcClt, cfg.UserPrivateKey, gatewayURL, logger)
	return kwildDriver
}

func setupCommon(ctx context.Context, t *testing.T, cfg TestEnvCfg) (TestEnvCfg, deployer.Deployer) {
	// ganache container
	ganacheC := adapters.StartGanacheDockerService(t, ctx, cfg.ChainCode.ToChainId().String())
	exposedChainRPC, err := ganacheC.ExposedEndpoint(ctx)
	require.NoError(t, err, "failed to get exposed endpoint")
	fmt.Println("exposedChainRPC: ", exposedChainRPC)
	unexposedChainRPC, err := ganacheC.UnexposedEndpoint(ctx)
	require.NoError(t, err, "failed to get unexposed endpoint")
	fmt.Println("unexposedChainRPC: ", unexposedChainRPC)

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
		"COMET_BFT_HOME":                      "/app/comet-bft",
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
			Modifier: func(db *schema.Schema) {
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

		kwildDriver := kwild.NewKwildDriver(kwilClt, nil, cfg.UserPrivateKey, cfg.GatewayURL, logger)
		return kwildDriver, nil, cfg
	}

	updatedCfg, chainDeployer := setupCommon(ctx, t, cfg)

	t.Logf("create kwild driver to %s, (gateway: %s)", updatedCfg.NodeURL, updatedCfg.GatewayURL)

	kwilClt, err := client.New(ctx, updatedCfg.NodeURL,
		client.WithChainRpcUrl(updatedCfg.ChainRPCURL),
		client.WithPrivateKey(updatedCfg.UserPrivateKey),
	)
	require.NoError(t, err, "failed to create kwil client")

	kwildDriver := kwild.NewKwildDriver(kwilClt, nil, updatedCfg.UserPrivateKey, updatedCfg.GatewayURL, logger)
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

	kwildDriver := kwild.NewKwildDriver(kwilClt, nil, cfg.UserPrivateKey, cfg.GatewayURL, logger)
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
