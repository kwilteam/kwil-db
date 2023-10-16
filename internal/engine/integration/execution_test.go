package integration_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/kwilteam/kwil-db/internal/engine/execution"
	"github.com/kwilteam/kwil-db/internal/engine/types"
	"github.com/kwilteam/kwil-db/internal/engine/types/testdata"
	"github.com/kwilteam/kwil-db/internal/sql/registry"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_Engine(t *testing.T) {
	type testCase struct {
		name string
		// ses1 is the first round of execution
		ses1 func(t *testing.T, global *execution.GlobalContext, reg *registry.Registry)

		// ses2 is the second round of execution
		ses2 func(t *testing.T, global *execution.GlobalContext, reg *registry.Registry)
		// after is called after the second round
		// It is not called in a session
		after func(t *testing.T, global *execution.GlobalContext, reg *registry.Registry)
	}

	tests := []testCase{
		{
			name: "create database",
			ses1: func(t *testing.T, global *execution.GlobalContext, reg *registry.Registry) {
				ctx := context.Background()
				err := global.CreateDataset(ctx, testdata.TestSchema, testdata.TestSchema.Owner)
				require.NoError(t, err)
			},
			after: func(t *testing.T, global *execution.GlobalContext, reg *registry.Registry) {
				ctx := context.Background()
				schema, err := global.GetSchema(ctx, testdata.TestSchema.DBID())
				require.NoError(t, err)

				require.EqualValues(t, testdata.TestSchema, schema)

				dbs, err := global.ListDatasets(ctx, testdata.TestSchema.Owner)
				require.NoError(t, err)

				require.Equal(t, 1, len(dbs))
				require.Equal(t, testdata.TestSchema.Name, dbs[0])

				regDbs, err := reg.List(ctx)
				require.NoError(t, err)

				require.Equal(t, 1, len(regDbs))
				require.Equal(t, testdata.TestSchema.DBID(), regDbs[0])
			},
		},
		{
			name: "drop database",
			ses1: func(t *testing.T, global *execution.GlobalContext, reg *registry.Registry) {
				ctx := context.Background()
				err := global.CreateDataset(ctx, testdata.TestSchema, testdata.TestSchema.Owner)
				require.NoError(t, err)

			},
			ses2: func(t *testing.T, global *execution.GlobalContext, reg *registry.Registry) {
				ctx := context.Background()
				err := global.DeleteDataset(ctx, testdata.TestSchema.DBID(), testdata.TestSchema.Owner)
				require.NoError(t, err)
			},
			after: func(t *testing.T, global *execution.GlobalContext, reg *registry.Registry) {
				ctx := context.Background()
				dbs, err := global.ListDatasets(ctx, testdata.TestSchema.Owner)
				require.NoError(t, err)

				require.Equal(t, 0, len(dbs))

				regDbs, err := reg.List(ctx)
				require.NoError(t, err)

				require.Equal(t, 0, len(regDbs))
			},
		},
		{
			name: "execute procedures",
			ses1: func(t *testing.T, global *execution.GlobalContext, reg *registry.Registry) {
				ctx := context.Background()
				err := global.CreateDataset(ctx, testdata.TestSchema, testdata.TestSchema.Owner)
				require.NoError(t, err)
			},
			ses2: func(t *testing.T, global *execution.GlobalContext, reg *registry.Registry) {
				signer := []byte("signer")

				ctx := context.Background()
				err := global.Execute(ctx, testdata.TestSchema.DBID(), testdata.ProcedureCreateUser.Name, signer, signer, []any{1, "satoshi", 42})
				require.NoError(t, err)

				err = global.Execute(ctx, testdata.TestSchema.DBID(), testdata.ProcedureCreatePost.Name, signer, signer, []any{1, "Bitcoin!", "The Bitcoin Whitepaper", "9/31/2008"})
				require.NoError(t, err)
			},
			after: func(t *testing.T, global *execution.GlobalContext, reg *registry.Registry) {
				ctx := context.Background()

				res, err := global.Call(ctx, testdata.TestSchema.DBID(), testdata.ProcedureGetPosts.Name, []byte("signer"), []byte("signer"), []any{"satoshi"})
				require.NoError(t, err)

				require.Equal(t, res.Columns, []string{"id", "title", "content", "post_date", "author"})
				require.Equal(t, len(res.Rows), 1)

				row1 := res.Rows[0]

				require.Equal(t, row1[0], int64(1))
				require.Equal(t, row1[1], "Bitcoin!")
				require.Equal(t, row1[2], "The Bitcoin Whitepaper")
				require.Equal(t, row1[3], "9/31/2008")
				require.Equal(t, row1[4], "satoshi")

				res2, err := global.Query(ctx, testdata.TestSchema.DBID(), `SELECT * from posts;`)
				require.NoError(t, err)

				require.Equal(t, res2.Columns, []string{"id", "title", "content", "author_id", "post_date"})
				require.Equal(t, len(res2.Rows), 1)
				require.Equal(t, res2.Rows[0], []any{int64(1), "Bitcoin!", "The Bitcoin Whitepaper", int64(1), "9/31/2008"})
			},
		},
		{
			name: "executing outside of a commit",
			ses1: func(t *testing.T, global *execution.GlobalContext, reg *registry.Registry) {
				ctx := context.Background()
				err := global.CreateDataset(ctx, testdata.TestSchema, testdata.TestSchema.Owner)
				require.NoError(t, err)
			},
			after: func(t *testing.T, global *execution.GlobalContext, reg *registry.Registry) {
				ctx := context.Background()

				err := global.Execute(ctx, testdata.TestSchema.DBID(), testdata.ProcedureCreatePost.Name, []byte("signer"), []byte("signer"), []any{1, "Bitcoin!", "The Bitcoin Whitepaper", "9/31/2008"})
				require.ErrorIs(t, err, registry.ErrRegistryNotWritable)
			},
		},
		{
			name: "calling outside of a commit",
			ses1: func(t *testing.T, global *execution.GlobalContext, reg *registry.Registry) {
				ctx := context.Background()
				err := global.CreateDataset(ctx, testdata.TestSchema, testdata.TestSchema.Owner)
				require.NoError(t, err)

				err = global.Execute(ctx, testdata.TestSchema.DBID(), testdata.ProcedureCreateUser.Name, []byte("signer"), []byte("signer"), []any{1, "satoshi", 42})
				require.NoError(t, err)
			},
			after: func(t *testing.T, global *execution.GlobalContext, reg *registry.Registry) {
				ctx := context.Background()

				users, err := global.Call(ctx, testdata.TestSchema.DBID(), testdata.ProcedureGetUserByAddress.Name, []byte("signer"), []byte("signer"), []any{[]byte("signer")})
				require.NoError(t, err)

				require.Equal(t, len(users.Rows), 1)
				require.Equal(t, []any{int64(1), "satoshi", int64(42)}, []any{users.Rows[0][0], users.Rows[0][1], users.Rows[0][2]})
			},
		},
		{
			name: "deploying database and immediately calling procedure",
			ses1: func(t *testing.T, global *execution.GlobalContext, reg *registry.Registry) {
				ctx := context.Background()

				err := global.CreateDataset(ctx, testdata.TestSchema, testdata.TestSchema.Owner)
				require.NoError(t, err)

				err = global.Execute(ctx, testdata.TestSchema.DBID(), testdata.ProcedureCreateUser.Name, []byte("signer"), []byte("signer"), []any{1, "satoshi", 42})
				require.NoError(t, err)
			},
			after: func(t *testing.T, global *execution.GlobalContext, reg *registry.Registry) {
				ctx := context.Background()

				users, err := global.Call(ctx, testdata.TestSchema.DBID(), testdata.ProcedureGetUserByAddress.Name, []byte("signer"), []byte("signer"), []any{[]byte("signer")})
				require.NoError(t, err)

				require.Equal(t, len(users.Rows), 1)
				require.Equal(t, []any{int64(1), "satoshi", int64(42)}, []any{users.Rows[0][0], users.Rows[0][1], users.Rows[0][2]})
			},
		},
		{
			name: "test failed extension init",
			ses1: func(t *testing.T, global *execution.GlobalContext, reg *registry.Registry) {
				ctx := context.Background()

				oldExtensions := []*types.Extension{}
				copy(oldExtensions, testdata.TestSchema.Extensions)

				testdata.TestSchema.Extensions = append(testdata.TestSchema.Extensions,
					&types.Extension{
						Name: "math",
						Initialization: []*types.ExtensionConfig{
							{
								Key:   "fail",
								Value: "true",
							},
						},
						Alias: "fail_math",
					},
				)

				err := global.CreateDataset(ctx, testdata.TestSchema, testdata.TestSchema.Owner)
				require.Error(t, err)

				testdata.TestSchema.Extensions = oldExtensions

				err = global.CreateDataset(ctx, testdata.TestSchema, testdata.TestSchema.Owner)
				assert.NoError(t, err)
			},
		},
		{
			name: "owner only action",
			ses1: func(t *testing.T, global *execution.GlobalContext, reg *registry.Registry) {
				ctx := context.Background()

				err := global.CreateDataset(ctx, testdata.TestSchema, testdata.TestSchema.Owner)
				require.NoError(t, err)

				err = global.Execute(ctx, testdata.TestSchema.DBID(), testdata.ProcedureAdminDeleteUser.Name, []byte("wrong_signer"), []byte("wrong_signer"), []any{1})
				require.Error(t, err)

				err = global.Execute(ctx, testdata.TestSchema.DBID(), testdata.ProcedureAdminDeleteUser.Name, testdata.TestSchema.Owner, testdata.TestSchema.Owner, []any{1})
				require.NoError(t, err)
			},
		},
		{
			name: "private action",
			ses1: func(t *testing.T, global *execution.GlobalContext, reg *registry.Registry) {
				ctx := context.Background()

				err := global.CreateDataset(ctx, testdata.TestSchema, testdata.TestSchema.Owner)
				require.NoError(t, err)

				// calling private fails
				err = global.Execute(ctx, testdata.TestSchema.DBID(), testdata.ProcedurePrivate.Name, testdata.TestSchema.Owner, testdata.TestSchema.Owner, []any{})
				require.Error(t, err)

				// calling a public which calls private succeeds
				err = global.Execute(ctx, testdata.TestSchema.DBID(), testdata.ProcedureCallsPrivate.Name, testdata.TestSchema.Owner, testdata.TestSchema.Owner, []any{})
				require.NoError(t, err)
			},
		},
		{
			name: "deploy and call at the same time",
			ses1: func(t *testing.T, global *execution.GlobalContext, reg *registry.Registry) {
				ctx := context.Background()

				err := global.CreateDataset(ctx, testdata.TestSchema, testdata.TestSchema.Owner)
				require.NoError(t, err)

				err = global.Execute(ctx, testdata.TestSchema.DBID(), testdata.ProcedureCreateUser.Name, []byte("signer"), []byte("signer"), []any{1, "satoshi", 42})
				require.NoError(t, err)

				_, err = global.Call(ctx, testdata.TestSchema.DBID(), testdata.ProcedureGetUserByAddress.Name, []byte("signer"), []byte("signer"), []any{[]byte("signer")})
				require.Error(t, err)
			},
		},
		{
			name: "deploy many databases",
			ses1: func(t *testing.T, global *execution.GlobalContext, reg *registry.Registry) {
				ctx := context.Background()

				for i := 0; i < 10; i++ {
					newSchema := *testdata.TestSchema
					newSchema.Name = testdata.TestSchema.Name + fmt.Sprint(i)

					err := global.CreateDataset(ctx, &newSchema, testdata.TestSchema.Owner)
					require.NoError(t, err)
				}
			},
		},
		{
			name: "deploying and immediately dropping",
			ses1: func(t *testing.T, global *execution.GlobalContext, reg *registry.Registry) {
				ctx := context.Background()

				err := global.CreateDataset(ctx, testdata.TestSchema, testdata.TestSchema.Owner)
				require.NoError(t, err)

				err = global.DeleteDataset(ctx, testdata.TestSchema.DBID(), testdata.TestSchema.Owner)
				require.NoError(t, err)
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if test.ses1 == nil {
				test.ses1 = func(t *testing.T, global *execution.GlobalContext, reg *registry.Registry) {}
			}
			if test.ses2 == nil {
				test.ses2 = func(t *testing.T, global *execution.GlobalContext, reg *registry.Registry) {}
			}
			if test.after == nil {
				test.after = func(t *testing.T, global *execution.GlobalContext, reg *registry.Registry) {}
			}

			global, reg, err := setup(t)
			if err != nil {
				t.Fatal(err)
			}

			ctx := context.Background()

			idempotencyKey1 := []byte("idempotencyKey1")

			err = reg.Begin(ctx, idempotencyKey1)
			require.NoError(t, err)

			test.ses1(t, global, reg)

			_, err = reg.Commit(ctx, idempotencyKey1)
			require.NoError(t, err)

			idempotencyKey2 := []byte("idempotencyKey2")

			err = reg.Begin(ctx, idempotencyKey2)
			require.NoError(t, err)

			test.ses2(t, global, reg)

			_, err = reg.Commit(ctx, idempotencyKey2)
			require.NoError(t, err)

			test.after(t, global, reg)

			err = reg.Close()
			require.NoError(t, err)
		})
	}
}
