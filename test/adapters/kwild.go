package adapters

import (
	"context"
	"fmt"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"kwil/pkg/kwil-client"
	"sync"
	"testing"
	"time"
)

const (
	KwildPort     = "50051"
	kwildDatabase = "kwil"
	kwildImage    = "kwild:latest"
)

var buildKwildOnce sync.Once

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

func GetGrpcDriver(t *testing.T, ctx context.Context, addr string, cfg *kwil_client.Config, envs map[string]string, dbUrl string) *kwil_client.Driver {
	t.Helper()

	if addr != "" {
		t.Logf("create grpc driver to %s", addr)
		cfg.Node.Endpoint = addr
		return kwil_client.NewDriver(cfg, dbUrl)
	}

	dbFiles := map[string]string{
		"../../scripts/pg-init-scripts/initdb.sh": "/docker-entrypoint-initdb.d/initdb.sh"}
	dc := StartDBDockerService(t, ctx, dbFiles)
	unexposedEndpoint, err := dc.UnexposedEndpoint(ctx)
	require.NoError(t, err)
	exposedEndpoint, err := dc.ExposedEndpoint(ctx)
	require.NoError(t, err)

	unexposedPgUrl := fmt.Sprintf(
		"postgres://%s:%s@%s/%s?sslmode=disable", pgUser, pgPassword, unexposedEndpoint, kwildDatabase)
	exposedPgUrl := fmt.Sprintf(
		"postgres://%s:%s@%s/%s?sslmode=disable", pgUser, pgPassword, exposedEndpoint, kwildDatabase)

	envs["KWIL_DB_URL"] = unexposedPgUrl
	envs["KWIL_LOG_LEVEL"] = "info"
	envs["KWIL_SERVER_ADDR"] = ":50051"

	// for specification verify
	kc := StartKwildDockerService(t, ctx, envs)
	endpoint, err := kc.ExposedEndpoint(ctx)
	require.NoError(t, err)
	t.Logf("create grpc driver to %s", endpoint)
	cfg.Node.Endpoint = endpoint
	return kwil_client.NewDriver(cfg, exposedPgUrl)
}
