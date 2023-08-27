// package integration is a package for integration tests.
// For now this package has a lot duplicated code from acceptance package because they share similar but same setup,
// this will be refactored in the future.
//
// This package also deliberately use different environment variables for configuration from acceptance package.

package integration

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/kwilteam/kwil-db/pkg/log"

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

// envFile is the default env file path
// It will pass values among different stages of the test setup
var envFile = runner.GetEnv("KINT_ENV_FILE", "./.env")

type IntTestConfig struct {
	acceptance.ActTestCfg

	NValidator int
}

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

// LoadConfig loads config from system env and .env file.
// Envs defined in envFile will not overwrite existing env vars.
func (r *IntHelper) LoadConfig() {
	ef, err := os.OpenFile(envFile, os.O_RDWR|os.O_CREATE, 0666)
	require.NoError(r.t, err, "failed to open env file")
	defer ef.Close()

	err = godotenv.Load(envFile)
	require.NoError(r.t, err, "failed to parse env file")

	// default wallet mnemonic: test test test test test test test test test test test junk
	// default wallet hd path : m/44'/60'/0'
	cfg := &IntTestConfig{
		ActTestCfg: acceptance.ActTestCfg{
			AliceRawPK:        runner.GetEnv("KINT_ALICE_PK", "f1aa5a7966c3863ccde3047f6a1e266cdc0c76b399e256b8fede92b1c69e4f4e"),
			BobRawPK:          runner.GetEnv("KINT_BOB_PK", "43f149de89d64bf9a9099be19e1b1f7a4db784af8fa07caf6f08dc86ba65636b"),
			SchemaFile:        runner.GetEnv("KINT_SCHEMA", "./test-data/test_db.kf"),
			LogLevel:          runner.GetEnv("KINT_LOG_LEVEL", "debug"),
			GWEndpoint:        runner.GetEnv("KINT_GATEWAY_ENDPOINT", "localhost:8080"),
			GrpcEndpoint:      runner.GetEnv("KINT_GRPC_ENDPOINT", "localhost:50051"),
			DockerComposeFile: runner.GetEnv("KINT_DOCKER_COMPOSE_FILE", "./docker-compose.yml"),
		},
	}

	waitTimeout := runner.GetEnv("KACT_WAIT_TIMEOUT", "10s")
	cfg.WaitTimeout, err = time.ParseDuration(waitTimeout)
	require.NoError(r.t, err, "invalid wait timeout")

	nodeNum := runner.GetEnv("KINT_VALIDATOR_NUM", "3")
	cfg.NValidator, err = strconv.Atoi(nodeNum)
	require.NoError(r.t, err, "invalid node number")

	alicePk, err := crypto.Secp256k1PrivateKeyFromHex(cfg.AliceRawPK)
	require.NoError(r.t, err, "invalid alice private key")
	cfg.AlicePK = crypto.DefaultSigner(alicePk)

	bobPk, err := crypto.Secp256k1PrivateKeyFromHex(cfg.BobRawPK)
	require.NoError(r.t, err, "invalid bob private key")
	cfg.BobPk = crypto.DefaultSigner(bobPk)

	r.cfg = cfg
	cfg.DumpToEnv()
}

func (r *IntHelper) updateGeneratedConfigHome(home string) {
	envs, err := godotenv.Read(envFile)
	require.NoError(r.t, err, "failed to read env file")

	envs["KWIL_HOME"] = home

	err = godotenv.Write(envs, envFile)
	require.NoError(r.t, err, "failed to write env vars to file")
}

func (r *IntHelper) generateNodeConfig() {
	r.t.Logf("generate testnet config")
	tmpPath := r.t.TempDir()
	r.t.Logf("create test temp directory: %s", tmpPath)

	err := nodecfg.GenerateTestnetConfig(&nodecfg.TestnetGenerateConfig{
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

	r.updateGeneratedConfigHome(tmpPath)
}

func (r *IntHelper) runDockerCompose(ctx context.Context) {
	r.t.Logf("run in docker compose")

	//setSchemaLoader(r.cfg.AliceAddr())

	envs, err := godotenv.Read(envFile)
	require.NoError(r.t, err, "failed to parse .env file")

	dc, err := compose.NewDockerCompose(r.cfg.DockerComposeFile)
	require.NoError(r.t, err, "failed to create docker compose object for kwild cluster")

	r.teardown = append(r.teardown, func() {
		r.t.Log("teardown docker compose")
		dc.Down(ctx, compose.RemoveOrphans(true), compose.RemoveImagesLocal)
	})

	r.t.Cleanup(func() {
		r.Teardown()
	})

	err = dc.
		WithEnv(envs).
		WaitForService("ext1",
			wait.NewLogStrategy("listening on").WithStartupTimeout(r.cfg.WaitTimeout)).
		WaitForService("k1",
			wait.NewLogStrategy("grpc server started").WithStartupTimeout(r.cfg.WaitTimeout)).
		WaitForService("k2",
			wait.NewLogStrategy("grpc server started").WithStartupTimeout(r.cfg.WaitTimeout)).
		WaitForService("k3",
			wait.NewLogStrategy("grpc server started").WithStartupTimeout(r.cfg.WaitTimeout)).
		Up(ctx)
	r.t.Log("docker compose up")

	require.NoError(r.t, err, "failed to start kwild cluster")

	serviceNames := dc.Services()
	r.t.Log("serviceNames", serviceNames)
	for _, name := range serviceNames {
		// skip ext1
		if name == "ext1" {
			continue
		}
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

	logger := log.New(log.Config{Level: r.cfg.LogLevel})
	kwilClt, err := client.New(r.cfg.GrpcEndpoint,
		client.WithSigner(r.cfg.AlicePK),
		client.WithLogger(logger),
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
