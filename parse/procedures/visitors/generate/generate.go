// package generate generates the plpgsql code for a procedure.
package generate

import (
	"fmt"
	"strings"

	"github.com/kwilteam/kwil-db/core/types"
	"github.com/kwilteam/kwil-db/core/utils/order"
	parser "github.com/kwilteam/kwil-db/parse/procedures/parser"
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
// Unlike other visitors, it does not need an error listener, since all errors
// should already be caught. There are some checks it does internally, but this
// should be treated as a last-resort check, and failure to catch these earlier
// should be considered a bug.
func GenerateProcedure(stmts []parser.Statement, proc *types.Procedure, procedureInputs []*types.NamedType,
	anonymousReceiverTypes []*types.DataType, pgSchemaName string) (g *GeneratedProcedure, err error) {
	defer func() {
		if e := recover(); e != nil {
			err = fmt.Errorf("panic: %v", e)
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
		anonymousTypes:   anonymousReceiverTypes,
		pgSchemaName:     pgSchemaName,
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

	// sanity check in case we have too many/little
	// anonymous receivers declared
	if len(anonymousReceiverTypes) != v.anonymousReceiverCount {
		return nil, fmt.Errorf("internal bug: expected %d anonymous receivers, got %d", v.anonymousReceiverCount, len(anonymousReceiverTypes))
	}

	return &GeneratedProcedure{
		Body:              body.String(),
		DeclaredVariables: declared,
		OutVariables:      v.returnedVariables,
		LoopTargets:       v.loopTargets,
	}, nil
}
