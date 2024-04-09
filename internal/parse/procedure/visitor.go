package procedure

import (
	"encoding/hex"
	"strconv"
	"strings"

	"github.com/antlr4-go/antlr/v4"
	"github.com/kwilteam/kwil-db/core/types"
	"github.com/kwilteam/kwil-db/internal/parse/procedure/gen"
	sqlparser "github.com/kwilteam/kwil-db/internal/parse/sql"
)

type proceduralLangVisitor struct {
	*gen.BaseProcedureParserVisitor
}

func (p *proceduralLangVisitor) Visit(tree antlr.ParseTree) interface{} {
	return tree.Accept(p)
}

func (p *proceduralLangVisitor) VisitCall_expression(ctx *gen.Call_expressionContext) interface {
} {
	e := &ExpressionCall{
		Name: ctx.IDENTIFIER().GetText(),
	}

	if ctx.Expression_list() != nil {
		e.Arguments = ctx.Expression_list().Accept(p).([]Expression)
	}

	return e
}

func (p *proceduralLangVisitor) VisitExpr_arithmetic(ctx *gen.Expr_arithmeticContext) interface {
} {
	expr := &ExpressionArithmetic{
		Left:  p.Visit(ctx.Expression(0)).(Expression),
		Right: p.Visit(ctx.Expression(1)).(Expression),
	}

	switch {
	case ctx.PLUS() != nil:
		expr.Operator = ArithmeticOperatorAdd
	case ctx.MINUS() != nil:
		expr.Operator = ArithmeticOperatorSub
	case ctx.MUL() != nil:
		expr.Operator = ArithmeticOperatorMul
	case ctx.DIV() != nil:
		expr.Operator = ArithmeticOperatorDiv
	case ctx.MOD() != nil:
		expr.Operator = ArithmeticOperatorMod
	default:
		panic("invalid arithmetic operator")
	}

	return expr
}

func (p *proceduralLangVisitor) VisitExpr_array_access(ctx *gen.Expr_array_accessContext) interface {
} {
	return &ExpressionArrayAccess{
		Target: p.Visit(ctx.Expression(0)).(Expression),
		Index:  p.Visit(ctx.Expression(1)).(Expression),
	}
}

func (p *proceduralLangVisitor) VisitExpr_blob_literal(ctx *gen.Expr_blob_literalContext) interface {
} {
	b := ctx.BLOB_LITERAL().GetText()
	// trim off beginning 0x
	if b[:2] != "0x" {
		panic("blob literals must start with 0x")
	}

	b = b[2:]

	decoded, err := hex.DecodeString(b)
	if err != nil {
		panic("invalid blob literal")
	}

	return &ExpressionBlobLiteral{
		Value: decoded,
	}
}

func (p *proceduralLangVisitor) VisitExpr_boolean_literal(ctx *gen.Expr_boolean_literalContext) interface {
} {
	return &ExpressionBooleanLiteral{
		Value: strings.ToLower(ctx.GetText()) == "true",
	}
}

func (p *proceduralLangVisitor) VisitExpr_call(ctx *gen.Expr_callContext) interface {
} {
	return p.Visit(ctx.Call_expression())
}

func (p *proceduralLangVisitor) VisitExpr_comparison(ctx *gen.Expr_comparisonContext) interface {
} {
	c := &ExpressionComparison{
		Left:  p.Visit(ctx.Expression(0)).(Expression),
		Right: p.Visit(ctx.Expression(1)).(Expression),
	}

	switch {
	case ctx.EQ() != nil:
		c.Operator = ComparisonOperatorEqual
	case ctx.NEQ() != nil:
		c.Operator = ComparisonOperatorNotEqual
	case ctx.LT() != nil:
		c.Operator = ComparisonOperatorLessThan
	case ctx.LT_EQ() != nil:
		c.Operator = ComparisonOperatorLessThanOrEqual
	case ctx.GT() != nil:
		c.Operator = ComparisonOperatorGreaterThan
	case ctx.GT_EQ() != nil:
		c.Operator = ComparisonOperatorGreaterThanOrEqual
	default:
		panic("invalid comparison operator")
	}

	return c
}

func (p *proceduralLangVisitor) VisitExpr_field_access(ctx *gen.Expr_field_accessContext) interface {
} {
	return &ExpressionFieldAccess{
		Target: p.Visit(ctx.Expression()).(Expression),
		Field:  ctx.IDENTIFIER().GetText(),
	}
}

func (p *proceduralLangVisitor) VisitExpr_int_literal(ctx *gen.Expr_int_literalContext) interface {
} {
	textVal := ctx.INT_LITERAL().GetText()
	i, err := strconv.ParseInt(textVal, 10, 64)
	if err != nil {
		panic("invalid int literal")
	}

	return &ExpressionIntLiteral{
		Value: i,
	}
}

func (p *proceduralLangVisitor) VisitExpr_make_array(ctx *gen.Expr_make_arrayContext) interface {
} {
	return p.Visit(ctx.Expression_make_array())
}

func (p *proceduralLangVisitor) VisitExpr_null_literal(ctx *gen.Expr_null_literalContext) interface {
} {
	return &ExpressionNullLiteral{}
}

func (p *proceduralLangVisitor) VisitExpr_parenthesized(ctx *gen.Expr_parenthesizedContext) interface {
} {
	return &ExpressionParenthesized{
		Expression: p.Visit(ctx.Expression()).(Expression),
	}
}

func (p *proceduralLangVisitor) VisitExpr_text_literal(ctx *gen.Expr_text_literalContext) interface {
} {

	// parse out the quotes
	if len(ctx.TEXT_LITERAL().GetText()) < 2 {
		panic("invalid text literal")
	}

	if ctx.TEXT_LITERAL().GetText()[0] != '\'' || ctx.TEXT_LITERAL().GetText()[len(ctx.TEXT_LITERAL().GetText())-1] != '\'' {
		panic("invalid text literal")
	}

	text := ctx.TEXT_LITERAL().GetText()[1 : len(ctx.TEXT_LITERAL().GetText())-1]

	return &ExpressionTextLiteral{
		Value: text,
	}
}

func (p *proceduralLangVisitor) VisitExpr_variable(ctx *gen.Expr_variableContext) interface {
} {
	return &ExpressionVariable{
		Name: getVariable(ctx.VARIABLE()),
	}
}

func (p *proceduralLangVisitor) VisitExpression_list(ctx *gen.Expression_listContext) interface {
} {
	exprs := make([]Expression, len(ctx.AllExpression()))
	for i, expr := range ctx.AllExpression() {
		exprs[i] = p.Visit(expr).(Expression)
	}

	return exprs
}

func (p *proceduralLangVisitor) VisitExpression_make_array(ctx *gen.Expression_make_arrayContext) interface {
} {
	exprs := p.Visit(ctx.Expression_list()).([]Expression)

	return &ExpressionMakeArray{
		Values: exprs,
	}
}

func (p *proceduralLangVisitor) VisitProgram(ctx *gen.ProgramContext) interface {
} {
	var clauses []Statement
	for _, statement := range ctx.AllStatement() {
		res := p.Visit(statement)
		if res != nil {
			clauses = append(clauses, res.(Statement))
		}
	}

	return clauses
}

func (p *proceduralLangVisitor) VisitRange(ctx *gen.RangeContext) interface {
} {
	return &LoopTargetRange{
		Start: p.Visit(ctx.Expression(0)).(Expression),
		End:   p.Visit(ctx.Expression(1)).(Expression),
	}
}

func (p *proceduralLangVisitor) VisitStmt_procedure_call(ctx *gen.Stmt_procedure_callContext) interface {
} {
	proc := ctx.Call_expression().Accept(p).(*ExpressionCall)

	vars := make([]string, len(ctx.AllVARIABLE()))
	for i, arg := range ctx.AllVARIABLE() {
		vars[i] = getVariable(arg)
	}

	return &StatementProcedureCall{
		Variables: vars,
		Call: &ExpressionCall{
			Name:      proc.Name,
			Arguments: proc.Arguments,
		},
	}
}

func (p *proceduralLangVisitor) VisitStmt_for_loop(ctx *gen.Stmt_for_loopContext) interface {
} {
	forLoop := &StatementForLoop{
		Variable: getVariable(ctx.VARIABLE(0)),
	}

	switch {
	case ctx.Range_() != nil:
		forLoop.Target = ctx.Range_().Accept(p).(*LoopTargetRange)
	case ctx.Call_expression() != nil:
		forLoop.Target = &LoopTargetCall{
			Call: ctx.Call_expression().Accept(p).(*ExpressionCall),
		}
	case ctx.VARIABLE(1) != nil: // TODO: check if this will this panic, or return nil?
		forLoop.Target = &LoopTargetVariable{
			Variable: &ExpressionVariable{
				Name: getVariable(ctx.VARIABLE(1)),
			},
		}
	case ctx.ANY_SQL() != nil:
		ast, err := sqlparser.Parse(ctx.ANY_SQL().GetText())
		if err != nil {
			panic("invalid SQL statement")
		}

		forLoop.Target = &LoopTargetSQL{
			Statement: ast,
		}
	}

	stmts := make([]Statement, len(ctx.AllStatement()))
	for i, stmt := range ctx.AllStatement() {
		stmts[i] = p.Visit(stmt).(Statement)
	}
	forLoop.Body = stmts

	return forLoop
}

func (p *proceduralLangVisitor) VisitStmt_if(ctx *gen.Stmt_ifContext) interface {
} {
	ifThens := ctx.AllIf_then_block()
	ifClause := &StatementIf{
		IfThens: make([]*IfThen, len(ifThens)),
	}

	for i, ifThen := range ifThens {
		ifClause.IfThens[i] = p.Visit(ifThen).(*IfThen)
	}

	if ctx.ELSE() != nil {
		stmts := make([]Statement, len(ctx.AllStatement()))
		for i, stmt := range ctx.AllStatement() {
			stmts[i] = p.Visit(stmt).(Statement)
		}

		ifClause.Else = stmts
	}

	return ifClause
}

func (p *proceduralLangVisitor) VisitIf_then_block(ctx *gen.If_then_blockContext) interface {
} {

	stmts := make([]Statement, len(ctx.AllStatement()))
	for i, stmt := range ctx.AllStatement() {
		stmts[i] = p.Visit(stmt).(Statement)
	}

	return &IfThen{
		If:   p.Visit(ctx.Expression()).(Expression),
		Then: stmts,
	}
}

func (p *proceduralLangVisitor) VisitStmt_return(ctx *gen.Stmt_returnContext) interface {
} {
	if ctx.Expression_list() != nil {
		return &StatementReturn{
			Values: ctx.Expression_list().Accept(p).([]Expression),
		}
	}

	if ctx.ANY_SQL() != nil {
		ast, err := sqlparser.Parse(ctx.ANY_SQL().GetText())
		if err != nil {
			panic("invalid SQL statement")
		}

		return &StatementReturn{
			SQL: ast,
		}
	}

	return &StatementReturn{}
}

func (p *proceduralLangVisitor) VisitStmt_return_next(ctx *gen.Stmt_return_nextContext) interface {
} {
	return &StatementReturnNext{
		Variable: getVariable(ctx.VARIABLE()),
	}
}

func (p *proceduralLangVisitor) VisitStmt_break(ctx *gen.Stmt_breakContext) interface{} {
	return &StatementBreak{}
}

func (p *proceduralLangVisitor) VisitStmt_sql(ctx *gen.Stmt_sqlContext) interface {
} {
	ast, err := sqlparser.Parse(ctx.ANY_SQL().GetText())
	if err != nil {
		panic("invalid SQL statement")
	}

	return &StatementSQL{
		Statement: ast,
	}
}

func (p *proceduralLangVisitor) VisitStmt_variable_assignment(ctx *gen.Stmt_variable_assignmentContext) interface {
} {
	return &StatementVariableAssignment{
		Name:  getVariable(ctx.VARIABLE()),
		Value: p.Visit(ctx.Expression()).(Expression),
	}
}

func (p *proceduralLangVisitor) VisitStmt_variable_assignment_with_declaration(ctx *gen.Stmt_variable_assignment_with_declarationContext) interface {
} {
	return &StatementVariableAssignmentWithDeclaration{
		Name:  getVariable(ctx.VARIABLE()),
		Type:  getType(ctx.Type_()),
		Value: p.Visit(ctx.Expression()).(Expression),
	}
}

func (p *proceduralLangVisitor) VisitStmt_variable_declaration(ctx *gen.Stmt_variable_declarationContext) interface {
} {
	return &StatementVariableDeclaration{
		Name: getVariable(ctx.VARIABLE()),
		Type: getType(ctx.Type_()),
	}
}

func getVariable(v antlr.TerminalNode) string {
	t := v.GetText()
	// trim off beginning $
	if t[0] != '$' && t[0] != '@' {
		panic("variable names must start with $ or @")
	}

	return t
}

func getType(t gen.ITypeContext) *types.DataType {
	return &types.DataType{
		Name:    t.IDENTIFIER().GetText(),
		IsArray: t.LBRACKET() != nil,
	}
}
