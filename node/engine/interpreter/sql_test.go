//go:build pglive

package interpreter

import (
	"context"
	"testing"

	"github.com/kwilteam/kwil-db/common"
	"github.com/kwilteam/kwil-db/core/types"
	"github.com/kwilteam/kwil-db/core/types/decimal"
	"github.com/kwilteam/kwil-db/extensions/precompiles"
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
				vals := func() map[string]precompiles.Value {
					return map[string]precompiles.Value{
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

			interp, err := NewInterpreter(ctx, tx, &common.Service{}, nil, nil)
			require.NoError(t, err)
			_ = interp

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

func mustNewVal(v any) precompiles.Value {
	val, err := precompiles.NewValue(v)
	if err != nil {
		panic(err)
	}
	return val
}

func mustDec(s string) *decimal.Decimal {
	d, err := decimal.NewFromString(s)
	if err != nil {
		panic(err)
	}
	return d
}

func mustUUID(s string) *types.UUID {
	u, err := types.ParseUUID(s)
	if err != nil {
		panic(err)
	}
	return u
}

// This tests how the engine stores and queries all sorts of different DDL.
// It tests by:
// 1. Creating a bunch of different types of metadata
// 2. Querying the metadata from the info schema
// 3. Ensuring that the interpreter's in-memory metadata matches the database's metadata
func Test_Metadata(t *testing.T) {
	ctx := context.Background()

	db, err := pg.NewDB(ctx, &pg.DBConfig{
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
	})
	require.NoError(t, err)
	defer db.Close()
	tx, err := db.BeginTx(ctx)
	require.NoError(t, err)
	defer tx.Rollback(ctx) // always rollback to avoid cleanup

	interp, err := NewInterpreter(ctx, tx, &common.Service{}, nil, nil)
	require.NoError(t, err)
	_ = interp

	require.NoError(t, err)

	// 1. Create a bunch of different types of metadata

	// 1.1 Tables: one with a composite foreign key, and one with a composite primary key. both have indexes
	err = interp.ExecuteWithoutEngineCtx(ctx, tx, `
		-- ========================
		-- Table: users
		-- - Single Primary Key (user_id)
		-- - CHECK constraint on balance >= 0
		-- - Single index on (email)
		-- - Demonstrates usage of: int8, text, bool, numeric, uuid, bytea
		-- ========================
		CREATE TABLE users (
		    user_id       UUID	PRIMARY KEY,
		    first_name    TEXT	NOT NULL,
		    last_name     TEXT	NOT NULL,
		    email         TEXT	NOT NULL,
		    is_active     BOOL	NOT NULL DEFAULT TRUE,
		    balance       NUMERIC(10, 2) NOT NULL DEFAULT 0,
		    avatar        BYTEA,
		    CONSTRAINT check_balance CHECK (balance >= 0)
		);
		
		-- Single index on email
		CREATE INDEX idx_users_email
		    ON users (email);
		
		-- ========================
		-- Table: products
		-- - Single Primary Key (product_id)
		-- - Composite index on (name, price)
		-- - Demonstrates usage of: int8 (BIGINT), text, numeric, uuid, bytea
		-- ========================
		CREATE TABLE products (
		    product_id    UUID PRIMARY KEY,
		    name          TEXT NOT NULL,
		    description   TEXT,
		    price         NUMERIC(10, 2) NOT NULL,
		    product_uuid  UUID NOT NULL,
		    product_image BYTEA,
		    is_active     BOOL NOT NULL DEFAULT TRUE
		);
		
		-- Composite index on (name, price)
		CREATE INDEX idx_products_name_price
		    ON products (name, price);
		
		-- ========================
		-- Table: orders
		-- - Single Primary Key (order_id)
		-- - Single Foreign Key (user_id -> users.user_id)
		-- - Demonstrates usage of: int8 (BIGINT), numeric
		-- ========================
		CREATE TABLE orders (
		    order_id    UUID PRIMARY KEY,
		    user_id     UUID NOT NULL,
		    total_amt   NUMERIC(10, 2) NOT NULL DEFAULT 0,
		    -- Single FK referencing users
		    CONSTRAINT fk_orders_users
		        FOREIGN KEY (user_id)
		        REFERENCES users(user_id)
		        ON DELETE CASCADE
		);
		
		-- ========================
		-- Table: order_details
		-- - Composite Primary Key (order_id, line_item)
		-- - Single Foreign Key on order_id -> orders(order_id)
		-- - Single Foreign Key on product_id -> products(product_id)
		-- - Demonstrates usage of: int8 (BIGINT), numeric
		-- ========================
		CREATE TABLE order_details (
		    order_id   UUID NOT NULL,
		    line_item  INT NOT NULL,
		    product_id UUID NOT NULL,
		
		    -- Composite PK
		    CONSTRAINT pk_order_details
		        PRIMARY KEY (order_id, line_item),
		
		    -- Single FKs
		    CONSTRAINT fk_order_details_orders
		        FOREIGN KEY (order_id)
		        REFERENCES orders(order_id)
		        ON DELETE CASCADE,
		    CONSTRAINT fk_order_details_products
		        FOREIGN KEY (product_id)
		        REFERENCES products(product_id)
		        ON DELETE RESTRICT
		);
		
		-- ========================
		-- Table: shipment
		-- - Single Primary Key (shipment_id)
		-- - Composite Foreign Key (order_id, line_item -> order_details)
		-- - Demonstrates usage of: int8 (BIGINT)
		-- ========================
		CREATE TABLE shipment (
		    shipment_id  UUID PRIMARY KEY,
		    order_id     UUID NOT NULL,
		    line_item    INT NOT NULL,
		
		    -- Composite FK referencing order_details
		    CONSTRAINT fk_shipment_order_details
		        FOREIGN KEY (order_id, line_item)
		        REFERENCES order_details(order_id, line_item)
		        ON DELETE CASCADE
		);
	`, nil, nil)
	require.NoError(t, err)

	// 1.2 Actions:
	// 	- one with no params and returns a single row
	//	- one with many params no returns
	//	- one that takes one params returns a table
	//	- one that has neither params nor returns
	err = interp.ExecuteWithoutEngineCtx(ctx, tx, `
	CREATE ACTION no_params_returns_single() public view returns (id int, name text) { return 1, 'hello'; };
	CREATE ACTION many_params_no_returns($a int, $b text, $c bool, $d numeric) public view {};
	CREATE ACTION one_param_returns_table($a int) system view returns table (id int, name text) { return select 1 as id, 'hello' as name; };
	CREATE ACTION no_params_no_returns() private owner {};
	`, nil, nil)
	require.NoError(t, err)

	// 1.3 Roles
	// 	- one with no permissions
	//	- one with some permissions
	err = interp.ExecuteWithoutEngineCtx(ctx, tx, `
	CREATE ROLE no_perms;
	CREATE ROLE some_perms;
	GRANT INSERT TO some_perms;
	GRANT SELECT TO some_perms;
	`, nil, nil)
	require.NoError(t, err)

	// 1.4 Extensions
	// This extension will also create a namespace
}

// var testSchemaExt = precompiles.PrecompileExtension[struct{}]{
// 	Methods: []precompiles.Method[struct{}]{
// 		{
// 			Name:            "test_schema",
// 			AccessModifiers: []precompiles.Modifier{precompiles.PUBLIC},
// 		},
// 	},
// }
