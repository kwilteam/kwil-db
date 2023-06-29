package actparser

import (
	"fmt"
	"github.com/antlr/antlr4/runtime/Go/antlr/v4"
	"github.com/kwilteam/action-grammar-go/actgrammar"
	"reflect"
	"strings"
)

type KFActionVisitor struct {
	actgrammar.BaseActionParserVisitor
	// @yaiba NOTE: may need schema to distinguish extension and action

	trace bool
}

type KFActionVisitorOption func(*KFActionVisitor)

func KFActionVisitorWithTrace(on bool) KFActionVisitorOption {
	return func(l *KFActionVisitor) {
		l.trace = on
	}
}

var _ actgrammar.ActionParserVisitor = &KFActionVisitor{}

func NewKFActionVisitor(opts ...KFActionVisitorOption) *KFActionVisitor {
	k := &KFActionVisitor{}
	for _, opt := range opts {
		opt(k)
	}
	return k
}

// Visit dispatch to the visit method of the ctx
// e.g. if the tree is a ParseContext, then dispatch call VisitParse.
// Overwrite is needed,
// refer to https://github.com/antlr/antlr4/pull/1841#issuecomment-576791512
func (v *KFActionVisitor) Visit(tree antlr.ParseTree) interface{} {
	if v.trace {
		fmt.Printf("visit tree: %v, %s\n", reflect.TypeOf(tree), tree.GetText())
	}
	return tree.Accept(v)
}

// VisitChildren visits the children of the specified node.
// Overwrite is needed,
// refer to https://github.com/antlr/antlr4/pull/1841#issuecomment-576791512
// calling function need to convert the result to asts
func (v *KFActionVisitor) VisitChildren(node antlr.RuleNode) interface{} {
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

func (v *KFActionVisitor) shouldVisitNextChild(node antlr.Tree, currentResult interface{}) bool {
	if _, ok := node.(antlr.TerminalNode); ok {
		return false
	}

	return true
}

// VisitStatement is called when start parsing, return []types.ActionStmt
func (v *KFActionVisitor) VisitStatement(ctx *actgrammar.StatementContext) interface{} {
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
func (v *KFActionVisitor) VisitSql_stmt(ctx *actgrammar.Sql_stmtContext) interface{} {
	stmt := ctx.GetText()
	return &DMLStmt{Statement: stmt}
}

// VisitCall_stmt is called when parse call statement, return *types.CallStmt
func (v *KFActionVisitor) VisitCall_stmt(ctx *actgrammar.Call_stmtContext) interface{} {
	// `a.b` is only for extension calls for now
	fnName := ctx.Call_body().Fn_name().GetText()
	if strings.Contains(fnName, ".") {
		// extension call
		stmt := &ExtensionCallStmt{
			Extension: strings.Split(fnName, ".")[0],
			Method:    strings.Split(fnName, ".")[1],
		}

		if ctx.Call_receivers() != nil {
			stmt.Receivers = v.Visit(ctx.Call_receivers()).([]string)
		}

		if len(ctx.Call_body().Fn_arg_list().AllFn_arg()) > 0 {
			stmt.Args = v.Visit(ctx.Call_body().Fn_arg_list()).([]string)
		}

		return stmt

	} else {
		// action call
		stmt := &ActionCallStmt{
			Method: fnName,
		}

		if len(ctx.Call_body().Fn_arg_list().AllFn_arg()) > 0 {
			stmt.Args = v.Visit(ctx.Call_body().Fn_arg_list()).([]string)
		}

		return stmt
	}
}

// VisitCall_receivers is called when parse call receivers, return []string
func (v *KFActionVisitor) VisitCall_receivers(ctx *actgrammar.Call_receiversContext) interface{} {
	receivers := make([]string, len(ctx.AllVariable_name()))
	for i, varCtx := range ctx.AllVariable_name() {
		receivers[i] = varCtx.GetText()
	}
	return receivers
}

// VisitFn_arg_list is called when parse function argument list, return []string
func (v *KFActionVisitor) VisitFn_arg_list(ctx *actgrammar.Fn_arg_listContext) interface{} {
	args := make([]string, len(ctx.AllFn_arg()))
	for i, argCtx := range ctx.AllFn_arg() {
		args[i] = argCtx.GetText()
	}
	return args
}
