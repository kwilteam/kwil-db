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
	GanachePort  = "8545"
	GanacheImage = "trufflesuite/ganache:v7.7.3"

	WalletMnemonic    = "test test test test test test test test test test test junk"
	WalletHDPath      = "m/44'/60'/0'"
	DeployerAccountPK = "dd23ca549a97cb330b011aebb674730df8b14acaee42d211ab45692699ab8ba5" // address: 0x1e59ce931B4CFea3fe4B875411e280e173cB7A9C
	UserAccountPK     = "f1aa5a7966c3863ccde3047f6a1e266cdc0c76b399e256b8fede92b1c69e4f4e" // address: 0xc89D42189f0450C2b2c3c61f58Ec5d628176A1E7
)

// ganacheContainer represents the ganache container type used in the module
type ganacheContainer struct {
	TContainer
}

func (c *ganacheContainer) ExposedEndpoint(ctx context.Context) (string, error) {
	endpoint, err := c.TContainer.ExposedEndpoint(ctx)
	if err != nil {
		return "", err
	}

	return "ws://" + endpoint, nil
}

func (c *ganacheContainer) UnexposedEndpoint(ctx context.Context) (string, error) {
	endpoint, err := c.TContainer.UnexposedEndpoint(ctx)
	if err != nil {
		return "", err
	}

	return "ws://" + endpoint, nil
}

// setupGanache creates an instance of the ganache container type
func setupGanache(ctx context.Context, chainId string, opts ...containerOption) (*ganacheContainer, error) {
	req := testcontainers.ContainerRequest{
		Name:         fmt.Sprintf("ganache-%d", time.Now().Unix()),
		Image:        GanacheImage,
		Env:          map[string]string{},
		Files:        []testcontainers.ContainerFile{},
		Networks:     []string{"test-network"},
		ExposedPorts: []string{},
		Cmd: []string{`--wallet.hdPath`, WalletHDPath,
			`--wallet.mnemonic`, WalletMnemonic,
			`--chain.chainId`, chainId},
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

	return &ganacheContainer{TContainer{
		Container:     container,
		ContainerPort: GanachePort,
	}}, nil
}

func StartGanacheDockerService(t *testing.T, ctx context.Context, chainId string) *ganacheContainer {
	//t.Helper()

	container, err := setupGanache(ctx,
		chainId,
		WithNetwork(kwilTestNetworkName),
		WithExposedPorts([]string{GanachePort}),
		WithWaitStrategy(
			wait.ForLog("RPC Listening on 0.0.0.0:8545")))

	require.NoError(t, err, "Could not start ganache container")

	t.Cleanup(func() {
		require.NoError(t, container.Terminate(ctx), "Could not stop ganache container")
	})

	err = container.ShowPortInfo(ctx)
	require.NoError(t, err)

	return container
}
