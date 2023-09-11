// package integration is a package for integration tests.
// For now this package has a lot duplicated code from acceptance package because they share similar but same setup,
// this will be refactored in the future.
//
// This package also deliberately use different environment variables for configuration from acceptance package.

package integration

import (
	"context"
	"encoding/hex"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"sort"
	"syscall"
	"testing"
	"time"

	"github.com/kwilteam/kwil-db/pkg/log"

	"github.com/kwilteam/kwil-db/internal/app/kwild"
	"github.com/kwilteam/kwil-db/internal/pkg/nodecfg"
	"github.com/kwilteam/kwil-db/pkg/client"
	"github.com/kwilteam/kwil-db/pkg/crypto"
	"github.com/kwilteam/kwil-db/test/acceptance"
	"github.com/kwilteam/kwil-db/test/runner"

	"github.com/cometbft/cometbft/crypto/ed25519"

	"github.com/joho/godotenv"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/compose"
	"github.com/testcontainers/testcontainers-go/wait"
)

// envFile is the default env file path
// It will pass values among different stages of the test setup
var envFile = runner.GetEnv("KIT_ENV_FILE", "./.env")

var defaultWaitStrategies = map[string]string{
	"ext1":  "listening on",
	"node0": "Starting Node service",
	"node1": "Starting Node service",
	"node2": "Starting Node service",
	"node3": "Starting Node service",
}

const ExtContainer = "ext1"

type IntTestConfig struct {
	acceptance.ActTestCfg

	NValidator    int
	NNonValidator int
}

type IntHelper struct {
	t           *testing.T
	cfg         *IntTestConfig
	home        string
	teardown    []func()
	containers  map[string]*testcontainers.DockerContainer
	privateKeys map[string]ed25519.PrivKey
}

func NewIntHelper(t *testing.T, opts ...HelperOpt) *IntHelper {
	helper := &IntHelper{
		t:           t,
		privateKeys: make(map[string]ed25519.PrivKey),
		containers:  make(map[string]*testcontainers.DockerContainer),
		cfg: &IntTestConfig{
			ActTestCfg: acceptance.ActTestCfg{},
		},
	}

	for _, opt := range opts {
		opt(helper)
	}

	return helper
}

type HelperOpt func(*IntHelper)

func WithValidators(n int) HelperOpt {
	return func(r *IntHelper) {
		r.cfg.NValidator = n
	}
}

func WithNonValidators(n int) HelperOpt {
	return func(r *IntHelper) {
		r.cfg.NNonValidator = n
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
	r.cfg.ActTestCfg = acceptance.ActTestCfg{
		AliceRawPK:        runner.GetEnv("KIT_ALICE_PK", "f1aa5a7966c3863ccde3047f6a1e266cdc0c76b399e256b8fede92b1c69e4f4e"),
		BobRawPK:          runner.GetEnv("KIT_BOB_PK", "43f149de89d64bf9a9099be19e1b1f7a4db784af8fa07caf6f08dc86ba65636b"),
		SchemaFile:        runner.GetEnv("KIT_SCHEMA", "./test-data/test_db.kf"),
		LogLevel:          runner.GetEnv("KIT_LOG_LEVEL", "debug"),
		GWEndpoint:        runner.GetEnv("KIT_GATEWAY_ENDPOINT", "localhost:8080"),
		GrpcEndpoint:      runner.GetEnv("KIT_GRPC_ENDPOINT", "localhost:50051"),
		DockerComposeFile: runner.GetEnv("KIT_DOCKER_COMPOSE_FILE", "./docker-compose.yml"),
	}

	waitTimeout := runner.GetEnv("KIT_WAIT_TIMEOUT", "10s")
	r.cfg.WaitTimeout, err = time.ParseDuration(waitTimeout)
	require.NoError(r.t, err, "invalid wait timeout")

	alicePk, err := crypto.Secp256k1PrivateKeyFromHex(r.cfg.AliceRawPK)
	require.NoError(r.t, err, "invalid alice private key")
	r.cfg.AlicePK = crypto.DefaultSigner(alicePk)

	bobPk, err := crypto.Secp256k1PrivateKeyFromHex(r.cfg.BobRawPK)
	require.NoError(r.t, err, "invalid bob private key")
	r.cfg.BobPk = crypto.DefaultSigner(bobPk)

	r.cfg.DumpToEnv()
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
		NNonValidators:          r.cfg.NNonValidator,
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
	r.home = tmpPath
	r.ExtractPrivateKeys()
	r.updateGeneratedConfigHome(tmpPath)
}

func (r *IntHelper) RunDockerComposeWithServices(ctx context.Context, services []string) {
	r.t.Logf("run in docker compose")

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

	stack := dc.WithEnv(envs)
	for _, service := range services {
		waitMsg := defaultWaitStrategies[service]
		stack = stack.WaitForService(service, wait.NewLogStrategy(waitMsg).WithStartupTimeout(r.cfg.WaitTimeout))
	}
	err = stack.Up(ctx, compose.RunServices(services...))
	r.t.Log("docker compose up")
	require.NoError(r.t, err, "failed to start kwild cluster")

	for _, name := range services {
		// skip ext1
		if name == ExtContainer {
			continue
		}
		container, err := dc.ServiceContainer(ctx, name)
		require.NoError(r.t, err, "failed to get container for service %s", name)
		r.containers[name] = container
	}

}

func (r *IntHelper) Setup(ctx context.Context, services []string) {
	r.generateNodeConfig()
	r.RunDockerComposeWithServices(ctx, services)
}

func (r *IntHelper) Teardown() {
	r.t.Log("teardown test environment")
	for _, fn := range r.teardown {
		fn()
	}
}

func (r *IntHelper) WaitForSignals(t *testing.T) {
	done := make(chan os.Signal, 1)
	signal.Notify(done, syscall.SIGINT, syscall.SIGTERM)

	// block waiting for a signal
	s := <-done
	t.Logf("Got signal: %v\n", s)
	r.Teardown()
	t.Logf("Teardown done\n")
}

func (r *IntHelper) ExtractPrivateKeys() {
	regexPath := filepath.Join(r.home, "*/private_key.txt")

	files, err := filepath.Glob(regexPath)
	require.NoError(r.t, err, "failed to get private key files")

	sort.Strings(files)

	for idx, file := range files {
		name := fmt.Sprintf("node%d", idx)
		pkeyBytes, err := os.ReadFile(file)
		require.NoError(r.t, err, "failed to read private key file")

		pkey, err := decodePrivateKey(string(pkeyBytes))
		require.NoError(r.t, err, "failed to decode private key")

		r.privateKeys[name] = pkey
	}
}

func decodePrivateKey(pkey string) (ed25519.PrivKey, error) {
	privB, err := hex.DecodeString(pkey)
	if err != nil {
		return nil, fmt.Errorf("error decoding private key: %v", err)
	}
	return ed25519.PrivKey(privB), nil
}

func (r *IntHelper) GetDriver(ctx context.Context, ctr *testcontainers.DockerContainer) KwilIntDriver {
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
		drivers = append(drivers, r.GetDriver(ctx, ctr))
	}
	return drivers
}

// name: containerName
// Creates a kwildriver for a specific node. Used for node initiated requests
func (r *IntHelper) GetNodeDriver(ctx context.Context, name string) KwilIntDriver {
	ctr := r.containers[name]

	nodeURL, err := ctr.PortEndpoint(ctx, "50051", "")
	require.NoError(r.t, err, "failed to get node url")
	gatewayURL, err := ctr.PortEndpoint(ctx, "8080", "")
	require.NoError(r.t, err, "failed to get gateway url")
	cometBftURL, err := ctr.PortEndpoint(ctx, "26657", "tcp")
	require.NoError(r.t, err, "failed to get cometBft url")

	r.t.Logf("nodeURL: %s gatewayURL: %s cometBftURL: %s for container name: %s", nodeURL, gatewayURL, cometBftURL, name)

	privKeyB := r.privateKeys[name].Bytes()
	privKey, err := crypto.Ed25519PrivateKeyFromBytes(privKeyB)
	require.NoError(r.t, err, "invalid private key")

	signer := crypto.DefaultSigner(privKey)
	logger := log.New(log.Config{Level: r.cfg.LogLevel})

	options := []client.ClientOpt{client.WithSigner(signer), client.WithLogger(logger)}

	kwilClt, err := client.New(r.cfg.GrpcEndpoint, options...)
	require.NoError(r.t, err, "failed to create kwil client")

	return kwild.NewKwildDriver(kwilClt)
}

func (r *IntHelper) NodePrivateKey(name string) ed25519.PrivKey {
	return r.privateKeys[name]
}

func (r *IntHelper) ServiceContainer(name string) *testcontainers.DockerContainer {
	return r.containers[name]
}
