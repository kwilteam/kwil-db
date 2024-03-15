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
		str.WriteString(fmt.Sprintf("%s %s", field.Name, field.Type.String()))
	}

	hasOutReturns := false
	// we need to write return types if there are any
	if returns != nil && len(returns.Types) > 0 {
		hasOutReturns = true
		if i != -1 {
			str.WriteString(", ")
		}

		if len(returns.Types) != len(outParams) {
			return "", fmt.Errorf("number of return types and out parameters do not match")
		}

		for i, field := range outParams {
			if i != 0 {
				str.WriteString(", ")
			}
			str.WriteString(fmt.Sprintf("OUT %s %s", field.Name, field.Type.String()))
		}
	}

	str.WriteString(") ")

	// writing the return type
	if returns != nil && returns.Table != nil {
		str.WriteString("\nRETURNS ")

		str.WriteString("TABLE(")
		for i, field := range returns.Table {
			if i != 0 {
				str.WriteString(", ")
			}
			// TODO: we need to give return types some sort of unique identifier.
			// this needs to match what we give other variables (could even just be underscores).
			str.WriteString(fmt.Sprintf("%s %s", field.Name, field.Type.String()))
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

			declaresTypes = true
			declareSection.WriteString(fmt.Sprintf("%s %s;\n", declare.Name, declare.Type.String()))
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
