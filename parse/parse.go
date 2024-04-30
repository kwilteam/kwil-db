// package parse contains logic for parsing Kuneiform schemas, procedures, actions,
// and SQL.
package parse

import (
	"github.com/kwilteam/kwil-db/core/types"
	"github.com/kwilteam/kwil-db/parse/actions"
	"github.com/kwilteam/kwil-db/parse/kuneiform"
	procedures "github.com/kwilteam/kwil-db/parse/procedures"
)

// ValidateSchema validates the syntax of the action and procedure bodies
// in a schema.
func ValidateSchema(schema *types.Schema) error {
	// to validate, we will simply call the analyzers and discard
	// the results. If there is an error, we will return it.
	_, err := actions.AnalyzeActions(schema, schema.DBID())
	if err != nil {
		return err
	}

	_, err = procedures.AnalyzeProcedures(schema, schema.DBID(), &procedures.AnalyzeOptions{
		LogProcedureNameOnError: true,
	})
	if err != nil {
		return err
	}

	return nil
}

// ParseKuneiform parses a Kuneiform schema. It will also perform
// checks on the procedure and SQL syntax, such as syntax validation,
// type checking, and ensuring that all view procedures/actions do not modify state.
func ParseKuneiform(kf string) (*types.Schema, error) {
	schema, err := kuneiform.Parse(kf)
	if err != nil {
		return nil, err
	}

	err = schema.Clean()
	if err != nil {
		return nil, err
	}

	if err := ValidateSchema(schema); err != nil {
		return nil, err
	}

	return schema, nil
}
