package parse

import "github.com/kwilteam/kwil-db/core/types"

// TODO: this does not actually do what we think it does.
// Recursive calls in the embedded sqlAnalyzer will call local sqlAnalyzer methods.
// This means that users can call unallowed expressions in actions.
// actionAnalyzer analyzes actions.
type actionAnalyzer struct {
	sqlAnalyzer
	// Mutative is true if the action mutates state.
	Mutative bool
	// schema is the database schema that contains the action.
	schema *types.Schema
	// inSQL is true if the visitor is in an SQL statement.
	inSQL bool
}

var _ Visitor = (*actionAnalyzer)(nil)

func (a *actionAnalyzer) VisitActionStmtSQL(p0 *ActionStmtSQL) any {
	// we simply need to call the sql analyzer to make it check the statement
	// and rewrite it to be deterministic. We can ignore the result.
	a.inSQL = true
	p0.SQL.Accept(&a.sqlAnalyzer)
	a.inSQL = false

	if a.sqlAnalyzer.sqlResult.Mutative {
		a.Mutative = true
	}

	return nil
}

func (a *actionAnalyzer) VisitActionStmtExtensionCall(p0 *ActionStmtExtensionCall) any {
	for _, arg := range p0.Args {
		arg.Accept(&a.sqlAnalyzer)
	}

	_, ok := a.schema.FindExtensionImport(p0.Extension)
	if !ok {
		a.errs.AddErr(p0, ErrActionNotFound, p0.Extension)
	}

	// we need to add all receivers to the known variables
	for _, rec := range p0.Receivers {
		a.blockContext.variables[rec] = types.UnknownType
	}

	return nil
}

func (a *actionAnalyzer) VisitActionStmtActionCall(p0 *ActionStmtActionCall) any {
	for _, arg := range p0.Args {
		arg.Accept(&a.sqlAnalyzer)
	}

	act, ok := a.schema.FindAction(p0.Action)
	if !ok {
		a.errs.AddErr(p0, ErrActionNotFound, p0.Action)
		return nil
	}

	if !act.IsView() {
		a.Mutative = true
	}

	return nil
}
