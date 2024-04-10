package kuneiform

import (
	"fmt"

	"github.com/kwilteam/kwil-db/core/types"
	"github.com/kwilteam/kwil-db/internal/engine/execution"
	"github.com/kwilteam/kwil-db/internal/engine/procedures"
	"github.com/kwilteam/kwil-db/internal/engine/sqlanalyzer"
)

// ValidateStatements validates the syntax of the action and procedure bodies
// in a schema.
func ValidateStatements(schema *types.Schema) error {
	for _, action := range schema.Actions {
		res, err := sqlanalyzer.ApplyRules(action.Body, sqlanalyzer.AllRules, schema.Tables, "pg_schema")
		if err != nil {
			return err
		}

		if action.IsView() && res.Mutative {
			return fmt.Errorf(`action "%s" is marked view, but contains mutative statements`, action.Name)
		}
	}

	if len(schema.Procedures) > 0 {
		_, err := procedures.GeneratePLPGSQL(schema, "pg_schema", "ctx", execution.PgSessionVars)
		if err != nil {
			return err
		}
	}

	return nil
}
