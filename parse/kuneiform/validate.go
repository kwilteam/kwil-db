package kuneiform

import (
	"github.com/kwilteam/kwil-db/core/types"
	actparser "github.com/kwilteam/kwil-db/parse/actions/parser"
	procedures "github.com/kwilteam/kwil-db/parse/procedures"
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
		_, err := procedures.AnalyzeProcedures(schema, "pg_schema", &procedures.AnalyzeOptions{
			LogProcedureNameOnError: true,
		})
		if err != nil {
			return err
		}
	}

	return nil
}
