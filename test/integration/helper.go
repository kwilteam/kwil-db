package integration

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"testing"

	"github.com/kwilteam/kwil-db/internal/app/kwild"
	"github.com/kwilteam/kwil-db/internal/pkg/nodecfg"
	"github.com/kwilteam/kwil-db/pkg/client"
	"github.com/kwilteam/kwil-db/pkg/crypto"
	"github.com/kwilteam/kwil-db/test/acceptance"
	"github.com/kwilteam/kwil-db/test/runner"

	"github.com/joho/godotenv"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/compose"
	"github.com/testcontainers/testcontainers-go/wait"
)

const DefaultContainerWaitTimeout = 10 * time.Second

type IntTestConfig struct {
	acceptance.ActTestCfg

	NValidator int
}

//// NewClient creates a new client that is a KwilAcceptanceDriver
//// this can be used to simulate several "wallets" in the same test
//func newGRPCClient(ctx context.Context, t *testing.T, cfg *IntTestConfig) KwilIntDriver {
//	kwilClt, err := client.New(ctx, cfg.GrpcEndpoint,
//		client.WithPrivateKey(cfg.AlicePK),
//		client.WithCometBftUrl(cfg.ChainEndpoint),
//	)
//	require.NoError(t, err, "failed to create kwil client")
//
//	kwildDriver := kwild.NewKwildDriver(kwilClt)
//	return kwildDriver
//}
//
//func NewClient(ctx context.Context, t *testing.T, driverType string, cfg *IntTestConfig) KwilIntDriver {
//	// sort of hacky, but we want to use the second user's private key
//	cfg.AliceRawPK = cfg.BobRawPK
//	cfg.AlicePK = cfg.BobPk
//	switch driverType {
//	//case "cli":fs
//	//	return setupCliDriver(ctx, t, cfg, logger)
//	case "grpc":
//		return newGRPCClient(ctx, t, cfg)
//	default:
//		panic("unknown driver type")
//	}
//}

type IntHelper struct {
	t          *testing.T
	cfg        *IntTestConfig
	teardown   []func()
	containers []*testcontainers.DockerContainer
}

func NewIntHelper(t *testing.T) *IntHelper {
	return &IntHelper{
		t: t,
	}
}

func (r *IntHelper) LoadConfig() {
	// default wallet mnemonic: test test test test test test test test test test test junk
	// default wallet hd path : m/44'/60'/0'
	cfg := &IntTestConfig{
		ActTestCfg: acceptance.ActTestCfg{
			AliceRawPK:        runner.GetEnv("KINT_ALICE_PK", "f1aa5a7966c3863ccde3047f6a1e266cdc0c76b399e256b8fede92b1c69e4f4e"),
			BobRawPK:          runner.GetEnv("KINT_BOB_PK", "43f149de89d64bf9a9099be19e1b1f7a4db784af8fa07caf6f08dc86ba65636b"),
			SchemaFile:        runner.GetEnv("KINT_SCHEMA", "./test-data/test_db.kf"),
			LogLevel:          runner.GetEnv("KINT_CHAIN_ENDPOINT", "http://localhost:26657"),
			GWEndpoint:        runner.GetEnv("KINT_GRPC_ENDPOINT", "localhost:9090"),
			GrpcEndpoint:      runner.GetEnv("KINT_GATEWAY_ENDPOINT", "localhost:8080"),
			DockerComposeFile: runner.GetEnv("KINT_DOCKER_COMPOSE_FILE", "./docker-compose.yml"),
		},
	}

	var err error
	nodeNum := runner.GetEnv("KWIL_INT_VALIDATOR_NUM", "3")
	cfg.NValidator, err = strconv.Atoi(nodeNum)
	require.NoError(r.t, err, "invalid node number")

	cfg.AlicePK, err = crypto.PrivateKeyFromHex(cfg.AliceRawPK)
	require.NoError(r.t, err, "invalid alice private key")

	cfg.BobPk, err = crypto.PrivateKeyFromHex(cfg.BobRawPK)
	require.NoError(r.t, err, "invalid bob private key")
	r.cfg = cfg

	cfg.DumpToEnv()
}

func (r *IntHelper) generateNodeConfig() {
	r.t.Logf("generate node config")
	tmpPath := r.t.TempDir()
	r.t.Logf("create test temp directory: %s", tmpPath)
	envVars, err := godotenv.Unmarshal("KWIL_HOME=" + tmpPath)
	require.NoError(r.t, err, "failed to unmarshal env vars")
	err = godotenv.Write(envVars, "./.env")
	require.NoError(r.t, err, "failed to write env vars to file")

	err = nodecfg.GenerateTestnetConfig(&nodecfg.TestnetGenerateConfig{
		NValidators:             r.cfg.NValidator,
		InitialHeight:           0,
		ConfigFile:              "",
		OutputDir:               tmpPath,
		NodeDirPrefix:           "node",
		PopulatePersistentPeers: true,
		HostnamePrefix:          "kwil-",
		HostnameSuffix:          "",
		StartingIPAddress:       "172.10.100.2",
		P2pPort:                 26656,
	})
	require.NoError(r.t, err, "failed to generate testnet config")
}

func (r *IntHelper) runDockerCompose(ctx context.Context) {
	r.t.Logf("run in docker compose")

	//setSchemaLoader(r.cfg.AliceAddr())

	fEnv, err := os.Open("./.env")
	require.NoError(r.t, err, "failed to open .env file")

	envs, err := godotenv.Parse(fEnv)
	require.NoError(r.t, err, "failed to parse .env file")

	dc, err := compose.NewDockerCompose(r.cfg.DockerComposeFile)
	require.NoError(r.t, err, "failed to create docker compose object for kwild cluster")
	err = dc.
		WithEnv(envs).
		WaitForService("ext1",
			wait.NewLogStrategy("listening on").WithStartupTimeout(DefaultContainerWaitTimeout)).
		WaitForService("k1",
			wait.NewLogStrategy("grpc server started").WithStartupTimeout(DefaultContainerWaitTimeout)).
		WaitForService("k2",
			wait.NewLogStrategy("grpc server started").WithStartupTimeout(DefaultContainerWaitTimeout)).
		WaitForService("k3",
			wait.NewLogStrategy("grpc server started").WithStartupTimeout(DefaultContainerWaitTimeout)).
		Up(ctx)
	r.t.Log("docker compose up")

	r.teardown = append(r.teardown, func() {
		r.t.Log("teardown docker compose")
		dc.Down(ctx, compose.RemoveOrphans(true), compose.RemoveImagesLocal)
	})

	r.t.Cleanup(func() {
		r.Teardown()
	})

	require.NoError(r.t, err, "failed to start kwild cluster")

	serviceNames := dc.Services()
	r.t.Log("serviceNames", serviceNames)
	for _, name := range serviceNames {
		container, err := dc.ServiceContainer(ctx, name)
		require.NoError(r.t, err, "failed to get container for service %s", name)
		r.containers = append(r.containers, container)
	}

}

func (r *IntHelper) Setup(ctx context.Context) {
	r.generateNodeConfig()
	r.runDockerCompose(ctx)
}

func (r *IntHelper) Teardown() {
	r.t.Log("teardown test environment")
	for _, fn := range r.teardown {
		fn()
	}
}

func (r *IntHelper) getDriver(ctx context.Context, ctr *testcontainers.DockerContainer) KwilIntDriver {
	// NOTE: maybe get from docker-compose.yml ? the port mapping is already there
	nodeURL, err := ctr.PortEndpoint(ctx, "50051", "")
	require.NoError(r.t, err, "failed to get node url")
	gatewayURL, err := ctr.PortEndpoint(ctx, "8080", "")
	require.NoError(r.t, err, "failed to get gateway url")
	cometBftURL, err := ctr.PortEndpoint(ctx, "26657", "tcp")
	fmt.Println("cometBftURL: ", cometBftURL)
	require.NoError(r.t, err, "failed to get cometBft url")

	name, err := ctr.Name(ctx)
	require.NoError(r.t, err, "failed to get container name")

	r.t.Logf("nodeURL: %s gatewayURL: %s for container name: %s", nodeURL, gatewayURL, name)
	kwilClt, err := client.New(ctx, r.cfg.GrpcEndpoint,
		client.WithPrivateKey(r.cfg.AlicePK),
		client.WithCometBftUrl(r.cfg.ChainEndpoint),
	)
	require.NoError(r.t, err, "failed to create kwil client")

	return kwild.NewKwildDriver(kwilClt)
}

func (r *IntHelper) GetDrivers(ctx context.Context) []KwilIntDriver {
	drivers := make([]KwilIntDriver, 0, len(r.containers))

	for _, ctr := range r.containers {
		drivers = append(drivers, r.getDriver(ctx, ctr))
	}
	return drivers
}
