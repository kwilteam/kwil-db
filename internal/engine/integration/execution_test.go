//go:build pglive

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
				require.Equal(t, testdata.TestSchema.Name, dbs[0].Name)

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
				signer := "signer"

				ctx := context.Background()
				_, err := global.Execute(ctx, &types.ExecutionData{
					Dataset:   testdata.TestSchema.DBID(),
					Procedure: testdata.ProcedureCreateUser.Name,
					Mutative:  true,
					Args:      []any{1, "satoshi", 42},
					Signer:    []byte(signer),
					Caller:    signer,
				})
				require.NoError(t, err)

				_, err = global.Execute(ctx, &types.ExecutionData{
					Dataset:   testdata.TestSchema.DBID(),
					Procedure: testdata.ProcedureCreatePost.Name,
					Mutative:  true,
					Args:      []any{1, "Bitcoin!", "The Bitcoin Whitepaper", "9/31/2008"},
					Signer:    []byte(signer),
					Caller:    signer,
				})
				require.NoError(t, err)
			},
			after: func(t *testing.T, global *execution.GlobalContext, reg *registry.Registry) {
				ctx := context.Background()

				res, err := global.Execute(ctx, &types.ExecutionData{
					Dataset:   testdata.TestSchema.DBID(),
					Procedure: testdata.ProcedureGetPosts.Name,
					Mutative:  false,
					Args:      []any{"satoshi"},
					Signer:    []byte("signer"),
					Caller:    "signer",
				})
				require.NoError(t, err)

				require.Equal(t, res.ReturnedColumns, []string{"id", "title", "content", "post_date", "author"})
				require.Equal(t, len(res.Rows), 1)

				row1 := res.Rows[0]

				require.Equal(t, row1[0], int64(1))
				require.Equal(t, row1[1], "Bitcoin!")
				require.Equal(t, row1[2], "The Bitcoin Whitepaper")
				require.Equal(t, row1[3], "9/31/2008")
				require.Equal(t, row1[4], "satoshi")

				dbid := testdata.TestSchema.DBID()
				// pgSchema := types.DBIDSchema(dbid)
				res2, err := global.Query(ctx, dbid, `SELECT * from posts;`) // or do we require callers to set qualify schema like `SELECT * from `+pgSchema+`.posts;` ?
				require.NoError(t, err)

				require.Equal(t, res2.ReturnedColumns, []string{"id", "title", "content", "author_id", "post_date"})
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

				_, err := global.Execute(ctx, &types.ExecutionData{
					Dataset:   testdata.TestSchema.DBID(),
					Procedure: testdata.ProcedureCreatePost.Name,
					Mutative:  true,
					Args:      []any{1, "Bitcoin!", "The Bitcoin Whitepaper", "9/31/2008"},
					Signer:    []byte("signer"),
					Caller:    "signer",
				})
				require.ErrorIs(t, err, registry.ErrRegistryNotWritable)
			},
		},
		{
			name: "calling outside of a commit",
			ses1: func(t *testing.T, global *execution.GlobalContext, reg *registry.Registry) {
				ctx := context.Background()
				err := global.CreateDataset(ctx, testdata.TestSchema, testdata.TestSchema.Owner)
				require.NoError(t, err)

				_, err = global.Execute(ctx, &types.ExecutionData{
					Dataset:   testdata.TestSchema.DBID(),
					Procedure: testdata.ProcedureCreateUser.Name,
					Mutative:  true,
					Args:      []any{1, "satoshi", 42},
					Signer:    []byte("signer"),
					Caller:    "signer",
				})
				require.NoError(t, err)
			},
			after: func(t *testing.T, global *execution.GlobalContext, reg *registry.Registry) {
				ctx := context.Background()

				users, err := global.Execute(ctx, &types.ExecutionData{
					Dataset:   testdata.TestSchema.DBID(),
					Procedure: testdata.ProcedureGetUserByAddress.Name,
					Mutative:  false,
					Args:      []any{"signer"},
					Signer:    []byte("signer"),
					Caller:    "signer",
				})
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

				_, err = global.Execute(ctx, &types.ExecutionData{
					Dataset:   testdata.TestSchema.DBID(),
					Procedure: testdata.ProcedureCreateUser.Name,
					Mutative:  true,
					Args:      []any{1, "satoshi", 42},
					Signer:    []byte("signer"),
					Caller:    "signer",
				})
				require.NoError(t, err)
			},
			after: func(t *testing.T, global *execution.GlobalContext, reg *registry.Registry) {
				ctx := context.Background()

				users, err := global.Execute(ctx, &types.ExecutionData{
					Dataset:   testdata.TestSchema.DBID(),
					Procedure: testdata.ProcedureGetUserByAddress.Name,
					Mutative:  false,
					Args:      []any{"signer"},
					Signer:    []byte("signer"),
					Caller:    "signer",
				})
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

				_, err = global.Execute(ctx, &types.ExecutionData{
					Dataset:   testdata.TestSchema.DBID(),
					Procedure: testdata.ProcedureAdminDeleteUser.Name,
					Mutative:  true,
					Args:      []any{1},
					Signer:    []byte("signer"),
					Caller:    "signer",
				})
				require.Error(t, err)

				_, err = global.Execute(ctx, &types.ExecutionData{
					Dataset:   testdata.TestSchema.DBID(),
					Procedure: testdata.ProcedureAdminDeleteUser.Name,
					Mutative:  true,
					Args:      []any{1},
					Signer:    testdata.TestSchema.Owner,
					Caller:    string(testdata.TestSchema.Owner),
				})
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
				_, err = global.Execute(ctx, &types.ExecutionData{
					Dataset:   testdata.TestSchema.DBID(),
					Procedure: testdata.ProcedurePrivate.Name,
					Mutative:  true,
					Args:      []any{},
					Signer:    []byte("signer"),
					Caller:    "signer",
				})
				require.Error(t, err)

				// calling a public which calls private succeeds
				_, err = global.Execute(ctx, &types.ExecutionData{
					Dataset:   testdata.TestSchema.DBID(),
					Procedure: testdata.ProcedureCallsPrivate.Name,
					Mutative:  true,
					Args:      []any{},
					Signer:    []byte("signer"),
					Caller:    "signer",
				})
				require.NoError(t, err)
			},
		},
		{
			name: "deploy and call at the same time",
			ses1: func(t *testing.T, global *execution.GlobalContext, reg *registry.Registry) {
				ctx := context.Background()

				err := global.CreateDataset(ctx, testdata.TestSchema, testdata.TestSchema.Owner)
				require.NoError(t, err)

				_, err = global.Execute(ctx, &types.ExecutionData{
					Dataset:   testdata.TestSchema.DBID(),
					Procedure: testdata.ProcedureCreateUser.Name,
					Mutative:  true,
					Args:      []any{1, "satoshi", 42},
					Signer:    []byte("signer"),
					Caller:    "signer",
				})
				require.NoError(t, err)

				_, err = global.Execute(ctx, &types.ExecutionData{
					Dataset:   testdata.TestSchema.DBID(),
					Procedure: testdata.ProcedureGetUserByAddress.Name,
					Mutative:  false,
					Args:      []any{"signer"},
					Signer:    []byte("signer"),
					Caller:    "signer",
				})
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
		{
			name: "case insensitive",
			ses1: func(t *testing.T, global *execution.GlobalContext, reg *registry.Registry) {
				ctx := context.Background()

				schema := *caseSchema

				err := global.CreateDataset(ctx, &schema, testdata.TestSchema.Owner)
				require.NoError(t, err)

				caller := "signer"
				signer := []byte("signer")

				_, err = global.Execute(ctx, &types.ExecutionData{
					Dataset:   schema.DBID(),
					Procedure: "CREATE_USER",
					Mutative:  true,
					Args:      []any{1, "satoshi"},
					Signer:    []byte(caller),
					Caller:    string(signer),
				})
				require.NoError(t, err)

				_, err = global.Execute(ctx, &types.ExecutionData{
					Dataset:   schema.DBID(),
					Procedure: "CREATE_USER",
					Mutative:  true,
					Args:      []any{"2", "vitalik"},
					Signer:    []byte(caller),
					Caller:    string(signer),
				})
				require.NoError(t, err)

				_, err = global.Execute(ctx, &types.ExecutionData{
					Dataset:   schema.DBID(),
					Procedure: "CREATE_FOLLOWER",
					Mutative:  true,
					Args:      []any{"satoshi", "vitalik"},
					Signer:    []byte(caller),
					Caller:    string(signer),
				})
				require.NoError(t, err)

				res, err := global.Execute(ctx, &types.ExecutionData{
					Dataset:   schema.DBID(),
					Procedure: "USE_EXTENSION",
					Mutative:  true,
					Args:      []any{1, "2"}, // math_ext.add($arg1 + $arg2, 1)
					Signer:    []byte(caller),
					Caller:    string(signer),
				})
				require.NoError(t, err)

				// "SELECT $rES as res;" will be a string because arg type
				// inference based on Go variables is only used for inline
				// expressions since postgres prepare/describe is desirable for
				// statements that actually reference a table (but this one does
				// not).
				require.Equal(t, "4", res.Rows[0][0])
				require.Equal(t, []string{"res"}, res.ReturnedColumns) // without the `AS res`, it would be `?column?`
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

			global, reg, _, err := setup(t)
			if err != nil {
				t.Fatal(err)
			}

			ctx := context.Background()

			idempotencyKey1 := []byte("idempotencyKey1")

			err = reg.Begin(ctx, idempotencyKey1)
			require.NoError(t, err)

			defer reg.Cancel(ctx)

			test.ses1(t, global, reg)

			id, err := reg.Precommit(ctx)
			require.NoError(t, err)
			require.NotEmpty(t, id)

			err = reg.Commit(ctx, idempotencyKey1)
			require.NoError(t, err)

			idempotencyKey2 := []byte("idempotencyKey2")

			err = reg.Begin(ctx, idempotencyKey2)
			require.NoError(t, err)

			test.ses2(t, global, reg)

			id, err = reg.Precommit(ctx)
			require.NoError(t, err)
			require.NotEmpty(t, id)

			err = reg.Commit(ctx, idempotencyKey2)
			require.NoError(t, err)

			test.after(t, global, reg)

			err = reg.Close(ctx)
			require.NoError(t, err)
		})
	}
}

var (
	caseSchema = &types.Schema{
		Name: "case_insensITive",
		Tables: []*types.Table{
			{
				Name: "usErs",
				Columns: []*types.Column{
					{
						Name: "id",
						Type: types.INT,
						Attributes: []*types.Attribute{
							{
								Type: types.PRIMARY_KEY,
							},
						},
					},
					{
						Name: "nAMe",
						Type: types.TEXT,
					},
				},
				Indexes: []*types.Index{
					{
						Name: "usErs_name",
						Columns: []string{
							"nAmE",
						},
						Type: types.BTREE,
					},
				},
			},
			{
				Name: "fOllOwers",
				Columns: []*types.Column{
					{
						Name: "foLlOwer_id",
						Type: types.INT,
						Attributes: []*types.Attribute{
							{
								Type: types.NOT_NULL,
							},
						},
					},
					{
						Name: "fOllOwee_id",
						Type: types.INT,
						Attributes: []*types.Attribute{
							{
								Type: types.NOT_NULL,
							},
						},
					},
				},
				Indexes: []*types.Index{
					{
						Name: "fOllOwers_pk",
						Columns: []string{
							"foLlowEr_id",
							"fOllOwee_Id",
						},
						Type: types.PRIMARY,
					},
				},
				ForeignKeys: []*types.ForeignKey{
					{
						ChildKeys: []string{
							"FoLlOwer_id",
						},
						ParentKeys: []string{
							"iD",
						},
						ParentTable: "useRS",
					},
					{
						ChildKeys: []string{
							"FoLlOweE_id",
						},
						ParentKeys: []string{
							"ID",
						},
						ParentTable: "UseRS",
					},
				},
			},
		},
		Procedures: []*types.Procedure{
			{
				Name: "CrEaTe_UsEr",
				Args: []string{
					"$Id",
					"$nAmE",
				},
				Public: true,
				Statements: []string{
					"INSERT INTO UseRs (ID, nAme) VALUES ($iD, $nAME);",
				},
			},
			{
				Name: "CrEaTe_FoLlOwEr",
				Args: []string{
					"$FoLlOwer_nAme",
					"$FoLlOwee_nAme",
				},
				Public: true,
				Statements: []string{
					`INSERT INTO FollOweRS (FOLlOwer_id, FOLlOwee_id)
					VALUES (
						(SELECT ID FROM USErs WHERE NAmE = $FoLlOwer_nAME),
						(SELECT ID FROM UsErS WHERE nAME = $FoLlOwee_nAME)
					);`,
				},
			},
			{
				Name: "use_ExTension",
				Args: []string{
					"$vAl1",
					"$vAl2",
				},
				Public: true,
				Statements: []string{
					"$rEs = Math_Ext.AdD($VAl1 + $VAl2, 1);",
					"SELECT $rES as res;", // type? procedure execution is not strongly typed...
				},
			},
		},
		Extensions: []*types.Extension{
			{
				Name:  "maTh",
				Alias: "Math_Ext",
			},
		},
	}
)
