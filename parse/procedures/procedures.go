// package procedures ties together the different visitors to generate the plpgsql code for a procedure.
package procedures

import (
	"fmt"

	"github.com/kwilteam/kwil-db/core/types"
	"github.com/kwilteam/kwil-db/parse/metadata"
	procedure "github.com/kwilteam/kwil-db/parse/procedures/parser"
	"github.com/kwilteam/kwil-db/parse/procedures/visitors/clean"
	"github.com/kwilteam/kwil-db/parse/procedures/visitors/generate"
	"github.com/kwilteam/kwil-db/parse/procedures/visitors/typing"
)

// AnalyzeOptions allows for setting options for the generation of plpgsql functions.
// These are useful to control generation behavior and error messages, depending on
// the context in which the generation is being used. E.g., certain errors may be
// more useful in an IDE than in the transaction logs.
type AnalyzeOptions struct {
	// LogProcedureNameOnError will log the procedure name in the error message.
	LogProcedureNameOnError bool
}

// AnalyzeProcedures will analyze a procedure, parse the body, and identify the inputs, returns, variables, and body
// that need to be generated for the procedure.
func AnalyzeProcedures(schema *types.Schema, pgSchemaName string, options *AnalyzeOptions) (stmts []*AnalyzedProcedure, err error) {
	stmts = make([]*AnalyzedProcedure, len(schema.Procedures))

	if options == nil {
		options = &AnalyzeOptions{}
	}

	// declaring variables here to allow us to add additional
	// information to the error message.
	var proc *types.Procedure
	var i int
	defer func() {
		if err != nil && options.LogProcedureNameOnError {
			err = fmt.Errorf(`error on procedure "%s": %w`, proc.Name, err)
		}
	}()

	for i, proc = range schema.Procedures {
		parsed, err := procedure.Parse(proc.Body)
		if err != nil {
			return nil, err
		}

		cleanedParams, cleanedSessionVars, err := clean.CleanProcedure(parsed, proc, schema, pgSchemaName, metadata.PgSessionVars)
		if err != nil {
			return nil, err
		}

		anonReceivers, err := typing.EnsureTyping(parsed, proc, schema, cleanedParams, cleanedSessionVars)
		if err != nil {
			return nil, err
		}

		generated, err := generate.GenerateProcedure(parsed, proc, cleanedParams, anonReceivers, pgSchemaName)
		if err != nil {
			return nil, err
		}

		// If the procedure returns out variables (instead of a table), then we need to ensure that the generated names
		// are used, instead of the user-provided names.
		returns := types.ProcedureReturn{}
		if proc.Returns != nil {
			returns = *proc.Returns.Copy()
			if generated.OutVariables != nil {
				returns.Fields = generated.OutVariables
			}
		}

		analyzed := &AnalyzedProcedure{
			Name:              proc.Name,
			Parameters:        cleanedParams,
			Returns:           returns,
			DeclaredVariables: generated.DeclaredVariables,
			LoopTargets:       generated.LoopTargets,
			Body:              generated.Body,
			IsView:            proc.IsView(),
			OwnerOnly:         proc.IsOwnerOnly(),
		}

		stmts[i] = analyzed
	}

	return stmts, nil
}

// AnalyzedProcedure is the result of analyzing a procedure.
type AnalyzedProcedure struct {
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
