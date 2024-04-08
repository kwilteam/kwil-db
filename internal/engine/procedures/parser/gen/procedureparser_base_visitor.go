// Code generated from ProcedureParser.g4 by ANTLR 4.13.1. DO NOT EDIT.

package gen // ProcedureParser
import "github.com/antlr4-go/antlr/v4"

type BaseProcedureParserVisitor struct {
	*antlr.BaseParseTreeVisitor
}

func (v *BaseProcedureParserVisitor) VisitProgram(ctx *ProgramContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseProcedureParserVisitor) VisitStmt_variable_declaration(ctx *Stmt_variable_declarationContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseProcedureParserVisitor) VisitStmt_variable_assignment(ctx *Stmt_variable_assignmentContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseProcedureParserVisitor) VisitStmt_variable_assignment_with_declaration(ctx *Stmt_variable_assignment_with_declarationContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseProcedureParserVisitor) VisitStmt_procedure_call(ctx *Stmt_procedure_callContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseProcedureParserVisitor) VisitStmt_for_loop(ctx *Stmt_for_loopContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseProcedureParserVisitor) VisitStmt_if(ctx *Stmt_ifContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseProcedureParserVisitor) VisitStmt_sql(ctx *Stmt_sqlContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseProcedureParserVisitor) VisitStmt_break(ctx *Stmt_breakContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseProcedureParserVisitor) VisitStmt_return(ctx *Stmt_returnContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseProcedureParserVisitor) VisitStmt_return_next(ctx *Stmt_return_nextContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseProcedureParserVisitor) VisitType(ctx *TypeContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseProcedureParserVisitor) VisitExpr_array_access(ctx *Expr_array_accessContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseProcedureParserVisitor) VisitExpr_arithmetic(ctx *Expr_arithmeticContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseProcedureParserVisitor) VisitExpr_variable(ctx *Expr_variableContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseProcedureParserVisitor) VisitExpr_null_literal(ctx *Expr_null_literalContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseProcedureParserVisitor) VisitExpr_blob_literal(ctx *Expr_blob_literalContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseProcedureParserVisitor) VisitExpr_comparison(ctx *Expr_comparisonContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseProcedureParserVisitor) VisitExpr_boolean_literal(ctx *Expr_boolean_literalContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseProcedureParserVisitor) VisitExpr_call(ctx *Expr_callContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseProcedureParserVisitor) VisitExpr_make_array(ctx *Expr_make_arrayContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseProcedureParserVisitor) VisitExpr_field_access(ctx *Expr_field_accessContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseProcedureParserVisitor) VisitExpr_int_literal(ctx *Expr_int_literalContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseProcedureParserVisitor) VisitExpr_text_literal(ctx *Expr_text_literalContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseProcedureParserVisitor) VisitExpr_parenthesized(ctx *Expr_parenthesizedContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseProcedureParserVisitor) VisitExpression_list(ctx *Expression_listContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseProcedureParserVisitor) VisitExpression_make_array(ctx *Expression_make_arrayContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseProcedureParserVisitor) VisitCall_expression(ctx *Call_expressionContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseProcedureParserVisitor) VisitRange(ctx *RangeContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseProcedureParserVisitor) VisitIf_then_block(ctx *If_then_blockContext) interface{} {
	return v.VisitChildren(ctx)
}
