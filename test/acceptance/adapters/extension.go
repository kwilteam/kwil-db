package adapters

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
)

const (
	MathExtImage = "kwilbrennan/extensions-math:latest"
	MathExtPort  = "50051"
)

type mathExtensionContainer struct {
	TContainer
}

func (c *mathExtensionContainer) ExposedEndpoint(ctx context.Context) (string, error) {
	endpoint, err := c.TContainer.ExposedEndpoint(ctx)
	if err != nil {
		return "", err
	}

	return "ws://" + endpoint, nil
}

func (c *mathExtensionContainer) UnexposedEndpoint(ctx context.Context) (string, error) {
	endpoint, err := c.TContainer.UnexposedEndpoint(ctx)
	if err != nil {
		return "", fmt.Errorf("could not get unexposed endpoint for math extension: %w", err)
	}

	return endpoint, nil
}

func newExtensionMath(ctx context.Context) (*mathExtensionContainer, error) {
	req := testcontainers.ContainerRequest{
		Name:         fmt.Sprintf("math-extension-%d", time.Now().Unix()),
		Image:        MathExtImage,
		Env:          map[string]string{},
		Files:        []testcontainers.ContainerFile{},
		ExposedPorts: []string{MathExtPort},
		//Cmd:          []string{"-h"},
		//WaitingFor: wait.ForListeningPort(MathExtPort),
	}

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		return nil, err
	}

	return &mathExtensionContainer{TContainer{
		Container:     container,
		ContainerPort: MathExtPort,
	}}, nil
}

func StartMathExtensionDockerService(t *testing.T, ctx context.Context) *mathExtensionContainer {
	container, err := newExtensionMath(ctx)
	if err != nil {
		panic(err)
	}
	require.NoError(t, err, "Could not start extension container")

	t.Cleanup(func() {
		require.NoError(t, container.Terminate(ctx), "Could not stop extension container")
	})

	err = container.ShowPortInfo(ctx)
	require.NoError(t, err)
	return container
}
