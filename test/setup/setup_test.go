package setup_test

import (
	"context"
	"testing"

	"github.com/kwilteam/kwil-db/config"
	"github.com/kwilteam/kwil-db/test/setup"
	"github.com/stretchr/testify/require"
)

// this is a simple test to ensure that the setup package is working
// TODO: we can probably delete this because it is implicitly tested by the other tests
// It is just here while I test this out
func Test_Setup(t *testing.T) {
	p := setup.SetupTests(t, &setup.TestConfig{
		ClientDriver: setup.Go,
		Network: &setup.NetworkConfig{
			Nodes: []*setup.NodeConfig{
				setup.DefaultNodeConfig(),
				setup.CustomNodeConfig(func(nc *setup.NodeConfig) {
					nc.Configure = func(c *config.Config) {
						c.P2P.Pex = false
					}
				}),
			},
			DBOwner: "0xabc",
			ExtraServices: []*setup.CustomService{
				{
					ServiceName:  "eth",
					DockerImage:  "kwildb/hardhat:latest",
					ExposedPort:  "8545",
					InternalPort: "8545",
				},
			},
		},
	})

	ctx := context.Background()

	client, err := p.Nodes[0].JSONRPCClient(t, ctx, false)
	require.NoError(t, err)

	ping, err := client.Ping(ctx)
	require.NoError(t, err)

	require.Equal(t, "pong", ping)
}
