// Code generated from ActionParser.g4 by ANTLR 4.13.1. DO NOT EDIT.

package actgrammar // ActionParser
import "github.com/antlr4-go/antlr/v4"

type BaseActionParserVisitor struct {
	*antlr.BaseParseTreeVisitor
}

func (v *BaseActionParserVisitor) VisitStatement(ctx *StatementContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseActionParserVisitor) VisitLiteral_value(ctx *Literal_valueContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseActionParserVisitor) VisitAction_name(ctx *Action_nameContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseActionParserVisitor) VisitStmt(ctx *StmtContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseActionParserVisitor) VisitSql_stmt(ctx *Sql_stmtContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseActionParserVisitor) VisitCall_stmt(ctx *Call_stmtContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseActionParserVisitor) VisitCall_receivers(ctx *Call_receiversContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseActionParserVisitor) VisitCall_body(ctx *Call_bodyContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseActionParserVisitor) VisitVariable(ctx *VariableContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseActionParserVisitor) VisitBlock_var(ctx *Block_varContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseActionParserVisitor) VisitExtension_call_name(ctx *Extension_call_nameContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseActionParserVisitor) VisitFn_name(ctx *Fn_nameContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseActionParserVisitor) VisitSfn_name(ctx *Sfn_nameContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseActionParserVisitor) VisitFn_arg_list(ctx *Fn_arg_listContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseActionParserVisitor) VisitFn_arg_expr(ctx *Fn_arg_exprContext) interface{} {
	return v.VisitChildren(ctx)
}
