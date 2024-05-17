package generate

import (
	"fmt"
	"strings"

	"github.com/kwilteam/kwil-db/core/types"
	"github.com/kwilteam/kwil-db/core/utils/order"
	"github.com/kwilteam/kwil-db/parse"
)

// GenerateProcedure generates the plpgsql code for a procedure.
func GenerateProcedure(proc *types.Procedure, schema *types.Schema, pgSchema string) (ddl string, err error) {
	defer func() {
		if e := recover(); e != nil {
			err = fmt.Errorf("panic: %v", e)
		}
	}()

	res, err := parse.ParseProcedure(proc, schema)
	if err != nil {
		return "", err
	}

	if res.ParseErrs.Err() != nil {
		return "", res.ParseErrs.Err()
	}

	vars := make([]*types.NamedType, len(proc.Parameters))
	for i, param := range proc.Parameters {
		vars[i] = &types.NamedType{
			Name: formatParameterName(param.Name[1:]),
			Type: param.Type,
		}
	}

	// we copy the return as to not modify it
	// we need to write return types if there are any.
	// If it returns a table, we do not want to change the column names,
	// since it will change the result. However, if there are out variables,
	// we want to format them
	var ret types.ProcedureReturn
	if proc.Returns != nil {
		ret.IsTable = proc.Returns.IsTable
		ret.Fields = make([]*types.NamedType, len(proc.Returns.Fields))
		for i, field := range proc.Returns.Fields {

			if ret.IsTable {
				ret.Fields[i] = field
			} else {
				ret.Fields[i] = &types.NamedType{
					Name: formatReturnVar(i),
					Type: field.Type,
				}
			}
		}
	}

	analyzed := &analyzedProcedure{
		Name:       proc.Name,
		Parameters: vars,
		Returns:    ret,
		IsView:     proc.IsView(),
		OwnerOnly:  proc.IsOwnerOnly(),
	}

	// we need to get the variables and anonymous variables (loop targets)
	for _, v := range order.OrderMap(res.Variables) {
		analyzed.DeclaredVariables = append(analyzed.DeclaredVariables, &types.NamedType{
			Name: v.Key,
			Type: v.Value,
		})
	}

	for _, v := range order.OrderMap(res.AnonymousVariables) {
		analyzed.LoopTargets = append(analyzed.LoopTargets, v.Key)
	}

	// we need to visit the AST to get the generated body
	sqlGen := &procedureGenerator{
		sqlGenerator: sqlGenerator{
			pgSchema: pgSchema,
		},
	}

	str := strings.Builder{}
	for _, stmt := range res.AST {
		str.WriteString(stmt.Accept(sqlGen).(string))
	}

	// little sanity check:
	if len(res.AnonymousVariables) != sqlGen.anonymousReceivers {
		return "", fmt.Errorf("internal bug: expected %d anonymous variables, got %d", sqlGen.anonymousReceivers, len(res.AnonymousVariables))
	}

	analyzed.Body = str.String()

	return generateProcedureWrapper(analyzed, pgSchema)
}

type analyzedProcedure struct {
	// Name is the name of the procedure.
	Name string
	// Parameters are the parameters, in order, that the procedure is expecting.
	// If no parameters are expected, this will be nil.
	Parameters []*types.NamedType
	// Returns is the expected return type(s) of the procedure.
	// If no return is expected, this will be nil.
	Returns types.ProcedureReturn
	// DeclaredVariables are the variables that need to be declared.
	DeclaredVariables []*types.NamedType
	// LoopTargets is a list of all variables that are loop targets.
	// They should be declared as RECORD in plpgsql.
	LoopTargets []string
	// Body is the plpgsql code for the procedure.
	Body string
	// IsView is true if the procedure is a view.
	IsView bool
	// OwnerOnly is true if the procedure is owner-only.
	OwnerOnly bool
}

// generateProcedureWrapper generates the plpgsql code for a procedure, not including the body.
// It takes a procedure and the body of the procedure and returns the plpgsql code that creates
// the procedure.
func generateProcedureWrapper(proc *analyzedProcedure, pgSchema string) (string, error) {
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
