//go:build pglive

package integration_test

import (
	"context"
	"testing"

	"github.com/kwilteam/kwil-db/common"
	"github.com/kwilteam/kwil-db/parse"
	"github.com/stretchr/testify/require"
)

// TestDeployment tests the negative cases for deployment of schemas
func Test_Deployment(t *testing.T) {
	type testCase struct {
		name string
		// either schema or procedure should be set
		// If schema is set, it will deploy the whole schema.
		// If only procedure is set, it expects only a procedure
		// body, and will wrap the procedure in a schema.
		schema    string
		procedure string
		err       error
	}

	testCases := []testCase{
		{
			name: "view procedure mutates",
			schema: `
		database mutative_view;

		table users {
		    id int primary key
		}

		procedure mutate_in_view() public view {
		    INSERT INTO users (id) VALUES (1);
		}`,
			err: parse.ErrViewMutatesState,
		},
		{
			name: "view procedure calls non-view",
			schema: `
		database view_calls_non_view;

		table users {
			id int primary key
		}

		procedure view_calls_non_view() public view {
			not_a_view();
		}

		procedure not_a_view() public {
			INSERT INTO users (id) VALUES (1);
		}`,
			err: parse.ErrViewMutatesState,
		},
		{
			name: "empty procedure",
			schema: `
		database empty_procedure;

		procedure empty_procedure() public {}
			`,
		},
		{
			name:      "untyped variable",
			procedure: `$intval := 1;`, // this can infer the type
		},
		{
			name:      "undeclared variable",
			procedure: `$intval int := $a;`,
			err:       parse.ErrUndeclaredVariable,
		},
		{
			name:      "non-existent @ variable",
			procedure: `$id int := @ethereum_height;`,
			err:       parse.ErrUnknownContextualVariable,
		},
		{
			name: "unknown function",
			procedure: `
			$int int := unknown_function();
			`,
			err: parse.ErrUnknownFunctionOrProcedure,
		},
		{
			name: "known procedure",
			schema: `
			database known_procedure;

			procedure known_procedure() public returns table(id int) {
				select 1 as id;
			}

			procedure known_procedure_2() public {
				for $row in select * from known_procedure() as k {

				}
			}
			`,
		},
		{
			name: "unknown function in SQL",
			procedure: `
			for $row in select * from unknown_function() {
				break;
			}
			`,
			err: parse.ErrUnknownFunctionOrProcedure,
		},
		{
			name: "various foreign procedures",
			schema: `database foreign_procedures;

			foreign procedure get_tbl() returns table(id int)
			foreign procedure get_scalar(int) returns (int)
			foreign procedure get_named_scalar(int) returns (id int)

			procedure call_all() public returns table(id int) {
				$int1 int := get_scalar['dbid', 'get_scalar'](1);
				$int2 int := get_named_scalar['dbid', 'get_scalar'](1);

				return select * from get_tbl['dbid', 'get_table']() as u;
			}
			`,
		},
		{
			name: "procedure returns select join from others",
			schema: `database select_join;

			table users {
				id int primary key,
				name text
			}

			foreign procedure get_tbl() returns table(id int)

			procedure get_users() public returns table(id int, name text) {
				return select * from users;
			}

			// get_all joins the users table with the result of get_tbl
			procedure get_all() public returns table(id int, name text) {
				return select a.id as id, u.name as name from get_tbl['dbid', 'get_tbl']() AS a
				INNER JOIN get_users() AS u ON a.id = u.id;
			}
			`,
		},
		{
			name: "action references foreign procedure and local procedure",
			schema: `database select_join;

			table users {
				id int primary key,
				name text
			}

			foreign procedure get_tbl() returns table(id int)

			procedure get_users() public returns table(id int, name text) {
				return select * from users;
			}

			// get_all joins the users table with the result of get_tbl
			action get_all() public view {
				select a.id as id, u.name as name from get_tbl['dbid', 'get_tbl']() AS a
				INNER JOIN get_users() AS u ON a.id = u.id;
			}
			`,
		},
		{
			name: "action references unknown foreign procedure",
			schema: `database select_join;

			action get_all() public view {
				select * from get_tbl['dbid', 'get_tbl']();
			}
			`,
			err: parse.ErrUnknownFunctionOrProcedure,
		},
		{
			name: "schema with max_len blob",
			schema: `database max_len_blob;

			table users {
				id int primary key,
				data blob max_len(10)
			}`,
		},
		{
			name: "max on uint256",
			schema: `database max_uint256;

			table users {
				id int primary key,
				data uint256 max(1000000000000000000000000000000000000000)
			}`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.schema != "" && tc.procedure != "" {
				t.Fatal("both schema and procedure set")
			}

			schema := tc.schema
			if tc.procedure != "" {
				schema = `database t;
				
				procedure t() public {
					` + tc.procedure + `
				}`
			}

			global, db, err := setup(t)
			if err != nil {
				t.Fatal(err)
			}
			defer cleanup(t, db)

			ctx := context.Background()

			tx, err := db.BeginTx(ctx)
			require.NoError(t, err)
			defer tx.Rollback(ctx)

			readonly, err := db.BeginReadTx(ctx)
			require.NoError(t, err)
			defer readonly.Rollback(ctx)

			// we intentionally use the bare kuneiform parser and don't
			// perform extra checks because we want to test that the engine
			// catches these errors
			parsed, err := parse.ParseSchemaWithoutValidation([]byte(schema))
			require.NoError(t, err)
			parsed.Schema.Owner = owner

			err = global.CreateDataset(&common.TxContext{
				BlockContext: &common.BlockContext{},
				Signer:       owner,
				Caller:       string(owner),
				TxID:         "test",
				Ctx:          ctx,
			}, tx, parsed.Schema)
			if tc.err != nil {
				require.ErrorIs(t, err, tc.err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
