package acceptance

import (
	"context"
	_ "embed"
	"fmt"
	"math/big"
	"testing"

	"github.com/kwilteam/kwil-db/config"
	ctypes "github.com/kwilteam/kwil-db/core/client/types"
	"github.com/kwilteam/kwil-db/core/crypto"
	"github.com/kwilteam/kwil-db/core/crypto/auth"
	"github.com/kwilteam/kwil-db/core/types"
	"github.com/kwilteam/kwil-db/test"
	"github.com/kwilteam/kwil-db/test/setup"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TODO:
// - transfer
// - engine
// - type roundtripping
// - private rpc

var (
	//go:embed users.sql
	usersSchema  string
	UserPrivkey1 = func() *crypto.Secp256k1PrivateKey {
		privk, err := crypto.Secp256k1PrivateKeyFromHex("f1aa5a7966c3863ccde3047f6a1e266cdc0c76b399e256b8fede92b1c69e4f4e")
		if err != nil {
			panic(err)
		}
		return privk
	}()
)

// setupSingleNodeClient creates a single node network for testing,
// and returns the client
func setupSingleNodeClient(t *testing.T, ctx context.Context, d setup.ClientDriver, usingKGW bool) setup.JSONRPCClient {
	t.Helper()

	ident, err := auth.EthSecp256k1Authenticator{}.Identifier(crypto.EthereumAddressFromPubKey(UserPrivkey1.Public().(*crypto.Secp256k1PublicKey)))
	require.NoError(t, err)

	testnet := setup.SetupTests(t, &setup.TestConfig{
		ClientDriver: d,
		Network: &setup.NetworkConfig{
			DBOwner: ident,
			Nodes: []*setup.NodeConfig{
				setup.DefaultNodeConfig(),
			},
		},
	})
	return testnet.Nodes[0].JSONRPCClient(t, ctx, &setup.ClientOptions{
		UsingKGW:   usingKGW,
		PrivateKey: UserPrivkey1,
	})
}
func Test_Transfer(t *testing.T) {
	ctx := context.Background()

	for _, driver := range setup.AllDrivers {
		if driver == setup.CLI {
			continue // TODO: delete this once it works for jsonrpc
		}
		t.Run("transfer_"+driver.String(), func(t *testing.T) {
			userPrivateKey, _, err := crypto.GenerateSecp256k1Key(nil)
			require.NoError(t, err)

			// helper function for getting the address of a private key
			address := func(p crypto.PrivateKey) types.HexBytes {
				secp, ok := p.(*crypto.Secp256k1PrivateKey)
				require.True(t, ok)

				return crypto.EthereumAddressFromPubKey(secp.Public().(*crypto.Secp256k1PublicKey))
			}

			stringAddress := func(p crypto.PrivateKey) string {
				addr := address(p)
				val, err := auth.EthSecp256k1Authenticator{}.Identifier(addr)
				require.NoError(t, err)
				return val
			}

			testnet := setup.SetupTests(t, &setup.TestConfig{
				ClientDriver: driver,
				Network: &setup.NetworkConfig{
					ConfigureGenesis: func(gc *config.GenesisConfig) {
						gc.DisabledGasCosts = false

						// giving gas to the user
						gc.Allocs = append(gc.Allocs, config.GenesisAlloc{
							ID:      config.KeyHexBytes{HexBytes: address(userPrivateKey)},
							KeyType: crypto.KeyTypeSecp256k1.String(),
							Amount:  big.NewInt(1000000000000000000),
						})
					},
					Nodes: []*setup.NodeConfig{
						setup.DefaultNodeConfig(),
					},
					DBOwner: stringAddress(userPrivateKey),
				},
			})

			// user 1 will send funds to user 2. User 2 will check that they received the funds
			user1 := testnet.Nodes[0].JSONRPCClient(t, ctx, &setup.ClientOptions{
				PrivateKey: userPrivateKey,
			})

			// user 1 creates an action, which user 2 will call to test they have funds
			tx, err := user1.ExecuteSQL(ctx, "CREATE ACTION do_something() public {}", nil, opts)
			require.NoError(t, err)
			test.ExpectTxSuccess(t, user1, ctx, tx)

			// auto-generate the private key for user 2
			user2 := testnet.Nodes[0].JSONRPCClient(t, ctx, nil)

			// user 2 tries to execute, gets rejected from mempool because no gas
			_, err = user2.Execute(ctx, "", "do_something", nil, opts)
			require.Error(t, err)
			require.Contains(t, err.Error(), "insufficient balances")

			tx, err = user1.Transfer(ctx, &types.AccountID{
				Identifier: address(user2.PrivateKey()),
				KeyType:    crypto.KeyTypeSecp256k1,
			}, big.NewInt(100000000000000000))
			require.NoError(t, err)
			test.ExpectTxSuccess(t, user1, ctx, tx)

			// user 2 tries to execute, works because they have gas
			tx, err = user2.Execute(ctx, "", "do_something", nil, opts)
			require.NoError(t, err)
			test.ExpectTxSuccess(t, user2, ctx, tx)
		})
	}
}

// TODO: delete opts once we fix the race condition where query tx returns an error
var opts = ctypes.WithSyncBroadcast(true)

func Test_Engine(t *testing.T) {
	for _, driver := range setup.AllDrivers {
		// if driver == setup.Go {
		// 	continue // TODO: delete this once it works for jsonrpc
		// }
		t.Run("engine_"+driver.String(), func(t *testing.T) {
			ctx := context.Background()
			client := setupSingleNodeClient(t, ctx, driver, false)

			// deploy the schema
			tx, err := client.ExecuteSQL(ctx, usersSchema, nil, opts)
			require.NoError(t, err)
			test.ExpectTxSuccess(t, client, ctx, tx)

			// create two profiles: satoshi and megatron
			tx, err = client.Execute(ctx, "", "create_profile", [][]any{
				{"satoshi", 32, "father of $btc"},
				{"megatron", 1000000, "leader of the decepticons"},
			}, opts)
			require.NoError(t, err)
			test.ExpectTxSuccess(t, client, ctx, tx)

			// create three posts, all responding to each other
			tx, err = client.Execute(ctx, "", "create_post", [][]any{
				{"satoshi", "hello world", nil},
			}, opts)
			require.NoError(t, err)
			test.ExpectTxSuccess(t, client, ctx, tx)

			satoshiPostUUID, err := getLatestPostID(ctx, client, "satoshi")
			require.NoError(t, err)

			tx, err = client.Execute(ctx, "", "create_post", [][]any{
				{"megatron", "hello satoshi", satoshiPostUUID},
			}, opts)
			require.NoError(t, err)
			test.ExpectTxSuccess(t, client, ctx, tx)

			megatronPostUUID, err := getLatestPostID(ctx, client, "megatron")
			require.NoError(t, err)

			tx, err = client.Execute(ctx, "", "create_post", [][]any{
				{"satoshi", "go back to cybertron", megatronPostUUID},
			}, opts)
			require.NoError(t, err)
			test.ExpectTxSuccess(t, client, ctx, tx)

			// testing recursive CTEs by getting the post chain
			res, err := client.Call(ctx, "", "get_thread", []any{satoshiPostUUID, 5})
			require.NoError(t, err)

			// 3 posts in the chain, and get_thread does not include the root post
			require.Len(t, res.QueryResult.Values, 2)

			assert.Equal(t, "hello satoshi", res.QueryResult.Values[0][1])
			assert.Equal(t, "go back to cybertron", res.QueryResult.Values[1][1])
		})
	}
}

// getLatestPostID returns the latest post from a user
func getLatestPostID(ctx context.Context, client setup.JSONRPCClient, user string) (id *types.UUID, err error) {
	res, err := client.Call(ctx, "", "get_posts", []any{user})
	if err != nil {
		return nil, err
	}

	if len(res.QueryResult.Values) == 0 {
		return nil, fmt.Errorf("no posts found for user %s", user)
	}

	str, ok := res.QueryResult.Values[0][0].(string)
	if !ok {
		return nil, fmt.Errorf("unexpected type for post ID: %T", res.QueryResult.Values[0][0])
	}

	return types.ParseUUID(str)
}
