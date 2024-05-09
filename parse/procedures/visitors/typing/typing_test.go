package typing_test

import (
	"testing"

	"github.com/kwilteam/kwil-db/core/types"
	"github.com/kwilteam/kwil-db/core/types/testdata"

	"github.com/kwilteam/kwil-db/parse/metadata"
	parser "github.com/kwilteam/kwil-db/parse/procedures/parser"
	"github.com/kwilteam/kwil-db/parse/procedures/visitors/typing"
	parseTypes "github.com/kwilteam/kwil-db/parse/types"

	"github.com/stretchr/testify/require"
)

// TODO: we need better tests, but due to how much these concerns are intertwined with the rest of
// the procedural language, I am going to leave this as is for now, since we will have robust
// procedural language tests.
func Test_Typing(t *testing.T) {
	type testcase struct {
		name string
		body string
		err  error // can be nil
	}

	/*
		the bodies below are inputted into a default procedure.
		It has two parameters: $id and $name.
	*/

	testcases := []testcase{
		{
			name: "declare and assign",
			body: `
			$id1 int := 1;
			`,
		},
		{
			name: "declare, then assign",
			body: `
			$id1 int;
			$id1 := 1;
			`,
		},
		{
			name: "double declare",
			body: `
			$id1 int;
			$id1 text;
			`,
			err: parseTypes.ErrVariableAlreadyDeclared,
		},
		{
			name: "redeclare input",
			// there is already a parameter named $id
			body: `
			$id int;
			`,
			err: parseTypes.ErrVariableAlreadyDeclared,
		},
		{
			name: "math",
			body: `
			$sum int := $id + 1;
			`,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			proc := &types.Procedure{
				Name: "simple",
				Parameters: []*types.ProcedureParameter{
					{
						Name: "$id",
						Type: types.IntType,
					},
					{
						Name: "$name",
						Type: types.TextType,
					},
				},
				Body: tc.body,
			}

			stmts, err := parser.Parse(tc.body)
			require.NoError(t, err)

			errListener := parseTypes.NewErrorListener()

			// named types match the parameters of the procedure
			_, err = typing.EnsureTyping(stmts, proc, testdata.TestSchema, []*types.NamedType{
				{
					Name: "$id",
					Type: types.IntType,
				},
				{
					Name: "$name",
					Type: types.TextType,
				},
			}, metadata.PgSessionVars, errListener)
			// if the error listener has an error, we should use that error
			if errListener.Err() != nil {
				err = errListener.Err()
			}

			if tc.err != nil {
				require.ErrorAs(t, err, &tc.err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
