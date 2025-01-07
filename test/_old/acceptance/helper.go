package acceptance

import (
	"context"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"math/rand/v2"
	"os"
	"os/signal"
	"path"
	"runtime"
	"strconv"
	"syscall"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/compose"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/kwilteam/kwil-db/app/setup"
	"github.com/kwilteam/kwil-db/config"
	"github.com/kwilteam/kwil-db/core/client"
	clientType "github.com/kwilteam/kwil-db/core/client/types"
	"github.com/kwilteam/kwil-db/core/crypto"
	"github.com/kwilteam/kwil-db/core/crypto/auth"
	"github.com/kwilteam/kwil-db/core/log"
	"github.com/kwilteam/kwil-db/core/types"
	"github.com/kwilteam/kwil-db/node"

	"github.com/kwilteam/kwil-db/test/driver"
	"github.com/kwilteam/kwil-db/test/utils"
)

var (
	getEnv = driver.GetEnv
)

const TestChainID = "kwil-test-chain"

// ActTestCfg is the config for acceptance test
type ActTestCfg struct {
	JSONRPCEndpoint string
	P2PAddress      string // p2p address
	AdminRPC        string // tcp or unix socket

	SchemaFile                string
	DockerComposeFile         string
	DockerComposeOverrideFile string

	WaitTimeout time.Duration
	LogLevel    log.Level

	CreatorRawPk  string
	VisitorRawPK  string
	CreatorSigner auth.Signer
	VisitorSigner auth.Signer

	GasEnabled bool
	PrivateRPC bool
}

func (e *ActTestCfg) CreatorIdent() []byte {
	return e.CreatorSigner.Identity()
}

func (e *ActTestCfg) VisitorIdent() []byte {
	return e.VisitorSigner.Identity()
}

func (e *ActTestCfg) DumpToEnv() error {
	var envTemplage = `
JSONRPC_ENDPOINT=%s
CHAIN_ENDPOINT=%s
CREATOR_PRIVATE_KEY=%s
CREATOR_PUBLIC_KEY=%x
VISITOR_PRIVATE_KEY=%s
VISITOR_PUBLIC_KEY=%x
`
	content := fmt.Sprintf(envTemplage,
		e.JSONRPCEndpoint,
		e.P2PAddress,
		e.CreatorRawPk,
		e.CreatorIdent(),
		e.VisitorRawPK,
		e.VisitorIdent(),
	)

	err := os.WriteFile("../../.local_env", []byte(content), 0644)
	if err != nil {
		return fmt.Errorf("failed to dump config to env: %w", err)
	}
	return nil
}

type ActHelper struct {
	t   *testing.T
	cfg *ActTestCfg

	container *testcontainers.DockerContainer // kwild node container

	// envs is used to store dynamically generated envs later used in docker-compose
	// e.g. `dc.WithEnv(r.envs)`
	// for now one env are used:
	// - KWIL_HOME: the home directory for the test
	envs map[string]string
}

func NewActHelper(t *testing.T) *ActHelper {
	return &ActHelper{
		t:    t,
		envs: make(map[string]string),
	}
}

func (r *ActHelper) GetConfig() *ActTestCfg {
	return r.cfg
}

// LoadConfig loads config from system env and env file.
// Envs defined in envFile will not overwrite existing env vars.
func (r *ActHelper) LoadConfig() *ActTestCfg {
	var err error
	logLevel, _ := log.ParseLevel(getEnv("KIT_LOG_LEVEL", "info"))

	// default wallet mnemonic: test test test test test test test test test test test junk
	// default wallet hd path : m/44'/60'/0'
	cfg := &ActTestCfg{
		// NOTE: these ENVs are used to test remote services
		CreatorRawPk:              getEnv("KACT_CREATOR_PK", "f1aa5a7966c3863ccde3047f6a1e266cdc0c76b399e256b8fede92b1c69e4f4e"),
		VisitorRawPK:              getEnv("KACT_VISITOR_PK", "43f149de89d64bf9a9099be19e1b1f7a4db784af8fa07caf6f08dc86ba65636b"),
		SchemaFile:                getEnv("KACT_SCHEMA", "./test-data/test_db.kf"),
		LogLevel:                  logLevel,
		JSONRPCEndpoint:           getEnv("KACT_JSONRPC_ENDPOINT", "http://127.0.0.1:8484"),
		P2PAddress:                getEnv("KACT_CHAIN_ENDPOINT", "tcp://0.0.0.0:26656"),
		AdminRPC:                  getEnv("KACT_ADMIN_RPC", "/tmp/admin.socket"),
		DockerComposeFile:         getEnv("KACT_DOCKER_COMPOSE_FILE", "./docker-compose.yml"),
		DockerComposeOverrideFile: getEnv("KACT_DOCKER_COMPOSE_OVERRIDE_FILE", "./docker-compose.override.yml"),
	}

	cfg.PrivateRPC, err = strconv.ParseBool(getEnv("KACT_PRIVATE_RPC", "false"))
	require.NoError(r.t, err, "invalid privateRPC bool")

	cfg.GasEnabled, err = strconv.ParseBool(getEnv("KACT_GAS_ENABLED", "false"))
	require.NoError(r.t, err, "invalid gasEnabled bool")

	// value is in format of "10s" or "1m"
	waitTimeout := getEnv("KACT_WAIT_TIMEOUT", "10s")
	cfg.WaitTimeout, err = time.ParseDuration(waitTimeout)
	require.NoError(r.t, err, "invalid wait timeout")

	creatorPrivKeyBts, err := hex.DecodeString(cfg.CreatorRawPk)
	require.NoError(r.t, err)
	creatorPk, err := crypto.UnmarshalPrivateKey(creatorPrivKeyBts, crypto.KeyTypeSecp256k1) // require secp for now
	require.NoError(r.t, err, "invalid creator private key")
	// for "user" stuff not node (validator) tx signing.  Need different signer for admin svc signer
	cfg.CreatorSigner = auth.GetUserSigner(creatorPk)

	bobPrivKeyBts, err := hex.DecodeString(cfg.VisitorRawPK)
	require.NoError(r.t, err)
	bobPk, err := crypto.UnmarshalPrivateKey(bobPrivKeyBts, crypto.KeyTypeSecp256k1) // require secp for now
	require.NoError(r.t, err, "invalid creator private key")
	cfg.VisitorSigner = auth.GetUserSigner(bobPk)

	r.cfg = cfg
	//cfg.DumpToEnv()

	return cfg
}

func (r *ActHelper) updateEnv(k, v string) {
	r.envs[k] = v
}

func (r *ActHelper) GenerateTestnetConfigs() {
	r.t.Logf("generate node config")
	tmpPath, err := os.MkdirTemp("", "TestKwilAct")
	if err != nil {
		r.t.Fatal(err)
	}
	r.t.Cleanup(func() {
		if r.t.Failed() {
			r.t.Logf("Retaining data for failed test at path %v", tmpPath)
			return
		}
		os.RemoveAll(tmpPath)
	})

	r.t.Logf("created test temp directory: %s", tmpPath)

	/*bal, ok := big.NewInt(0).SetString("1000000000000000000000000000", 10)
	if !ok {
		r.t.Fatal("failed to parse balance")
	}
	creatorIdent := hex.EncodeToString(r.cfg.CreatorSigner.Identity())*/

	nodeKey, genesisCfg := makeSingleNodeNet(6600)

	err = setup.GenerateNodeRoot(&setup.NodeGenConfig{
		IP:         "0.0.0.0",
		PortOffset: 0,
		DBPort:     5454,  // test/acceptance/docker-compose-dev.yml
		NoPEX:      false, // pointless for one node but run the goroutines
		RootDir:    tmpPath,
		NodeKey:    nodeKey,
		Genesis:    genesisCfg,
	})
	/*err = nodecfg.GenerateTestnetConfigs(&nodecfg.NodeGenerateConfig{
		ChainID:       TestChainID,
		BlockInterval: time.Second,
		// InitialHeight: 0,
		OutputDir:        tmpPath,
		JoinExpiry:       14400,
		WithoutGasCosts:  !r.cfg.GasEnabled,
		AuthenticateRPCs: r.cfg.PrivateRPC,
		Allocs: map[string]*big.Int{
			creatorIdent: bal,
		},
	})*/
	require.NoError(r.t, err, "failed to generate node config")

	r.updateEnv("KWIL_HOME", tmpPath)
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func makeSingleNodeNet(seed uint64) (crypto.PrivateKey, *config.GenesisConfig) {
	// generate Keys, so that the connection strings and the validator set can be generated before the node config files are generated
	var seedArr [32]byte
	binary.LittleEndian.PutUint64(seedArr[:], seed)
	seedArr = sha256.Sum256(seedArr[:])
	rr := rand.NewChaCha8(seedArr)
	priv := node.NewKey(&deterministicPRNG{ChaCha8: rr})

	pubBts := priv.Public().Bytes()

	genConfig := &config.GenesisConfig{
		ChainID:          "kwil-testnet",
		Leader:           config.EncodePubKeyAndType(pubBts, priv.Type()),
		DisabledGasCosts: true,
		JoinExpiry:       14400,
		VoteExpiry:       108000,
		MaxBlockSize:     6 * 1024 * 1024,
		MaxVotesPerTx:    200,
		Validators: []*types.Validator{
			{
				PubKey:     pubBts,
				PubKeyType: priv.Type(),
				Power:      1,
			},
		},
	}
	return priv, genConfig
}

type deterministicPRNG struct {
	readBuf [8]byte
	readLen int // 0 <= readLen <= 8
	*rand.ChaCha8
}

// Read is a bad replacement for the actual Read method added in Go 1.23
func (dr *deterministicPRNG) Read(p []byte) (n int, err error) {
	// fill p by calling Uint64 in a loop until we have enough bytes
	if dr.readLen > 0 {
		n = copy(p, dr.readBuf[len(dr.readBuf)-dr.readLen:])
		dr.readLen -= n
		p = p[n:]
	}
	for len(p) >= 8 {
		binary.LittleEndian.PutUint64(p, dr.ChaCha8.Uint64())
		p = p[8:]
		n += 8
	}
	if len(p) > 0 {
		binary.LittleEndian.PutUint64(dr.readBuf[:], dr.Uint64())
		n += copy(p, dr.readBuf[:])
		dr.readLen = 8 - len(p)
	}
	return n, nil
}

func (r *ActHelper) runDockerCompose(ctx context.Context) {
	r.t.Logf("setup test environment")

	composeFiles := []string{r.cfg.DockerComposeFile}
	if r.cfg.DockerComposeOverrideFile != "" && fileExists(r.cfg.DockerComposeOverrideFile) {
		composeFiles = append(composeFiles, r.cfg.DockerComposeOverrideFile)
	}

	r.t.Logf("use compose files: %v", composeFiles)
	dc, err := compose.NewDockerCompose(composeFiles...)
	require.NoError(r.t, err, "failed to create docker compose object for single kwild node")

	r.t.Cleanup(func() {
		r.t.Logf("teardown docker compose")
		// err := dc.Down(ctx)
		// require.NoErrorf(r.t, err, "failed to teardown")
	})

	// NOTE: if you run with debugger image, you need to attach to the debugger
	// before the timeout
	err = dc.
		WithEnv(r.envs).
		WaitForService(
			"pg",
			wait.NewLogStrategy(`listening on IPv4 address "0.0.0.0", port 5432`).
				WithStartupTimeout(r.cfg.WaitTimeout)).
		WaitForService(
			"kwild",
			wait.NewLogStrategy("CONS: Committed Block").WithStartupTimeout(r.cfg.WaitTimeout)).
		Up(ctx, compose.Wait(true))
	r.t.Log("docker compose up")

	require.NoError(r.t, err, "failed to start kwild node")

	// NOTE: not sure how to get a container if we have multiple services with
	// same image
	container, err := dc.ServiceContainer(ctx, "kwild")
	require.NoError(r.t, err, "failed to get kwild node container")
	r.container = container
}

func (r *ActHelper) Setup(ctx context.Context) {
	r.GenerateTestnetConfigs()
	r.runDockerCompose(ctx)

	// update configured endpoints, so that we can still test against remote services
	jsonrpcEndpoint, _, err := utils.KwildJSONRPCEndpoints(r.container, ctx)
	require.NoError(r.t, err, "failed to get json-rpc endpoint")
	r.cfg.JSONRPCEndpoint = jsonrpcEndpoint
}

func (r *ActHelper) WaitUntilInterrupt() {
	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGTERM)

	// block waiting for a signal
	s := <-done
	r.t.Logf("Got signal: %v, teardown\n", s)
}

// GetDriver returns a concrete driver for acceptance test, based on the driver
// type and user. By default, the driver is created with the creator's private key.
func (r *ActHelper) GetDriver(driveType string, user string) KwilAcceptanceDriver {
	pk := r.cfg.CreatorRawPk
	signer := r.cfg.CreatorSigner
	id := signer.Identity()
	if user == "visitor" {
		signer = r.cfg.VisitorSigner
		pk = r.cfg.VisitorRawPK
		id = signer.Identity()
	} else if user == "" {
		// to run tests without private key
		signer = nil
		pk = ""
		id = nil
	}

	switch driveType {
	case "jsonrpc":
		return r.getJSONRPCClientDriver(signer, r.cfg.JSONRPCEndpoint)
	case "cli":
		return r.getCliDriver(pk, id, r.cfg.JSONRPCEndpoint)
	default:
		panic("unsupported driver type")
	}
}

func (r *ActHelper) getJSONRPCClientDriver(signer auth.Signer, endpoint string) KwilAcceptanceDriver {
	logger := log.New(log.WithFormat(log.FormatUnstructured), log.WithLevel(r.cfg.LogLevel))

	kwilClt, err := client.NewClient(context.TODO(), endpoint, &clientType.Options{
		Signer:  signer,
		ChainID: TestChainID,
		Logger:  logger,
	})
	require.NoError(r.t, err, "failed to create json-rpc client")

	return driver.NewKwildClientDriver(kwilClt, signer, nil, logger)
}

func (r *ActHelper) getCliDriver(privKey string, identifier []byte, endpoint string) KwilAcceptanceDriver {
	logger := log.New(log.WithFormat(log.FormatUnstructured), log.WithLevel(r.cfg.LogLevel))

	_, currentFilePath, _, _ := runtime.Caller(1)
	cliBinPath := path.Join(path.Dir(currentFilePath), "../../.build/kwil-cli")

	return driver.NewKwilCliDriver(cliBinPath, endpoint, privKey, TestChainID, identifier, false, nil, logger)
}
