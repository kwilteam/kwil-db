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
	"net/url"
	"os"
	"os/signal"
	"path"
	"path/filepath"
	"runtime"
	"slices"
	"sort"
	"strconv"
	"syscall"
	"testing"
	"time"

	"github.com/kwilteam/kwil-db/cmd/kwil-admin/nodecfg"
	"github.com/kwilteam/kwil-db/core/adminclient"
	"github.com/kwilteam/kwil-db/core/client"
	"github.com/kwilteam/kwil-db/core/crypto"
	"github.com/kwilteam/kwil-db/core/crypto/auth"
	"github.com/kwilteam/kwil-db/core/gatewayclient"
	"github.com/kwilteam/kwil-db/core/log"
	http "github.com/kwilteam/kwil-db/core/rpc/client/user/http"
	clientType "github.com/kwilteam/kwil-db/core/types/client"
	ethdeposits "github.com/kwilteam/kwil-db/extensions/listeners/eth_deposits"
	"github.com/kwilteam/kwil-db/test/driver"
	"github.com/kwilteam/kwil-db/test/driver/operator"
	ethdeployer "github.com/kwilteam/kwil-db/test/integration/eth-deployer"
	"github.com/kwilteam/kwil-db/test/utils"

	"github.com/cometbft/cometbft/crypto/ed25519"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/compose"
	"github.com/testcontainers/testcontainers-go/wait"
)

var (
	getEnv = driver.GetEnv
)

var logWaitStrategies = map[string]string{
	ExtContainer:  "listening on",
	Ext3Container: "listening on",
	"node0":       "finalized block",
	"node1":       "finalized block",
	"node2":       "finalized block",
	"node3":       "finalized block",
	"node4":       "finalized block",
	"node5":       "finalized block",
	"kgw":         "KGW Server started",
	"ganache":     "RPC Listening on 0.0.0.0:8545",
	"pg0":         `listening on IPv4 address "0.0.0.0", port 5432`,
	"pg1":         `listening on IPv4 address "0.0.0.0", port 5432`,
	"pg2":         `listening on IPv4 address "0.0.0.0", port 5432`,
	"pg3":         `listening on IPv4 address "0.0.0.0", port 5432`,
	"pg4":         `listening on IPv4 address "0.0.0.0", port 5432`,
	"pg5":         `listening on IPv4 address "0.0.0.0", port 5432`,
}

const (
	ExtContainer  = "ext1"
	Ext3Container = "ext3"
	testChainID   = "kwil-test-chain"
)

// IntTestConfig is the config for integration test
// This is totally separate from acceptance test
type IntTestConfig struct {
	JSONRPCEndpoint string
	HTTPEndpoint    string
	ChainEndpoint   string
	AdminRPC        string // Should be of form /var/run/kwil/admin.sock or 127.0.0.1:8485

	SchemaFile                string
	DockerComposeFile         string
	DockerComposeOverrideFile string
	GanacheComposeFile        string
	WithGanache               bool
	ExposedHTTPPorts          bool

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

	SpamOracleEnabled bool

	Snapshots SnapshotConfig
}

type IntHelper struct {
	t           *testing.T
	cfg         *IntTestConfig
	containers  map[string]*testcontainers.DockerContainer
	privateKeys map[string]ed25519.PrivKey
	// envs is used to store dynamically generated envs later used in docker-compose
	// e.g. `dc.WithEnv(r.envs)`
	// for now two envs are used:
	// - KWIL_HOME: the home directory for the test
	// - KWIL_NETWORK: the network name for the test
	envs map[string]string

	// Extensions
	ethDeposit EthDepositOracle
}

type EthDepositOracle struct {
	Enabled           bool
	UnexposedChainRPC string
	ExposedChainRPC   string
	Deployer          *ethdeployer.Deployer
	ByzDeployer       *ethdeployer.Deployer
	EscrowAddress     string

	confirmations int64 // we always use 0, so not very useful at present

	NumByzantineExpiryNodes int
	ByzantineEscrowAddr     string
}

type SnapshotConfig struct {
	Enabled         bool
	MaxSnapshots    uint64
	RecurringHeight uint64
}

func NewIntHelper(t *testing.T, opts ...HelperOpt) *IntHelper {
	helper := &IntHelper{
		t:           t,
		privateKeys: make(map[string]ed25519.PrivKey),
		containers:  make(map[string]*testcontainers.DockerContainer),
		cfg: &IntTestConfig{
			Allocs: make(map[string]*big.Int),
			Snapshots: SnapshotConfig{
				Enabled:         false,
				MaxSnapshots:    3,
				RecurringHeight: 10,
			},
		},
		envs: make(map[string]string),
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

func WithExposedHTTPPorts() HelperOpt {
	return func(r *IntHelper) {
		r.cfg.ExposedHTTPPorts = true
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

// WithConfirmations overrides the default required confirmations (0).
func WithConfirmations(n int64) HelperOpt {
	return func(r *IntHelper) {
		r.ethDeposit.confirmations = n // note: only useful for for non-zero confs
	}
}

func WithGenesisAlloc(allocs map[string]*big.Int) HelperOpt {
	return func(r *IntHelper) {
		r.cfg.Allocs = allocs
	}
}

func WithNumByzantineExpiryNodes(n int) HelperOpt {
	return func(r *IntHelper) {
		r.ethDeposit.NumByzantineExpiryNodes = n
	}
}

func WithGanache() HelperOpt {
	return func(r *IntHelper) {
		r.cfg.WithGanache = true
	}
}

func WithSpamOracle() HelperOpt {
	return func(r *IntHelper) {
		r.cfg.SpamOracleEnabled = true
	}
}

func WithSnapshots() HelperOpt {
	return func(r *IntHelper) {
		r.cfg.Snapshots.Enabled = true
	}
}

func WithMaxSnapshots(num uint64) HelperOpt {
	return func(r *IntHelper) {
		r.cfg.Snapshots.MaxSnapshots = num
	}
}

func WithRecurringHeight(heights uint64) HelperOpt {
	return func(r *IntHelper) {
		r.cfg.Snapshots.RecurringHeight = heights
	}
}

// LoadConfig loads config from system env and .env file.
// Envs defined in envFile will not overwrite existing env vars.
func (r *IntHelper) LoadConfig() {
	var err error

	// default wallet mnemonic: test test test test test test test test test test test junk
	// default wallet hd path : m/44'/60'/0'
	r.cfg = &IntTestConfig{
		CreatorRawPk:              getEnv("KIT_CREATOR_PK", "f1aa5a7966c3863ccde3047f6a1e266cdc0c76b399e256b8fede92b1c69e4f4e"),
		VisitorRawPK:              getEnv("KIT_VISITOR_PK", "43f149de89d64bf9a9099be19e1b1f7a4db784af8fa07caf6f08dc86ba65636b"),
		SchemaFile:                getEnv("KIT_SCHEMA", "./test-data/test_db.kf"),
		LogLevel:                  getEnv("KIT_LOG_LEVEL", "info"),
		HTTPEndpoint:              getEnv("KIT_HTTP_ENDPOINT", "http://localhost:8080/"),
		JSONRPCEndpoint:           getEnv("KIT_JSONRPC_ENDPOINT", "http://localhost:8484"),
		AdminRPC:                  getEnv("KIT_ADMIN_RPC", "/tmp/admin.socket"),
		DockerComposeFile:         getEnv("KIT_DOCKER_COMPOSE_FILE", "./docker-compose.yml"),
		DockerComposeOverrideFile: getEnv("KIT_DOCKER_COMPOSE_OVERRIDE_FILE", "./docker-compose.override.yml"),
		GanacheComposeFile:        getEnv("KIT_GANACHE_COMPOSE_FILE", "./ganache-docker-compose.yml"),
	}

	waitTimeout := getEnv("KIT_WAIT_TIMEOUT", "20s")
	r.cfg.WaitTimeout, err = time.ParseDuration(waitTimeout)
	require.NoError(r.t, err, "invalid wait timeout")

	creatorPk, err := crypto.Secp256k1PrivateKeyFromHex(r.cfg.CreatorRawPk)
	require.NoError(r.t, err, "invalid creator private key")

	r.cfg.CreatorSigner = &auth.EthPersonalSigner{Key: *creatorPk}

	bobPk, err := crypto.Secp256k1PrivateKeyFromHex(r.cfg.VisitorRawPK)
	require.NoError(r.t, err, "invalid visitor private key")
	r.cfg.VisitorSigner = &auth.EthPersonalSigner{Key: *bobPk}

	// Overwritten using helperOpts
	r.cfg.VoteExpiry = 14400
	r.cfg.JoinExpiry = 14400
}

func (r *IntHelper) updateEnv(k, v string) {
	r.envs[k] = v
}

func (r *IntHelper) generateNodeConfig(homeDir string) {
	r.t.Logf("generate testnet config at %s", homeDir)

	extensionConfigs := make([]map[string]map[string]string, r.cfg.NValidator)
	for i := range extensionConfigs {
		extensionConfigs[i] = make(map[string]map[string]string)
		if r.ethDeposit.Enabled {
			address := r.ethDeposit.EscrowAddress
			if i < r.ethDeposit.NumByzantineExpiryNodes {
				address = r.ethDeposit.ByzantineEscrowAddr
			}

			cfg := ethdeposits.EthDepositConfig{
				RPCProvider:     r.ethDeposit.UnexposedChainRPC,
				ContractAddress: address,
				// setting values here since we cannot have the defaults, since
				// local ganache is a new network
				StartingHeight:        0,
				RequiredConfirmations: r.ethDeposit.confirmations, // TODO: remove this from the r.ethDeposit struct. it is not needed
				ReconnectionInterval:  30,
				MaxRetries:            20,
				BlockSyncChunkSize:    1000,
			}

			extensionConfigs[i][ethdeposits.ListenerName] = cfg.Map()
		}
		extensionConfigs[i]["spammer"] = map[string]string{
			"enabled": strconv.FormatBool(r.cfg.SpamOracleEnabled),
		}
	}

	var allocs map[string]*big.Int
	if r.cfg.SpamOracleEnabled {
		bal, ok := big.NewInt(0).SetString("100000000000000000000000000000000", 10)
		if !ok {
			r.t.Fatal("failed to parse balance")
		}
		creatorIdent := hex.EncodeToString(r.cfg.CreatorSigner.Identity())
		allocs = map[string]*big.Int{
			creatorIdent: bal,
		}
	}

	err := nodecfg.GenerateTestnetConfig(&nodecfg.TestnetGenerateConfig{
		ChainID:       testChainID,
		BlockInterval: r.cfg.BlockInterval,
		// InitialHeight:           0,
		NValidators:             r.cfg.NValidator,
		NNonValidators:          r.cfg.NNonValidator,
		ConfigFile:              "",
		OutputDir:               homeDir,
		NodeDirPrefix:           "node",
		PopulatePersistentPeers: true,
		HostnamePrefix:          "kwil-",
		HostnameSuffix:          "",
		// use this to ease the process running test parallel
		// NOTE: need to match docker-compose kwild service name
		DnsNamePrefix:     "node",
		P2pPort:           26656,
		JoinExpiry:        r.cfg.JoinExpiry,
		WithoutGasCosts:   !r.cfg.WithGas,
		VoteExpiry:        r.cfg.VoteExpiry,
		WithoutNonces:     false,
		Allocs:            allocs,
		FundNonValidators: r.cfg.WithGas, // when gas is required, also give the non-validators some for tests
		Extensions:        extensionConfigs,
		SnapshotsEnabled:  r.cfg.Snapshots.Enabled,
		MaxSnapshots:      r.cfg.Snapshots.MaxSnapshots,
		SnapshotHeights:   r.cfg.Snapshots.RecurringHeight,
	}, &nodecfg.ConfigOpts{
		DnsHost: true,
	})
	require.NoError(r.t, err, "failed to generate testnet config")
	r.ExtractPrivateKeys(homeDir)
	r.updateEnv("KWIL_HOME", homeDir)
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

func (r *IntHelper) RunGanache(ctx context.Context) {
	r.RunDockerComposeWithServices(ctx, []string{"ganache"})
	// Get the Escrow address and the ChainRPCURL
	ctr, ok := r.containers["ganache"]
	require.True(r.t, ok, "failed to get container for service ganache")

	exposedChainRPC, unexposedChainRPC, err := utils.GanacheWSEndpoints(ctr, ctx)
	require.NoError(r.t, err, "failed to get ganache endpoints")

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

	r.t.Logf("Endpoint info: %s \n\tUnexposed: %s \n\tEscrowAddr: %s",
		r.ethDeposit.ExposedChainRPC,
		r.ethDeposit.UnexposedChainRPC,
		r.ethDeposit.EscrowAddress)

	time.Sleep(5 * time.Second) // wait for contracts to be deployed
	// Probably start mining here
	if r.ethDeposit.NumByzantineExpiryNodes > 0 {
		// Deploy Byzantine contracts
		byzantineDeployer, err := ethdeployer.NewDeployer(exposedChainRPC, r.cfg.CreatorRawPk, 5)
		require.NoError(r.t, err, "failed to get deployer")
		r.ethDeposit.ByzDeployer = byzantineDeployer

		err = byzantineDeployer.Deploy()
		require.NoError(r.t, err, "failed to deploy contracts")

		r.ethDeposit.ByzantineEscrowAddr = byzantineDeployer.EscrowAddress()
	}
}

func (r *IntHelper) EthDeployer(byzMode bool) *ethdeployer.Deployer {
	if byzMode {
		return r.ethDeposit.ByzDeployer
	}
	return r.ethDeposit.Deployer
}

func (r *IntHelper) RunDockerComposeWithServices(ctx context.Context, services []string) {
	r.t.Logf("run in docker compose")
	time.Sleep(time.Second) // sometimes docker compose fails if previous test had some slow async clean up (no idea)

	composeFiles := []string{r.cfg.DockerComposeFile}
	if r.cfg.DockerComposeOverrideFile != "" && fileExists(r.cfg.DockerComposeOverrideFile) {
		composeFiles = append(composeFiles, r.cfg.DockerComposeOverrideFile)
	}
	r.t.Logf("use compose files: %v", composeFiles)
	dc, err := compose.NewDockerCompose(composeFiles...)
	require.NoError(r.t, err, "failed to create docker compose object for kwild cluster")

	ctxUp, cancel := context.WithCancel(ctx)

	r.t.Cleanup(func() {
		if r.t.Failed() {
			r.t.Logf("Stopping but keeping containers for inspection after failed test: %v", dc.Services())
			cancel() // Stop, not Down, which would remove the containers too --- this doesn't work, dang
			time.Sleep(5 * time.Second)

			// There is no dc.Stop, but there should be! Do this instead:
			svcs := dc.Services()
			slices.Sort(svcs)
			for _, svc := range svcs { // sort is silly, but I just want to stop nodes before pgs
				ct, err := dc.ServiceContainer(ctx, svc)
				if err != nil {
					r.t.Logf("could not get container %v: %v", svc, err)
					continue
				}
				err = ct.Stop(ctx, nil)
				if err != nil {
					r.t.Logf("could not stop container %v: %v", svc, err)
				}
			}
			return
		}
		r.t.Logf("teardown %s", dc.Services())
		err := dc.Down(ctx, compose.RemoveVolumes(true))
		require.NoErrorf(r.t, err, "failed to teardown %s", dc.Services())
		cancel() // no context leak
	})

	stack := dc.WithEnv(r.envs)
	for _, service := range services {
		waitMsg, ok := logWaitStrategies[service]
		if ok {
			stack = stack.WaitForService(service, wait.NewLogStrategy(waitMsg).WithStartupTimeout(r.cfg.WaitTimeout))
			continue
		}
	}
	// Use compose.Wait to wait for containers to become "healthy" according to
	// their defined healthchecks.

	// NOTE: services will be sorted by docker-compose here.
	err = stack.Up(ctxUp, compose.Wait(true), compose.RunServices(services...))
	r.t.Log("docker compose up")

	time.Sleep(3 * time.Second) // RPC errors with chain_info and other stuff... trying anything now

	require.NoError(r.t, err, "failed to start kwild cluster services %v", services)

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

// Setup sets up the test environment
// Following steps are done:
// 1. Create a temporary directory for current test
// 2. Prepare files for docker-compose to run
// 3. Run Ganache ahead if required(for the purpose to populate config for eth-deposit)
// 4. Generate node configuration files
// 5. Run docker-compose with the given services
func (r *IntHelper) Setup(ctx context.Context, services []string) {
	tmpDir, err := os.MkdirTemp("", "TestKwilInt")
	if err != nil {
		r.t.Fatal(err)
	}
	r.t.Cleanup(func() {
		if r.t.Failed() {
			r.t.Logf("Retaining data for failed test at path %v", tmpDir)
			return
		}
		os.RemoveAll(tmpDir)
	})

	r.t.Logf("create test directory: %s for %s", tmpDir, r.t.Name())

	r.prepareDockerCompose(ctx, tmpDir)

	if r.cfg.WithGanache {
		// NOTE: it's more natural and easier if able to configure oracle
		// through kwild cli flags
		r.RunGanache(ctx)
	}

	r.generateNodeConfig(tmpDir)

	r.RunDockerComposeWithServices(ctx, services)
}

// prepareDockerCompose prepares the docker-compose.yml file for the test.
// It does the following:
// 1. Create a new network for current test
// 2. Generate new docker-compose.yml using newly generated network
// 3. Copy pginit.sql to the same directory as docker-compose.yml
//
// NOTE:
// By default, the subnet pool assigned by docker is too big. Since we create
// a new network for each test, docker may complain not be able to create a new
// network. If ever this happens, a different setting `default-address-pools`
// for docker daemon should be used. For example, CI server is using the following
// setting in /etc/docker/daemon.json:
//
//	"default-address-pools": [
//	  {
//	    "base": "10.10.0.0/16",
//	    "size": 24
//	  }
//	]
//
// Another approach to make parallel tests work is using the same network for all tests,
// assuming the subnet pool is big enough for all containers at a time. It's still
// relevant to `default-address-pools` setting, so I'll leave it as is for now.
func (r *IntHelper) prepareDockerCompose(ctx context.Context, tmpDir string) {
	// create a new network for each test to avoid container DNS name conflicts
	// for parallel running
	testName := r.t.Name()
	localNetwork, err := utils.EnsureNetworkExist(context.Background(), testName)
	require.NoError(r.t, err, "failed to create network")
	localNetworkName := localNetwork.Name

	r.t.Cleanup(func() {
		if localNetwork != nil && !r.t.Failed() {
			r.t.Logf("teardown docker network %s from %s", localNetworkName, testName)
			err := localNetwork.Remove(ctx)
			require.NoError(r.t, err, "failed to remove network")
		}
	})

	// another seemingly possible way to do this is instead of using template
	// docker-compose file is to use envs in docker-compose.yml, but it doesn't work
	//r.updateEnv("KWIL_NETWORK", localNetworkName)

	var exposedHTTPPorts []int
	if r.cfg.ExposedHTTPPorts {
		// Actually only need this as long as the number of nodes defined in
		// the docker-compose.yml.template file. This is not related to
		// NValidators and NNValidators.
		for i := 0; i < 20; i++ { // more than enough, which is 6 presently
			exposedHTTPPorts = append(exposedHTTPPorts, i+8080+1)
		}
	}

	// here we generate a new docker-compose.yml file with the new network from template
	composeFile := filepath.Join(tmpDir, "docker-compose.yml")
	dockerImageName := utils.DefaultDockerImage
	if r.cfg.SpamOracleEnabled {
		dockerImageName = "kwild-spammer:latest"
	}
	err = utils.CreateComposeFile(composeFile, "./docker-compose.yml.template",
		utils.ComposeConfig{
			Network:          localNetworkName,
			ExposedHTTPPorts: exposedHTTPPorts,
			DockerImage:      dockerImageName,
		})
	require.NoError(r.t, err, "failed to create docker compose file")

	r.t.Logf("generated compose file: %s, network: %s, test: %s",
		composeFile, localNetworkName, testName)

	// copy pginit.sql to same directory as docker-compose.yml
	// so it can be mounted into the pg containers
	pgInitSQL, err := os.ReadFile("./pginit.sql")
	require.NoError(r.t, err, "failed to read pginit.sql")
	pgInitFile := filepath.Join(tmpDir, "pginit.sql")
	err = os.WriteFile(pgInitFile, pgInitSQL, 0644)
	require.NoError(r.t, err, "failed to write pginit.sql")

	// copy docker-compose.override.yml if exists
	if fileExists(r.cfg.DockerComposeOverrideFile) {
		overrideCompose, err := os.ReadFile(r.cfg.DockerComposeOverrideFile)
		require.NoError(r.t, err, "failed to read docker-compose.override.yml")
		overrideFile := filepath.Join(tmpDir, "docker-compose.override.yml")
		err = os.WriteFile(overrideFile, overrideCompose, 0644)
		require.NoError(r.t, err, "failed to write docker-compose.override.yml")
		r.cfg.DockerComposeOverrideFile = overrideFile
	}

	//config to use generated compose file
	r.cfg.DockerComposeFile = composeFile
}

func (r *IntHelper) WaitForSignals(t *testing.T) {
	done := make(chan os.Signal, 1)
	signal.Notify(done, syscall.SIGINT, syscall.SIGTERM)

	// block waiting for a signal
	s := <-done
	t.Logf("Got signal: %v, teardown\n", s)
}

func (r *IntHelper) ExtractPrivateKeys(home string) {
	regexPath := filepath.Join(home, "*", "private_key")

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

	switch driverType { // NOTE: REST api(for kwild) is discarded since kgw v0.3
	case "jsonrpc":
		return r.getJSONRPCClientDriver(signer, gatewayURL, gatewayProvider, nil)
	case "cli":
		return r.getCliDriver(gatewayURL, pk, signer.Identity(), gatewayProvider, nil)
	default:
		panic("unsupported driver type")
	}
}

// GetUserDriver returns an integration driver connected to the given rpc node
// using the private key
func (r *IntHelper) GetUserDriver(ctx context.Context, nodeName string, driverType string, deployer *ethdeployer.Deployer) KwilIntDriver {
	gatewayProvider := false

	ctr := r.containers[nodeName]
	jsonrpcURL, _, err := utils.KwildJSONRPCEndpoints(ctr, ctx)
	require.NoError(r.t, err, "failed to get json-rpc url")
	httpURL, _, err := utils.KwildHTTPEndpoints(ctr, ctx)
	require.NoError(r.t, err, "failed to get http url")
	cometBftURL, _, err := utils.KwildRpcEndpoints(ctr, ctx)
	require.NoError(r.t, err, "failed to get cometBft url")
	r.t.Logf("httpURL: %s jsonrpcURL: %s cometBftURL: %s for container: %s",
		httpURL, jsonrpcURL, cometBftURL, nodeName)

	signer := r.cfg.CreatorSigner
	pk := r.cfg.CreatorRawPk

	switch driverType {
	case "jsonrpc":
		return r.getJSONRPCClientDriver(signer, jsonrpcURL, gatewayProvider, deployer)
	case "http":
		return r.getHTTPClientDriver(signer, httpURL, gatewayProvider, deployer)
	case "cli":
		return r.getCliDriver(jsonrpcURL, pk, signer.Identity(), gatewayProvider, deployer)
	default:
		panic("unsupported driver type")
	}
}

// GetOperatorDriver returns a integration driver connected to the given kwil node,
// using the private key of the operator.
// The passed nodeName needs to be the same as the name of the container in docker-compose.yml
func (r *IntHelper) GetOperatorDriver(ctx context.Context, nodeName string, driverType string) operator.KwilOperatorDriver {
	switch driverType {
	case "jsonrpc":
		// Only cli is used presently, running from *within the container*.

		c, ok := r.containers[nodeName]
		if !ok {
			r.t.Fatalf("container %s not found", nodeName)
		}

		adminJSONRPCURL, err := c.PortEndpoint(ctx, "8485", "http")
		require.NoError(r.t, err, "failed to get admin json-rpc url")

		clt, err := adminclient.NewClient(ctx, adminJSONRPCURL)
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

	case "http", "grpc":
		r.t.Fatalf("driver not supported for node operator: %v", driverType)
		return nil
	default:
		r.t.Fatalf("unknown node operator driver: %v", driverType)
		return nil
	}
}

func (r *IntHelper) getJSONRPCClientDriver(signer auth.Signer, endpoint string, gatewayProvider bool, deployer *ethdeployer.Deployer) *driver.KwildClientDriver {
	logger := log.New(log.Config{Level: r.cfg.LogLevel})
	logger = *logger.With(log.String("testCase", r.t.Name()))

	var kwilClt clientType.Client
	var err error

	if gatewayProvider {
		// TODO: make gatewayclient use the JSON-RPC client! It's still the old HTTP one and it won't work
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

	return driver.NewKwildClientDriver(kwilClt, signer, deployer, logger)
}

func (r *IntHelper) getHTTPClientDriver(signer auth.Signer, endpoint string, gatewayProvider bool, deployer *ethdeployer.Deployer) *driver.KwildClientDriver {
	logger := log.New(log.Config{Level: r.cfg.LogLevel})
	logger = *logger.With(log.String("testCase", r.t.Name()))

	parsedURL, err := url.Parse(endpoint)
	require.NoError(r.t, err, "bad url")

	var kwilClt clientType.Client

	if gatewayProvider {
		kwilClt, err = gatewayclient.NewClient(context.TODO(), endpoint, &gatewayclient.GatewayOptions{
			Options: clientType.Options{
				Signer:  signer,
				ChainID: testChainID,
				Logger:  logger,
			},
		})
	} else {
		httpClient := http.NewClient(parsedURL)
		kwilClt, err = client.WrapClient(context.TODO(), httpClient, &clientType.Options{
			Signer:  signer,
			ChainID: testChainID,
			Logger:  logger,
		})
	}
	require.NoError(r.t, err, "wrapping http client failed")

	return driver.NewKwildClientDriver(kwilClt, signer, deployer, logger)
}

// getCLIAdminClientDriver returns a kwil-admin client driver connected to the
// given kwil node's container. The adminSvcServer is passed to kwil-admin's
// --rpcserver flag.
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

func (r *IntHelper) getCliDriver(endpoint, privKey string, identity []byte,
	gatewayProvider bool, deployer *ethdeployer.Deployer) *driver.KwilCliDriver {
	logger := log.New(log.Config{Level: r.cfg.LogLevel})
	logger = *logger.With(log.String("testCase", r.t.Name()))

	_, currentFilePath, _, _ := runtime.Caller(1)
	cliBinPath := path.Join(path.Dir(currentFilePath), "../../.build/kwil-cli")

	return driver.NewKwilCliDriver(cliBinPath, endpoint, privKey, testChainID, identity, gatewayProvider, deployer, logger)
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
