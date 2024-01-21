package actparser

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/antlr/antlr4/runtime/Go/antlr/v4"
	"github.com/kwilteam/action-grammar-go/actgrammar"
	"github.com/kwilteam/kwil-db/parse/internal/util"
	"github.com/kwilteam/kwil-db/parse/sql/tree"
)

// astBuilder is the visitor to build the ast from the parse tree
type astBuilder struct {
	actgrammar.BaseActionParserVisitor
	// @yaiba NOTE: may need schema to distinguish extension and action

	trace    bool
	trackPos bool
}

var _ actgrammar.ActionParserVisitor = &astBuilder{}

func newAstBuilder(trace bool, trackPos bool) *astBuilder {
	k := &astBuilder{
		trace:    trace,
		trackPos: trackPos,
	}

	return k
}

func (v *astBuilder) getPos(ctx antlr.ParserRuleContext) *tree.Position {
	if !v.trackPos {
		return nil
	}

	return &tree.Position{
		StartLine:   ctx.GetStart().GetLine(),
		StartColumn: ctx.GetStart().GetColumn(),
		EndLine:     ctx.GetStop().GetLine(),
		EndColumn:   ctx.GetStop().GetColumn(),
	}
}

// Visit dispatch to the visit method of the ctx
// e.g. if the tree is a ParseContext, then dispatch call VisitParse.
// Overwrite is needed,
// refer to https://github.com/antlr/antlr4/pull/1841#issuecomment-576791512
func (v *astBuilder) Visit(tree antlr.ParseTree) interface{} {
	if v.trace {
		fmt.Printf("visit tree: %v, %s\n", reflect.TypeOf(tree), tree.GetText())
	}
	return tree.Accept(v)
}

// VisitChildren visits the children of the specified node.
// Overwrite is needed,
// refer to https://github.com/antlr/antlr4/pull/1841#issuecomment-576791512
// calling function need to convert the result to asts
func (v *astBuilder) VisitChildren(node antlr.RuleNode) interface{} {
	var result []ActionStmt
	n := node.GetChildCount()
	for i := 0; i < n; i++ {
		child := node.GetChild(i)
		if !v.shouldVisitNextChild(child, result) {
			if v.trace {
				fmt.Printf("should not visit next child: %v,\n", reflect.TypeOf(child))
			}
			break
		}
		c := child.(antlr.ParseTree)
		childResult := v.Visit(c).(ActionStmt)
		result = append(result, childResult)
	}
	return result
}

func (v *astBuilder) shouldVisitNextChild(node antlr.Tree, currentResult interface{}) bool {
	if _, ok := node.(antlr.TerminalNode); ok {
		return false
	}

	return true
}

// VisitStatement is called when start parsing, return []types.ActionStmt
func (v *astBuilder) VisitStatement(ctx *actgrammar.StatementContext) interface{} {
	stmtCount := len(ctx.AllStmt())
	stmts := make([]ActionStmt, stmtCount)

	for i, stmtCtx := range ctx.AllStmt() {
		if stmtCtx.Call_stmt() != nil {
			stmts[i] = v.Visit(stmtCtx.Call_stmt()).(ActionStmt)
		} else {
			stmts[i] = v.Visit(stmtCtx.Sql_stmt()).(ActionStmt)
		}
	}

	return stmts
}

// VisitSql_stmt is called when parse sql statement, return *types.DMLStmt
func (v *astBuilder) VisitSql_stmt(ctx *actgrammar.Sql_stmtContext) interface{} {
	stmt := ctx.GetText()
	return &DMLStmt{Statement: stmt}
}

// VisitCall_stmt is called when parse call statement, return *types.CallStmt
func (v *astBuilder) VisitCall_stmt(ctx *actgrammar.Call_stmtContext) interface{} {
	// `a.b` is only for extension calls for now
	fnName := ctx.Call_body().Fn_name().GetText()
	if ctx.Call_body().Fn_name().Extension_call_name() != nil {
		// NOTE: in the future, if we support call external action, then the
		// extension_call and action_call could be the same syntax, need a better
		// way to distinguish them. A naive way is to check the function name with
		// the list of extensions we have in current Kuneiform.

		// extension call
		stmt := &ExtensionCallStmt{
			Extension: strings.Split(fnName, ".")[0],
			Method:    strings.Split(fnName, ".")[1],
		}

		if ctx.Call_receivers() != nil {
			stmt.Receivers = v.Visit(ctx.Call_receivers()).([]string)
		}

		if len(ctx.Call_body().Fn_arg_list().AllFn_arg_expr()) > 0 {
			stmt.Args = v.Visit(ctx.Call_body().Fn_arg_list()).([]tree.Expression)
		}

		return stmt

	} else {
		// action call
		stmt := &ActionCallStmt{
			Method: fnName,
		}

		if len(ctx.Call_body().Fn_arg_list().AllFn_arg_expr()) > 0 {
			stmt.Args = v.Visit(ctx.Call_body().Fn_arg_list()).([]tree.Expression)
		}

		return stmt
	}
}

// VisitCall_receivers is called when parse call receivers, return []string
func (v *astBuilder) VisitCall_receivers(ctx *actgrammar.Call_receiversContext) interface{} {
	receivers := make([]string, len(ctx.AllVariable()))
	for i, varCtx := range ctx.AllVariable() {
		receivers[i] = varCtx.GetText()
	}
	return receivers
}

// VisitFn_arg_list is called when parse function argument list, return []tree.Expression
func (v *astBuilder) VisitFn_arg_list(ctx *actgrammar.Fn_arg_listContext) interface{} {
	args := make([]tree.Expression, len(ctx.AllFn_arg_expr()))
	for i, argCtx := range ctx.AllFn_arg_expr() {
		args[i] = v.Visit(argCtx).(tree.Expression)
	}
	return args
}

// VisitFn_arg_expr is called when parse function argument expression return tree.Expression
// NOTE: this is a subset of util.KFSqliteVisitor.VisitExpr
func (v *astBuilder) VisitFn_arg_expr(ctx *actgrammar.Fn_arg_exprContext) interface{} {
	return v.visitFn_arg_expr(ctx)
}

func (v *astBuilder) visitFn_arg_expr(ctx actgrammar.IFn_arg_exprContext) tree.Expression {
	if ctx == nil {
		return nil
	}

	// order is important, map to expr definition in Antlr sql-grammar(not exactly)
	switch {
	// primary expressions
	case ctx.Literal_value() != nil:
		return &tree.ExpressionLiteral{Value: ctx.Literal_value().GetText()}
	// sql bind parameter
	case ctx.Variable() != nil:
		return &tree.ExpressionBindParameter{Parameter: ctx.Variable().GetText()}
	case ctx.Block_var() != nil:
		return &tree.ExpressionBindParameter{Parameter: ctx.Block_var().GetText()}
	case ctx.GetElevate_expr() != nil:
		expr := v.visitFn_arg_expr(ctx.GetElevate_expr())
		switch t := expr.(type) {
		case *tree.ExpressionLiteral:
			t.Wrapped = true
		case *tree.ExpressionBindParameter:
			t.Wrapped = true
		case *tree.ExpressionUnary:
			t.Wrapped = true
		case *tree.ExpressionBinaryComparison:
			t.Wrapped = true
		case *tree.ExpressionFunction:
			t.Wrapped = true
		case *tree.ExpressionArithmetic:
			t.Wrapped = true
		default:
			panic(fmt.Sprintf("unknown expression type %T", expr))
		}
		return expr
	// unary operators
	case ctx.MINUS() != nil && ctx.GetUnary_expr() != nil:
		return &tree.ExpressionUnary{
			Operator: tree.UnaryOperatorMinus,
			Operand:  v.visitFn_arg_expr(ctx.GetUnary_expr()),
		}
	case ctx.PLUS() != nil && ctx.GetUnary_expr() != nil:
		return &tree.ExpressionUnary{
			Operator: tree.UnaryOperatorPlus,
			Operand:  v.visitFn_arg_expr(ctx.GetUnary_expr()),
		}
	case ctx.TILDE() != nil && ctx.GetUnary_expr() != nil:
		return &tree.ExpressionUnary{
			Operator: tree.UnaryOperatorBitNot,
			Operand:  v.visitFn_arg_expr(ctx.GetUnary_expr()),
		}
	// binary opertors
	// artithmetic operators
	case ctx.PIPE2() != nil:
		return &tree.ExpressionArithmetic{
			Left:     v.visitFn_arg_expr(ctx.Fn_arg_expr(0)),
			Right:    v.visitFn_arg_expr(ctx.Fn_arg_expr(1)),
			Operator: tree.ArithmeticConcat,
		}
	// TODO: this was where ctx.STAR() != nil was
	case ctx.STAR() != nil:
		return &tree.ExpressionArithmetic{
			Left:     v.visitFn_arg_expr(ctx.Fn_arg_expr(0)),
			Right:    v.visitFn_arg_expr(ctx.Fn_arg_expr(1)),
			Operator: tree.ArithmeticOperatorMultiply,
		}
	case ctx.DIV() != nil:
		return &tree.ExpressionArithmetic{
			Left:     v.visitFn_arg_expr(ctx.Fn_arg_expr(0)),
			Right:    v.visitFn_arg_expr(ctx.Fn_arg_expr(1)),
			Operator: tree.ArithmeticOperatorDivide,
		}
	case ctx.MOD() != nil:
		return &tree.ExpressionArithmetic{
			Left:     v.visitFn_arg_expr(ctx.Fn_arg_expr(0)),
			Right:    v.visitFn_arg_expr(ctx.Fn_arg_expr(1)),
			Operator: tree.ArithmeticOperatorModulus,
		}
	case ctx.PLUS() != nil:
		return &tree.ExpressionArithmetic{
			Left:     v.visitFn_arg_expr(ctx.Fn_arg_expr(0)),
			Right:    v.visitFn_arg_expr(ctx.Fn_arg_expr(1)),
			Operator: tree.ArithmeticOperatorAdd,
		}
	case ctx.MINUS() != nil:
		return &tree.ExpressionArithmetic{
			Left:     v.visitFn_arg_expr(ctx.Fn_arg_expr(0)),
			Right:    v.visitFn_arg_expr(ctx.Fn_arg_expr(1)),
			Operator: tree.ArithmeticOperatorSubtract,
		}
	case ctx.LT2() != nil:
		return &tree.ExpressionArithmetic{
			Left:     v.visitFn_arg_expr(ctx.Fn_arg_expr(0)),
			Right:    v.visitFn_arg_expr(ctx.Fn_arg_expr(1)),
			Operator: tree.ArithmeticOperatorBitwiseLeftShift,
		}
	case ctx.GT2() != nil:
		return &tree.ExpressionArithmetic{
			Left:     v.visitFn_arg_expr(ctx.Fn_arg_expr(0)),
			Right:    v.visitFn_arg_expr(ctx.Fn_arg_expr(1)),
			Operator: tree.ArithmeticOperatorBitwiseRightShift,
		}
	case ctx.AMP() != nil:
		return &tree.ExpressionArithmetic{
			Left:     v.visitFn_arg_expr(ctx.Fn_arg_expr(0)),
			Right:    v.visitFn_arg_expr(ctx.Fn_arg_expr(1)),
			Operator: tree.ArithmeticOperatorBitwiseAnd,
		}
	case ctx.PIPE() != nil:
		return &tree.ExpressionArithmetic{
			Left:     v.visitFn_arg_expr(ctx.Fn_arg_expr(0)),
			Right:    v.visitFn_arg_expr(ctx.Fn_arg_expr(1)),
			Operator: tree.ArithmeticOperatorBitwiseOr,
		}
	// compare operators
	case ctx.LT() != nil:
		return &tree.ExpressionBinaryComparison{
			Left:     v.visitFn_arg_expr(ctx.Fn_arg_expr(0)),
			Right:    v.visitFn_arg_expr(ctx.Fn_arg_expr(1)),
			Operator: tree.ComparisonOperatorLessThan,
		}
	case ctx.LT_EQ() != nil:
		return &tree.ExpressionBinaryComparison{
			Left:     v.visitFn_arg_expr(ctx.Fn_arg_expr(0)),
			Right:    v.visitFn_arg_expr(ctx.Fn_arg_expr(1)),
			Operator: tree.ComparisonOperatorLessThanOrEqual,
		}
	case ctx.GT() != nil:
		return &tree.ExpressionBinaryComparison{
			Left:     v.visitFn_arg_expr(ctx.Fn_arg_expr(0)),
			Right:    v.visitFn_arg_expr(ctx.Fn_arg_expr(1)),
			Operator: tree.ComparisonOperatorGreaterThan,
		}
	case ctx.GT_EQ() != nil:
		return &tree.ExpressionBinaryComparison{
			Left:     v.visitFn_arg_expr(ctx.Fn_arg_expr(0)),
			Right:    v.visitFn_arg_expr(ctx.Fn_arg_expr(1)),
			Operator: tree.ComparisonOperatorGreaterThanOrEqual,
		}
	case ctx.ASSIGN() != nil:
		return &tree.ExpressionBinaryComparison{
			Left:     v.visitFn_arg_expr(ctx.Fn_arg_expr(0)),
			Right:    v.visitFn_arg_expr(ctx.Fn_arg_expr(1)),
			Operator: tree.ComparisonOperatorEqual,
		}
	case ctx.EQ() != nil:
		return &tree.ExpressionBinaryComparison{
			Left:     v.visitFn_arg_expr(ctx.Fn_arg_expr(0)),
			Right:    v.visitFn_arg_expr(ctx.Fn_arg_expr(1)),
			Operator: tree.ComparisonOperatorDoubleEqual,
		}
	case ctx.SQL_NOT_EQ1() != nil:
		return &tree.ExpressionBinaryComparison{
			Left:     v.visitFn_arg_expr(ctx.Fn_arg_expr(0)),
			Right:    v.visitFn_arg_expr(ctx.Fn_arg_expr(1)),
			Operator: tree.ComparisonOperatorNotEqual,
		}
	case ctx.SQL_NOT_EQ2() != nil:
		return &tree.ExpressionBinaryComparison{
			Left:     v.visitFn_arg_expr(ctx.Fn_arg_expr(0)),
			Right:    v.visitFn_arg_expr(ctx.Fn_arg_expr(1)),
			Operator: tree.ComparisonOperatorNotEqualDiamond,
		}
	// unary op NOT
	case ctx.NOT_() != nil && ctx.GetUnary_expr() != nil:
		return &tree.ExpressionUnary{
			Operator: tree.UnaryOperatorNot,
			Operand:  v.visitFn_arg_expr(ctx.GetUnary_expr()),
		}
	case ctx.AND_() != nil:
		return &tree.ExpressionBinaryComparison{
			Left:     v.visitFn_arg_expr(ctx.Fn_arg_expr(0)),
			Operator: tree.LogicalOperatorAnd,
			Right:    v.visitFn_arg_expr(ctx.Fn_arg_expr(1)),
		}
	case ctx.OR_() != nil:
		return &tree.ExpressionBinaryComparison{
			Left:     v.visitFn_arg_expr(ctx.Fn_arg_expr(0)),
			Operator: tree.LogicalOperatorOr,
			Right:    v.visitFn_arg_expr(ctx.Fn_arg_expr(1)),
		}
	case ctx.Sfn_name() != nil:
		expr := &tree.ExpressionFunction{
			Inputs: make([]tree.Expression, len(ctx.AllFn_arg_expr())),
		}
		funcName := util.ExtractSQLName(ctx.Sfn_name().GetText())
		f, ok := tree.SQLFunctionGetterMap[strings.ToLower(funcName)]
		if !ok {
			panic(fmt.Sprintf("unsupported function '%s'", funcName))
		}
		expr.Function = f(v.getPos(ctx))

		for i, e := range ctx.AllFn_arg_expr() {
			expr.Inputs[i] = v.visitFn_arg_expr(e)
		}
		return expr
	//case ctx.STAR() != nil:
	//	return &tree.ExpressionArithmetic{
	//		Left:     v.visitFn_arg_expr(ctx.Fn_arg_expr(0)),
	//		Right:    v.visitFn_arg_expr(ctx.Fn_arg_expr(1)),
	//		Operator: tree.ArithmeticOperatorMultiply,
	//	}
	default:
		panic(fmt.Sprintf("cannot recognize expr '%s'", ctx.GetText()))
	}
}
