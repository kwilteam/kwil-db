//go:build pglive

package integration_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/kwilteam/kwil-db/common"
	"github.com/kwilteam/kwil-db/common/sql"
	"github.com/kwilteam/kwil-db/core/types"
	"github.com/kwilteam/kwil-db/core/types/testdata"
	"github.com/kwilteam/kwil-db/internal/engine/execution"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_Engine(t *testing.T) {
	type testCase struct {
		name string
		// ses1 is the first round of execution
		ses1 func(t *testing.T, global *execution.GlobalContext, tx sql.DB)

		// ses2 is the second round of execution
		ses2 func(t *testing.T, global *execution.GlobalContext, tx sql.DB)
		// after is called after the second round
		// It is not called in a session, and therefore can only read from the database.
		after func(t *testing.T, global *execution.GlobalContext, tx sql.DB)
	}

	tests := []testCase{
		{
			name: "create database",
			ses1: func(t *testing.T, global *execution.GlobalContext, tx sql.DB) {
				ctx := context.Background()
				err := global.CreateDataset(ctx, tx, testdata.TestSchema, &common.TransactionData{
					Signer: testdata.TestSchema.Owner,
					Caller: string(testdata.TestSchema.Owner),
					TxID:   "txid1",
				})
				require.NoError(t, err)
			},
			after: func(t *testing.T, global *execution.GlobalContext, tx sql.DB) {
				schema, err := global.GetSchema(testdata.TestSchema.DBID())
				require.NoError(t, err)

				require.EqualValues(t, testdata.TestSchema, schema)

				dbs, err := global.ListDatasets(testdata.TestSchema.Owner)
				require.NoError(t, err)

				require.Equal(t, 1, len(dbs))
				require.Equal(t, testdata.TestSchema.Name, dbs[0].Name)
			},
		},
		{
			name: "drop database",
			ses1: func(t *testing.T, global *execution.GlobalContext, tx sql.DB) {
				ctx := context.Background()
				err := global.CreateDataset(ctx, tx, testdata.TestSchema, &common.TransactionData{
					Signer: testdata.TestSchema.Owner,
					Caller: string(testdata.TestSchema.Owner),
					TxID:   "txid1",
				})
				require.NoError(t, err)

			},
			ses2: func(t *testing.T, global *execution.GlobalContext, tx sql.DB) {
				ctx := context.Background()
				err := global.DeleteDataset(ctx, tx, testdata.TestSchema.DBID(), &common.TransactionData{
					Signer: testdata.TestSchema.Owner,
					Caller: string(testdata.TestSchema.Owner),
					TxID:   "txid2",
				})
				require.NoError(t, err)
			},
			after: func(t *testing.T, global *execution.GlobalContext, tx sql.DB) {
				dbs, err := global.ListDatasets(testdata.TestSchema.Owner)
				require.NoError(t, err)

				require.Equal(t, 0, len(dbs))
			},
		},
		{
			name: "execute procedures",
			ses1: func(t *testing.T, global *execution.GlobalContext, tx sql.DB) {
				ctx := context.Background()
				err := global.CreateDataset(ctx, tx, testdata.TestSchema, &common.TransactionData{
					Signer: testdata.TestSchema.Owner,
					Caller: string(testdata.TestSchema.Owner),
					TxID:   "txid1",
				})
				require.NoError(t, err)
			},
			ses2: func(t *testing.T, global *execution.GlobalContext, tx sql.DB) {
				signer := "signer"

				ctx := context.Background()
				_, err := global.Procedure(ctx, tx, &common.ExecutionData{
					Dataset:   testdata.TestSchema.DBID(),
					Procedure: testdata.ActionCreateUser.Name,
					Args:      []any{1, "satoshi", 42},
					TransactionData: common.TransactionData{
						Signer: []byte(signer),
						Caller: signer,
					},
				})
				require.NoError(t, err)

				_, err = global.Procedure(ctx, tx, &common.ExecutionData{
					Dataset:   testdata.TestSchema.DBID(),
					Procedure: testdata.ActionCreatePost.Name,
					Args:      []any{1, "Bitcoin!", "The Bitcoin Whitepaper", "9/31/2008"},
					TransactionData: common.TransactionData{
						Signer: []byte(signer),
						Caller: signer,
					},
				})
				require.NoError(t, err)
			},
			after: func(t *testing.T, global *execution.GlobalContext, tx sql.DB) {
				ctx := context.Background()

				res, err := global.Procedure(ctx, tx, &common.ExecutionData{
					Dataset:   testdata.TestSchema.DBID(),
					Procedure: testdata.ActionGetPosts.Name,
					Args:      []any{"satoshi"},
					TransactionData: common.TransactionData{
						Signer: []byte("signer"),
						Caller: "signer",
					},
				})
				require.NoError(t, err)

				require.Equal(t, res.Columns, []string{"id", "title", "content", "post_date", "author"})
				require.Equal(t, len(res.Rows), 1)

				row1 := res.Rows[0]

				require.Equal(t, row1[0], int64(1))
				require.Equal(t, row1[1], "Bitcoin!")
				require.Equal(t, row1[2], "The Bitcoin Whitepaper")
				require.Equal(t, row1[3], "9/31/2008")
				require.Equal(t, row1[4], "satoshi")

				dbid := testdata.TestSchema.DBID()
				// pgSchema := common.DBIDSchema(dbid)
				res2, err := global.Execute(ctx, tx, dbid, `SELECT * from posts;`, nil) // or do we require callers to set qualify schema like `SELECT * from `+pgSchema+`.posts;` ?
				require.NoError(t, err)

				require.Equal(t, res2.Columns, []string{"id", "title", "content", "author_id", "post_date"})
				require.Equal(t, len(res2.Rows), 1)
				require.Equal(t, res2.Rows[0], []any{int64(1), "Bitcoin!", "The Bitcoin Whitepaper", int64(1), "9/31/2008"})
			},
		},
		{
			name: "executing outside of a commit",
			ses1: func(t *testing.T, global *execution.GlobalContext, tx sql.DB) {
				ctx := context.Background()
				err := global.CreateDataset(ctx, tx, testdata.TestSchema, &common.TransactionData{
					Signer: testdata.TestSchema.Owner,
					Caller: string(testdata.TestSchema.Owner),
					TxID:   "txid1",
				})
				require.NoError(t, err)
			},
			after: func(t *testing.T, global *execution.GlobalContext, tx sql.DB) {
				ctx := context.Background()

				_, err := global.Procedure(ctx, tx, &common.ExecutionData{
					Dataset:   testdata.TestSchema.DBID(),
					Procedure: testdata.ActionCreatePost.Name,
					Args:      []any{1, "Bitcoin!", "The Bitcoin Whitepaper", "9/31/2008"},
					TransactionData: common.TransactionData{
						Signer: []byte("signer"),
						Caller: "signer",
					},
				})
				require.NotNil(t, err)
			},
		},
		{
			name: "calling outside of a commit",
			ses1: func(t *testing.T, global *execution.GlobalContext, tx sql.DB) {
				ctx := context.Background()
				err := global.CreateDataset(ctx, tx, testdata.TestSchema, &common.TransactionData{
					Signer: testdata.TestSchema.Owner,
					Caller: string(testdata.TestSchema.Owner),
					TxID:   "txid1",
				})
				require.NoError(t, err)

				_, err = global.Procedure(ctx, tx, &common.ExecutionData{
					Dataset:   testdata.TestSchema.DBID(),
					Procedure: testdata.ActionCreateUser.Name,
					Args:      []any{1, "satoshi", 42},
					TransactionData: common.TransactionData{
						Signer: []byte("signer"),
						Caller: "signer",
					},
				})
				require.NoError(t, err)
			},
			after: func(t *testing.T, global *execution.GlobalContext, tx sql.DB) {
				ctx := context.Background()

				users, err := global.Procedure(ctx, tx, &common.ExecutionData{
					Dataset:   testdata.TestSchema.DBID(),
					Procedure: testdata.ActionGetUserByAddress.Name,
					Args:      []any{"signer"},
					TransactionData: common.TransactionData{
						Signer: []byte("signer"),
						Caller: "signer",
					},
				})
				require.NoError(t, err)

				require.Equal(t, len(users.Rows), 1)
				require.Equal(t, []any{int64(1), "satoshi", int64(42)}, []any{users.Rows[0][0], users.Rows[0][1], users.Rows[0][2]})
			},
		},
		{
			name: "deploying database and immediately calling procedure",
			ses1: func(t *testing.T, global *execution.GlobalContext, tx sql.DB) {
				ctx := context.Background()

				err := global.CreateDataset(ctx, tx, testdata.TestSchema, &common.TransactionData{
					Signer: testdata.TestSchema.Owner,
					Caller: string(testdata.TestSchema.Owner),
					TxID:   "txid1",
				})
				require.NoError(t, err)

				_, err = global.Procedure(ctx, tx, &common.ExecutionData{
					Dataset:   testdata.TestSchema.DBID(),
					Procedure: testdata.ActionCreateUser.Name,
					Args:      []any{1, "satoshi", 42},
					TransactionData: common.TransactionData{
						Signer: []byte("signer"),
						Caller: "signer",
					},
				})
				require.NoError(t, err)
			},
			after: func(t *testing.T, global *execution.GlobalContext, tx sql.DB) {
				ctx := context.Background()

				users, err := global.Procedure(ctx, tx, &common.ExecutionData{
					Dataset:   testdata.TestSchema.DBID(),
					Procedure: testdata.ActionGetUserByAddress.Name,
					Args:      []any{"signer"},
					TransactionData: common.TransactionData{
						Signer: []byte("signer"),
						Caller: "signer",
					},
				})
				require.NoError(t, err)

				require.Equal(t, len(users.Rows), 1)
				require.Equal(t, []any{int64(1), "satoshi", int64(42)}, []any{users.Rows[0][0], users.Rows[0][1], users.Rows[0][2]})
			},
		},
		{
			name: "test failed extension init",
			ses1: func(t *testing.T, global *execution.GlobalContext, tx sql.DB) {
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

				err := global.CreateDataset(ctx, tx, testdata.TestSchema, &common.TransactionData{
					Signer: testdata.TestSchema.Owner,
					Caller: string(testdata.TestSchema.Owner),
					TxID:   "txid1",
				})
				require.Error(t, err)

				testdata.TestSchema.Extensions = oldExtensions

				err = global.CreateDataset(ctx, tx, testdata.TestSchema, &common.TransactionData{
					Signer: testdata.TestSchema.Owner,
					Caller: string(testdata.TestSchema.Owner),
					TxID:   "txid1",
				})
				assert.NoError(t, err)
			},
		},
		{
			name: "owner only action",
			ses1: func(t *testing.T, global *execution.GlobalContext, tx sql.DB) {
				ctx := context.Background()

				err := global.CreateDataset(ctx, tx, testdata.TestSchema, &common.TransactionData{
					Signer: testdata.TestSchema.Owner,
					Caller: string(testdata.TestSchema.Owner),
					TxID:   "txid1",
				})
				require.NoError(t, err)

				_, err = global.Procedure(ctx, tx, &common.ExecutionData{
					Dataset:   testdata.TestSchema.DBID(),
					Procedure: testdata.ActionAdminDeleteUser.Name,
					Args:      []any{1},
					TransactionData: common.TransactionData{
						Signer: []byte("signer"),
						Caller: "signer",
					},
				})
				require.Error(t, err)

				_, err = global.Procedure(ctx, tx, &common.ExecutionData{
					Dataset:   testdata.TestSchema.DBID(),
					Procedure: testdata.ActionAdminDeleteUser.Name,
					Args:      []any{1},
					TransactionData: common.TransactionData{
						Signer: testdata.TestSchema.Owner,
						Caller: string(testdata.TestSchema.Owner),
					},
				})
				require.NoError(t, err)
			},
		},
		{
			name: "private action",
			ses1: func(t *testing.T, global *execution.GlobalContext, tx sql.DB) {
				ctx := context.Background()

				err := global.CreateDataset(ctx, tx, testdata.TestSchema, &common.TransactionData{
					Signer: testdata.TestSchema.Owner,
					Caller: string(testdata.TestSchema.Owner),
					TxID:   "txid1",
				})
				require.NoError(t, err)

				// calling private fails
				_, err = global.Procedure(ctx, tx, &common.ExecutionData{
					Dataset:   testdata.TestSchema.DBID(),
					Procedure: testdata.ActionPrivate.Name,
					Args:      []any{},
					TransactionData: common.TransactionData{
						Signer: []byte("signer"),
						Caller: "signer",
					},
				})
				require.Error(t, err)

				// calling a public which calls private succeeds
				_, err = global.Procedure(ctx, tx, &common.ExecutionData{
					Dataset:   testdata.TestSchema.DBID(),
					Procedure: testdata.ActionCallsPrivate.Name,
					Args:      []any{},
					TransactionData: common.TransactionData{
						Signer: []byte("signer"),
						Caller: "signer",
					},
				})
				require.NoError(t, err)
			},
		},
		{
			// this test used to track that this was not possible, because it was necessary
			// to protect our old SQLite atomicity model. This is no longer necessary,
			// and it's actually preferable that we can support this. Logically, it makes sense
			// that a deploy tx followed by an execute tx in the same block should work.
			name: "deploy and call at the same time",
			ses1: func(t *testing.T, global *execution.GlobalContext, tx sql.DB) {
				ctx := context.Background()

				err := global.CreateDataset(ctx, tx, testdata.TestSchema, &common.TransactionData{
					Signer: testdata.TestSchema.Owner,
					Caller: string(testdata.TestSchema.Owner),
					TxID:   "txid1",
				})
				require.NoError(t, err)

				_, err = global.Procedure(ctx, tx, &common.ExecutionData{
					Dataset:   testdata.TestSchema.DBID(),
					Procedure: testdata.ActionCreateUser.Name,
					Args:      []any{1, "satoshi", 42},
					TransactionData: common.TransactionData{
						Signer: []byte("signer"),
						Caller: "signer",
					},
				})
				require.NoError(t, err)

				_, err = global.Procedure(ctx, tx, &common.ExecutionData{
					Dataset:   testdata.TestSchema.DBID(),
					Procedure: testdata.ActionGetUserByAddress.Name,
					Args:      []any{"signer"},
					TransactionData: common.TransactionData{
						Signer: []byte("signer"),
						Caller: "signer",
					},
				})
				require.NoError(t, err)
			},
		},
		{
			name: "deploy many databases",
			ses1: func(t *testing.T, global *execution.GlobalContext, tx sql.DB) {
				ctx := context.Background()

				for i := 0; i < 10; i++ {
					newSchema := *testdata.TestSchema
					newSchema.Name = testdata.TestSchema.Name + fmt.Sprint(i)

					err := global.CreateDataset(ctx, tx, &newSchema, &common.TransactionData{
						Signer: testdata.TestSchema.Owner,
						Caller: string(testdata.TestSchema.Owner),
						TxID:   "txid" + fmt.Sprint(i),
					})
					require.NoError(t, err)
				}
			},
		},
		{
			name: "deploying and immediately dropping",
			ses1: func(t *testing.T, global *execution.GlobalContext, tx sql.DB) {
				ctx := context.Background()

				err := global.CreateDataset(ctx, tx, testdata.TestSchema, &common.TransactionData{
					Signer: testdata.TestSchema.Owner,
					Caller: string(testdata.TestSchema.Owner),
					TxID:   "txid1",
				})
				require.NoError(t, err)

				err = global.DeleteDataset(ctx, tx, testdata.TestSchema.DBID(), &common.TransactionData{
					Signer: testdata.TestSchema.Owner,
					Caller: string(testdata.TestSchema.Owner),
					TxID:   "txid3",
				})
				require.NoError(t, err)
			},
		},
		{
			name: "case insensitive",
			ses1: func(t *testing.T, global *execution.GlobalContext, tx sql.DB) {
				ctx := context.Background()

				schema := *caseSchema

				err := global.CreateDataset(ctx, tx, &schema, &common.TransactionData{
					Signer: testdata.TestSchema.Owner,
					Caller: string(testdata.TestSchema.Owner),
					TxID:   "txid1",
				})
				require.NoError(t, err)

				caller := "signer"
				signer := []byte("signer")

				_, err = global.Procedure(ctx, tx, &common.ExecutionData{
					Dataset:   schema.DBID(),
					Procedure: "CREATE_USER",
					Args:      []any{1, "satoshi"},
					TransactionData: common.TransactionData{
						Signer: []byte(caller),
						Caller: string(signer),
					},
				})
				require.NoError(t, err)

				_, err = global.Procedure(ctx, tx, &common.ExecutionData{
					Dataset:   schema.DBID(),
					Procedure: "CREATE_USER",
					Args:      []any{"2", "vitalik"},
					TransactionData: common.TransactionData{
						Signer: []byte(caller),
						Caller: string(signer),
					},
				})
				require.NoError(t, err)

				_, err = global.Procedure(ctx, tx, &common.ExecutionData{
					Dataset:   schema.DBID(),
					Procedure: "CREATE_FOLLOWER",
					Args:      []any{"satoshi", "vitalik"},
					TransactionData: common.TransactionData{
						Signer: []byte(caller),
						Caller: string(signer),
					},
				})
				require.NoError(t, err)

				res, err := global.Procedure(ctx, tx, &common.ExecutionData{
					Dataset:   schema.DBID(),
					Procedure: "USE_EXTENSION",
					Args:      []any{1, "2"}, // math_ext.add($arg1 + $arg2, 1)
					TransactionData: common.TransactionData{
						Signer: []byte(caller),
						Caller: string(signer),
					},
				})
				require.NoError(t, err)

				// "SELECT $rES as res;" will be a string because arg type
				// inference based on Go variables is only used for inline
				// expressions since postgres prepare/describe is desirable for
				// statements that actually reference a table (but this one does
				// not).
				require.Equal(t, "4", res.Rows[0][0])
				require.Equal(t, []string{"res"}, res.Columns) // without the `AS res`, it would be `?column?`
			},
		},
		{
			name: "procedure",
			ses1: func(t *testing.T, global *execution.GlobalContext, tx sql.DB) {
				ctx := context.Background()

				err := global.CreateDataset(ctx, tx, testdata.TestSchema, &common.TransactionData{
					Signer: testdata.TestSchema.Owner,
					Caller: string(testdata.TestSchema.Owner),
					TxID:   "txid1",
				})
				require.NoError(t, err)
			},
			ses2: func(t *testing.T, global *execution.GlobalContext, tx sql.DB) {
				ctx := context.Background()

				_, err := global.Procedure(ctx, tx, &common.ExecutionData{
					Dataset:   testdata.TestSchema.DBID(),
					Procedure: testdata.ProcCreateUser.Name,
					Args:      []any{1, "satoshi", 42},
					TransactionData: common.TransactionData{
						Signer: []byte("signer"),
						Caller: "signer",
					},
				})
				require.NoError(t, err)

				user, err := global.Procedure(ctx, tx, &common.ExecutionData{
					Dataset:   testdata.TestSchema.DBID(),
					Procedure: testdata.ProcGetUserByAddress.Name,
					Args:      []any{"signer"},
					TransactionData: common.TransactionData{
						Signer: []byte("signer"),
						Caller: "signer",
					},
				})
				require.NoError(t, err)

				require.Equal(t, len(user.Rows), 1)

				require.Equal(t, []any{int64(1), "satoshi", int64(42)}, []any{user.Rows[0][0], user.Rows[0][1], user.Rows[0][2]})
			},
		},
		// TODO: test procedure returns table
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if test.ses1 == nil {
				test.ses1 = func(t *testing.T, global *execution.GlobalContext, tx sql.DB) {}
			}
			if test.ses2 == nil {
				test.ses2 = func(t *testing.T, global *execution.GlobalContext, tx sql.DB) {}
			}
			if test.after == nil {
				test.after = func(t *testing.T, global *execution.GlobalContext, tx sql.DB) {}
			}

			global, db, err := setup(t)
			if err != nil {
				t.Fatal(err)
			}
			defer cleanup(t, db)

			ctx := context.Background()

			tx, err := db.BeginOuterTx(ctx)
			require.NoError(t, err)
			defer tx.Rollback(ctx)

			test.ses1(t, global, tx)

			id, err := tx.Precommit(ctx) // not needed, but test how txApp would use the engine
			require.NoError(t, err)
			require.NotEmpty(t, id)

			err = tx.Commit(ctx)
			require.NoError(t, err)

			tx2, err := db.BeginTx(ctx)
			require.NoError(t, err)
			defer tx2.Rollback(ctx)

			test.ses2(t, global, tx2)

			// Omit Precommit here, just to test that it's allowed even though
			// txApp would want the commit ID.

			err = tx2.Commit(ctx)
			require.NoError(t, err)

			readOnly, err := db.BeginTx(ctx)
			require.NoError(t, err)
			defer readOnly.Rollback(ctx)

			test.after(t, global, readOnly)
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
						Type: types.IntType,
						Attributes: []*types.Attribute{
							{
								Type: types.PRIMARY_KEY,
							},
						},
					},
					{
						Name: "nAMe",
						Type: types.TextType,
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
						Type: types.IntType,
						Attributes: []*types.Attribute{
							{
								Type: types.NOT_NULL,
							},
						},
					},
					{
						Name: "fOllOwee_id",
						Type: types.IntType,
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
		Actions: []*types.Action{
			{
				Name: "CrEaTe_UsEr",
				Parameters: []string{
					"$Id",
					"$nAmE",
				},
				Public: true,
				Body:   "INSERT INTO UseRs (ID, nAme) VALUES ($iD, $nAME);",
			},
			{
				Name: "CrEaTe_FoLlOwEr",
				Parameters: []string{
					"$FoLlOwer_nAme",
					"$FoLlOwee_nAme",
				},
				Public: true,
				Body: `INSERT INTO FollOweRS (FOLlOwer_id, FOLlOwee_id)
					VALUES (
						(SELECT ID FROM USErs WHERE NAmE = $FoLlOwer_nAME),
						(SELECT ID FROM UsErS WHERE nAME = $FoLlOwee_nAME)
					);`,
			},
			{
				Name: "use_ExTension",
				Parameters: []string{
					"$vAl1",
					"$vAl2",
				},
				Public: true,
				Body:   "$rEs = Math_Ext.AdD($VAl1 + $VAl2, 1); SELECT $rES as res;", // type? procedure execution is not strongly typed...

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
