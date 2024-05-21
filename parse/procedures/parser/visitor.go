package parser

import (
	"encoding/hex"
	"errors"
	"fmt"
	"math"
	"math/big"
	"strconv"
	"strings"

	"github.com/antlr4-go/antlr/v4"
	"github.com/holiman/uint256"
	"github.com/kwilteam/kwil-db/core/types"
	"github.com/kwilteam/kwil-db/core/types/decimal"
	"github.com/kwilteam/kwil-db/parse/procedures/gen"
	sqlparser "github.com/kwilteam/kwil-db/parse/sql"
	"github.com/kwilteam/kwil-db/parse/sql/tree"
	parseTypes "github.com/kwilteam/kwil-db/parse/types"
)

type proceduralLangVisitor struct {
	*gen.BaseProcedureParserVisitor
	errs parseTypes.AntlrErrorListener
}

func (p *proceduralLangVisitor) Visit(tree antlr.ParseTree) interface{} {
	return tree.Accept(p)
}

func (p *proceduralLangVisitor) VisitNormal_call(ctx *gen.Normal_callContext) any {
	e := &ExpressionCall{
		Name: ctx.IDENTIFIER().GetText(),
	}

	if ctx.Expression_list() != nil {
		e.Arguments = ctx.Expression_list().Accept(p).([]Expression)
	}

	e.Set(ctx)

	return e
}

func (p *proceduralLangVisitor) VisitForeign_call(ctx *gen.Foreign_callContext) any {
	e := &ExpressionForeignCall{
		Name: ctx.IDENTIFIER().GetText(),
	}

	dbid := ctx.GetDbid()
	if dbid == nil {
		// this should get caught by the parser
		p.errs.RuleErr(ctx, parseTypes.ParseErrorTypeSyntax, errors.New("missing dbid"))
	}

	procedure := ctx.GetProcedure()
	if procedure == nil {
		// this should get caught by the parser
		p.errs.RuleErr(ctx, parseTypes.ParseErrorTypeSyntax, errors.New("missing procedure"))
	}

	e.ContextArgs = []Expression{dbid.Accept(p).(Expression), procedure.Accept(p).(Expression)}

	if ctx.Expression_list() != nil {
		e.Arguments = ctx.Expression_list().Accept(p).([]Expression)
	}

	e.Set(ctx)

	return e

}

func (p *proceduralLangVisitor) VisitExpr_arithmetic(ctx *gen.Expr_arithmeticContext) any {
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

	expr.Set(ctx)

	return expr
}

func (p *proceduralLangVisitor) VisitExpr_array_access(ctx *gen.Expr_array_accessContext) any {
	e := &ExpressionArrayAccess{
		Target: p.Visit(ctx.Expression(0)).(Expression),
		Index:  p.Visit(ctx.Expression(1)).(Expression),
	}

	if ctx.Type_cast() != nil {
		e.TypeCast = ctx.Type_cast().Accept(p).(*types.DataType)
	}

	e.Set(ctx)

	return e
}

func (p *proceduralLangVisitor) VisitExpr_literal(ctx *gen.Expr_literalContext) any {
	l := ctx.Literal().Accept(p).(interface{ Cast(*types.DataType) })
	if ctx.Type_cast() != nil {
		l.Cast(ctx.Type_cast().Accept(p).(*types.DataType))
	}

	return l
}

var (
	maxInt64 = big.NewInt(math.MaxInt64)
	minInt64 = big.NewInt(math.MinInt64)
)

func (p *proceduralLangVisitor) VisitLiteral(ctx *gen.LiteralContext) any {
	switch {
	case ctx.INT_LITERAL() != nil:
		txt := ctx.INT_LITERAL().GetText()

		// parse as big int, then if it is too large for int64, we will
		// make it a uint256
		bigNum := new(big.Int)
		_, ok := bigNum.SetString(txt, 10)
		if !ok {
			// this should get caught by the parser
			p.errs.RuleErr(ctx, parseTypes.ParseErrorTypeSyntax, errors.New("invalid integer literal"))
		}

		if bigNum.Cmp(maxInt64) > 0 {
			// make it a uint256
			u256, ok := uint256.FromBig(bigNum)
			if !ok {
				// this should get caught by the parser
				p.errs.RuleErr(ctx, parseTypes.ParseErrorTypeSyntax, errors.New("invalid integer literal"))
				return &ExpressionNullLiteral{}
			}

			return &ExpressionUint256Literal{
				Value: u256,
			}
		}
		if bigNum.Cmp(minInt64) < 0 {
			// error
			p.errs.RuleErr(ctx, parseTypes.ParseErrorTypeSyntax, errors.New("invalid integer literal"))
			return &ExpressionNullLiteral{}
		}

		return &ExpressionIntLiteral{
			Value: bigNum.Int64(),
		}
	case ctx.TEXT_LITERAL() != nil:
		txt := ctx.TEXT_LITERAL().GetText()
		// trim off beginning and ending quotes
		if txt[0] != '\'' || txt[len(txt)-1] != '\'' {
			// this should get caught by the parser
			p.errs.RuleErr(ctx, parseTypes.ParseErrorTypeSyntax, errors.New("invalid text literal"))
		}

		return &ExpressionTextLiteral{
			Value: txt[1 : len(txt)-1],
		}
	case ctx.BLOB_LITERAL() != nil:
		b := ctx.BLOB_LITERAL().GetText()
		// trim off beginning 0x
		if b[:2] != "0x" {
			// this should get caught by the parser
			p.errs.RuleErr(ctx, parseTypes.ParseErrorTypeSyntax, errors.New("invalid blob literal"))
		}

		b = b[2:]

		decoded, err := hex.DecodeString(b)
		if err != nil {
			// this should get caught by the parser
			p.errs.RuleErr(ctx, parseTypes.ParseErrorTypeSyntax, errors.New("invalid blob literal"))
		}

		return &ExpressionBlobLiteral{
			Value: decoded,
		}
	case ctx.BOOLEAN_LITERAL() != nil:
		return &ExpressionBooleanLiteral{
			Value: strings.ToLower(ctx.BOOLEAN_LITERAL().GetText()) == "true",
		}
	case ctx.NULL_LITERAL() != nil:
		return &ExpressionNullLiteral{}
	case ctx.FIXED_LITERAL() != nil:
		txt := ctx.FIXED_LITERAL().GetText()
		// should be of form 123.456
		parts := strings.Split(txt, ".")
		if len(parts) != 2 {
			// this should get caught by the parser
			p.errs.RuleErr(ctx, parseTypes.ParseErrorTypeSyntax, errors.New("invalid fixed literal"))
			return &ExpressionNullLiteral{}
		}

		dec, err := decimal.NewFromString(ctx.FIXED_LITERAL().GetText())
		if err != nil {
			p.errs.TokenErr(ctx.FIXED_LITERAL().GetSymbol(), parseTypes.ParseErrorTypeSemantic, err)
			return &ExpressionNullLiteral{}
		}

		return &ExpressionDecimalLiteral{
			Value: dec,
		}
	default:
		// this should never happen
		panic("invalid literal")
	}
}

func (p *proceduralLangVisitor) VisitLiteral_list(ctx *gen.Literal_listContext) any {
	literals := make([]Expression, len(ctx.AllLiteral()))
	for i, lit := range ctx.AllLiteral() {
		literals[i] = lit.Accept(p).(Expression)
	}

	return literals
}

func (p *proceduralLangVisitor) VisitExpr_call(ctx *gen.Expr_callContext) any {
	e := p.Visit(ctx.Call_expression())

	var tc *types.DataType
	if ctx.Type_cast() != nil {
		tc = ctx.Type_cast().Accept(p).(*types.DataType)
	}

	switch v := e.(type) {
	case *ExpressionCall:
		v.Set(ctx)
		v.TypeCast = tc
	case *ExpressionForeignCall:
		v.Set(ctx)
		v.TypeCast = tc
	default:
		// should never happen
		panic(fmt.Sprintf("invalid type cast for %T", e))
	}

	return e
}

func (p *proceduralLangVisitor) VisitExpr_comparison(ctx *gen.Expr_comparisonContext) any {
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

	c.Set(ctx)

	return c
}

func (p *proceduralLangVisitor) VisitExpr_field_access(ctx *gen.Expr_field_accessContext) any {
	e := &ExpressionFieldAccess{
		Target: p.Visit(ctx.Expression()).(Expression),
		Field:  ctx.IDENTIFIER().GetText(),
	}

	if ctx.Type_cast() != nil {
		e.TypeCast = ctx.Type_cast().Accept(p).(*types.DataType)
	}

	e.Set(ctx)

	return e
}

func (p *proceduralLangVisitor) VisitExpr_make_array(ctx *gen.Expr_make_arrayContext) any {
	e := p.Visit(ctx.Expression_make_array())

	var tc *types.DataType
	if ctx.Type_cast() != nil {
		tc = ctx.Type_cast().Accept(p).(*types.DataType)
	}

	e2, ok := e.(*ExpressionMakeArray)
	if !ok {
		p.errs.RuleErr(ctx, parseTypes.ParseErrorTypeUnknown, errors.New("expected array literal"))
	}

	e2.Set(ctx)
	e2.TypeCast = tc

	return e
}

func (p *proceduralLangVisitor) VisitExpr_parenthesized(ctx *gen.Expr_parenthesizedContext) any {
	e := &ExpressionParenthesized{
		Expression: p.Visit(ctx.Expression()).(Expression),
	}

	if ctx.Type_cast() != nil {
		e.TypeCast = ctx.Type_cast().Accept(p).(*types.DataType)
	}

	e.Set(ctx)

	return e
}

func (p *proceduralLangVisitor) VisitExpr_variable(ctx *gen.Expr_variableContext) any {
	e := &ExpressionVariable{
		Name: getVariable(ctx.VARIABLE()),
	}

	if ctx.Type_cast() != nil {
		e.TypeCast = ctx.Type_cast().Accept(p).(*types.DataType)
	}

	e.Set(ctx)

	return e
}

func (p *proceduralLangVisitor) VisitExpression_list(ctx *gen.Expression_listContext) any {
	exprs := make([]Expression, len(ctx.AllExpression()))
	for i, expr := range ctx.AllExpression() {
		exprs[i] = p.Visit(expr).(Expression)
	}

	return exprs
}

func (p *proceduralLangVisitor) VisitExpression_make_array(ctx *gen.Expression_make_arrayContext) any {
	exprs := p.Visit(ctx.Expression_list()).([]Expression)

	e := &ExpressionMakeArray{
		Values: exprs,
	}

	// we do not handle type casts here, they are handled in the parent
	e.Set(ctx)

	return e
}

func (p *proceduralLangVisitor) VisitProgram(ctx *gen.ProgramContext) any {
	var clauses []Statement
	for _, statement := range ctx.AllStatement() {
		res := p.Visit(statement)
		if res != nil {
			clauses = append(clauses, res.(Statement))
		}
	}

	return clauses
}

func (p *proceduralLangVisitor) VisitRange(ctx *gen.RangeContext) any {
	e := &LoopTargetRange{
		Start: p.Visit(ctx.Expression(0)).(Expression),
		End:   p.Visit(ctx.Expression(1)).(Expression),
	}

	e.Set(ctx)

	return e
}

func (p *proceduralLangVisitor) VisitStmt_procedure_call(ctx *gen.Stmt_procedure_callContext) any {
	s := &StatementProcedureCall{
		Call: ctx.Call_expression().Accept(p).(ICallExpression),
	}

	if len(ctx.AllVariable_or_underscore()) > 0 {
		s.Variables = make([]*string, len(ctx.AllVariable_or_underscore()))
	}
	for i, arg := range ctx.AllVariable_or_underscore() {
		// can either be *string or nil
		v := arg.Accept(p)
		varStr, ok := v.(*string)
		if ok {
			s.Variables[i] = varStr
		} else {
			// check if it's nil
			if v != nil {
				// this would be a bug
				p.errs.RuleErr(ctx, parseTypes.ParseErrorTypeUnknown, errors.New("invalid variable"))
			}
			s.Variables[i] = nil
		}
	}

	s.Set(ctx)
	return s
}

func (p *proceduralLangVisitor) VisitVariable_or_underscore(ctx *gen.Variable_or_underscoreContext) any {
	if ctx.UNDERSCORE() != nil {
		return nil
	}
	if ctx.VARIABLE() != nil {
		v := getVariable(ctx.VARIABLE())
		return &v
	}

	// this should never happen
	p.errs.RuleErr(ctx, parseTypes.ParseErrorTypeUnknown, errors.New("invalid variable"))
	return nil
}

func (p *proceduralLangVisitor) VisitStmt_for_loop(ctx *gen.Stmt_for_loopContext) any {
	forLoop := &StatementForLoop{
		Variable: getVariable(ctx.VARIABLE(0)),
	}

	switch {
	case ctx.Range_() != nil:
		forLoop.Target = ctx.Range_().Accept(p).(*LoopTargetRange)
	case ctx.Call_expression() != nil:
		forLoop.Target = &LoopTargetCall{
			Call: ctx.Call_expression().Accept(p).(ICallExpression),
		}
	case len(ctx.AllVARIABLE()) > 1 && ctx.VARIABLE(1) != nil:
		forLoop.Target = &LoopTargetVariable{
			Variable: &ExpressionVariable{
				Name: getVariable(ctx.VARIABLE(1)),
			},
		}
	case ctx.ANY_SQL() != nil:
		sqlLoc := &parseTypes.Node{}
		sqlLoc.SetToken(ctx.ANY_SQL().GetSymbol())
		forLoop.Target = &LoopTargetSQL{
			Statement:         p.parseSQLToken(ctx.ANY_SQL()),
			StatementLocation: sqlLoc,
		}
	}

	stmts := make([]Statement, len(ctx.AllStatement()))
	for i, stmt := range ctx.AllStatement() {
		stmts[i] = p.Visit(stmt).(Statement)
	}
	forLoop.Body = stmts

	forLoop.Set(ctx)

	return forLoop
}

func (p *proceduralLangVisitor) VisitStmt_if(ctx *gen.Stmt_ifContext) any {
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

	ifClause.Set(ctx)

	return ifClause
}

func (p *proceduralLangVisitor) VisitIf_then_block(ctx *gen.If_then_blockContext) any {

	stmts := make([]Statement, len(ctx.AllStatement()))
	for i, stmt := range ctx.AllStatement() {
		stmts[i] = p.Visit(stmt).(Statement)
	}

	e := &IfThen{
		If:   p.Visit(ctx.Expression()).(Expression),
		Then: stmts,
	}

	e.Set(ctx)

	return e
}

func (p *proceduralLangVisitor) VisitStmt_return(ctx *gen.Stmt_returnContext) any {
	if ctx.Expression_list() != nil {
		s := &StatementReturn{
			Values: ctx.Expression_list().Accept(p).([]Expression),
		}

		s.Set(ctx)

		return s
	}

	if ctx.ANY_SQL() != nil {
		sqlLoc := &parseTypes.Node{}
		sqlLoc.SetToken(ctx.ANY_SQL().GetSymbol())
		s := &StatementReturn{
			SQL:         p.parseSQLToken(ctx.ANY_SQL()),
			SQLLocation: sqlLoc,
		}
		s.Set(ctx)
		return s
	}

	s := &StatementReturn{}
	s.Set(ctx)
	return s
}

func (p *proceduralLangVisitor) VisitStmt_return_next(ctx *gen.Stmt_return_nextContext) any {
	s := &StatementReturnNext{
		Returns: ctx.Expression_list().Accept(p).([]Expression),
	}

	s.Set(ctx)
	return s
}

func (p *proceduralLangVisitor) VisitStmt_break(ctx *gen.Stmt_breakContext) interface{} {
	s := &StatementBreak{}
	s.Set(ctx)
	return s
}

func (p *proceduralLangVisitor) VisitStmt_sql(ctx *gen.Stmt_sqlContext) any {
	sqlLoc := &parseTypes.Node{}
	sqlLoc.SetToken(ctx.ANY_SQL().GetSymbol())
	s := &StatementSQL{
		Statement:         p.parseSQLToken(ctx.ANY_SQL()),
		StatementLocation: sqlLoc,
	}
	s.Set(ctx)
	return s
}

// ParseSQLToken parses a SQL statement token.
// Since SQL statements are defined as entire tokens in the procedural language,
// they get lexed as a single token. This function will parse the token into an AST.
// It will attempt to parse exactly one sql statement. If more than one statement is found,
// it will return the first statement and log an error. If no statements are found, it will
// return an empty select statement.
func (p *proceduralLangVisitor) parseSQLToken(tok antlr.TerminalNode) tree.AstNode {
	stmt := tok.GetText()
	errLis := p.errs.ChildFromToken("sql-parse", tok.GetSymbol())
	ast, err := sqlparser.ParseWithErrorListener(stmt, errLis)
	if err != nil {
		panic(fmt.Errorf("error parsing SQL statement: %s: %s ", stmt, err.Error()))
	}
	if errLis.Err() != nil {
		p.errs.Add(errLis.Errors()...)
	}

	if len(ast) != 1 {
		p.errs.TokenErr(tok.GetSymbol(), parseTypes.ParseErrorTypeSyntax, errors.New("expected single SQL statement, found "+strconv.Itoa(len(ast))))

		if len(ast) > 0 {
			return ast[0]
		}

		return &tree.SelectStmt{} // just to avoid nil pointer dereference
	}

	return ast[0]
}

func (p *proceduralLangVisitor) VisitStmt_variable_assignment(ctx *gen.Stmt_variable_assignmentContext) any {
	s := &StatementVariableAssignment{
		Name:  getVariable(ctx.VARIABLE()),
		Value: p.Visit(ctx.Expression()).(Expression),
	}

	s.Set(ctx)
	return s
}

func (p *proceduralLangVisitor) VisitStmt_variable_assignment_with_declaration(ctx *gen.Stmt_variable_assignment_with_declarationContext) any {
	s := &StatementVariableAssignmentWithDeclaration{
		Name:  getVariable(ctx.VARIABLE()),
		Type:  getType(ctx.Type_()),
		Value: p.Visit(ctx.Expression()).(Expression),
	}

	s.Set(ctx)
	return s
}

func (p *proceduralLangVisitor) VisitStmt_variable_declaration(ctx *gen.Stmt_variable_declarationContext) any {
	s := &StatementVariableDeclaration{
		Name: getVariable(ctx.VARIABLE()),
		Type: getType(ctx.Type_()),
	}

	s.Set(ctx)
	return s
}

func getVariable(v antlr.TerminalNode) string {
	t := v.GetText()
	// trim off beginning $
	if t[0] != '$' && t[0] != '@' {
		// this should never happen
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

func (p *proceduralLangVisitor) VisitType_cast(ctx *gen.Type_castContext) any {
	return ctx.Type_().Accept(p)
}

func (p *proceduralLangVisitor) VisitType(ctx *gen.TypeContext) any {
	dt := &types.DataType{
		Name: ctx.IDENTIFIER().GetText(),
	}

	if ctx.Literal_list() != nil {
		literals := ctx.Literal_list().Accept(p).([]Expression)

		// check here that the data type can support metadata
		if strings.ToLower(dt.Name) == types.DecimalStr {
			if len(literals) != 2 {
				p.errs.RuleErr(ctx, parseTypes.ParseErrorTypeSyntax, errors.New("invalid metadata for type"))
			}

			precision, ok := literals[0].(*ExpressionIntLiteral)
			if !ok {
				p.errs.RuleErr(ctx, parseTypes.ParseErrorTypeSyntax, errors.New("invalid metadata for type"))
			}

			scale, ok := literals[1].(*ExpressionIntLiteral)
			if !ok {
				p.errs.RuleErr(ctx, parseTypes.ParseErrorTypeSyntax, errors.New("invalid metadata for type"))
			}

			// create a new fixed data type, which will validate the metadata
			var err error
			dt, err = types.NewDecimalType(uint16(precision.Value), uint16(scale.Value))
			if err != nil {
				p.errs.RuleErr(ctx, parseTypes.ParseErrorTypeSemantic, err)
			}
		} else {
			p.errs.RuleErr(ctx, parseTypes.ParseErrorTypeSyntax, errors.New("invalid metadata for type"))
		}
	}

	if ctx.LBRACKET() != nil {
		dt.IsArray = true
	}

	return dt
}
