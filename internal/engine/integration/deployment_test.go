package integration_test

import (
	"context"
	"testing"

	"github.com/kwilteam/kwil-db/internal/engine"
	"github.com/kwilteam/kwil-db/kuneiform"
	"github.com/stretchr/testify/require"
)

var _ = engine.ErrReadOnlyProcedureContainsDML // hack to keep this imported while i comment out the tests

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
			err: engine.ErrReadOnlyProcedureContainsDML,
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
			err: engine.ErrReadOnlyProcedureCallsMutative,
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
			procedure: `$intval := 1;`,
			err:       engine.ErrUntypedVariable,
		},
		{
			name:      "undeclared variable",
			procedure: `$intval int := $a;`,
			err:       engine.ErrUndeclaredVariable,
		},
		{
			name:      "non-existent @ variable",
			procedure: `$id int := @ethereum_height;`,
			err:       engine.ErrUnknownContextualVariable,
		},
		{
			name: "unknown function",
			procedure: `
			$int int := unknown_function();
			`,
			err: engine.ErrUnknownFunctionOrProcedure,
		},
		{
			name: "known procedure",
			schema: `
			database known_procedure;

			procedure known_procedure() public returns table(id int) {
				select 1 as id;
			}

			procedure known_procedure_2() public {
				for $row in select * from known_procedure() {

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
			err: engine.ErrUnknownFunctionOrProcedure,
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

			tx, err := db.BeginOuterTx(ctx)
			require.NoError(t, err)
			defer tx.Rollback(ctx)

			readonly, err := db.BeginReadTx(ctx)
			require.NoError(t, err)
			defer readonly.Rollback(ctx)

			parsed, err := kuneiform.Parse(schema)
			require.NoError(t, err)

			err = global.CreateDataset(ctx, tx, parsed, owner)
			if tc.err != nil {
				require.ErrorIs(t, err, tc.err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
