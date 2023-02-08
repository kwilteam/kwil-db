package adapters

import (
	"context"
	"fmt"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"kwil/internal/app/kwild"
	"kwil/pkg/kclient"
	"testing"
	"time"
)

const (
	KwildPort     = "50051"
	kwildDatabase = "kwil"
	kwildImage    = "kwild:latest"
)

// kwildContainer represents the kwild container type used in the module
type kwildContainer struct {
	TContainer
}

// setupKwild creates an instance of the kwild container type
func setupKwild(ctx context.Context, opts ...containerOption) (*kwildContainer, error) {
	req := testcontainers.ContainerRequest{
		Name:         fmt.Sprintf("kwild-%d", time.Now().Unix()),
		Image:        kwildImage,
		Env:          map[string]string{},
		Files:        []testcontainers.ContainerFile{},
		ExposedPorts: []string{},
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

	return &kwildContainer{TContainer{
		Container:     container,
		ContainerPort: KwildPort,
	}}, nil
}

func StartKwildDockerService(t *testing.T, ctx context.Context, envs map[string]string) *kwildContainer {
	//t.Helper()

	container, err := setupKwild(ctx,
		//WithDockerFile("kwil"),
		WithNetwork(kwilTestNetworkName),
		WithExposedPort(KwildPort),
		WithEnv(envs),
		// ForListeningPort requires image has /bin/sh
		WithWaitStrategy(wait.ForLog("grpc server started"), wait.ForLog("deposits synced")),
	)

	require.NoError(t, err, "Could not start kwil container")

	t.Cleanup(func() {
		require.NoError(t, container.Terminate(ctx), "Could not stop kwil container")
	})

	err = container.ShowPortInfo(ctx)
	require.NoError(t, err)
	return container
}

func GetKwildDriver(ctx context.Context, t *testing.T, addr string, cfg *kclient.Config, envs map[string]string) *kwild.Driver {
	t.Helper()

	if addr != "" {
		t.Logf("create grpc driver to %s", addr)
		cfg.Node.Endpoint = addr
		return kwild.NewDriver(cfg)
	}

	dc := StartDBDockerService(t, ctx)
	unexposedEndpoint, err := dc.UnexposedEndpoint(ctx)
	require.NoError(t, err)

	unexposedPgURL := fmt.Sprintf(
		"postgres://%s:%s@%s/%s?sslmode=disable", pgUser, pgPassword, unexposedEndpoint, kwildDatabase)

	envs["KWILD_DB_URL"] = unexposedPgURL
	envs["KWILD_LOG_LEVEL"] = "info"
	envs["KWILD_SERVER_ADDR"] = ":50051"

	// for specification verify
	kc := StartKwildDockerService(t, ctx, envs)
	endpoint, err := kc.ExposedEndpoint(ctx)
	require.NoError(t, err)
	t.Logf("create grpc driver to %s", endpoint)
	cfg.Node.Endpoint = endpoint
	return kwild.NewDriver(cfg)
}
