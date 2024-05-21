package generate

import (
	"fmt"
	"strings"

	"github.com/kwilteam/kwil-db/core/types"
)

/*
Generated foreign procedure calls are responsible for making calls to procedures in other schemas.
The dbid and procedure can be passed in dynamically. It should follow the following general structure, with variations
for inputs/outputs:

create or replace function _fp_template(_dbid TEXT, _procedure TEXT, _arg1 TEXT, _arg2 INT8, OUT _out_1 TEXT) as $$
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

    -- Schema owner cannot be nil, and will only be nil
    -- if the procedure is not found
    IF _schema_owner IS NULL THEN
        RAISE EXCEPTION 'Procedure "%" not found in schema "%"', _procedure, _dbid;
    END IF;

    -- we now ensure that:
    -- 1. if we are in a read-only connection, that the procedure is view
    -- 2. if it is owner only, the signer is the owner
    -- 3. it is public
    -- 4. our input types match the expected input types
    -- 5. our return types match the expected return types

    -- 1. if we are in a read-only connection, that the procedure is view
    IF _is_view = FALSE AND current_setting('is_read_only') = 'on' THEN
        RAISE EXCEPTION 'Non-view procedure "%" called in view-only connection', _procedure;
    END IF;

    -- 2. if it is owner only, the signer is the owner
    IF _is_owner_only = TRUE AND _schema_owner != current_setting('ctx.signer')::BYTEA THEN
        RAISE EXCEPTION 'Procedure "%" is owner-only and cannot be called by user "%" in schema "%"', _procedure, current_setting('ctx.signer'), _dbid;
    END IF;

    -- 3. it is public
    IF _is_public = FALSE THEN
        RAISE EXCEPTION 'Procedure "%" is not public and cannot be foreign called', _procedure;
    END IF;

    -- 4. our input types match the expected input types
    IF array_length(_expected_input_types, 1) != 1 THEN
        RAISE EXCEPTION 'Procedure "%" expects exactly one input type, but got %', _procedure, array_length(_expected_input_types, 1);
    END IF;

    -- since _arg1 is text, we want to ensure that the input type is text
    IF _expected_input_types[1] != 'TEXT' THEN
        RAISE EXCEPTION 'Procedure "%" expects input type "TEXT", but got "%"', _procedure, _expected_input_types[1];
    END IF;

    IF _expected_input_types[2] != 'INT' THEN
        RAISE EXCEPTION 'Procedure "%" expects input type "INT", but got "%"', _procedure, _expected_input_types[2];
    END IF;

    -- 5. our return types match the expected return types
    IF array_length(_expected_return_types, 1) != 1 THEN
        RAISE EXCEPTION 'Procedure "%" expects exactly one return type, but got %', _procedure, array_length(_expected_return_types, 1);
    END IF;

    -- since _out_1 is text, we want to ensure that the return type is text
    IF _expected_return_types[1] != 'TEXT' THEN
        RAISE EXCEPTION 'Procedure "%" expects return type "TEXT", but got "%"', _procedure, _expected_return_types[1];
    END IF;

    -- we now call the procedure. we prefix ds_ to the dbid, as per the Kwil rules.
    EXECUTE format('SELECT * FROM ds_%I.%I($1, $2)', _dbid, _procedure) INTO _out_1 USING _arg1, _arg2;

	-- or, to return a table, we can do:
	-- RETURN QUERY EXECUTE format('SELECT * FROM ds_%I.%I($1, $2)', _dbid, _procedure) USING _arg1, _arg2;

END;
$$ LANGUAGE plpgsql;
*/

// This is implicitly
// coupled to the schema defined in internal/engine.execution/queries.go, and therefore is implicitly
// a circular dependency. I am unsure how to resolve this, but am punting on it for now since the structure
// of the new parts of the engine are still in flux.

// GenerateForeignProcedure generates a plpgsql function that allows the schema to dynamically
// call procedures in other schemas, expecting certain inputs and return values. It will prefix
// the generated function with _fp_ (for "foreign procedure").
func GenerateForeignProcedure(proc *types.ForeignProcedure, pgSchema string) (string, error) {
	str := strings.Builder{}

	// first write the header
	str.WriteString(fmt.Sprintf(`CREATE OR REPLACE FUNCTION %s._fp_%s(_dbid TEXT, _procedure TEXT`, pgSchema, proc.Name))

	// we now need to format the inputs. Inputs will be named _arg1, _arg2, etc.
	// we start at 1 since postgres is 1-indexed.
	argList := make([]string, len(proc.Parameters))
	for i, in := range proc.Parameters {
		pgStr, err := in.PGString()
		if err != nil {
			return "", err
		}
		name := fmt.Sprintf("_arg%d", i+1)
		str.WriteString(fmt.Sprintf(", %s %s", name, pgStr))
		argList[i] = name
	}

	var outList []string
	// if there are non-table outputs, we need to format them.
	// we can ignore the name of the output, since it is not a table
	if proc.Returns != nil && !proc.Returns.IsTable {
		for i, out := range proc.Returns.Fields {
			str.WriteString(", OUT ")
			pgStr, err := out.Type.PGString()
			if err != nil {
				return "", err
			}

			name := fmt.Sprintf("_out_%d", i+1)
			str.WriteString(name)
			str.WriteString(" ")
			str.WriteString(pgStr)
			outList = append(outList, name)
		}
	}

	str.WriteString(`)`)

	// If the return type is a table, we need to format the returns as a table.
	if proc.Returns != nil && proc.Returns.IsTable {
		str.WriteString(" RETURNS TABLE(")
		for i, out := range proc.Returns.Fields {
			if i > 0 {
				str.WriteString(", ")
			}

			str.WriteString(out.Name)
			str.WriteString(" ")
			pgStr, err := out.Type.PGString()
			if err != nil {
				return "", err
			}

			str.WriteString(pgStr)
		}
		str.WriteString(")")
	} else if proc.Returns != nil && len(proc.Returns.Fields) == 0 {
		// if we are returning nothing, we need to specify that we are returning nothing.
		str.WriteString(" RETURNS VOID")
	} else if proc.Returns == nil {
		// if we are returning nothing, we need to specify that we are returning nothing.
		str.WriteString(" RETURNS VOID")
	} //if none of the above trigger, then there must be OUT variables, so we do not need to specify a return type.

	str.WriteString(` AS $$ `)

	// declare variables
	str.WriteString(`DECLARE
    _schema_owner BYTEA;
    _is_view BOOLEAN;
    _is_owner_only BOOLEAN;
    _is_public BOOLEAN;
	_returns_table BOOLEAN;
    _expected_input_types TEXT[];
    _expected_return_types TEXT[];`)

	// begin block
	str.WriteString("\nBEGIN")

	// select the procedure info, and perform checks 1-3
	str.WriteString(`
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
	`)

	// check the length of the expected input types
	// if no proc inputs, we check that inputs in the schema should be nil.
	// If there are proc inputs, we first check that the array_length is not null,
	// and then that it is equal, and then that the types match.
	if len(proc.Parameters) == 0 {
		// first check the length of the array
		str.WriteString(fmt.Sprintf(`
	IF array_length(_expected_input_types, 1) IS NOT NULL THEN
		RAISE EXCEPTION 'Foreign procedure definition "%s" expects no inputs, but got procedure "%%" requires %% inputs', _procedure, array_length(_expected_input_types, 1);
	END IF;
	`, proc.Name))
	} else {
		str.WriteString(fmt.Sprintf(`
	IF array_length(_expected_input_types, 1) IS NULL THEN
		RAISE EXCEPTION 'Foreign procedure definition "%s" expects %d inputs, but procedure "%%" requires no inputs', _procedure;
	END IF;

	IF array_length(_expected_input_types, 1) != %d THEN
		RAISE EXCEPTION 'Foreign procedure definition "%s" expects %d inputs, but procedure "%%" requires %% inputs', _procedure, array_length(_expected_input_types, 1);
	END IF;`, proc.Name, len(proc.Parameters), len(proc.Parameters), proc.Name, len(proc.Parameters)))
	}

	// now we check that the types match
	for i, in := range proc.Parameters {
		str.WriteString(fmt.Sprintf(`
	IF _expected_input_types[%d] != '%s' THEN
		RAISE EXCEPTION 'Foreign procedure definition "%s" expects input type "%s", but procedure "%%" requires %%', _procedure, _expected_input_types[%d];
	END IF;`, i+1, in.String(), proc.Name, in.String(), i+1))
	}

	// if the proc returns a table, ensure that the called procedure returns a table
	if proc.Returns != nil && proc.Returns.IsTable {
		str.WriteString(fmt.Sprintf(`
		IF _returns_table = FALSE THEN
			RAISE EXCEPTION 'Foreign procedure definition "%s" expects a table return, but procedure "%%" does not return a table', _procedure;
		END IF;`, proc.Name))

		// If a table is returned, we also need to ensure that it returns the exact correct column names and types.
		for i, out := range proc.Returns.Fields {
			// check the return type
			str.WriteString(fmt.Sprintf(`
			IF _expected_return_types[%d] != '%s' THEN
				RAISE EXCEPTION 'Foreign procedure definition "%s" expects return type "%s" at column position %d, but procedure "%%" requires %%', _procedure, _expected_return_types[%d];
			END IF;`, i+1, out.Type.String(), proc.Name, out.Type.String(), i+1, i+1))

			// check the return name
			str.WriteString(fmt.Sprintf(`
			IF _expected_return_types[%d] != '%s' THEN
				RAISE EXCEPTION 'Foreign procedure definition "%s" expects return name "%s" at column position %d, but procedure "%%" requires %%', _procedure, _expected_return_types[%d];
			END IF;`, i+1, out.Name, proc.Name, out.Name, i+1, i+1))
		}
	} else {
		// else check that the called procedure does not return a table
		str.WriteString(fmt.Sprintf(`
		IF _returns_table = TRUE THEN
			RAISE EXCEPTION 'Foreign procedure definition "%s" expects a non-table return, but procedure "%%" returns a table', _procedure;
		END IF;`, proc.Name))

		// since we are not returning a table, if we are returning anything, we need to check that the return types match.
		// we do not care about the return names, since they are not tables.
		if proc.Returns != nil {
			for i, out := range proc.Returns.Fields {
				str.WriteString(fmt.Sprintf(`
				IF _expected_return_types[%d] != '%s' THEN
					RAISE EXCEPTION 'Foreign procedure definition "%s" expects return type "%s" at return position %d, but procedure "%%" requires %%', _procedure, _expected_return_types[%d];
				END IF;`, i+1, out.Type.String(), proc.Name, out.Type.String(), i+1, i+1))
			}
		}
	}

	// now we call the procedure.
	// If we are calling a table procedure, we need to use RETURN QUERY EXECUTE.
	// Otherwise, we can just use EXECUTE INTO.
	// we only have to worry about SQL injection for the DBID and the procedure name.
	// Everything else is a string variable defined in this function
	if proc.Returns != nil && proc.Returns.IsTable {
		// if it returns a table, we need to use RETURN QUERY EXECUTE
		str.WriteString(fmt.Sprintf(`
	RETURN QUERY EXECUTE format('SELECT * FROM ds_%%I.%%I(`))
		str.WriteString(dollarsignVars(argList))
		str.WriteString(`)', _dbid, _procedure)`)
	} else {
		// if it returns nothing, we do not need to worry
		// about selecting INTO
		str.WriteString(fmt.Sprintf(`
	EXECUTE format('SELECT * FROM ds_%%I.%%I(`))
		str.WriteString(dollarsignVars(argList))
		str.WriteString(`)', _dbid, _procedure)`)

		if proc.Returns != nil {
			str.WriteString(` INTO `)
			str.WriteString(formatStringList(outList))
		}
	}
	if len(argList) > 0 {
		str.WriteString(" USING ")
		str.WriteString(formatStringList(argList))
	}

	// end block
	str.WriteString(`; END; $$ LANGUAGE plpgsql;`)

	return str.String(), nil
}

// dollarsignVars returns enough dollar signs to be used as a variable in a plpgsql function.
func dollarsignVars(strs []string) string {
	str := strings.Builder{}
	for i := range strs {
		if i > 0 {
			str.WriteString(", ")
		}

		str.WriteString(fmt.Sprintf("$%d", i+1))
	}

	return str.String()
}

func formatStringList(strs []string) string {
	str := strings.Builder{}
	for i, s := range strs {
		if i > 0 {
			str.WriteString(", ")
		}

		str.WriteString(s)
	}

	return str.String()
}
