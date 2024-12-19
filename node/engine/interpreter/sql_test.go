//go:build pglive

package interpreter

import (
	"context"
	"testing"

	"github.com/kwilteam/kwil-db/common"
	"github.com/kwilteam/kwil-db/core/types"
	"github.com/kwilteam/kwil-db/core/types/decimal"
	"github.com/kwilteam/kwil-db/node/engine"
	"github.com/kwilteam/kwil-db/node/pg"
	"github.com/kwilteam/kwil-db/node/types/sql"
	"github.com/stretchr/testify/require"
)

func Test_built_in_sql(t *testing.T) {
	type testcase struct {
		name string
		fn   func(ctx context.Context, db sql.DB)
	}
	tests := []testcase{
		{
			name: "test store and load actions",
			fn: func(ctx context.Context, db sql.DB) {
				for _, act := range all_test_actions {
					err := storeAction(ctx, db, "main", act)
					require.NoError(t, err)
				}

				actions, err := listActionsInNamespace(ctx, db, "main")
				require.NoError(t, err)

				actMap := map[string]*Action{}
				for _, act := range actions {
					actMap[act.Name] = act
				}

				require.Equal(t, len(all_test_actions), len(actMap))
				for _, act := range all_test_actions {
					stored, ok := actMap[act.Name]
					require.True(t, ok)
					require.Equal(t, act.Name, stored.Name)
					require.Equal(t, act.RawStatement, stored.RawStatement)
					require.Equal(t, act.Modifiers, stored.Modifiers)
					namedTypesEq(t, act.Parameters, stored.Parameters)

					if act.Returns != nil {
						require.NotNil(t, stored.Returns)
						require.Equal(t, act.Returns.IsTable, stored.Returns.IsTable)
						namedTypesEq(t, act.Returns.Fields, stored.Returns.Fields)
					} else {
						require.Nil(t, stored.Returns)
					}

					require.Equal(t, len(act.Body), len(stored.Body))
				}
			},
		},
		{
			name: "test store and load tables",
			fn: func(ctx context.Context, db sql.DB) {
				_, err := db.Execute(ctx, `
				CREATE TABLE main.users (
					id UUID PRIMARY KEY,
					name TEXT NOT NULL CHECK (name <> '' AND length(name) <= 100),
 					age INT CHECK (age >= 0),
					wallet_address TEXT UNIQUE NOT NULL
				);`)
				require.NoError(t, err)

				_, err = db.Execute(ctx, `
				CREATE TABLE main.posts (
					id UUID PRIMARY KEY,
					title TEXT NOT NULL,
					author_id UUID REFERENCES main.users (id) ON DELETE CASCADE
				);
				`)
				require.NoError(t, err)

				_, err = db.Execute(ctx, `CREATE UNIQUE INDEX ON main.users (name);`)
				require.NoError(t, err)

				_, err = db.Execute(ctx, `CREATE INDEX user_ages ON main.users (age);`)
				require.NoError(t, err)

				_, err = createNamespace(ctx, db, "other", namespaceTypeUser)
				require.NoError(t, err)

				_, err = db.Execute(ctx, `CREATE TABLE other.my_table (id UUID PRIMARY KEY);`)
				require.NoError(t, err)

				wantSchemas := map[string]map[string]*engine.Table{
					"main": {
						"users": {
							Name: "users",
							Columns: []*engine.Column{
								{
									Name:         "id",
									DataType:     types.UUIDType,
									IsPrimaryKey: true,
								},
								{
									Name:     "name",
									DataType: types.TextType,
								},
								{
									Name:     "age",
									DataType: types.IntType,
									Nullable: true,
								},
								{
									Name:     "wallet_address",
									DataType: types.TextType,
								},
							},
							Indexes: []*engine.Index{
								{
									Name:    "user_ages",
									Columns: []string{"age"},
									Type:    engine.BTREE,
								},
								{
									Name:    "users_name_idx",
									Columns: []string{"name"},
									Type:    engine.UNIQUE_BTREE,
								},
								{
									Name:    "users_pkey",
									Columns: []string{"id"},
									Type:    engine.PRIMARY,
								},
								{
									Name:    "users_wallet_address_key",
									Columns: []string{"wallet_address"},
									Type:    engine.UNIQUE_BTREE,
								},
							},
							Constraints: map[string]*engine.Constraint{
								"users_name_check": {
									Type:    engine.ConstraintCheck,
									Columns: []string{"name"},
								},
								"users_age_check": {
									Type:    engine.ConstraintCheck,
									Columns: []string{"age"},
								},
								"users_wallet_address_key": {
									Type:    engine.ConstraintUnique,
									Columns: []string{"wallet_address"},
								},
							},
						},
						"posts": {
							Name: "posts",
							Columns: []*engine.Column{
								{
									Name:         "id",
									DataType:     types.UUIDType,
									IsPrimaryKey: true,
								},
								{
									Name:     "title",
									DataType: types.TextType,
								},
								{
									Name:     "author_id",
									DataType: types.UUIDType,
									Nullable: true,
								},
							},
							Indexes: []*engine.Index{
								{
									Name:    "posts_pkey",
									Columns: []string{"id"},
									Type:    engine.PRIMARY,
								},
							},
							Constraints: map[string]*engine.Constraint{
								"posts_author_id_fkey": {
									Type:    engine.ConstraintFK,
									Columns: []string{"author_id"},
								},
							},
						},
					},
					"other": {
						"my_table": {
							Name: "my_table",
							Columns: []*engine.Column{
								{
									Name:         "id",
									DataType:     types.UUIDType,
									IsPrimaryKey: true,
								},
							},
							Indexes: []*engine.Index{
								{
									Name:    "my_table_pkey",
									Columns: []string{"id"},
									Type:    engine.PRIMARY,
								},
							},
						},
					},
				}

				tables := map[string]map[string]*engine.Table{}

				for schemaName := range wantSchemas {
					tbls, err := listTablesInNamespace(ctx, db, schemaName)
					require.NoError(t, err)
					tables[schemaName] = map[string]*engine.Table{}
					for _, tbl := range tbls {
						tables[schemaName][tbl.Name] = tbl
					}
				}

				require.Equal(t, len(wantSchemas), len(tables))
				for schemaName, wantSchema := range wantSchemas {
					storedTbls, ok := tables[schemaName]
					require.True(t, ok)
					for _, want := range wantSchema {
						stored, ok := storedTbls[want.Name]
						require.True(t, ok)
						require.Equal(t, want.Name, stored.Name)
						require.Equal(t, len(want.Columns), len(stored.Columns))
						for i, wc := range want.Columns {
							sc := stored.Columns[i]
							require.Equal(t, wc.Name, sc.Name)
							require.Equal(t, wc.DataType.String(), sc.DataType.String())
							require.Equal(t, wc.IsPrimaryKey, sc.IsPrimaryKey)
							require.Equal(t, wc.Nullable, sc.Nullable)
						}
						require.Equal(t, len(want.Indexes), len(stored.Indexes))
						for i, wi := range want.Indexes {
							si := stored.Indexes[i]
							require.Equal(t, wi.Columns, si.Columns)
							require.Equal(t, wi.Type, si.Type)
							require.Equal(t, wi.Name, si.Name)
						}
						require.Equal(t, len(stored.Constraints), len(want.Constraints))
						for i, wc := range want.Constraints {
							sc := stored.Constraints[i]
							require.Equal(t, wc.Type, sc.Type)
							require.Equal(t, wc.Columns, sc.Columns)
						}
					}
				}
			},
		},
		{
			name: "test store and load extensions",
			fn: func(ctx context.Context, db sql.DB) {
				vals := func() map[string]Value {
					return map[string]Value{
						"str":     mustNewVal("val1"),
						"int":     mustNewVal(123),
						"bool":    mustNewVal(true),
						"dec":     mustNewVal(mustDec("123.456")),
						"uuid":    mustNewVal(mustUUID("c7b6a54c-392c-48f9-803d-31cb97e76052")),
						"blob":    mustNewVal([]byte{1, 2, 3}),
						"strarr":  mustNewVal([]string{"a", "b", "c"}),
						"intarr":  mustNewVal([]int{1, 2, 3}),
						"boolarr": mustNewVal([]bool{true, false, true}),
						"decarr":  mustNewVal([]*decimal.Decimal{mustDec("1.23"), mustDec("4.56")}),
						"uuidarr": mustNewVal([]*types.UUID{mustUUID("c7b6a54c-392c-48f9-803d-31cb97e76052"), mustUUID("c7b6a54c-392c-48f9-803d-31cb97e76053")}),
						"blobarr": mustNewVal([][]byte{{1, 2, 3}, {4, 5, 6}}),
					}
				}

				err := registerExtensionInitialization(ctx, db, "ext1_init", "ext1", vals())
				require.NoError(t, err)

				err = registerExtensionInitialization(ctx, db, "ext2_init", "ext2", vals())
				require.NoError(t, err)

				exts, err := getExtensionInitializationMetadata(ctx, db)
				require.NoError(t, err)

				require.Equal(t, 2, len(exts))
				require.Equal(t, "ext1", exts[0].ExtName)
				require.Equal(t, "ext1_init", exts[0].Alias)
				require.EqualValues(t, vals(), exts[0].Metadata)

				require.Equal(t, "ext2", exts[1].ExtName)
				require.Equal(t, "ext2_init", exts[1].Alias)
				require.EqualValues(t, vals(), exts[1].Metadata)
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			cfg := &pg.DBConfig{
				PoolConfig: pg.PoolConfig{
					ConnConfig: pg.ConnConfig{
						Host:   "127.0.0.1",
						Port:   "5432",
						User:   "kwild",
						Pass:   "kwild", // would be ignored if pg_hba.conf set with trust
						DBName: "kwil_test_db",
					},
					MaxConns: 11,
				},
			}

			ctx := context.Background()

			db, err := pg.NewDB(ctx, cfg)
			require.NoError(t, err)
			defer db.Close()
			tx, err := db.BeginTx(ctx)
			require.NoError(t, err)
			defer tx.Rollback(ctx) // always rollback to avoid cleanup

			interp, err := NewInterpreter(ctx, tx, &common.Service{})
			require.NoError(t, err)

			err = interp.SetOwner(ctx, tx, "owner")
			require.NoError(t, err)

			test.fn(ctx, tx)
		})
	}
}

func namedTypesEq(t *testing.T, a, b []*NamedType) {
	require.Equal(t, len(a), len(b))
	for i, at := range a {
		require.Equal(t, at.Name, b[i].Name)
		require.Equal(t, at.Type.String(), b[i].Type.String())
	}
}

func mustNewVal(v any) Value {
	val, err := NewValue(v)
	if err != nil {
		panic(err)
	}
	return val
}
