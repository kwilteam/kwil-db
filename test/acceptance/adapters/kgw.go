package adapters

import (
	"context"
	"fmt"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
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
					HostPath: path.Join(path.Dir(currentFilePath), "../../acceptance/test-data/keys.json")},
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
