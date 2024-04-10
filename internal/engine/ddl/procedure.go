package ddl

import (
	"fmt"
	"strings"

	"github.com/kwilteam/kwil-db/core/types"
)

// GenerateProcedure is the plpgsql code for a procedure.
// It takes a procedure and the body of the procedure and returns the plpgsql code for the procedure.
func GenerateProcedure(fields []*types.NamedType, loopTargets []string, returns *types.ProcedureReturn,
	declarations []*types.NamedType, outParams []*types.NamedType, schema, name, body string) (string, error) {
	str := strings.Builder{}
	str.WriteString("CREATE OR REPLACE FUNCTION ")
	str.WriteString(fmt.Sprintf("%s.%s(", schema, name))

	// writing the function parameters

	// paramSet tracks the used params, and will not allow them
	// to be redeclared in the DECLARE section.
	paramSet := make(map[string]struct{})
	i := -1
	var field *types.NamedType
	for i, field = range fields {
		if i != 0 {
			str.WriteString(", ")
		}

		paramSet[field.Name] = struct{}{}

		typ, err := field.Type.PGString()
		if err != nil {
			return "", err
		}

		str.WriteString(fmt.Sprintf("%s %s", field.Name, typ))
	}

	hasOutReturns := false
	// we need to write return types if there are any
	if returns != nil && len(returns.Fields) > 0 && !returns.IsTable {
		hasOutReturns = true
		if i != -1 {
			str.WriteString(", ")
		}

		if len(returns.Fields) != len(outParams) {
			return "", fmt.Errorf("number of return types and out parameters do not match. expected %d, got %d", len(returns.Fields), len(outParams))
		}

		for i, field := range outParams {
			if i != 0 {
				str.WriteString(", ")
			}

			typ, err := field.Type.PGString()
			if err != nil {
				return "", err
			}

			str.WriteString(fmt.Sprintf("OUT %s %s", field.Name, typ))
		}
	}

	str.WriteString(") ")

	// writing the return type
	if returns != nil && returns.IsTable && len(returns.Fields) > 0 {
		str.WriteString("\nRETURNS ")

		str.WriteString("TABLE(")
		for i, field := range returns.Fields {
			if i != 0 {
				str.WriteString(", ")
			}
			// TODO: we need to give return types some sort of unique identifier.
			// this needs to match what we give other variables (could even just be underscores).
			typ, err := field.Type.PGString()
			if err != nil {
				return "", err
			}

			str.WriteString(fmt.Sprintf("%s %s", field.Name, typ))
		}
		str.WriteString(") ")
	} else if !hasOutReturns {
		str.WriteString("\nRETURNS void ")
	}

	str.WriteString("AS $$\n")

	// writing the variable declarations

	// declaresTypes tracks if the DECLARE section is needed.
	declaresTypes := false
	declareSection := strings.Builder{}
	if len(declarations) > 0 {
		for _, declare := range declarations {
			_, ok := paramSet[declare.Name]
			if ok {
				continue
			}

			typ, err := declare.Type.PGString()
			if err != nil {
				return "", err
			}

			declaresTypes = true
			declareSection.WriteString(fmt.Sprintf("%s %s;\n", declare.Name, typ))
		}
	}
	if len(loopTargets) > 0 {
		declaresTypes = true
		for _, loopTarget := range loopTargets {
			declareSection.WriteString(fmt.Sprintf("%s RECORD;\n", loopTarget))
		}
	}

	if declaresTypes {
		str.WriteString("DECLARE\n")
		str.WriteString(declareSection.String())
	}

	// finishing the function

	str.WriteString("BEGIN\n")
	str.WriteString(body)
	str.WriteString("\nEND;\n$$ LANGUAGE plpgsql;")

	return str.String(), nil
}
