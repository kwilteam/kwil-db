// package  implements a visitor that visits all children of every node in the AST.
// No visitor methods will return anything.
package traverse

import parser "github.com/kwilteam/kwil-db/internal/parse/procedure"

// r s all nodes in the AST.
// It can be given a set of functions to call when it visits a node of a certain type.
// It guarantees no order of traversal, and is non-deterministic.
type Traverser struct {
	StatementVariableDeclaration               func(*parser.StatementVariableDeclaration)
	StatementVariableAssignment                func(*parser.StatementVariableAssignment)
	StatementVariableAssignmentWithDeclaration func(*parser.StatementVariableAssignmentWithDeclaration)
	StatementProcedureCall                     func(*parser.StatementProcedureCall)
	StatementForLoop                           func(*parser.StatementForLoop)
	StatementIf                                func(*parser.StatementIf)
	StatementSQL                               func(*parser.StatementSQL)
	StatementReturn                            func(*parser.StatementReturn)
	StatementReturnNext                        func(*parser.StatementReturnNext)
	StatementBreak                             func(*parser.StatementBreak)
	ExpressionTextLiteral                      func(*parser.ExpressionTextLiteral)
	ExpressionBooleanLiteral                   func(*parser.ExpressionBooleanLiteral)
	ExpressionIntLiteral                       func(*parser.ExpressionIntLiteral)
	ExpressionNullLiteral                      func(*parser.ExpressionNullLiteral)
	ExpressionBlobLiteral                      func(*parser.ExpressionBlobLiteral)
	ExpressionMakeArray                        func(*parser.ExpressionMakeArray)
	ExpressionCall                             func(*parser.ExpressionCall)
	ExpressionVariable                         func(*parser.ExpressionVariable)
	ExpressionArrayAccess                      func(*parser.ExpressionArrayAccess)
	ExpressionFieldAccess                      func(*parser.ExpressionFieldAccess)
	ExpressionParenthesized                    func(*parser.ExpressionParenthesized)
	ExpressionComparison                       func(*parser.ExpressionComparison)
	ExpressionArithmetic                       func(*parser.ExpressionArithmetic)
	LoopTargetRange                            func(*parser.LoopTargetRange)
	LoopTargetCall                             func(*parser.LoopTargetCall)
	LoopTargetVariable                         func(*parser.LoopTargetVariable)
	LoopTargetSQL                              func(*parser.LoopTargetSQL)
}

var _ parser.Visitor = &Traverser{}

func (v *Traverser) VisitStatementVariableDeclaration(s *parser.StatementVariableDeclaration) interface{} {
	if v.StatementVariableDeclaration != nil {
		v.StatementVariableDeclaration(s)
	}
	return nil
}
func (v *Traverser) VisitStatementVariableAssignment(s *parser.StatementVariableAssignment) interface{} {
	if v.StatementVariableAssignment != nil {
		v.StatementVariableAssignment(s)
	}
	s.Value.Accept(v)
	return nil
}
func (v *Traverser) VisitStatementVariableAssignmentWithDeclaration(s *parser.StatementVariableAssignmentWithDeclaration) interface{} {
	if v.StatementVariableAssignmentWithDeclaration != nil {
		v.StatementVariableAssignmentWithDeclaration(s)
	}
	s.Value.Accept(v)
	return nil
}
func (v *Traverser) VisitStatementProcedureCall(s *parser.StatementProcedureCall) interface{} {
	if v.StatementProcedureCall != nil {
		v.StatementProcedureCall(s)
	}

	s.Call.Accept(v)

	return nil
}
func (v *Traverser) VisitStatementForLoop(s *parser.StatementForLoop) interface{} {
	if v.StatementForLoop != nil {
		v.StatementForLoop(s)
	}

	s.Target.Accept(v)

	for _, stmt := range s.Body {
		stmt.Accept(v)
	}

	return nil
}
func (v *Traverser) VisitStatementIf(s *parser.StatementIf) interface{} {
	if v.StatementIf != nil {
		v.StatementIf(s)
	}

	for _, ifthen := range s.IfThens {
		ifthen.If.Accept(v)

		for _, stmt := range ifthen.Then {
			stmt.Accept(v)
		}
	}

	for _, stmt := range s.Else {
		stmt.Accept(v)
	}

	return nil
}
func (v *Traverser) VisitStatementSQL(s *parser.StatementSQL) interface{} {
	if v.StatementSQL != nil {
		v.StatementSQL(s)
	}

	return nil
}
func (v *Traverser) VisitStatementReturn(s *parser.StatementReturn) interface{} {
	if v.StatementReturn != nil {
		v.StatementReturn(s)
	}

	for _, expr := range s.Values {
		expr.Accept(v)
	}

	return nil
}

func (v *Traverser) VisitStatementReturnNext(s *parser.StatementReturnNext) interface{} {
	if v.StatementReturnNext != nil {
		v.StatementReturnNext(s)
	}

	for _, expr := range s.Returns {
		expr.Accept(v)
	}

	return nil
}

func (v *Traverser) VisitStatementBreak(s *parser.StatementBreak) interface{} {
	if v.StatementBreak != nil {
		v.StatementBreak(s)
	}

	return nil
}
func (v *Traverser) VisitExpressionTextLiteral(s *parser.ExpressionTextLiteral) interface{} {
	if v.ExpressionTextLiteral != nil {
		v.ExpressionTextLiteral(s)
	}
	return nil
}
func (v *Traverser) VisitExpressionBooleanLiteral(s *parser.ExpressionBooleanLiteral) interface{} {
	if v.ExpressionBooleanLiteral != nil {
		v.ExpressionBooleanLiteral(s)
	}
	return nil
}
func (v *Traverser) VisitExpressionIntLiteral(s *parser.ExpressionIntLiteral) interface{} {
	if v.ExpressionIntLiteral != nil {
		v.ExpressionIntLiteral(s)
	}
	return nil
}
func (v *Traverser) VisitExpressionNullLiteral(s *parser.ExpressionNullLiteral) interface{} {
	if v.ExpressionNullLiteral != nil {
		v.ExpressionNullLiteral(s)
	}
	return nil
}
func (v *Traverser) VisitExpressionBlobLiteral(s *parser.ExpressionBlobLiteral) interface{} {
	if v.ExpressionBlobLiteral != nil {
		v.ExpressionBlobLiteral(s)
	}
	return nil
}
func (v *Traverser) VisitExpressionMakeArray(s *parser.ExpressionMakeArray) interface{} {
	if v.ExpressionMakeArray != nil {
		v.ExpressionMakeArray(s)
	}

	for _, elem := range s.Values {
		elem.Accept(v)
	}

	return nil
}
func (v *Traverser) VisitExpressionCall(s *parser.ExpressionCall) interface{} {
	if v.ExpressionCall != nil {
		v.ExpressionCall(s)
	}

	for _, arg := range s.Arguments {
		arg.Accept(v)
	}

	return nil
}
func (v *Traverser) VisitExpressionVariable(s *parser.ExpressionVariable) interface{} {
	if v.ExpressionVariable != nil {
		v.ExpressionVariable(s)
	}
	return nil
}
func (v *Traverser) VisitExpressionArrayAccess(s *parser.ExpressionArrayAccess) interface{} {
	if v.ExpressionArrayAccess != nil {
		v.ExpressionArrayAccess(s)
	}
	s.Target.Accept(v)
	s.Index.Accept(v)
	return nil
}
func (v *Traverser) VisitExpressionFieldAccess(s *parser.ExpressionFieldAccess) interface{} {
	if v.ExpressionFieldAccess != nil {
		v.ExpressionFieldAccess(s)
	}
	s.Target.Accept(v)
	return nil
}
func (v *Traverser) VisitExpressionParenthesized(s *parser.ExpressionParenthesized) interface{} {
	if v.ExpressionParenthesized != nil {
		v.ExpressionParenthesized(s)
	}
	s.Expression.Accept(v)
	return nil
}
func (v *Traverser) VisitExpressionComparison(s *parser.ExpressionComparison) interface{} {
	if v.ExpressionComparison != nil {
		v.ExpressionComparison(s)
	}
	s.Left.Accept(v)
	s.Right.Accept(v)
	return nil
}
func (v *Traverser) VisitExpressionArithmetic(s *parser.ExpressionArithmetic) interface{} {
	if v.ExpressionArithmetic != nil {
		v.ExpressionArithmetic(s)
	}
	s.Left.Accept(v)
	s.Right.Accept(v)
	return nil
}

func (v *Traverser) VisitLoopTargetRange(s *parser.LoopTargetRange) interface{} {
	if v.LoopTargetRange != nil {
		v.LoopTargetRange(s)
	}
	s.Start.Accept(v)
	s.End.Accept(v)
	return nil
}

func (v *Traverser) VisitLoopTargetCall(s *parser.LoopTargetCall) interface{} {
	if v.LoopTargetCall != nil {
		v.LoopTargetCall(s)
	}
	s.Call.Accept(v)
	return nil
}

func (v *Traverser) VisitLoopTargetVariable(s *parser.LoopTargetVariable) interface{} {
	if v.LoopTargetVariable != nil {
		v.LoopTargetVariable(s)
	}
	s.Variable.Accept(v)
	return nil
}

func (v *Traverser) VisitLoopTargetSQL(s *parser.LoopTargetSQL) interface{} {
	if v.LoopTargetSQL != nil {
		v.LoopTargetSQL(s)
	}
	return nil
}
