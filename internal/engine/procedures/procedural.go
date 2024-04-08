// package procedural ties together the different visitors to generate the plpgsql code for a procedure.
package procedural

import (
	"github.com/kwilteam/kwil-db/core/types"
	"github.com/kwilteam/kwil-db/internal/engine/ddl"
	"github.com/kwilteam/kwil-db/internal/engine/procedures/parser"
	"github.com/kwilteam/kwil-db/internal/engine/procedures/visitors/clean"
	"github.com/kwilteam/kwil-db/internal/engine/procedures/visitors/generate"
	"github.com/kwilteam/kwil-db/internal/engine/procedures/visitors/typing"
)

// PreparedProcedures will prepare the plpgsql functions for all procedures in the schema.
// It takes the schema, the desired postgres schema name, and the schema getter.
// It will return the prepared statements for the procedures.
func GeneratePLPGSQL(schema *types.Schema, pgSchemaName string,
	sessionVarPrefix string, sessionVarTypes map[string]*types.DataType) ([]string, error) {
	stmts := make([]string, len(schema.Procedures))

	for i, proc := range schema.Procedures {
		parsed, err := parser.Parse(proc.Body)
		if err != nil {
			return nil, err
		}

		cleanedParams, err := clean.CleanProcedure(parsed, proc, schema, pgSchemaName, sessionVarPrefix, sessionVarTypes)
		if err != nil {
			return nil, err
		}

		err = typing.EnsureTyping(parsed, proc, schema, cleanedParams)
		if err != nil {
			return nil, err
		}

		generated, err := generate.GenerateProcedure(parsed, proc, cleanedParams)
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
