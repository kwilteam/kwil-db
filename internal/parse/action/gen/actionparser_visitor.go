// Code generated from ActionParser.g4 by ANTLR 4.13.1. DO NOT EDIT.

package actgrammar // ActionParser
import "github.com/antlr4-go/antlr/v4"

// A complete Visitor for a parse tree produced by ActionParser.
type ActionParserVisitor interface {
	antlr.ParseTreeVisitor

	// Visit a parse tree produced by ActionParser#statement.
	VisitStatement(ctx *StatementContext) interface{}

	// Visit a parse tree produced by ActionParser#literal_value.
	VisitLiteral_value(ctx *Literal_valueContext) interface{}

	// Visit a parse tree produced by ActionParser#action_name.
	VisitAction_name(ctx *Action_nameContext) interface{}

	// Visit a parse tree produced by ActionParser#stmt.
	VisitStmt(ctx *StmtContext) interface{}

	// Visit a parse tree produced by ActionParser#sql_stmt.
	VisitSql_stmt(ctx *Sql_stmtContext) interface{}

	// Visit a parse tree produced by ActionParser#call_stmt.
	VisitCall_stmt(ctx *Call_stmtContext) interface{}

	// Visit a parse tree produced by ActionParser#call_receivers.
	VisitCall_receivers(ctx *Call_receiversContext) interface{}

	// Visit a parse tree produced by ActionParser#call_body.
	VisitCall_body(ctx *Call_bodyContext) interface{}

	// Visit a parse tree produced by ActionParser#variable.
	VisitVariable(ctx *VariableContext) interface{}

	// Visit a parse tree produced by ActionParser#block_var.
	VisitBlock_var(ctx *Block_varContext) interface{}

	// Visit a parse tree produced by ActionParser#extension_call_name.
	VisitExtension_call_name(ctx *Extension_call_nameContext) interface{}

	// Visit a parse tree produced by ActionParser#fn_name.
	VisitFn_name(ctx *Fn_nameContext) interface{}

	// Visit a parse tree produced by ActionParser#sfn_name.
	VisitSfn_name(ctx *Sfn_nameContext) interface{}

	// Visit a parse tree produced by ActionParser#fn_arg_list.
	VisitFn_arg_list(ctx *Fn_arg_listContext) interface{}

	// Visit a parse tree produced by ActionParser#fn_arg_expr.
	VisitFn_arg_expr(ctx *Fn_arg_exprContext) interface{}
}
