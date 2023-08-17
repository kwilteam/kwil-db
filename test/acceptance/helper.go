package acceptance

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/kwilteam/kwil-db/internal/app/kwild"
	"github.com/kwilteam/kwil-db/internal/pkg/nodecfg"
	"github.com/kwilteam/kwil-db/pkg/client"
	"github.com/kwilteam/kwil-db/pkg/crypto"
	"github.com/kwilteam/kwil-db/test/runner"

	"github.com/joho/godotenv"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go/modules/compose"
	"github.com/testcontainers/testcontainers-go/wait"
)

const DefaultContainerWaitTimeout = 10 * time.Second

type ActTestCfg struct {
	GWEndpoint    string // gateway endpoint
	GrpcEndpoint  string
	ChainEndpoint string

	SchemaFile        string
	DockerComposeFile string

	LogLevel string

	AliceRawPK string // Alice is the owner
	BobRawPK   string
	AlicePK    crypto.Signer
	BobPk      crypto.Signer
}

func (e *ActTestCfg) AliceAddr() string {
	return e.AlicePK.PubKey().Address().String()
}

func (e *ActTestCfg) BobAddr() string {
	return e.BobPk.PubKey().Address().String()
}

func (e *ActTestCfg) IsRemote() bool {
	return e.GrpcEndpoint != ""
}

func (e *ActTestCfg) DumpToEnv() error {
	var envTemplage = `
GRPC_ENDPOINT=%s
GATEWAY_ENDPOINT=%s
CHAIN_ENDPOINT=%s
ALICE_PK=%s
ALICE_ADDR=%s
BOB_PK=%s
BOB_ADDR=%s
`
	content := fmt.Sprintf(envTemplage,
		e.GrpcEndpoint,
		e.GWEndpoint,
		e.ChainEndpoint,
		e.AliceRawPK,
		e.AliceAddr(),
		e.BobRawPK,
		e.BobAddr(),
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

func (r *ActHelper) LoadConfig() {
	// default wallet mnemonic: test test test test test test test test test test test junk
	// default wallet hd path : m/44'/60'/0'
	cfg := &ActTestCfg{
		AliceRawPK:        runner.GetEnv("KACT_ALICE_PK", "f1aa5a7966c3863ccde3047f6a1e266cdc0c76b399e256b8fede92b1c69e4f4e"),
		BobRawPK:          runner.GetEnv("KACT_BOB_PK", "43f149de89d64bf9a9099be19e1b1f7a4db784af8fa07caf6f08dc86ba65636b"),
		SchemaFile:        runner.GetEnv("KACT_SCHEMA", "./test-data/test_db.kf"),
		LogLevel:          runner.GetEnv("KACT_CHAIN_ENDPOINT", "http://localhost:26657"),
		GWEndpoint:        runner.GetEnv("KACT_GRPC_ENDPOINT", "localhost:9090"),
		GrpcEndpoint:      runner.GetEnv("KACT_GATEWAY_ENDPOINT", "localhost:8080"),
		DockerComposeFile: runner.GetEnv("KACT_DOCKER_COMPOSE_FILE", "./docker-compose.yml"),
	}

	var err error
	cfg.AlicePK, err = crypto.Secp256k1PrivateKeyFromHex(cfg.AliceRawPK)
	require.NoError(r.t, err, "invalid alice private key")
	//if err != nil {
	//	return nil, fmt.Errorf("invalid alice private key: %v", err)
	//}

	cfg.BobPk, err = crypto.Secp256k1PrivateKeyFromHex(cfg.BobRawPK)
	require.NoError(r.t, err, "invalid bob private key")
	r.cfg = cfg

	cfg.DumpToEnv()
}

func (r *ActHelper) generateNodeConfig() {
	tmpPath := r.t.TempDir()
	r.t.Logf("create test temp directory: %s", tmpPath)
	envVars, err := godotenv.Unmarshal("KWIL_HOME=" + tmpPath)
	require.NoError(r.t, err, "failed to unmarshal env vars")
	err = godotenv.Write(envVars, "./.env")
	require.NoError(r.t, err, "failed to write env vars to file")

	err = nodecfg.GenerateNodeConfig(&nodecfg.NodeGenerateConfig{
		InitialHeight: 0,
		HomeDir:       tmpPath,
	})
	require.NoError(r.t, err, "failed to generate node config")
}

func (r *ActHelper) runDockerCompose(ctx context.Context) {
	// start test containers
	r.t.Logf("setup test environment")

	//setSchemaLoader(r.cfg.AliceAddr())

	fEnv, err := os.Open("./.env")
	require.NoError(r.t, err, "failed to open .env file")

	envs, err := godotenv.Parse(fEnv)
	require.NoError(r.t, err, "failed to parse .env file")

	dc, err := compose.NewDockerCompose(r.cfg.DockerComposeFile)
	require.NoError(r.t, err, "failed to create docker compose object for single kwild node")

	r.teardown = append(r.teardown, func() {
		r.t.Logf("teardown docker compose")
		dc.Down(ctx, compose.RemoveOrphans(true), compose.RemoveImagesLocal)
	})

	r.t.Cleanup(func() {
		r.Teardown()
	})

	err = dc.
		WithEnv(envs).
		WaitForService(
			"ext",
			wait.NewLogStrategy("listening on").WithStartupTimeout(DefaultContainerWaitTimeout)).
		WaitForService(
			"kwild",
			wait.NewLogStrategy("grpc server started").WithStartupTimeout(DefaultContainerWaitTimeout)).
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

func (r *ActHelper) GetAliceDriver(ctx context.Context) KwilAcceptanceDriver {
	kwilClt, err := client.New(ctx, r.cfg.GrpcEndpoint,
		client.WithSigner(r.cfg.AlicePK),
		client.WithCometBftUrl(r.cfg.ChainEndpoint),
	)
	require.NoError(r.t, err, "failed to create kwil client")

	return kwild.NewKwildDriver(kwilClt)
}

func (r *ActHelper) GetBobDriver(ctx context.Context) KwilAcceptanceDriver {
	kwilClt, err := client.New(ctx, r.cfg.GrpcEndpoint,
		client.WithSigner(r.cfg.BobPk),
		client.WithCometBftUrl(r.cfg.ChainEndpoint),
	)
	require.NoError(r.t, err, "failed to create kwil client")

	return kwild.NewKwildDriver(kwilClt)
}
