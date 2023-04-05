package adapters

import (
	"context"
	"fmt"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"testing"
	"time"
)

const (
	HasuraPort  = "8080"
	HasuraImage = "hasura/graphql-engine:v2.16.0"
)

// postgresContainer represents the postgres container type used in the module
type hasuraContainer struct {
	TContainer
}

func (c *hasuraContainer) ExposedEndpoint(ctx context.Context) (string, error) {
	endpoint, err := c.TContainer.ExposedEndpoint(ctx)
	if err != nil {
		return "", err
	}

	return "http://" + endpoint, nil
}

func (c *hasuraContainer) UnexposedEndpoint(ctx context.Context) (string, error) {
	endpoint, err := c.TContainer.UnexposedEndpoint(ctx)
	if err != nil {
		return "", err
	}

	return "http://" + endpoint, nil
}

// setupHasura creates an instance of the postgres container type
func setupHasura(ctx context.Context, opts ...containerOption) (*hasuraContainer, error) {
	req := testcontainers.ContainerRequest{
		Name:         fmt.Sprintf("hasura-%d", time.Now().Unix()),
		Image:        HasuraImage,
		Env:          map[string]string{},
		Networks:     []string{"test-network"},
		ExposedPorts: []string{}}

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

	return &hasuraContainer{TContainer{
		Container:     container,
		ContainerPort: HasuraPort,
	}}, nil
}

func StartHasuraDockerService(ctx context.Context, t *testing.T, envs map[string]string) *hasuraContainer {
	//t.Helper()

	container, err := setupHasura(ctx,
		WithNetwork(kwilTestNetworkName),
		WithExposedPort(HasuraPort),
		WithEnv(envs),
		WithWaitStrategy(wait.ForHTTP("/healthz").WithPort(HasuraPort)))

	require.NoError(t, err, "Could not start hasura container")

	t.Cleanup(func() {
		require.NoError(t, container.Terminate(ctx), "Could not stop hasura container")
	})

	err = container.ShowPortInfo(ctx)
	require.NoError(t, err)

	return container
}
