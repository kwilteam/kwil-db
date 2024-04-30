package ddl

import (
	"fmt"
	"strings"

	"github.com/kwilteam/kwil-db/core/types"
	"github.com/kwilteam/kwil-db/parse/procedures"
)

// GenerateProcedure generates the plpgsql code for a procedure.
// It takes a procedure and the body of the procedure and returns the plpgsql code for the procedure.
func GenerateProcedure(proc *procedures.AnalyzedProcedure, pgSchema string) (string, error) {
	str := strings.Builder{}
	str.WriteString("CREATE OR REPLACE FUNCTION ")
	str.WriteString(fmt.Sprintf("%s.%s(", pgSchema, proc.Name))

	// writing the function parameters

	// paramSet tracks the used params, and will not allow them
	// to be redeclared in the DECLARE section.
	paramSet := make(map[string]struct{})
	i := -1
	var field *types.NamedType
	for i, field = range proc.Parameters {
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
	if len(proc.Returns.Fields) > 0 && !proc.Returns.IsTable {
		hasOutReturns = true
		if i != -1 {
			str.WriteString(", ")
		}

		for i, field := range proc.Returns.Fields {
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
	if proc.Returns.IsTable && len(proc.Returns.Fields) > 0 {
		str.WriteString("\nRETURNS ")

		str.WriteString("TABLE(")
		for i, field := range proc.Returns.Fields {
			if i != 0 {
				str.WriteString(", ")
			}

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
	if len(proc.DeclaredVariables) > 0 {
		for _, declare := range proc.DeclaredVariables {
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
	if len(proc.LoopTargets) > 0 {
		declaresTypes = true
		for _, loopTarget := range proc.LoopTargets {
			declareSection.WriteString(fmt.Sprintf("%s RECORD;\n", loopTarget))
		}
	}

	if declaresTypes {
		str.WriteString("DECLARE\n")
		str.WriteString(declareSection.String())
	}

	// finishing the function

	str.WriteString("BEGIN\n")
	str.WriteString(proc.Body)
	str.WriteString("\nEND;\n$$ LANGUAGE plpgsql;")

	return str.String(), nil
}
