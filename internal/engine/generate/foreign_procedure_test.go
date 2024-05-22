package generate

import (
	"strings"
	"testing"
	"unicode"

	"github.com/kwilteam/kwil-db/core/types"
	"github.com/stretchr/testify/require"
)

var schemaName = "schema"

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
CREATE OR REPLACE FUNCTION schema._fp_test(_dbid TEXT, _procedure TEXT) RETURNS VOID AS $$
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
    ON p.schema_id = s.id
    WHERE p.name = _procedure AND s.dbid = _dbid;

    IF _schema_owner IS NULL THEN
        RAISE EXCEPTION 'Procedure "%" not found in schema "%"', _procedure, _dbid;
    END IF;

    IF _is_view = FALSE AND current_setting('is_read_only') = 'on' THEN
        RAISE EXCEPTION 'Non-view procedure "%" called in view-only connection', _procedure;
    END IF;

    IF _is_owner_only = TRUE AND _schema_owner != current_setting('ctx.signer')::BYTEA THEN
        RAISE EXCEPTION 'Procedure "%" is owner-only and cannot be called by signer "%" in schema "%"', _procedure, current_setting('ctx.signer'), _dbid;
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
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := GenerateForeignProcedure(test.procedure, schemaName)
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

func removeWhitespace(s string) string {
	return strings.Map(func(r rune) rune {
		if unicode.IsSpace(r) {
			return -1
		}
		return r
	}, s)
}
