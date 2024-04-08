// package generate generates the plpgsql code for a procedure.
package generate

import (
	"fmt"
	"strings"

	"github.com/kwilteam/kwil-db/core/types"
	"github.com/kwilteam/kwil-db/internal/engine/procedures/parser"
	"github.com/kwilteam/kwil-db/internal/utils/order"
)

type GeneratedProcedure struct {
	// Body is the plpgsql code for the procedure.
	Body string
	// DeclaredVariables are the variables that need to be declared.
	DeclaredVariables []*types.NamedType
	// LoopTargets are variables that are loop targets.
	// They should be declared as RECORD.
	LoopTargets []string
	// OutVariables are the variables that need to be declared.
	OutVariables []*types.NamedType
}

// GenerateProcedure generates the plpgsql code for a procedure.
// It returns the body, as well as the variables that need to be declared.
func GenerateProcedure(stmts []parser.Statement, proc *types.Procedure, procedureInputs []*types.NamedType) (g *GeneratedProcedure, err error) {
	defer func() {
		if e := recover(); e != nil {
			err = fmt.Errorf("panic: %v", e)
		}

		if err != nil {
			err = fmt.Errorf("generate: %w", err)
		}
	}()

	// add procedure inputs to the variables
	variables := make(map[string]*types.DataType)
	for _, arg := range procedureInputs {
		variables[arg.Name] = arg.Type
	}

	v := &generatorVisitor{
		variables:        variables,
		currentProcedure: proc,
	}

	body := strings.Builder{}
	for _, stmt := range stmts {
		stmt := stmt.Accept(v).(string)
		if stmt == "" {
			continue
		}
		body.WriteString(stmt)
		body.WriteString("\n")
	}

	declared := make([]*types.NamedType, 0, len(v.variables))
	for _, kv := range order.OrderMap(v.variables) {
		declared = append(declared, &types.NamedType{
			Name: kv.Key,
			Type: kv.Value.Copy(),
		})
	}

	return &GeneratedProcedure{
		Body:              body.String(),
		DeclaredVariables: declared,
		OutVariables:      v.returnedVariables,
		LoopTargets:       v.loopTargets,
	}, nil
}
