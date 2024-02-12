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
	"io"
	"math/big"
	"net"
	"os"
	"os/signal"
	"path"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/cometbft/cometbft/crypto/ed25519"
	"github.com/joho/godotenv"
	"github.com/kwilteam/kwil-db/cmd/kwil-admin/nodecfg"
	"github.com/kwilteam/kwil-db/core/adminclient"
	"github.com/kwilteam/kwil-db/core/client"
	"github.com/kwilteam/kwil-db/core/crypto"
	"github.com/kwilteam/kwil-db/core/crypto/auth"
	"github.com/kwilteam/kwil-db/core/gatewayclient"
	"github.com/kwilteam/kwil-db/core/log"
	gRPC "github.com/kwilteam/kwil-db/core/rpc/client/user/grpc"
	clientType "github.com/kwilteam/kwil-db/core/types/client"
	"github.com/kwilteam/kwil-db/test/driver"
	"github.com/kwilteam/kwil-db/test/driver/operator"
	ethdeployer "github.com/kwilteam/kwil-db/test/integration/eth-deployer"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/compose"
	"github.com/testcontainers/testcontainers-go/wait"
)

var (
	getEnv = driver.GetEnv

	// envFile is the default env file path
	// it will pass values among different stages of the test setup
	envFile = getEnv("KIT_ENV_FILE", "./.env")
)

var logWaitStrategies = map[string]string{
	ExtContainer:  "listening on",
	Ext3Container: "listening on",
	"node0":       "Starting Node service",
	"node1":       "Starting Node service",
	"node2":       "Starting Node service",
	"node3":       "Starting Node service",
	"kgw":         "KGW Server started",
}

var healthWaitStrategies = map[string]bool{
	"pg0": true,
	"pg1": true,
	"pg2": true,
	"pg3": true,
}

const (
	ExtContainer  = "ext1"
	Ext3Container = "ext3"
	testChainID   = "kwil-test-chain"
)

// IntTestConfig is the config for integration test
// This is totally separate from acceptance test
type IntTestConfig struct {
	HTTPEndpoint  string
	GrpcEndpoint  string
	ChainEndpoint string
	AdminRPC      string // either tcp or unix.  Should be of form unix://var/run/kwil/admin.sock or tcp://localhost:26657

	SchemaFile                string
	DockerComposeFile         string
	DockerComposeOverrideFile string
	GanacheComposeFile        string
	WithGanache               bool

	WaitTimeout time.Duration
	LogLevel    string

	CreatorRawPk  string
	VisitorRawPK  string
	CreatorSigner auth.Signer
	VisitorSigner auth.Signer

	BlockInterval time.Duration // timeout_commit i.e. minimum block interval

	Allocs map[string]*big.Int

	NValidator    int
	NNonValidator int
	JoinExpiry    int64
	VoteExpiry    int64
	WithGas       bool
}

type IntHelper struct {
	t           *testing.T
	cfg         *IntTestConfig
	home        string
	teardown    []func()
	containers  map[string]*testcontainers.DockerContainer
	privateKeys map[string]ed25519.PrivKey

	// Oracles
	ethDeposit EthDepositOracle
}

type EthDepositOracle struct {
	Enabled           bool
	UnexposedChainRPC string
	ExposedChainRPC   string
	Deployer          *ethdeployer.Deployer
	EscrowAddress     string

	confirmations string

	byzantine_expiry string
	byzantine_spam   string
}

func NewIntHelper(t *testing.T, opts ...HelperOpt) *IntHelper {
	helper := &IntHelper{
		t:           t,
		privateKeys: make(map[string]ed25519.PrivKey),
		containers:  make(map[string]*testcontainers.DockerContainer),
		cfg: &IntTestConfig{
			JoinExpiry: 14400,
			VoteExpiry: 14400,
			Allocs:     make(map[string]*big.Int),
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

func WithGas() HelperOpt {
	return func(r *IntHelper) {
		r.cfg.WithGas = true
	}
}

func WithVoteExpiry(expiry int64) HelperOpt {
	return func(r *IntHelper) {
		r.cfg.VoteExpiry = expiry
	}
}

func WithEthDepositOracle(enabled bool) HelperOpt {
	return func(r *IntHelper) {
		r.ethDeposit.Enabled = enabled
	}
}

func WithConfirmations(n string) HelperOpt {
	return func(r *IntHelper) {
		r.ethDeposit.confirmations = n
	}
}

func WithGenesisAlloc(allocs map[string]*big.Int) HelperOpt {
	return func(r *IntHelper) {
		r.cfg.Allocs = allocs
	}
}

func WithByzantineExpiry() HelperOpt {
	return func(r *IntHelper) {
		r.ethDeposit.byzantine_expiry = "true"
	}
}

func WithByzantineSpam() HelperOpt {
	return func(r *IntHelper) {
		r.ethDeposit.byzantine_spam = "true"
	}
}

func WithGanache() HelperOpt {
	return func(r *IntHelper) {
		r.cfg.WithGanache = true
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

	r.cfg = &IntTestConfig{
		CreatorRawPk:              getEnv("KIT_CREATOR_PK", "f1aa5a7966c3863ccde3047f6a1e266cdc0c76b399e256b8fede92b1c69e4f4e"),
		VisitorRawPK:              getEnv("KIT_VISITOR_PK", "43f149de89d64bf9a9099be19e1b1f7a4db784af8fa07caf6f08dc86ba65636b"),
		SchemaFile:                getEnv("KIT_SCHEMA", "./test-data/test_db.kf"),
		LogLevel:                  getEnv("KIT_LOG_LEVEL", "info"),
		HTTPEndpoint:              getEnv("KIT_HTTP_ENDPOINT", "http://localhost:8080"),
		GrpcEndpoint:              getEnv("KIT_GRPC_ENDPOINT", "localhost:50051"),
		AdminRPC:                  getEnv("KIT_ADMIN_RPC", "unix:///tmp/admin.sock"),
		DockerComposeFile:         getEnv("KIT_DOCKER_COMPOSE_FILE", "./docker-compose.yml"),
		DockerComposeOverrideFile: getEnv("KIT_DOCKER_COMPOSE_OVERRIDE_FILE", "./docker-compose.override.yml"),
		GanacheComposeFile:        getEnv("KIT_GANACHE_COMPOSE_FILE", "./ganache-docker-compose.yml"),
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
		WithoutGasCosts:         !r.cfg.WithGas,
		VoteExpiry:              r.cfg.VoteExpiry,
		WithoutNonces:           false,
		Allocs:                  r.cfg.Allocs,
		FundNonValidators:       r.cfg.WithGas, // when gas is required, also give the non-validators some for tests
		EthDeposits: nodecfg.EthDepositOracle{
			Enabled:               r.ethDeposit.Enabled,
			Endpoint:              r.ethDeposit.UnexposedChainRPC,
			RequiredConfirmations: r.ethDeposit.confirmations,
			EscrowAddress:         r.ethDeposit.EscrowAddress,
			ChainID:               "5",

			// ByzantineExpiry: r.ethDeposit.byzantine_expiry,
			// ByzantineSpam:   r.ethDeposit.byzantine_spam,
		},
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
	time.Sleep(time.Second) // sometimes docker compose fails if previous test had some slow async clean up (no idea)

	composeFiles := []string{r.cfg.GanacheComposeFile}
	dc, err := compose.NewDockerCompose(composeFiles...)
	require.NoError(r.t, err, "failed to create docker compose object for ganache")

	r.teardown = append(r.teardown, func() {
		r.t.Log("teardown ganache")
		dc.Down(ctx, compose.RemoveOrphans(true), compose.RemoveImagesLocal)
	})

	r.t.Cleanup(func() { // redundant if test defers Teardown()
		// NOTE: Cleanup functions will be called in last added, first called order.
		// but here we call Teardown(), will will call all the teardown fns, maybe
		// not the behavior we want.
		// this should call newly added teardown fns?
		r.Teardown()
	})

	// Use compose.Wait to wait for containers to become "healthy" according to
	// their defined healthchecks.
	dockerComposeId := fmt.Sprintf("%d", time.Now().Unix())
	stack := dc.WithEnv(map[string]string{
		"uid": dockerComposeId,
	})
	stack = stack.WaitForService("ganache",
		wait.NewLogStrategy("RPC Listening on 0.0.0.0:8545").WithStartupTimeout(r.cfg.WaitTimeout))

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
	unexposedChainRPC := fmt.Sprintf("ws://%s", net.JoinHostPort(ganacheIp, "8545"))
	r.t.Log("unexposedChainRPC", unexposedChainRPC)

	// Deploy contracts
	ethDeployer, err := ethdeployer.NewDeployer(exposedChainRPC, r.cfg.CreatorRawPk, 5)
	require.NoError(r.t, err, "failed to get deployer")

	// Deploy Token and Escrow contracts
	err = ethDeployer.Deploy()
	require.NoError(r.t, err, "failed to deploy contracts")

	r.ethDeposit.UnexposedChainRPC = unexposedChainRPC
	r.ethDeposit.ExposedChainRPC = exposedChainRPC
	r.ethDeposit.Deployer = ethDeployer
	r.ethDeposit.EscrowAddress = ethDeployer.EscrowAddress()

	fmt.Println("Endpoint info: ", r.ethDeposit.ExposedChainRPC, " \n Unexposed: ", r.ethDeposit.UnexposedChainRPC, "\n EscrowAddr: ", r.ethDeposit.EscrowAddress)

	time.Sleep(5 * time.Second) // wait for contracts to be deployed
	// Probably start mining here
}

func (r *IntHelper) EthDeployer() *ethdeployer.Deployer {
	return r.ethDeposit.Deployer
}

func (r *IntHelper) RunDockerComposeWithServices(ctx context.Context, services []string) {
	r.t.Logf("run in docker compose")
	time.Sleep(time.Second) // sometimes docker compose fails if previous test had some slow async clean up (no idea)

	envs, err := godotenv.Read(envFile)
	require.NoError(r.t, err, "failed to parse .env file")

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
		if r.ethDeposit.Enabled && strings.HasPrefix(service, "node") {
			waitMsg := "Started listening for new blocks on ethereum"
			stack = stack.WaitForService(service, wait.NewLogStrategy(waitMsg).WithStartupTimeout(r.cfg.WaitTimeout))
			continue
		}

		waitMsg, ok := logWaitStrategies[service]
		if ok {
			stack = stack.WaitForService(service, wait.NewLogStrategy(waitMsg).WithStartupTimeout(r.cfg.WaitTimeout))
			continue
		}

		if healthWaitStrategies[service] {
			stack = stack.WaitForService(service, wait.NewHealthStrategy().WithStartupTimeout(r.cfg.WaitTimeout))
			continue
		}

	}
	// Use compose.Wait to wait for containers to become "healthy" according to
	// their defined healthchecks.

	// NOTE: services will be sorted by docker-compose here.
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
	if r.cfg.WithGanache {
		r.RunGanache(ctx)
	}

	r.generateNodeConfig()
	r.RunDockerComposeWithServices(ctx, services)
}

func (r *IntHelper) Teardown() {
	// return // to not cleanup, which will break multiple tests and subtests
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

// GetUserGatewayDriver returns an integration driver connected to the given gateway node
func (r *IntHelper) GetUserGatewayDriver(ctx context.Context, driverType string, user string) KwilIntDriver {
	gatewayProvider := true

	ctr := r.containers["kgw"]
	gatewayURL, err := ctr.PortEndpoint(ctx, "8090", "http")
	require.NoError(r.t, err, "failed to get gateway url")
	r.t.Logf("gatewayURL: %s  for container: %s", gatewayURL, "kgw")
	// NOTE: gatewayURL should be http://localhost:8090, match the domain in docker-compose.yml

	signer := r.cfg.CreatorSigner
	pk := r.cfg.CreatorRawPk

	if user == "visitor" {
		signer = r.cfg.VisitorSigner
		pk = r.cfg.VisitorRawPK
	}

	switch driverType {
	case "http":
		return r.getHTTPClientDriver(signer, gatewayURL, gatewayProvider)
	case "cli":
		return r.getCliDriver(gatewayURL, pk, signer.Identity(), gatewayProvider)
	default:
		panic("unsupported driver type")
	}
}

// GetUserDriver returns an integration driver connected to the given rpc node
// using the private key
func (r *IntHelper) GetUserDriver(ctx context.Context, nodeName string, driverType string) KwilIntDriver {
	gatewayProvider := false

	ctr := r.containers[nodeName]
	// NOTE: maybe get from docker-compose.yml ? the port mapping is already there
	grpcURL, err := ctr.PortEndpoint(ctx, "50051", "tcp")
	require.NoError(r.t, err, "failed to get node url")
	httpURL, err := ctr.PortEndpoint(ctx, "8080", "http")
	require.NoError(r.t, err, "failed to get gateway url")
	cometBftURL, err := ctr.PortEndpoint(ctx, "26657", "tcp")
	require.NoError(r.t, err, "failed to get cometBft url")
	r.t.Logf("grpcURL: %s httpURL: %s cometBftURL: %s for container: %s", grpcURL, httpURL, cometBftURL, nodeName)

	signer := r.cfg.CreatorSigner
	pk := r.cfg.CreatorRawPk

	switch driverType {
	case "http":
		return r.getHTTPClientDriver(signer, httpURL, gatewayProvider)
	case "grpc":
		// should use grpcURL, r.cfg.GrpcEndpoint is not correct(it's intended
		// to be used for `remote` test mode, but integation tests don't use it)
		return r.getGRPCClientDriver(signer)
	case "cli":
		return r.getCliDriver(httpURL, pk, signer.Identity(), gatewayProvider)
	default:
		panic("unsupported driver type")
	}
}

// GetOperatorDriver returns a integration driver connected to the given kwil node,
// using the private key of the operator.
// The passed nodeName needs to be the same as the name of the container in docker-compose.yml
func (r *IntHelper) GetOperatorDriver(ctx context.Context, nodeName string, driverType string) operator.KwilOperatorDriver {
	switch driverType {
	case "http":
		r.t.Fatalf("http driver not supported for node operator")
		return nil
	case "grpc":
		c, ok := r.containers[nodeName]
		if !ok {
			r.t.Fatalf("container %s not found", nodeName)
		}

		adminGrpcUrl, err := c.PortEndpoint(ctx, "50151", "tcp")
		require.NoError(r.t, err, "failed to get admin grpc url")

		clt, err := adminclient.NewClient(ctx, adminGrpcUrl)
		if err != nil {
			r.t.Fatalf("failed to create admin client: %v", err)
		}

		return &operator.AdminClientDriver{
			Client: clt,
		}
	case "cli":
		c, ok := r.containers[nodeName]
		if !ok {
			r.t.Fatalf("container %s not found", nodeName)
		}

		return r.getCLIAdminClientDriver(r.cfg.AdminRPC, c)
	default:
		panic("unsupported driver type")
	}
}

func (r *IntHelper) getHTTPClientDriver(signer auth.Signer, endpoint string, gatewayProvider bool) *driver.KwildClientDriver {
	logger := log.New(log.Config{Level: r.cfg.LogLevel})

	var kwilClt clientType.Client
	var err error

	if gatewayProvider {
		kwilClt, err = gatewayclient.NewClient(context.TODO(), endpoint, &gatewayclient.GatewayOptions{
			Options: clientType.Options{
				Signer:  signer,
				ChainID: testChainID,
				Logger:  logger,
			},
		})
	} else {
		kwilClt, err = client.NewClient(context.TODO(), endpoint, &clientType.Options{
			Signer:  signer,
			ChainID: testChainID,
			Logger:  logger,
		})
	}

	require.NoError(r.t, err, "failed to create kwil client")

	return driver.NewKwildClientDriver(kwilClt, signer, r.ethDeposit.Deployer, logger)
}

func (r *IntHelper) getGRPCClientDriver(signer auth.Signer) *driver.KwildClientDriver {
	logger := log.New(log.Config{Level: r.cfg.LogLevel})

	gtOptions := []gRPC.Option{gRPC.WithTlsCert("")}
	gt, err := gRPC.New(context.Background(), r.cfg.GrpcEndpoint, gtOptions...)
	require.NoError(r.t, err, "failed to create grpc transport")

	kwilClt, err := client.WrapClient(context.TODO(), gt, &clientType.Options{
		Signer:  signer,
		ChainID: testChainID,
		Logger:  logger,
	})
	require.NoError(r.t, err, "failed to create grpc client")

	return driver.NewKwildClientDriver(kwilClt, signer, r.ethDeposit.Deployer, logger)
}

// getCLIAdminClientDriver returns a kwil-admin client driver connected to the given kwil node.
// the adminSvcServer should be either unix:// or tcp://
func (r *IntHelper) getCLIAdminClientDriver(adminSvcServer string, c *testcontainers.DockerContainer) operator.KwilOperatorDriver {
	return &operator.OperatorCLIDriver{
		Exec: func(ctx context.Context, args ...string) ([]byte, error) {
			_, reader, err := c.Exec(ctx, append([]string{"/app/kwil-admin"}, args...))
			if err != nil {
				return nil, err
			}
			bts, err := io.ReadAll(reader)
			if err != nil {
				return nil, err
			}

			// docker engine returns an 8 byte header as part of their response
			// https://docs.docker.com/engine/api/v1.43/#tag/Container/operation/ContainerAttach

			if len(bts) < 8 {
				return nil, fmt.Errorf("invalid response from docker engine")
			}

			return bts[8:], nil
		},
		RpcUrl: adminSvcServer,
	}
}

func (r *IntHelper) getCliDriver(endpoint string, privKey string, identity []byte, gatewayProvider bool) *driver.KwilCliDriver {
	logger := log.New(log.Config{Level: r.cfg.LogLevel})

	_, currentFilePath, _, _ := runtime.Caller(1)
	cliBinPath := path.Join(path.Dir(currentFilePath),
		"../../.build/kwil-cli")

	return driver.NewKwilCliDriver(cliBinPath, endpoint, privKey, testChainID, identity, gatewayProvider, r.ethDeposit.Deployer, logger)
}

func (r *IntHelper) NodePrivateKey(name string) ed25519.PrivKey {
	return r.privateKeys[name]
}

func (r *IntHelper) NodeKeys() map[string]ed25519.PrivKey {
	return r.privateKeys
}

func (r *IntHelper) ServiceContainer(name string) *testcontainers.DockerContainer {
	return r.containers[name]
}
