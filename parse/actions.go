package parse

import "github.com/kwilteam/kwil-db/core/types"

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

func (a *actionAnalyzer) VisitExtensionCallStmt(p0 *ActionStmtExtensionCall) any {
	for _, arg := range p0.Args {
		arg.Accept(&a.sqlAnalyzer)
	}

	_, ok := a.schema.FindExtensionImport(p0.Extension)
	if !ok {
		a.errs.AddErr(p0, ErrActionNotFound, p0.Extension)
	}

	return nil
}

func (a *actionAnalyzer) VisitActionCallStmt(p0 *ActionStmtActionCall) any {
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

// we need to selectively throw errors on dis-allowed expressions on non-sql visits
// actions only support:
// - Variables
// - Unary expressions
// - Binary expressions
// - Function calls
// - Arithmetic expressions
// - Parenthesized expressions
// Everything else must return errors when not in SQL

func (a *actionAnalyzer) VisitExpressionForeignCall(p0 *ExpressionForeignCall) any {
	if !a.inSQL {
		a.errs.AddErr(p0, ErrInvalidActionExpression, "in-line action statements do not support foreign calls")
	}

	return a.sqlAnalyzer.VisitExpressionForeignCall(p0)
}

func (a *actionAnalyzer) VisitExpressionArrayAccess(p0 *ExpressionArrayAccess) any {
	if !a.inSQL {
		a.errs.AddErr(p0, ErrInvalidActionExpression, "in-line action statements do not support arrays")
	}

	return a.sqlAnalyzer.VisitExpressionArrayAccess(p0)
}

func (a *actionAnalyzer) VisitExpressionMakeArray(p0 *ExpressionMakeArray) any {
	if !a.inSQL {
		a.errs.AddErr(p0, ErrInvalidActionExpression, "in-line action statements do not support array creation")
	}

	return a.sqlAnalyzer.VisitExpressionMakeArray(p0)
}

func (a *actionAnalyzer) VisitExpressionFieldAccess(p0 *ExpressionFieldAccess) any {
	if !a.inSQL {
		a.errs.AddErr(p0, ErrInvalidActionExpression, "in-line action statements do not support compound variables")
	}

	return a.sqlAnalyzer.VisitExpressionFieldAccess(p0)
}

func (a *actionAnalyzer) VisitExpressionLogical(p0 *ExpressionLogical) any {
	if !a.inSQL {
		a.errs.AddErr(p0, ErrInvalidActionExpression, "in-line action statements do not support logical expressions")
	}

	return a.sqlAnalyzer.VisitExpressionLogical(p0)
}

func (a *actionAnalyzer) VisitExpressionColumn(p0 *ExpressionColumn) any {
	if !a.inSQL {
		a.errs.AddErr(p0, ErrInvalidActionExpression, "in-line action statements do not support column references")
	}

	return a.sqlAnalyzer.VisitExpressionColumn(p0)
}

func (a *actionAnalyzer) VisitExpressionCollate(p0 *ExpressionCollate) any {
	if !a.inSQL {
		a.errs.AddErr(p0, ErrInvalidActionExpression, "in-line action statements do not support collation")
	}

	return a.sqlAnalyzer.VisitExpressionCollate(p0)
}

func (a *actionAnalyzer) VisitExpressionStringComparison(p0 *ExpressionStringComparison) any {
	if !a.inSQL {
		a.errs.AddErr(p0, ErrInvalidActionExpression, "in-line action statements do not support string comparisons")
	}

	return a.sqlAnalyzer.VisitExpressionStringComparison(p0)
}

func (a *actionAnalyzer) VisitExpressionIs(p0 *ExpressionIs) any {
	if !a.inSQL {
		a.errs.AddErr(p0, ErrInvalidActionExpression, "in-line action statements do not support IS expressions")
	}

	return a.sqlAnalyzer.VisitExpressionIs(p0)
}

func (a *actionAnalyzer) VisitExpressionIn(p0 *ExpressionIn) any {
	if !a.inSQL {
		a.errs.AddErr(p0, ErrInvalidActionExpression, "in-line action statements do not support IN expressions")
	}

	return a.sqlAnalyzer.VisitExpressionIn(p0)
}

func (a *actionAnalyzer) VisitExpressionBetween(p0 *ExpressionBetween) any {
	if !a.inSQL {
		a.errs.AddErr(p0, ErrInvalidActionExpression, "in-line action statements do not support BETWEEN expressions")
	}

	return a.sqlAnalyzer.VisitExpressionBetween(p0)
}

func (a *actionAnalyzer) VisitExpressionSubquery(p0 *ExpressionSubquery) any {
	if !a.inSQL {
		a.errs.AddErr(p0, ErrInvalidActionExpression, "in-line action statements do not support subqueries")
	}

	return a.sqlAnalyzer.VisitExpressionSubquery(p0)
}

func (a *actionAnalyzer) VisitExpressionCase(p0 *ExpressionCase) any {
	if !a.inSQL {
		a.errs.AddErr(p0, ErrInvalidActionExpression, "in-line action statements do not support CASE expressions")
	}

	return a.sqlAnalyzer.VisitExpressionCase(p0)
}
