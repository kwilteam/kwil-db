package kuneiform

import (
	"github.com/kwilteam/kwil-db/core/types"
	"github.com/kwilteam/kwil-db/internal/engine/execution"
	"github.com/kwilteam/kwil-db/internal/engine/procedures"
	actparser "github.com/kwilteam/kwil-db/internal/parse/action"
)

// ValidateStatements validates the syntax of the action and procedure bodies
// in a schema.
func ValidateStatements(schema *types.Schema) error {
	for _, action := range schema.Actions {
		_, err := actparser.Parse(action.Body)
		if err != nil {
			return err
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
