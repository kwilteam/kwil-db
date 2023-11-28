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
	"path"
	"path/filepath"
	"runtime"
	"sort"
	"syscall"
	"testing"
	"time"

	"github.com/cometbft/cometbft/crypto/ed25519"
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/compose"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/kwilteam/kwil-db/cmd/kwil-admin/nodecfg"
	bClient "github.com/kwilteam/kwil-db/core/bridge/client"
	"github.com/kwilteam/kwil-db/core/client"
	"github.com/kwilteam/kwil-db/core/crypto"
	"github.com/kwilteam/kwil-db/core/crypto/auth"
	"github.com/kwilteam/kwil-db/core/log"
	"github.com/kwilteam/kwil-db/core/types/chain"
	deployer "github.com/kwilteam/kwil-db/test/deployer"
	ethDeployer "github.com/kwilteam/kwil-db/test/deployer/evm"
	"github.com/kwilteam/kwil-db/test/driver"

	ec "github.com/ethereum/go-ethereum/crypto"
)

var (
	getEnv = driver.GetEnv

	// envFile is the default env file path
	// it will pass values among different stages of the test setup
	envFile = getEnv("KIT_ENV_FILE", "./.env")
)

var defaultWaitStrategies = map[string]string{
	ExtContainer:  "listening on",
	Ext3Container: "listening on",
	"node0":       "Starting Node service",
	"node1":       "Starting Node service",
	"node2":       "Starting Node service",
	"node3":       "Starting Node service",
}

const (
	ExtContainer  = "ext1"
	Ext3Container = "ext3"
	testChainID   = "kwil-test-chain"
)

// IntTestConfig is the config for integration test
// This is totally separate from acceptance test
type IntTestConfig struct {
	GWEndpoint      string // gateway endpoint
	GrpcEndpoint    string
	ChainEndpoint   string
	ExposedChainRPC string

	SchemaFile                string
	DockerComposeFile         string
	DockerComposeOverrideFile string
	GanacheComposeFile        string

	WaitTimeout time.Duration
	LogLevel    string

	CreatorRawPk  string
	VisitorRawPK  string
	CreatorSigner auth.Signer
	VisitorSigner auth.Signer

	BlockInterval time.Duration // timeout_commit i.e. minimum block interval

	NValidator    int
	NNonValidator int
	JoinExpiry    int64

	EscrowAddress    string
	NumConfirmations int64
	WithoutGasCosts  bool
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
			JoinExpiry:       14400,
			NumConfirmations: 12,
			WithoutGasCosts:  true,
		},
	}

	helper.LoadConfig()

	for _, opt := range opts {
		opt(helper)
	}

	return helper
}

type HelperOpt func(*IntHelper)

func WithBlockInterval(d time.Duration) HelperOpt {
	return func(r *IntHelper) {
		r.cfg.BlockInterval = d
	}
}

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

func WithJoinExpiry(expiry int64) HelperOpt {
	return func(r *IntHelper) {
		r.cfg.JoinExpiry = expiry
	}
}

func WithNumConfirmations(numConfirmations int64) HelperOpt {
	return func(r *IntHelper) {
		r.cfg.NumConfirmations = numConfirmations
	}
}

func WithoutGasCosts(withoutGas bool) HelperOpt {
	return func(r *IntHelper) {
		r.cfg.WithoutGasCosts = withoutGas
	}
}

func (r *IntHelper) EscrowAddress() string {
	return r.cfg.EscrowAddress
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

	r.cfg = &IntTestConfig{
		CreatorRawPk:              getEnv("KIT_CREATOR_PK", "f1aa5a7966c3863ccde3047f6a1e266cdc0c76b399e256b8fede92b1c69e4f4e"),
		VisitorRawPK:              getEnv("KIT_VISITOR_PK", "43f149de89d64bf9a9099be19e1b1f7a4db784af8fa07caf6f08dc86ba65636b"),
		SchemaFile:                getEnv("KIT_SCHEMA", "./test-data/test_db.kf"),
		LogLevel:                  getEnv("KIT_LOG_LEVEL", "info"),
		GWEndpoint:                getEnv("KIT_GATEWAY_ENDPOINT", "localhost:8080"),
		GrpcEndpoint:              getEnv("KIT_GRPC_ENDPOINT", "localhost:50051"),
		DockerComposeFile:         getEnv("KIT_DOCKER_COMPOSE_FILE", "./docker-compose.yml"),
		DockerComposeOverrideFile: getEnv("KIT_DOCKER_COMPOSE_OVERRIDE_FILE", "./docker-compose.override.yml"),
		GanacheComposeFile:        getEnv("KIT_GANACHE_COMPOSE_FILE", "./ganache/docker-compose.yml"),
	}

	waitTimeout := getEnv("KIT_WAIT_TIMEOUT", "10s")
	r.cfg.WaitTimeout, err = time.ParseDuration(waitTimeout)
	require.NoError(r.t, err, "invalid wait timeout")

	creatorPk, err := crypto.Secp256k1PrivateKeyFromHex(r.cfg.CreatorRawPk)
	require.NoError(r.t, err, "invalid creator private key")

	r.cfg.CreatorSigner = &auth.EthPersonalSigner{Key: *creatorPk}

	bobPk, err := crypto.Secp256k1PrivateKeyFromHex(r.cfg.VisitorRawPK)
	require.NoError(r.t, err, "invalid visitor private key")
	r.cfg.VisitorSigner = &auth.EthPersonalSigner{Key: *bobPk}
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
	// To prevent go test from cleaning up (TODO: consider making this an option):
	// tmpPath, err := os.MkdirTemp("", "TestKwilInt")
	// if err != nil {
	// 	r.t.Fatal(err)
	// }
	r.t.Logf("create test temp directory: %s", tmpPath)

	err := nodecfg.GenerateTestnetConfig(&nodecfg.TestnetGenerateConfig{
		ChainID:       testChainID,
		BlockInterval: r.cfg.BlockInterval,
		// InitialHeight:           0,
		NValidators:             r.cfg.NValidator,
		NNonValidators:          r.cfg.NNonValidator,
		ConfigFile:              "",
		OutputDir:               tmpPath,
		NodeDirPrefix:           "node",
		PopulatePersistentPeers: true,
		HostnamePrefix:          "kwil-",
		HostnameSuffix:          "",
		StartingIPAddress:       "172.10.100.2",
		P2pPort:                 26656,
		JoinExpiry:              r.cfg.JoinExpiry,
		WithoutGasCosts:         r.cfg.WithoutGasCosts,
		WithoutNonces:           false,
		RequiredConfirmations:   r.cfg.NumConfirmations,
		ChainRPCURL:             r.cfg.ChainEndpoint,
		EscrowAddress:           r.cfg.EscrowAddress,
	}, nil)
	require.NoError(r.t, err, "failed to generate testnet config")
	r.home = tmpPath
	r.ExtractPrivateKeys()
	r.updateGeneratedConfigHome(tmpPath)
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

func (r *IntHelper) RunGanache(ctx context.Context) {
	r.t.Logf("run ganache")
	composeFiles := []string{r.cfg.GanacheComposeFile}
	dc, err := compose.NewDockerCompose(composeFiles...)
	require.NoError(r.t, err, "failed to create docker compose object for ganache")

	r.teardown = append(r.teardown, func() {
		r.t.Log("teardown ganache")
		dc.Down(ctx, compose.RemoveOrphans(true), compose.RemoveImagesLocal)
	})

	r.t.Cleanup(func() { // redundant if test defers Teardown()
		r.Teardown()
	})

	// Use compose.Wait to wait for containers to become "healthy" according to
	// their defined healthchecks.
	dockerComposeId := fmt.Sprintf("%d", time.Now().Unix())
	stack := dc.WithEnv(map[string]string{
		"uid": dockerComposeId,
	})
	stack = stack.WaitForService("ganache", wait.NewLogStrategy("RPC Listening on 0.0.0.0:8545").WithStartupTimeout(r.cfg.WaitTimeout))

	err = stack.Up(ctx, compose.Wait(true))
	r.t.Log("ganache up")
	require.NoError(r.t, err, "failed to start ganache")

	// Get the Escrow address and the ChainRPCURL
	ctr, err := dc.ServiceContainer(ctx, "ganache")
	require.NoError(r.t, err, "failed to get container for service ganache")

	exposedChainRPC, err := ctr.PortEndpoint(ctx, "8545", "ws")
	r.t.Log("exposedChainRPC", exposedChainRPC)
	require.NoError(r.t, err, "failed to get exposed endpoint")
	ganacheIp, err := ctr.ContainerIP(ctx)
	require.NoError(r.t, err, "failed to get ganache container ip")
	unexposedChainRPC := fmt.Sprintf("ws://%s:8545", ganacheIp)
	r.t.Log("unexposedChainRPC", unexposedChainRPC)

	// Deploy contracts
	// chainDeployer := GetDeployer("eth", exposedChainRPC, cfg.DatabaseDeployerPrivateKeyString, cfg.denomination)
	chainDeployer, err := GetDeployer(ctx, "eth", exposedChainRPC, r.cfg.CreatorRawPk)
	require.NoError(r.t, err, "failed to get deployer")

	tokenAddress, err := chainDeployer.DeployToken(ctx)
	require.NoError(r.t, err, "failed to deploy token")
	r.t.Log("Token address: ", tokenAddress)

	escrowAddress, err := chainDeployer.DeployEscrow(ctx, tokenAddress)
	require.NoError(r.t, err, "failed to deploy contract")
	r.t.Logf("Escrow address: %s\n", escrowAddress)

	r.cfg.ChainEndpoint = unexposedChainRPC
	// TODO: get escrow address from contract deployment
	r.cfg.EscrowAddress = escrowAddress
	r.cfg.ExposedChainRPC = exposedChainRPC
	// sleep

}

func (r *IntHelper) RunDockerComposeWithServices(ctx context.Context, services []string) {
	r.t.Logf("run in docker compose")
	time.Sleep(time.Second) // sometimes docker compose fails if previous test had some slow async clean up (no idea)

	envs, err := godotenv.Read(envFile)
	require.NoError(r.t, err, "failed to parse .env file")

	// Run Kwil cluster
	composeFiles := []string{r.cfg.DockerComposeFile}
	if r.cfg.DockerComposeOverrideFile != "" && fileExists(r.cfg.DockerComposeOverrideFile) {
		composeFiles = append(composeFiles, r.cfg.DockerComposeOverrideFile)
	}
	dc, err := compose.NewDockerCompose(composeFiles...)
	require.NoError(r.t, err, "failed to create docker compose object for kwild cluster")

	r.teardown = append(r.teardown, func() {
		r.t.Log("teardown docker compose")
		dc.Down(ctx, compose.RemoveOrphans(true), compose.RemoveImagesLocal)
	})

	r.t.Cleanup(func() { // redundant if test defers Teardown()
		r.Teardown()
	})

	stack := dc.WithEnv(envs)
	for _, service := range services {
		waitMsg := defaultWaitStrategies[service]
		stack = stack.WaitForService(service, wait.NewLogStrategy(waitMsg).WithStartupTimeout(r.cfg.WaitTimeout))
	}
	// Use compose.Wait to wait for containers to become "healthy" according to
	// their defined healthchecks.
	err = stack.Up(ctx, compose.Wait(true), compose.RunServices(services...))
	r.t.Log("docker compose up")
	require.NoError(r.t, err, "failed to start kwild cluster")

	for _, name := range services {
		// skip ext containers
		if name == ExtContainer || name == Ext3Container {
			continue
		}
		container, err := dc.ServiceContainer(ctx, name)
		require.NoError(r.t, err, "failed to get container for service %s", name)
		r.containers[name] = container
	}

}

func (r *IntHelper) Setup(ctx context.Context, services []string) {
	// Run Ganache container
	r.RunGanache(ctx)

	// generate config for testnet based on the ganache container
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
	regexPath := filepath.Join(r.home, "*", "private_key")

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

// GetUserDriver returns a integration driver connected to the given kwil node
// using the creator's private key
func (r *IntHelper) GetUserDriver(ctx context.Context, name string, driverType string) KwilIntDriver {
	ctr := r.containers[name]

	// NOTE: maybe get from docker-compose.yml ? the port mapping is already there
	nodeURL, err := ctr.PortEndpoint(ctx, "50051", "")
	require.NoError(r.t, err, "failed to get node url")
	gatewayURL, err := ctr.PortEndpoint(ctx, "8080", "")
	require.NoError(r.t, err, "failed to get gateway url")
	cometBftURL, err := ctr.PortEndpoint(ctx, "26657", "tcp")
	require.NoError(r.t, err, "failed to get cometBft url")
	r.t.Logf("nodeURL: %s gatewayURL: %s cometBftURL: %s for container name: %s", nodeURL, gatewayURL, cometBftURL, name)

	signer := r.cfg.CreatorSigner
	pk := r.cfg.CreatorRawPk
	switch driverType {
	case "client":
		return r.getClientDriver(signer, pk)
	case "cli":
		return r.getCliDriver(pk, signer.PublicKey())
	default:
		panic("unsupported driver type")
	}
}

// GetOperatorDriver returns a integration driver connected to the given kwil node,
// using the private key of the operator
func (r *IntHelper) GetOperatorDriver(ctx context.Context, name string, driverType string) KwilIntDriver {
	ctr := r.containers[name]

	rpcURL, err := ctr.PortEndpoint(ctx, "50051", "")
	require.NoError(r.t, err, "failed to get node url")
	gatewayURL, err := ctr.PortEndpoint(ctx, "8080", "")
	require.NoError(r.t, err, "failed to get gateway url")
	p2pURL, err := ctr.PortEndpoint(ctx, "26656", "tcp")
	require.NoError(r.t, err, "failed to get p2p url")
	cometBftURL, err := ctr.PortEndpoint(ctx, "26657", "tcp")
	require.NoError(r.t, err, "failed to get cometBFT RPC url")

	r.t.Logf(`user RPC URL: "%s"
gateway URL: "%s"
p2p URL: "%s"
cometBFT URL: "%s"
container name: "%s"`,
		rpcURL, gatewayURL, cometBftURL, p2pURL, name)

	privKeyB := r.privateKeys[name].Bytes()
	privKeyHex := hex.EncodeToString(privKeyB)
	privKey, err := crypto.Ed25519PrivateKeyFromBytes(privKeyB)
	require.NoError(r.t, err, "invalid private key")
	signer := &auth.Ed25519Signer{Ed25519PrivateKey: *privKey}

	pk := privKeyHex
	switch driverType {
	case "client":
		return r.getClientDriver(signer, pk)
	case "cli":
		return r.getCliDriver(pk, signer.PubKey().Bytes())
	default:
		panic("unsupported driver type")
	}
}

func (r *IntHelper) getClientDriver(signer auth.Signer, pKey string) KwilIntDriver {
	logger := log.New(log.Config{Level: r.cfg.LogLevel})
	ctx := context.Background()

	options := []client.Option{client.WithSigner(signer, testChainID),
		client.WithLogger(logger),
		client.WithTLSCert("")} // TODO: handle cert
	kwilClt, err := client.Dial(ctx, r.cfg.GrpcEndpoint, options...)
	require.NoError(r.t, err, "failed to create kwil client")

	if !r.cfg.WithoutGasCosts {
		bridgeClt, err := bClient.New(ctx, r.cfg.ExposedChainRPC, chain.GOERLI, r.cfg.EscrowAddress)
		require.NoError(r.t, err, "failed to create bridge client")

		privKey, err := ec.HexToECDSA(pKey)
		require.NoError(r.t, err, "invalid private key")

		return driver.NewKwildClientDriver(kwilClt, bridgeClt, privKey, driver.WithLogger(logger))
	}

	return driver.NewKwildClientDriver(kwilClt, nil, nil, driver.WithLogger(logger))
}

func (r *IntHelper) getCliDriver(privKey string, pubKey []byte) KwilIntDriver {
	logger := log.New(log.Config{Level: r.cfg.LogLevel})

	_, currentFilePath, _, _ := runtime.Caller(1)
	cliBinPath := path.Join(path.Dir(currentFilePath),
		fmt.Sprintf("../../.build/kwil-cli-%s-%s", runtime.GOOS, runtime.GOARCH))
	adminBinPath := path.Join(path.Dir(currentFilePath),
		fmt.Sprintf("../../.build/kwil-admin-%s-%s", runtime.GOOS, runtime.GOARCH))

	return driver.NewKwilCliDriver(cliBinPath, adminBinPath, r.cfg.GrpcEndpoint, privKey, pubKey, logger)
}

func (r *IntHelper) NodePrivateKey(name string) ed25519.PrivKey {
	return r.privateKeys[name]
}

func (r *IntHelper) ServiceContainer(name string) *testcontainers.DockerContainer {
	return r.containers[name]
}

func GetDeployer(ctx context.Context, deployerType string, rpcURL string, privateKeyStr string) (deployer.Deployer, error) {
	switch deployerType {
	case "eth":
		return ethDeployer.New(ctx, rpcURL, privateKeyStr)
	default:
		panic("unknown deployer type")
	}
}
