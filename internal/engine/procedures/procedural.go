// package procedures ties together the different visitors to generate the plpgsql code for a procedure.
package procedures

import (
	"fmt"

	"github.com/kwilteam/kwil-db/core/types"
	"github.com/kwilteam/kwil-db/internal/engine/ddl"
	"github.com/kwilteam/kwil-db/internal/engine/procedures/clean"
	"github.com/kwilteam/kwil-db/internal/engine/procedures/generate"
	"github.com/kwilteam/kwil-db/internal/engine/procedures/typing"
	"github.com/kwilteam/kwil-db/internal/parse/procedure"
)

// GenerateOptions allows for setting options for the generation of plpgsql functions.
// These are useful to control generation behavior and error messages, depending on
// the context in which the generation is being used. E.g., certain errors may be
// more useful in an IDE than in the transaction logs.
type GenerateOptions struct {
	// LogProcedureNameOnError will log the procedure name in the error message.
	LogProcedureNameOnError bool
}

// GeneratePLPGSQL will prepare the plpgsql functions for all procedures in the schema.
// It takes the schema, the desired postgres schema name, and the schema getter.
// It will return the prepared statements for the procedures.
func GeneratePLPGSQL(schema *types.Schema, pgSchemaName string,
	sessionVarPrefix string, sessionVarTypes map[string]*types.DataType, options *GenerateOptions) (ddlStmts []string, err error) {
	stmts := make([]string, len(schema.Procedures))

	if options == nil {
		options = &GenerateOptions{}
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

		cleanedParams, cleanedSessionVars, err := clean.CleanProcedure(parsed, proc, schema, pgSchemaName, sessionVarPrefix, sessionVarTypes)
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

		sql, err := ddl.GenerateProcedure(cleanedParams, generated.LoopTargets, proc.Returns, generated.DeclaredVariables, generated.OutVariables, pgSchemaName, proc.Name, generated.Body)
		if err != nil {
			return nil, err
		}

		stmts[i] = sql
	}

	return stmts, nil
}
