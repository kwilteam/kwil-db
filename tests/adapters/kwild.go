package adapters

import (
	"context"
	"fmt"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/testcontainers/testcontainers-go"
	grpc_client "kwil/kwil/client/grpc-client"
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

func GetGrpcDriver(t *testing.T, ctx context.Context, addr string, dbFiles map[string]string, envs map[string]string) *grpc_client.Driver {
	t.Helper()

	if addr != "" {
		return &grpc_client.Driver{Addr: addr}
	}

	dc := StartDBDockerService(t, ctx, dbFiles)
	unexposedEndpoint, err := dc.UnexposedEndpoint(ctx)
	assert.NoError(t, err)
	exposedEndpoint, err := dc.ExposedEndpoint(ctx)
	assert.NoError(t, err)

	unexposedPgUrl := fmt.Sprintf(
		"postgres://%s:%s@%s/%s?sslmode=disable", pgUser, pgPassword, unexposedEndpoint, kwildDatabase)
	exposedPgUrl := fmt.Sprintf(
		"postgres://%s:%s@%s/%s?sslmode=disable", pgUser, pgPassword, exposedEndpoint, kwildDatabase)

	envs["PG_DATABASE_URL"] = unexposedPgUrl

	// for cli
	viper.Set("PG_DATABASE_URL", exposedPgUrl)

	kc := StartKwildDockerService(t, ctx, envs)
	endpoint, err := kc.ExposedEndpoint(ctx)
	assert.NoError(t, err)
	t.Logf("create grpc driver to %s", endpoint)
	return &grpc_client.Driver{Addr: endpoint}
}
