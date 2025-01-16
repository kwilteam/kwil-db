package acceptance

import (
	"context"
	_ "embed"
	"testing"

	"github.com/kwilteam/kwil-db/test"
	"github.com/kwilteam/kwil-db/test/setup"
	"github.com/stretchr/testify/require"
)

// TODO:
// - transfer
// - engine
// - type roundtripping
// - private rpc

const dbOwner = "db_owner"

// setupSingleNodeClient creates a single node network for testing,
// and returns the client
func setupSingleNodeClient(t *testing.T, ctx context.Context, d setup.ClientDriver, usingKGW bool) setup.JSONRPCClient {
	t.Helper()

	testnet := setup.SetupTests(t, &setup.TestConfig{
		ClientDriver: d,
		Network: &setup.NetworkConfig{
			Nodes: []*setup.NodeConfig{
				setup.DefaultNodeConfig(),
			},
		},
	})
	return testnet.Nodes[0].JSONRPCClient(t, ctx, &setup.ClientOptions{
		UsingKGW: usingKGW,
	})
}

// // TODO: come back to this once genesis allocs are fixed
// func Test_Transfer(t *testing.T) {
// 	ctx := context.Background()

// 	for _, driver := range setup.AllDrivers {
// 		t.Run("transfer_"+driver.String(), func(t *testing.T) {
// 			userPrivateKey, _, err := crypto.GenerateSecp256k1Key(nil)
// 			require.NoError(t, err)

// 			// helper function for getting the address of a private key
// 			address := func(p crypto.PrivateKey) string {
// 				secp, ok := p.(*crypto.Secp256k1PrivateKey)
// 				require.True(t, ok)

// 				ident, err := auth.Secp25k1Authenticator{}.Identifier(secp.Public().Bytes())
// 				require.NoError(t, err)

// 				return ident
// 			}

// 			testnet := setup.SetupTests(t, &setup.TestConfig{
// 				ClientDriver: driver,
// 				Network: &setup.NetworkConfig{
// 					ConfigureGenesis: func(gc *config.GenesisConfig) {
// 						gc.DisabledGasCosts = false

// 						// giving gas to the user
// 						gc.Allocs[address(userPrivateKey)] = big.NewInt(1000000000)
// 					},
// 					Nodes: []*setup.NodeConfig{
// 						setup.DefaultNodeConfig(),
// 					},
// 				},
// 			})

// 			// user 1 will send funds to user 2. User 2 will check that they received the funds
// 			user1 := testnet.Nodes[0].JSONRPCClient(t, ctx, &setup.ClientOptions{
// 				PrivateKey: userPrivateKey,
// 			})

// 			// auto-generate the private key for user 2
// 			user2 := testnet.Nodes[0].JSONRPCClient(t, ctx, nil)

// 			user2Address := address(user2.PrivateKey())

// 			tx, err := user1.Transfer(ctx, user2Address, big.NewInt(1000000))
// 			require.NoError(t, err)

// 			test.ExpectTxSuccess(t, user1, ctx, tx)
// 		})
// 	}
// }

//go:embed users.sql
var usersSchema string

func Test_Engine(t *testing.T) {
	for _, driver := range setup.AllDrivers {
		if driver == setup.CLI {
			continue // TODO: delete this once it works for jsonrpc
		}
		t.Run("engine_"+driver.String(), func(t *testing.T) {
			ctx := context.Background()
			client := setupSingleNodeClient(t, ctx, driver, false)

			tx, err := client.ExecuteSQL(ctx, usersSchema, nil)
			require.NoError(t, err)
			test.ExpectTxSuccess(t, client, ctx, tx)
		})
	}
}
