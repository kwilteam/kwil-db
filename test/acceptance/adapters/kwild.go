package adapters

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/docker/go-connections/nat"

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

	SecondContainerPort string
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

	return &kwildContainer{
		TContainer: TContainer{
			Container:     container,
			ContainerPort: KwildPort,
		},
		SecondContainerPort: KwildGatewayPort,
	}, nil
}

func StartKwildDockerService(t *testing.T, ctx context.Context, envs map[string]string) *kwildContainer {
	//t.Helper()

	container, err := setupKwild(ctx,
		//WithDockerFile("kwil"),
		WithNetwork(kwilTestNetworkName),
		WithExposedPort(KwildPort),
		WithExposedPort(KwildGatewayPort),
		WithEnv(envs),
		// ForListeningPort requires image has /bin/sh
		WithWaitStrategy(wait.ForLog("grpc server started") /*, wait.ForLog("deposits synced")*/),
	)

	require.NoError(t, err, "Could not start kwil container")

	getLogs := listenToLogs(ctx, container)
	defer getLogs(t)

	t.Cleanup(func() {
		require.NoError(t, container.Terminate(ctx), "Could not stop kwil container")
	})

	err = container.ShowPortInfo(ctx)
	require.NoError(t, err)

	return container
}

type TestLogConsumer struct {
	Msgs []string
}

func (g *TestLogConsumer) Accept(l testcontainers.Log) {
	g.Msgs = append(g.Msgs, string(l.Content))
}

func listenToLogs(ctx context.Context, c testcontainers.Container) func(t *testing.T) error {
	g := TestLogConsumer{
		Msgs: []string{},
	}

	c.FollowOutput(&g) // must be called before StarLogProducer

	err := c.StartLogProducer(ctx)
	if err != nil {
		fmt.Printf("could not start log producer: %v", err)
		return func(t *testing.T) error { return err }
	}

	return func(t *testing.T) error {
		err = c.StopLogProducer()
		if err != nil {
			t.Errorf("could not stop log producer: %v", err)
			return err
		}

		for _, msg := range g.Msgs {
			t.Logf("kwild log: %s", msg)
		}

		return nil
	}

}

func (c *kwildContainer) SecondExposedEndpoint(ctx context.Context) (string, error) {
	host, err := c.Host(ctx)
	if err != nil {
		return "", err
	}
	hostPort, err := c.MappedPort(context.Background(), nat.Port(c.SecondContainerPort))
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s:%s", host, hostPort.Port()), nil
}

func (c *kwildContainer) SecondUnexposedEndpoint(ctx context.Context) (string, error) {
	ipC, err := c.ContainerIP(ctx)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s:%s", ipC, c.SecondContainerPort), nil
}
