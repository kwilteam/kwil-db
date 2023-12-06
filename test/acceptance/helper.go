package acceptance

import (
	"context"
	"encoding/hex"
	"fmt"
	"math/big"
	"os"
	"os/signal"
	"path"
	"runtime"
	"strconv"
	"syscall"
	"testing"
	"time"

	"github.com/joho/godotenv"
	"github.com/kwilteam/kwil-db/cmd/kwil-admin/nodecfg"
	"github.com/kwilteam/kwil-db/core/client"
	"github.com/kwilteam/kwil-db/core/crypto"
	"github.com/kwilteam/kwil-db/core/crypto/auth"
	"github.com/kwilteam/kwil-db/core/log"
	gRPC "github.com/kwilteam/kwil-db/core/rpc/client/user/grpc"
	"github.com/kwilteam/kwil-db/test/driver"

	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go/modules/compose"
	"github.com/testcontainers/testcontainers-go/wait"
)

var (
	getEnv = driver.GetEnv

	// envFile is the default env file path
	// it will pass values among different stages of the test setup
	envFile = getEnv("KACT_ENV_FILE", "./.env")
)

const TestChainID = "kwil-test-chain"

// ActTestCfg is the config for acceptance test
type ActTestCfg struct {
	HTTPEndpoint          string
	GrpcEndpoint          string
	P2PAddress            string // cometbft p2p address
	AdminClientUnixSocket string

	SchemaFile                string
	DockerComposeFile         string
	DockerComposeOverrideFile string
	NoCleanup                 bool

	WaitTimeout time.Duration
	LogLevel    string

	CreatorRawPk  string
	VisitorRawPK  string
	CreatorSigner auth.Signer
	VisitorSigner auth.Signer

	GasEnabled bool
}

func (e *ActTestCfg) CreatorIdent() []byte {
	return e.CreatorSigner.Identity()
}

func (e *ActTestCfg) VisitorIdent() []byte {
	return e.VisitorSigner.Identity()
}

func (e *ActTestCfg) IsRemote() bool {
	return e.GrpcEndpoint != ""
}

func (e *ActTestCfg) DumpToEnv() error {
	var envTemplage = `
GRPC_ENDPOINT=%s
GATEWAY_ENDPOINT=%s
CHAIN_ENDPOINT=%s
CREATOR_PRIVATE_KEY=%s
CREATOR_PUBLIC_KEY=%x
VISITOR_PRIVATE_KEY=%s
VISITOR_PUBLIC_KEY=%x
`
	content := fmt.Sprintf(envTemplage,
		e.GrpcEndpoint,
		e.HTTPEndpoint,
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
	t        *testing.T
	cfg      *ActTestCfg
	teardown []func()
}

func NewActHelper(t *testing.T) *ActHelper {
	return &ActHelper{
		t: t,
	}
}

func (r *ActHelper) GetConfig() *ActTestCfg {
	return r.cfg
}

// LoadConfig loads config from system env and env file.
// Envs defined in envFile will not overwrite existing env vars.
func (r *ActHelper) LoadConfig() *ActTestCfg {
	ef, err := os.OpenFile(envFile, os.O_RDWR|os.O_CREATE, 0666)
	require.NoError(r.t, err, "failed to open env file")
	defer ef.Close()

	err = godotenv.Load(envFile)
	require.NoError(r.t, err, "failed to parse env file")

	// default wallet mnemonic: test test test test test test test test test test test junk
	// default wallet hd path : m/44'/60'/0'
	cfg := &ActTestCfg{
		CreatorRawPk:              getEnv("KACT_CREATOR_PK", "f1aa5a7966c3863ccde3047f6a1e266cdc0c76b399e256b8fede92b1c69e4f4e"),
		VisitorRawPK:              getEnv("KACT_VISITOR_PK", "43f149de89d64bf9a9099be19e1b1f7a4db784af8fa07caf6f08dc86ba65636b"),
		SchemaFile:                getEnv("KACT_SCHEMA", "./test-data/test_db.kf"),
		LogLevel:                  getEnv("KACT_LOG_LEVEL", "info"),
		HTTPEndpoint:              getEnv("KACT_HTTP_ENDPOINT", "http://localhost:8080"),
		GrpcEndpoint:              getEnv("KACT_GRPC_ENDPOINT", "localhost:50051"),
		P2PAddress:                getEnv("KACT_CHAIN_ENDPOINT", "tcp://0.0.0.0:26656"),
		AdminClientUnixSocket:     getEnv("KACT_ADMIN_CLIENT_UNIX_SOCKET", "tcp://localhost:50151"),
		DockerComposeFile:         getEnv("KACT_DOCKER_COMPOSE_FILE", "./docker-compose.yml"),
		DockerComposeOverrideFile: getEnv("KACT_DOCKER_COMPOSE_OVERRIDE_FILE", "./docker-compose.override.yml"),
	}

	cfg.GasEnabled, err = strconv.ParseBool(getEnv("KACT_GAS_ENABLED", "false"))
	require.NoError(r.t, err, "invalid gasEnabled bool")

	cfg.NoCleanup, err = strconv.ParseBool(getEnv("KACT_NO_CLEANUP", "false"))
	require.NoError(r.t, err, "invalid noCleanup bool")

	// value is in format of "10s" or "1m"
	waitTimeout := getEnv("KACT_WAIT_TIMEOUT", "10s")
	cfg.WaitTimeout, err = time.ParseDuration(waitTimeout)
	require.NoError(r.t, err, "invalid wait timeout")

	creatorPk, err := crypto.Secp256k1PrivateKeyFromHex(cfg.CreatorRawPk)
	require.NoError(r.t, err, "invalid creator private key")
	cfg.CreatorSigner = &auth.EthPersonalSigner{Key: *creatorPk}

	bobPk, err := crypto.Secp256k1PrivateKeyFromHex(cfg.VisitorRawPK)
	require.NoError(r.t, err, "invalid visitor private key")
	cfg.VisitorSigner = &auth.EthPersonalSigner{Key: *bobPk}

	r.cfg = cfg
	cfg.DumpToEnv()

	return cfg
}

func (r *ActHelper) updateGeneratedConfig(ks, vs []string) {
	require.Equal(r.t, len(ks), len(vs), "length of keys and values should be equal")

	envs, err := godotenv.Read(envFile)
	require.NoError(r.t, err, "failed to read env file")

	for i := range ks {
		envs[ks[i]] = vs[i]
	}

	err = godotenv.Write(envs, envFile)
	require.NoError(r.t, err, "failed to write env vars to file")
}

func (r *ActHelper) generateNodeConfig() {
	r.t.Logf("generate node config")
	var tmpPath string
	if r.cfg.NoCleanup {
		var err error
		tmpPath, err = os.MkdirTemp("", "TestKwilAct")
		if err != nil {
			r.t.Fatal(err)
		}
	} else {
		tmpPath = r.t.TempDir() // automatically removed by testing.T.Cleanup
	}

	r.t.Logf("created test temp directory: %s", tmpPath)

	bal, ok := big.NewInt(0).SetString("1000000000000000000000000000", 10)
	if !ok {
		r.t.Fatal("failed to parse balance")
	}
	creatorIdent := hex.EncodeToString(r.cfg.CreatorSigner.Identity())

	err := nodecfg.GenerateNodeConfig(&nodecfg.NodeGenerateConfig{
		ChainID:       TestChainID,
		BlockInterval: time.Second,
		// InitialHeight: 0,
		OutputDir:       tmpPath,
		JoinExpiry:      14400,
		WithoutGasCosts: !r.cfg.GasEnabled,
		WithoutNonces:   false,
		Allocs: map[string]*big.Int{
			creatorIdent: bal,
		},
	})
	require.NoError(r.t, err, "failed to generate node config")

	r.updateGeneratedConfig(
		[]string{"KWIL_HOME"},
		[]string{tmpPath})
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

func (r *ActHelper) runDockerCompose(ctx context.Context) {
	r.t.Logf("setup test environment")

	//setSchemaLoader(r.cfg.CreatorAddr())

	envs, err := godotenv.Read(envFile)
	require.NoError(r.t, err, "failed to parse .env file")

	composeFiles := []string{r.cfg.DockerComposeFile}
	if r.cfg.DockerComposeOverrideFile != "" && fileExists(r.cfg.DockerComposeOverrideFile) {
		composeFiles = append(composeFiles, r.cfg.DockerComposeOverrideFile)
	}
	dc, err := compose.NewDockerCompose(composeFiles...)
	require.NoError(r.t, err, "failed to create docker compose object for single kwild node")

	r.teardown = append(r.teardown, func() {
		r.t.Logf("teardown docker compose")
		dc.Down(ctx, compose.RemoveOrphans(true), compose.RemoveImagesLocal)
	})

	r.t.Cleanup(func() {
		r.Teardown()
	})

	// NOTE: if you run with debugger image, you need to attach to the debugger
	// before the timeout
	err = dc.
		WithEnv(envs).
		WaitForService(
			"ext",
			wait.NewLogStrategy("listening on").WithStartupTimeout(r.cfg.WaitTimeout)).
		WaitForService(
			"kwild",
			wait.NewLogStrategy("Starting Node service").WithStartupTimeout(r.cfg.WaitTimeout)).
		Up(ctx)
	r.t.Log("docker compose up")

	require.NoError(r.t, err, "failed to start kwild node")
}

func (r *ActHelper) Setup(ctx context.Context) {
	r.generateNodeConfig()
	r.runDockerCompose(ctx)
}

func (r *ActHelper) Teardown() {
	if r.cfg.NoCleanup {
		return
	}
	r.t.Log("teardown test environment")
	for _, fn := range r.teardown {
		fn()
	}
}

func (r *ActHelper) WaitUntilInterrupt() {
	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGTERM)

	// block waiting for a signal
	s := <-done
	r.t.Logf("Got signal: %v\n", s)
	r.Teardown()
	r.t.Logf("Teardown done\n")
}

// GetDriver returns a concrete driver for acceptance test, based on the driver
// type and user. By default, the driver is created with the creator's private key.
func (r *ActHelper) GetDriver(driveType string, user string) KwilAcceptanceDriver {
	pk := r.cfg.CreatorRawPk
	signer := r.cfg.CreatorSigner
	if user == "visitor" {
		signer = r.cfg.VisitorSigner
		pk = r.cfg.VisitorRawPK
	}

	switch driveType {
	case "http":
		return r.getHTTPClientDriver(signer)
	case "grpc":
		return r.getGRPCClientDriver(signer)
	case "cli":
		return r.getCliDriver(pk, signer.Identity())
	default:
		panic("unsupported driver type")
	}
}

func (r *ActHelper) getHTTPClientDriver(signer auth.Signer) KwilAcceptanceDriver {
	logger := log.New(log.Config{Level: r.cfg.LogLevel})

	kwilClt, err := client.NewClient(context.TODO(), r.cfg.HTTPEndpoint, &client.ClientOptions{
		Signer:  signer,
		ChainID: TestChainID,
		Logger:  logger,
	})
	require.NoError(r.t, err, "failed to create http client")

	return driver.NewKwildClientDriver(kwilClt, logger)
}

func (r *ActHelper) getGRPCClientDriver(signer auth.Signer) KwilAcceptanceDriver {
	logger := log.New(log.Config{Level: r.cfg.LogLevel})

	gtOptions := []gRPC.Option{gRPC.WithTlsCert("")}
	gt, err := gRPC.New(context.Background(), r.cfg.GrpcEndpoint, gtOptions...)
	require.NoError(r.t, err, "failed to create grpc transport")

	kwilClt, err := client.WrapClient(context.TODO(), gt, &client.ClientOptions{
		Signer: signer,
		Logger: logger,
		// we dont care about chain id here
	})
	require.NoError(r.t, err, "failed to create grpc client")

	return driver.NewKwildClientDriver(kwilClt, logger)
}

func (r *ActHelper) getCliDriver(privKey string, identifier []byte) KwilAcceptanceDriver {
	logger := log.New(log.Config{Level: r.cfg.LogLevel})

	_, currentFilePath, _, _ := runtime.Caller(1)
	cliBinPath := path.Join(path.Dir(currentFilePath),
		fmt.Sprintf("../../.build/kwil-cli-%s-%s", runtime.GOOS, runtime.GOARCH))

	return driver.NewKwilCliDriver(cliBinPath, r.cfg.HTTPEndpoint, privKey, TestChainID, identifier, logger)
}
