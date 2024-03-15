// Code generated from ProcedureParser.g4 by ANTLR 4.13.1. DO NOT EDIT.

package gen // ProcedureParser
import "github.com/antlr4-go/antlr/v4"

// A complete Visitor for a parse tree produced by ProcedureParser.
type ProcedureParserVisitor interface {
	antlr.ParseTreeVisitor

	// Visit a parse tree produced by ProcedureParser#program.
	VisitProgram(ctx *ProgramContext) interface{}

	// Visit a parse tree produced by ProcedureParser#stmt_variable_declaration.
	VisitStmt_variable_declaration(ctx *Stmt_variable_declarationContext) interface{}

	// Visit a parse tree produced by ProcedureParser#stmt_variable_assignment.
	VisitStmt_variable_assignment(ctx *Stmt_variable_assignmentContext) interface{}

	// Visit a parse tree produced by ProcedureParser#stmt_variable_assignment_with_declaration.
	VisitStmt_variable_assignment_with_declaration(ctx *Stmt_variable_assignment_with_declarationContext) interface{}

	// Visit a parse tree produced by ProcedureParser#stmt_procedure_call.
	VisitStmt_procedure_call(ctx *Stmt_procedure_callContext) interface{}

	// Visit a parse tree produced by ProcedureParser#stmt_for_loop.
	VisitStmt_for_loop(ctx *Stmt_for_loopContext) interface{}

	// Visit a parse tree produced by ProcedureParser#stmt_if.
	VisitStmt_if(ctx *Stmt_ifContext) interface{}

	// Visit a parse tree produced by ProcedureParser#stmt_sql.
	VisitStmt_sql(ctx *Stmt_sqlContext) interface{}

	// Visit a parse tree produced by ProcedureParser#stmt_break.
	VisitStmt_break(ctx *Stmt_breakContext) interface{}

	// Visit a parse tree produced by ProcedureParser#stmt_return.
	VisitStmt_return(ctx *Stmt_returnContext) interface{}

	// Visit a parse tree produced by ProcedureParser#stmt_return_next.
	VisitStmt_return_next(ctx *Stmt_return_nextContext) interface{}

	// Visit a parse tree produced by ProcedureParser#type.
	VisitType(ctx *TypeContext) interface{}

	// Visit a parse tree produced by ProcedureParser#expr_array_access.
	VisitExpr_array_access(ctx *Expr_array_accessContext) interface{}

	// Visit a parse tree produced by ProcedureParser#expr_arithmetic.
	VisitExpr_arithmetic(ctx *Expr_arithmeticContext) interface{}

	// Visit a parse tree produced by ProcedureParser#expr_variable.
	VisitExpr_variable(ctx *Expr_variableContext) interface{}

	// Visit a parse tree produced by ProcedureParser#expr_null_literal.
	VisitExpr_null_literal(ctx *Expr_null_literalContext) interface{}

	// Visit a parse tree produced by ProcedureParser#expr_blob_literal.
	VisitExpr_blob_literal(ctx *Expr_blob_literalContext) interface{}

	// Visit a parse tree produced by ProcedureParser#expr_comparison.
	VisitExpr_comparison(ctx *Expr_comparisonContext) interface{}

	// Visit a parse tree produced by ProcedureParser#expr_boolean_literal.
	VisitExpr_boolean_literal(ctx *Expr_boolean_literalContext) interface{}

	// Visit a parse tree produced by ProcedureParser#expr_call.
	VisitExpr_call(ctx *Expr_callContext) interface{}

	// Visit a parse tree produced by ProcedureParser#expr_make_array.
	VisitExpr_make_array(ctx *Expr_make_arrayContext) interface{}

	// Visit a parse tree produced by ProcedureParser#expr_field_access.
	VisitExpr_field_access(ctx *Expr_field_accessContext) interface{}

	// Visit a parse tree produced by ProcedureParser#expr_int_literal.
	VisitExpr_int_literal(ctx *Expr_int_literalContext) interface{}

	// Visit a parse tree produced by ProcedureParser#expr_text_literal.
	VisitExpr_text_literal(ctx *Expr_text_literalContext) interface{}

	// Visit a parse tree produced by ProcedureParser#expr_parenthesized.
	VisitExpr_parenthesized(ctx *Expr_parenthesizedContext) interface{}

	// Visit a parse tree produced by ProcedureParser#expression_list.
	VisitExpression_list(ctx *Expression_listContext) interface{}

	// Visit a parse tree produced by ProcedureParser#expression_make_array.
	VisitExpression_make_array(ctx *Expression_make_arrayContext) interface{}

	// Visit a parse tree produced by ProcedureParser#call_expression.
	VisitCall_expression(ctx *Call_expressionContext) interface{}

	// Visit a parse tree produced by ProcedureParser#range.
	VisitRange(ctx *RangeContext) interface{}

	// Visit a parse tree produced by ProcedureParser#if_then_block.
	VisitIf_then_block(ctx *If_then_blockContext) interface{}
}
