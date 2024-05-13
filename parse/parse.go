// package parse contains logic for parsing Kuneiform schemas, procedures, actions,
// and SQL.
package parse

import (
	"fmt"

	"github.com/kwilteam/kwil-db/core/types"
	"github.com/kwilteam/kwil-db/parse/actions"
	"github.com/kwilteam/kwil-db/parse/kuneiform"
	procedures "github.com/kwilteam/kwil-db/parse/procedures"
	parseTypes "github.com/kwilteam/kwil-db/parse/types"
)

// ParseKuneiform parses a Kuneiform schema. It will also perform
// checks on the procedure and SQL syntax, such as syntax validation,
// type checking, and ensuring that all view procedures/actions do not modify state.
func ParseKuneiform(kf string) (*ParseResult, error) {
	// we will return the schema on errors, since language servers will want to
	// have the schema even if there are errors.
	schema, info, errs, err := kuneiform.Parse(kf)
	if err != nil {
		return nil, fmt.Errorf("unknown schema error: %w", err)
	}

	res := &ParseResult{
		Schema:     schema,
		Errs:       errs,
		SchemaInfo: info,
	}
	if res.Err() != nil {
		if schema != nil {
			// try to clean, but ignore errors since the failed parse
			// may have left the schema in a bad state.
			schema.Clean()
		}

		return res, nil
	}

	err = schema.Clean()
	if err != nil {
		return nil, err
	}

	_, actionErrs, err := actions.AnalyzeActions(schema, &actions.AnalyzeOpts{
		PGSchemaName: schema.DBID(),
		SchemaInfo:   info,
	})
	if err != nil {
		return nil, fmt.Errorf("error analyzing actions: %w", err)
	}
	res.Errs.Add(actionErrs...)

	_, procErrs, err := procedures.AnalyzeProcedures(schema, schema.DBID(), &procedures.AnalyzeOptions{
		SchemaInfo: info,
	})
	if err != nil {
		return nil, fmt.Errorf("error analyzing procedures: %w", err)
	}
	res.Errs.Add(procErrs...)

	return res, nil
}

// ParseResult is the result of parsing a Kuneiform schema.
type ParseResult struct {
	Schema     *types.Schema          `json:"schema"`
	Errs       parseTypes.ParseErrors `json:"errors,omitempty"`
	SchemaInfo *parseTypes.SchemaInfo `json:"schema_info,omitempty"`
}

// Err returns all the errors as a single error.
func (r *ParseResult) Err() error {
	if r == nil {
		return nil
	}
	return r.Errs.Err()
}
