//go:build rely_docker

// package proxy is an example of how the testing package can be used. It tests
// three contracts that are used to form a proxy pattern. An explanation of
// proxy contracts in Solidity can be found here:
// https://www.cyfrin.io/blog/upgradeable-proxy-smart-contract-pattern
package proxy

import (
	"context"
	_ "embed"
	"encoding/json"
	"testing"

	"github.com/kwilteam/kwil-db/common"
	"github.com/kwilteam/kwil-db/core/utils"
	kwilTesting "github.com/kwilteam/kwil-db/testing"
	"github.com/stretchr/testify/require"
)

// Test_Impl_1 tests the impl_1.kf file.
func Test_Impl_1(t *testing.T) {
	kwilTesting.RunSchemaTest(t, kwilTesting.SchemaTest{
		Name:        "impl_1",
		SchemaFiles: []string{"./impl_1.kf"},
		SeedStatements: map[string][]string{
			"impl_1": {
				`INSERT INTO users (id, name, address)
				 VALUES ('42f856df-b212-4bdc-a396-f8fb6eae6901'::uuid, 'satoshi', '0xAddress'),
				 ('d68e737d-708f-45f8-9311-317afcaccc63'::uuid, 'zeus', 'zeus.eth')`,
			},
		},
		TestCases: []kwilTesting.TestCase{
			{
				// should create a user - happy case
				Name:     "create user - success",
				Database: "impl_1",
				Target:   "create_user",
				Args:     []any{"gilgamesh"},
			},
			{
				// conflicting with the name "satoshi" in the "name" column,
				// which is unique.
				Name:     "conflicting username - failure",
				Database: "impl_1",
				Target:   "create_user",
				Args:     []any{"satoshi"},
				ErrMsg:   "duplicate key value",
			},
			{
				// conflicting with the wallet address provided by @caller
				// in the "address" column, which is unique
				Name:     "conflicting wallet address - failure",
				Database: "impl_1",
				Target:   "create_user",
				Args:     []any{"poseidon"},
				Caller:   "0xAddress", // same address as satoshi
				ErrMsg:   "duplicate key value",
			},
			{
				// tests get_users, expecting the users that were seeded.
				Name:     "reading a table of users - success",
				Database: "impl_1",
				Target:   "get_users",
				Returns: [][]any{
					{
						"satoshi", "0xAddress",
					},
					{
						"zeus", "zeus.eth",
					},
				},
			},
		},
	})
}

//go:embed impl_2_test.json
var impl2TestJson []byte

// Test_Impl_2 tests the impl_2.kf file.
// It uses the impl_2_test.json file to show how tests
// can be done using json files as well.
func Test_Impl_2(t *testing.T) {
	var schemaTest kwilTesting.SchemaTest
	err := json.Unmarshal(impl2TestJson, &schemaTest)
	require.NoError(t, err)

	kwilTesting.RunSchemaTest(t, schemaTest)
}

// Test_Proxy tests proxy.kf to ensure that proxy functionality
// works as expected.
func Test_Proxy(t *testing.T) {
	kwilTesting.RunSchemaTest(t, kwilTesting.SchemaTest{
		Name:        "proxy",
		SchemaFiles: []string{"./proxy.kf", "./impl_1.kf", "./impl_2.kf"},
		SeedStatements: map[string][]string{
			"impl_1": {
				`INSERT INTO users (id, name, address)
				 VALUES ('42f856df-b212-4bdc-a396-f8fb6eae6901'::uuid, 'satoshi', '0xAddress')`,
			},
		},
		// since this is a more complex test, we use the function test to
		// allow us to code arbitrary logic against the engine.
		FunctionTests: []kwilTesting.TestFunc{
			func(ctx context.Context, platform *kwilTesting.Platform) error {
				proxyDbid := utils.GenerateDBID("proxy", platform.Deployer)
				impl1Dbid := utils.GenerateDBID("impl_1", platform.Deployer)
				impl2Dbid := utils.GenerateDBID("impl_2", platform.Deployer)

				// register the owner
				_, err := platform.Engine.Procedure(ctx, platform.DB, &common.ExecutionData{
					TransactionData: common.TransactionData{
						Signer: platform.Deployer,
						Caller: string(platform.Deployer),
						TxID:   platform.Txid(),
						Height: 1,
					},
					Dataset:   proxyDbid,
					Procedure: "register_owner",
				})
				require.NoError(t, err)

				// set the proxy to schema 1
				_, err = platform.Engine.Procedure(ctx, platform.DB, &common.ExecutionData{
					TransactionData: common.TransactionData{
						Signer: platform.Deployer,
						Caller: string(platform.Deployer),
						TxID:   platform.Txid(),
						Height: 1,
					},
					Dataset:   proxyDbid,
					Procedure: "set_target",
					Args:      []any{impl1Dbid},
				})
				require.NoError(t, err)

				// get the user from schema 1
				res, err := platform.Engine.Procedure(ctx, platform.DB, &common.ExecutionData{
					TransactionData: common.TransactionData{
						Signer: platform.Deployer,
						Caller: string(platform.Deployer),
						TxID:   platform.Txid(),
						Height: 1,
					},
					Dataset:   proxyDbid,
					Procedure: "get_users",
				})
				require.NoError(t, err)

				require.EqualValues(t, [][]any{
					{"satoshi", "0xAddress"},
				}, res.Rows)

				// set the proxy to schema 2
				_, err = platform.Engine.Procedure(ctx, platform.DB, &common.ExecutionData{
					TransactionData: common.TransactionData{
						Signer: platform.Deployer,
						Caller: string(platform.Deployer),
						TxID:   platform.Txid(),
						Height: 2,
					},
					Dataset:   proxyDbid,
					Procedure: "set_target",
					Args:      []any{impl2Dbid},
				})
				require.NoError(t, err)

				// migrate schema 2 from schema 1
				_, err = platform.Engine.Procedure(ctx, platform.DB, &common.ExecutionData{
					TransactionData: common.TransactionData{
						Signer: platform.Deployer,
						Caller: string(platform.Deployer),
						TxID:   platform.Txid(),
						Height: 2,
					},
					Dataset:   impl2Dbid,
					Procedure: "migrate",
					Args:      []any{impl1Dbid, "get_users"},
				})
				require.NoError(t, err)

				// drop the old schema
				err = platform.Engine.DeleteDataset(ctx, platform.DB, impl1Dbid,
					&common.TransactionData{
						Signer: platform.Deployer,
						Caller: string(platform.Deployer),
						TxID:   platform.Txid(),
						Height: 2,
					})
				require.NoError(t, err)

				// check that the users exist in schema 2
				res, err = platform.Engine.Procedure(ctx, platform.DB, &common.ExecutionData{
					TransactionData: common.TransactionData{
						Signer: platform.Deployer,
						Caller: string(platform.Deployer),
						TxID:   platform.Txid(),
						Height: 2,
					},
					Dataset:   proxyDbid,
					Procedure: "get_users",
				})
				require.NoError(t, err)

				require.EqualValues(t, [][]any{
					{"satoshi", "0xAddress"},
				}, res.Rows)

				return nil
			},
		},
	})
}
