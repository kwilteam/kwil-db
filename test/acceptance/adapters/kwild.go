package adapters

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

const (
	KwildPort        = "50051"
	KwildGatewayPort = "8080"
	KwildDatabase    = "kwil"
	kwildImage       = "kwild:latest"
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
		WithExposedPorts([]string{KwildPort, KwildGatewayPort}),
		WithEnv(envs),
		// ForListeningPort requires image has /bin/sh
		WithWaitStrategy(wait.ForLog("grpc server started") /*, wait.ForLog("deposits synced")*/),
	)

	require.NoError(t, err, "Could not start kwil container")

	t.Cleanup(func() {
		require.NoError(t, container.Terminate(ctx), "Could not stop kwil container")
	})

	err = container.ShowPortInfo(ctx)
	require.NoError(t, err)
	return container
}
