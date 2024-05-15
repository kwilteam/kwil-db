// package procedures ties together the different visitors to generate the plpgsql code for a procedure.
package procedures

import (
	"fmt"

	"github.com/kwilteam/kwil-db/core/types"
	procedure "github.com/kwilteam/kwil-db/parse/procedures/parser"
	"github.com/kwilteam/kwil-db/parse/procedures/visitors/clean"
	"github.com/kwilteam/kwil-db/parse/procedures/visitors/generate"
	"github.com/kwilteam/kwil-db/parse/procedures/visitors/typing"
	parseTypes "github.com/kwilteam/kwil-db/parse/types"
)

// AnalyzeOptions allows for setting options for the generation of plpgsql functions.
// These are useful to control generation behavior and error messages, depending on
// the context in which the generation is being used. E.g., certain errors may be
// more useful in an IDE than in the transaction logs.
type AnalyzeOptions struct {
	// LogProcedureNameOnError will log the procedure name in the error message.
	// This is generally useful if the parser is not being used in an IDE where positional
	// errors are displayed, and we want to tell the user which procedure caused the error.
	LogProcedureNameOnError bool
	// SchemaInfo is the schema information for the procedure.
	// If it is not nil, the schema information will be used to modify the position
	// of the error messages.
	SchemaInfo *parseTypes.SchemaInfo
}

// AnalyzeProcedures will analyze a procedure, parse the body, and identify the inputs, returns, variables, and body
// that need to be generated for the procedure.
func AnalyzeProcedures(schema *types.Schema, pgSchemaName string, options *AnalyzeOptions) (stmts []*AnalyzedProcedure, parseErrs parseTypes.ParseErrors, err error) {
	stmts = make([]*AnalyzedProcedure, len(schema.Procedures))
	parseErrs = make(parseTypes.ParseErrors, 0)

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

		// if schema info is provided, we will use it to modify the error messages.
		errorListener := parseTypes.NewErrorListener()
		startLine := 0
		startCol := 0
		if options.SchemaInfo != nil {
			procPos, ok := options.SchemaInfo.Blocks[proc.Name]
			if !ok {
				// should never happen, this would be a bug in our code
				return nil, nil, fmt.Errorf("could not find position for procedure %s", proc.Name)
			}
			startLine = procPos.StartLine
			startCol = procPos.StartCol
		}
		errorListener = errorListener.Child("procedure", startLine, startCol)

		parsed, err := procedure.ParseWithErrorListener(proc.Body, errorListener)
		if err != nil {
			return nil, nil, err
		}

		// if there are errors, we should not continue generating the procedure
		if errorListener.Err() != nil {
			parseErrs.Add(errorListener.Errors()...)
			continue
		}

		cleanedParams, cleanedSessionVars, err := clean.CleanProcedure(parsed, proc, schema, pgSchemaName, errorListener)
		if err != nil {
			return nil, nil, err
		}

		// if there are errors, we should not continue generating the procedure
		if errorListener.Err() != nil {
			parseErrs.Add(errorListener.Errors()...)
			continue
		}

		anonReceivers, err := typing.EnsureTyping(parsed, proc, schema, cleanedParams, cleanedSessionVars, errorListener)
		if err != nil {
			return nil, nil, err
		}

		// if there are errors, we should not continue generating the procedure
		if errorListener.Err() != nil {
			parseErrs.Add(errorListener.Errors()...)
			continue
		}

		generated, err := generate.GenerateProcedure(parsed, proc, cleanedParams, anonReceivers, pgSchemaName)
		if err != nil {
			return nil, nil, err
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

	return stmts, parseErrs, nil
}

// AnalyzedProcedure is the result of analyzing a procedure.
type AnalyzedProcedure struct {
	// Name is the name of the procedure.
	Name string `json:"name"`
	// Parameters are the parameters, in order, that the procedure is expecting.
	// If no parameters are expected, this will be nil.
	Parameters []*types.NamedType `json:"parameters"`
	// Returns is the expected return type(s) of the procedure.
	// If no return is expected, this will be nil.
	Returns types.ProcedureReturn `json:"returns"`
	// DeclaredVariables are the variables that need to be declared.
	DeclaredVariables []*types.NamedType `json:"declared_variables"`
	// LoopTargets is a list of all variables that are loop targets.
	// They should be declared as RECORD in plpgsql.
	LoopTargets []string `json:"loop_targets"`
	// Body is the plpgsql code for the procedure.
	Body string `json:"body"`
	// IsView is true if the procedure is a view.
	IsView bool `json:"is_view"`
	// OwnerOnly is true if the procedure is owner-only.
	OwnerOnly bool `json:"owner_only"`
}
