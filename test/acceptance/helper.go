package acceptance

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"path"
	"runtime"
	"syscall"
	"testing"
	"time"

	"github.com/kwilteam/kwil-db/pkg/client"
	"github.com/kwilteam/kwil-db/pkg/crypto"
	"github.com/kwilteam/kwil-db/pkg/log"
	"github.com/kwilteam/kwil-db/pkg/nodecfg"
	"github.com/kwilteam/kwil-db/test/driver"

	"github.com/joho/godotenv"
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

// ActTestCfg is the config for acceptance test
type ActTestCfg struct {
	GWEndpoint    string // gateway endpoint
	GrpcEndpoint  string
	ChainEndpoint string

	SchemaFile                string
	DockerComposeFile         string
	DockerComposeOverrideFile string

	WaitTimeout time.Duration
	LogLevel    string

	CreatorRawPk  string
	VisitorRawPK  string
	CreatorSigner crypto.Signer
	VisitorSigner crypto.Signer
}

func (e *ActTestCfg) CreatorAddr() string {
	return e.CreatorSigner.PubKey().Address().String()
}

func (e *ActTestCfg) VisitorAddr() string {
	return e.VisitorSigner.PubKey().Address().String()
}

func (e *ActTestCfg) IsRemote() bool {
	return e.GrpcEndpoint != ""
}

func (e *ActTestCfg) DumpToEnv() error {
	var envTemplage = `
GRPC_ENDPOINT=%s
GATEWAY_ENDPOINT=%s
CHAIN_ENDPOINT=%s
CREATOR_PK=%s
CREATOR_ADDR=%s
VISITOR_PK=%s
VISITOR_ADDR=%s
`
	content := fmt.Sprintf(envTemplage,
		e.GrpcEndpoint,
		e.GWEndpoint,
		e.ChainEndpoint,
		e.CreatorRawPk,
		e.CreatorAddr(),
		e.VisitorRawPK,
		e.VisitorAddr(),
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
func (r *ActHelper) LoadConfig() {
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
		GWEndpoint:                getEnv("KACT_GATEWAY_ENDPOINT", "localhost:8080"),
		GrpcEndpoint:              getEnv("KACT_GRPC_ENDPOINT", "localhost:50051"),
		DockerComposeFile:         getEnv("KACT_DOCKER_COMPOSE_FILE", "./docker-compose.yml"),
		DockerComposeOverrideFile: getEnv("KACT_DOCKER_COMPOSE_OVERRIDE_FILE", "./docker-compose.override.yml"),
	}

	// value is in format of "10s" or "1m"
	waitTimeout := getEnv("KACT_WAIT_TIMEOUT", "10s")
	cfg.WaitTimeout, err = time.ParseDuration(waitTimeout)
	require.NoError(r.t, err, "invalid wait timeout")

	creatorPk, err := crypto.Secp256k1PrivateKeyFromHex(cfg.CreatorRawPk)
	require.NoError(r.t, err, "invalid creator private key")
	cfg.CreatorSigner = crypto.DefaultSigner(creatorPk)

	bobPk, err := crypto.Secp256k1PrivateKeyFromHex(cfg.VisitorRawPK)
	require.NoError(r.t, err, "invalid visitor private key")
	cfg.VisitorSigner = crypto.DefaultSigner(bobPk)

	r.cfg = cfg
	cfg.DumpToEnv()
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
	tmpPath := r.t.TempDir() // automatically removed by testing.T.Cleanup
	// To prevent go test from cleaning up:
	// tmpPath, err := os.MkdirTemp("", "TestKwilAct")
	// if err != nil {
	// 	r.t.Fatal(err)
	// }
	r.t.Logf("create test temp directory: %s", tmpPath)

	err := nodecfg.GenerateNodeConfig(&nodecfg.NodeGenerateConfig{
		// InitialHeight: 0,
		OutputDir:       tmpPath,
		JoinExpiry:      86400,
		WithoutGasCosts: true,
		WithoutNonces:   false,
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
	case "client":
		return r.getClientDriver(signer)
	case "cli":
		return r.getCliDriver(pk, signer.PubKey().Bytes())
	default:
		panic("unsupported driver type")
	}
}

func (r *ActHelper) getClientDriver(signer crypto.Signer) KwilAcceptanceDriver {
	logger := log.New(log.Config{Level: r.cfg.LogLevel})

	options := []client.Option{client.WithSigner(signer),
		client.WithLogger(logger),
		client.WithTLSCert("")} // TODO: handle cert
	kwilClt, err := client.Dial(r.cfg.GrpcEndpoint, options...)
	require.NoError(r.t, err, "failed to create kwil client")

	return driver.NewKwildClientDriver(kwilClt, driver.WithLogger(logger))
}

func (r *ActHelper) getCliDriver(privKey string, pubKey []byte) KwilAcceptanceDriver {
	logger := log.New(log.Config{Level: r.cfg.LogLevel})

	_, currentFilePath, _, _ := runtime.Caller(1)
	cliBinPath := path.Join(path.Dir(currentFilePath),
		fmt.Sprintf("../../.build/kwil-cli-%s-%s", runtime.GOOS, runtime.GOARCH))
	adminBinPath := path.Join(path.Dir(currentFilePath),
		fmt.Sprintf("../../.build/kwil-admin-%s-%s", runtime.GOOS, runtime.GOARCH))

	return driver.NewKwilCliDriver(cliBinPath, adminBinPath, r.cfg.GrpcEndpoint, privKey, pubKey, logger)
}
