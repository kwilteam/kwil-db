package adapters

import (
	"context"
	"fmt"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"kwil/internal/app/kgw"
	"kwil/pkg/kclient"
	"path"
	"runtime"
	"testing"
	"time"
)

const (
	KgwPort  = "8082"
	kgwImage = "kwil-gateway:latest"
)

// kgwContainer represents the kwil-gateway container type used in the module
type kgwContainer struct {
	TContainer
}

// setupKgw creates an instance of the kgw container type
func setupKgw(ctx context.Context, opts ...containerOption) (*kgwContainer, error) {
	_, currentFilePath, _, _ := runtime.Caller(1)

	req := testcontainers.ContainerRequest{
		Name:         fmt.Sprintf("kgw-%d", time.Now().Unix()),
		Image:        kgwImage,
		Env:          map[string]string{},
		Files:        []testcontainers.ContainerFile{},
		ExposedPorts: []string{},
		Mounts: testcontainers.ContainerMounts{
			{
				Source: testcontainers.GenericBindMountSource{
					HostPath: path.Join(path.Dir(currentFilePath), "../acceptance/test-data/keys.json")},
				Target:   "/app/keys.json",
				ReadOnly: true,
			},
		},
		//Cmd:          []string{"-h"},
	}

	for _, opt := range opts {
		opt(&req)
	}

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		return nil, err
	}

	return &kgwContainer{TContainer{
		Container:     container,
		ContainerPort: KgwPort,
	}}, nil
}

func StartKgwDockerService(ctx context.Context, t *testing.T, envs map[string]string) *kgwContainer {
	//t.Helper()

	container, err := setupKgw(ctx,
		//WithDockerFile("kwil-gateway"),
		WithNetwork(kwilTestNetworkName),
		WithExposedPort(KgwPort),
		WithEnv(envs),
		WithWaitStrategy(wait.ForLog("kwil gateway started"), wait.ForLog("graphql initialized")),
		// won't work, since api_key is required
		//WithWaitStrategy(wait.ForHTTP("/healthz").WithPort(KgwPort)),
	)

	require.NoError(t, err, "Could not start kwil-gateway container")

	t.Cleanup(func() {
		require.NoError(t, container.Terminate(ctx), "Could not stop kwil-gateway container")
	})

	err = container.ShowPortInfo(ctx)
	require.NoError(t, err)
	return container
}

func GetKgwDriver(ctx context.Context, t *testing.T, kwildAddr string, graphqlAddr string, apiKey string, cfg *kclient.Config, fundEnvs map[string]string) *kgw.Driver {
	t.Helper()

	if kwildAddr != "" {
		t.Logf("create kgw driver to %s", kwildAddr)
		cfg.Node.Addr = kwildAddr
		return kgw.NewDriver(cfg, graphqlAddr, apiKey)
	}

	// db container
	dc := StartDBDockerService(t, ctx)
	unexposedEndpoint, err := dc.UnexposedEndpoint(ctx)
	require.NoError(t, err)
	unexposedPgURL := fmt.Sprintf(
		"postgres://%s:%s@%s/%s?sslmode=disable", pgUser, pgPassword, unexposedEndpoint, kwildDatabase)

	// hasura container
	unexposedHasuraPgURL := fmt.Sprintf(
		"postgres://%s:%s@%s/%s?sslmode=disable", pgUser, pgPassword, unexposedEndpoint, "postgres")
	hasuraEnvs := map[string]string{
		"PG_DATABASE_URL":                      unexposedPgURL,
		"HASURA_GRAPHQL_METADATA_DATABASE_URL": unexposedHasuraPgURL,
		"HASURA_GRAPHQL_ENABLE_CONSOLE":        "true",
		"HASURA_GRAPHQL_DEV_MODE":              "true",
		"HASURA_METADATA_DB":                   "postgres",
	}
	hasurac := StartHasuraDockerService(ctx, t, hasuraEnvs)
	unexposedHasuraEndpoint, err := hasurac.UnexposedEndpoint(ctx)
	require.NoError(t, err)

	// kwild container
	fundEnvs["KWILD_GRAPHQL_ADDR"] = unexposedHasuraEndpoint
	// @yaiba can't get addr here, because the gw container is not ready yet
	// need a hacky way to get the addr
	fundEnvs["KWILD_GATEWAY_ADDR"] = ""
	fundEnvs["KWILD_DB_URL"] = unexposedPgURL
	fundEnvs["KWILD_LOG_LEVEL"] = "info"
	kc := StartKwildDockerService(t, ctx, fundEnvs)
	exposedkwildEndpoint, err := kc.ExposedEndpoint(ctx)
	require.NoError(t, err)
	unexposedKwildEndpoint, err := kc.UnexposedEndpoint(ctx)
	require.NoError(t, err)

	// kgw container
	kgwEnvs := map[string]string{
		"KWILGW_KWILD_ADDR":         unexposedKwildEndpoint,
		"KWILGW_GRAPHQL_ADDR":       unexposedHasuraEndpoint,
		"KWILGW_LOG_LEVEL":          "info",
		"KWILGW_SERVER_LISTEN_ADDR": ":8082",
	}

	kgwc := StartKgwDockerService(ctx, t, kgwEnvs)
	kgwEndpoint, err := kgwc.ExposedEndpoint(ctx)
	require.NoError(t, err)

	cfg.Node.Addr = exposedkwildEndpoint
	// from test-data/keys.json
	testAPIKey := "testkwilkey"
	t.Logf("create kgw driver to %s", kgwEndpoint)
	return kgw.NewDriver(cfg, kgwEndpoint, testAPIKey)
}
