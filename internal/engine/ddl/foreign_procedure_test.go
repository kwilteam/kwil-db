package ddl

import (
	"strings"
	"testing"
	"unicode"

	"github.com/kwilteam/kwil-db/core/types"
	"github.com/stretchr/testify/require"
)

func Test_ForeignProcedureGen(t *testing.T) {
	type testcase struct {
		name      string
		procedure *types.ForeignProcedure
		want      string
	}

	tests := []testcase{
		{
			name: "procedure has no inputs and no outputs",
			procedure: &types.ForeignProcedure{
				Name: "test",
			},
			want: `
CREATE OR REPLACE FUNCTION _fp_test(_dbid TEXT, _procedure TEXT) RETURNS VOID AS $$
DECLARE
    _schema_owner BYTEA;
    _is_view BOOLEAN;
    _is_owner_only BOOLEAN;
    _is_public BOOLEAN;
	_returns_table BOOLEAN;
    _expected_input_types TEXT[];
    _expected_return_types TEXT[];
BEGIN

    SELECT p.param_types, p.return_types, p.is_view, p.owner_only, p.public, s.owner, p.returns_table
    INTO _expected_input_types, _expected_return_types, _is_view, _is_owner_only, _is_public, _schema_owner, _returns_table
    FROM kwild_internal.procedures as p INNER JOIN kwild_internal.kwil_schemas as s
    ON s.schema_id = p.id
    WHERE p.name = _procedure AND s.dbid = _dbid;

    IF _schema_owner IS NULL THEN
        RAISE EXCEPTION 'Procedure "%" not found in schema "%"', _procedure, _dbid;
    END IF;

    IF _is_view = FALSE AND current_setting('is_read_only') = 'on' THEN
        RAISE EXCEPTION 'Non-view procedure "%" called in view-only connection', _procedure;
    END IF;

    IF _is_owner_only = TRUE AND _schema_owner != current_setting('ctx.caller') THEN
        RAISE EXCEPTION 'Procedure "%" is owner-only and cannot be called by user "%" in schema "%"', _procedure, current_setting('ctx.caller'), _dbid;
    END IF;

    IF _is_public = FALSE THEN
        RAISE EXCEPTION 'Procedure "%" is not public and cannot be foreign called', _procedure;
    END IF;

	IF array_length(_expected_input_types, 1) IS NOT NULL THEN
		RAISE EXCEPTION 'Foreign procedure definition "test" expects no inputs, but got procedure "%" requires % inputs', _procedure, array_length(_expected_input_types, 1);
	END IF;

	IF _returns_table = TRUE THEN
		RAISE EXCEPTION 'Foreign procedure definition "test" expects a non-table return, but procedure "%" returns a table', _procedure;
	END IF;

EXECUTE format('SELECT * FROM ds_%I.%I()', _dbid, _procedure);

END;
$$ LANGUAGE plpgsql;`,
		},
		// {
		// 	name: "procedure has inputs and outputs (not table output)",
		// },
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := GenerateForeignProcedure(test.procedure)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			// fmt.Println("GOT:")
			// fmt.Println(got)
			// fmt.Printf("\n\n\nWANT:")
			// fmt.Println(test.want)
			// panic("")

			require.Equal(t, removeWhitespace(test.want), removeWhitespace(got))
		})
	}
}

// func Test_ForeignProcedureReturnsTable(t *testing.T) {
// 	type testcase struct {
// 		name      string
// 		procedure string
// 		ins       []*types.NamedType
// 		outs      []*types.NamedType
// 		want      string
// 	}

// 	tests := []testcase{
// 		{
// 			name:      "simple",
// 			procedure: "test",
// 			ins: []*types.NamedType{
// 				{
// 					Name: "a",
// 					Type: types.IntType,
// 				},
// 			},
// 			outs: []*types.NamedType{
// 				{
// 					Name: "b",
// 					Type: types.IntType,
// 				},
// 			},
// 			want: `
// CREATE OR REPLACE FUNCTION _fp_test(_dbid TEXT, _name TEXT, a INT8)
// RETURNS TABLE(b INT8)
// AS $$
// BEGIN
// 	RETURN QUERY EXECUTE format('SELECT * FROM %I.%I(a)', _dbid, _name);
// END;
// $$ LANGUAGE plpgsql;
// `,
// 		},
// 		{
// 			name:      "multiple",
// 			procedure: "test",
// 			ins: []*types.NamedType{
// 				{
// 					Name: "a",
// 					Type: types.IntType,
// 				},
// 				{
// 					Name: "b",
// 					Type: types.TextType,
// 				},
// 			},
// 			outs: []*types.NamedType{
// 				{
// 					Name: "c",
// 					Type: types.IntType,
// 				},
// 				{
// 					Name: "d",
// 					Type: types.TextType,
// 				},
// 			},
// 			want: `
// 			CREATE OR REPLACE FUNCTION _fp_test(_dbid TEXT, _name TEXT, a INT8, b TEXT)
// RETURNS TABLE(c INT8, d TEXT)
// AS $$
// BEGIN
// 	RETURN QUERY EXECUTE format('SELECT * FROM %I.%I(a, b)', _dbid, _name);
// END;
// $$ LANGUAGE plpgsql;
// 			`,
// 		},
// 		{
// 			// we only test for no inputs because a table return cannot have no outputs.
// 			name:      "no inputs",
// 			procedure: "test",
// 			ins:       nil,
// 			outs: []*types.NamedType{
// 				{
// 					Name: "b",
// 					Type: types.IntType,
// 				},
// 			},
// 			want: `
// 			CREATE OR REPLACE FUNCTION _fp_test(_dbid TEXT, _name TEXT)
// RETURNS TABLE(b INT8)
// AS $$
// BEGIN
// 	RETURN QUERY EXECUTE format('SELECT * FROM %I.%I()', _dbid, _name);
// END;
// $$ LANGUAGE plpgsql;
// 			`,
// 		},
// 	}

// 	for _, test := range tests {
// 		t.Run(test.name, func(t *testing.T) {
// 			got, err := formatForeignProcReturnsTable(test.procedure, test.ins, test.outs)
// 			if err != nil {
// 				t.Fatalf("unexpected error: %v", err)
// 			}

// 			fmt.Println(got)
// 			fmt.Printf("\n\n\n")
// 			fmt.Println(test.want)
// 			panic("")
// 			require.Equalf(t, removeWhitespace(test.want), removeWhitespace(got), "expected: %v, got: %v", test.want, got)
// 		})
// 	}
// }

func removeWhitespace(s string) string {
	return strings.Map(func(r rune) rune {
		if unicode.IsSpace(r) {
			return -1
		}
		return r
	}, s)
}
