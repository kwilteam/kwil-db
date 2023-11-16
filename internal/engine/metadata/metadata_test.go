package metadata_test

import (
	"testing"

	"context"

	"github.com/kwilteam/kwil-db/internal/engine/metadata"
	"github.com/kwilteam/kwil-db/internal/engine/types"
	"github.com/kwilteam/kwil-db/internal/engine/types/testdata"
	sql "github.com/kwilteam/kwil-db/internal/sql"
	"github.com/stretchr/testify/require"
)

func Test_MetadataStore(t *testing.T) {
	type testCase struct {
		name string
		fn   func(t *testing.T, exec sql.ResultSetFunc, kv metadata.KV)
	}

	tests := []testCase{
		{
			name: "store tables",
			fn: func(t *testing.T, exec sql.ResultSetFunc, kv metadata.KV) {
				tbls := []*types.Table{
					testdata.TableUsers,
					testdata.TablePosts,
				}

				ctx := context.Background()
				err := metadata.CreateTables(ctx, tbls, kv, exec)
				require.NoError(t, err)

				tables, err := metadata.ListTables(ctx, kv)
				require.NoError(t, err)

				require.ElementsMatch(t, tbls, tables)
			},
		},
		{
			name: "store procedures",
			fn: func(t *testing.T, exec sql.ResultSetFunc, kv metadata.KV) {
				procs := []*types.Procedure{
					testdata.ProcedureCreateUser,
					testdata.ProcedureCreatePost,
					testdata.ProcedureGetPosts,
				}

				ctx := context.Background()
				err := metadata.StoreProcedures(ctx, procs, kv)
				require.NoError(t, err)

				procedures, err := metadata.ListProcedures(ctx, kv)
				require.NoError(t, err)

				require.ElementsMatch(t, procs, procedures)
			},
		},
		{
			name: "store extensions",
			fn: func(t *testing.T, exec sql.ResultSetFunc, kv metadata.KV) {
				exts := []*types.Extension{
					testdata.ExtensionErc20,
				}

				ctx := context.Background()
				err := metadata.StoreExtensions(ctx, exts, kv)
				require.NoError(t, err)

				extensions, err := metadata.ListExtensions(ctx, kv)
				require.NoError(t, err)

				require.ElementsMatch(t, exts, extensions)
			},
		},
	}

	for _, tt := range tests {
		ctx := context.Background()
		kv := mockKV{
			data: map[string][]byte{},
		}

		err := metadata.RunMigration(ctx, kv)
		if err != nil {
			t.Fatal(err)
		}

		tt.fn(t, func(ctx context.Context, stmt string, args map[string]any) (*sql.ResultSet, error) {
			return nil, nil
		}, kv)
	}
}

type mockKV struct {
	data map[string][]byte
}

var _ metadata.KV = mockKV{}

func (m mockKV) Set(ctx context.Context, key []byte, value []byte) error {
	m.data[string(key)] = value
	return nil
}

func (m mockKV) Get(ctx context.Context, key []byte) ([]byte, error) {
	val, ok := m.data[string(key)]
	if !ok {
		return nil, nil
	}

	return val, nil
}
