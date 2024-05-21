// Code generated from KuneiformParser.g4 by ANTLR 4.13.1. DO NOT EDIT.

package gen // KuneiformParser
import "github.com/antlr4-go/antlr/v4"

type BaseKuneiformParserVisitor struct {
	*antlr.BaseParseTreeVisitor
}

func (v *BaseKuneiformParserVisitor) VisitEntry(ctx *EntryContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseKuneiformParserVisitor) VisitString_literal(ctx *String_literalContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseKuneiformParserVisitor) VisitInteger_literal(ctx *Integer_literalContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseKuneiformParserVisitor) VisitDecimal_literal(ctx *Decimal_literalContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseKuneiformParserVisitor) VisitBoolean_literal(ctx *Boolean_literalContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseKuneiformParserVisitor) VisitNull_literal(ctx *Null_literalContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseKuneiformParserVisitor) VisitBinary_literal(ctx *Binary_literalContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseKuneiformParserVisitor) VisitIdentifier(ctx *IdentifierContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseKuneiformParserVisitor) VisitIdentifier_list(ctx *Identifier_listContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseKuneiformParserVisitor) VisitType(ctx *TypeContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseKuneiformParserVisitor) VisitType_cast(ctx *Type_castContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseKuneiformParserVisitor) VisitVariable(ctx *VariableContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseKuneiformParserVisitor) VisitVariable_list(ctx *Variable_listContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseKuneiformParserVisitor) VisitSchema(ctx *SchemaContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseKuneiformParserVisitor) VisitAnnotation(ctx *AnnotationContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseKuneiformParserVisitor) VisitDatabase_declaration(ctx *Database_declarationContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseKuneiformParserVisitor) VisitUse_declaration(ctx *Use_declarationContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseKuneiformParserVisitor) VisitTable_declaration(ctx *Table_declarationContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseKuneiformParserVisitor) VisitColumn_def(ctx *Column_defContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseKuneiformParserVisitor) VisitIndex_def(ctx *Index_defContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseKuneiformParserVisitor) VisitForeign_key_def(ctx *Foreign_key_defContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseKuneiformParserVisitor) VisitForeign_key_action(ctx *Foreign_key_actionContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseKuneiformParserVisitor) VisitType_list(ctx *Type_listContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseKuneiformParserVisitor) VisitNamed_type_list(ctx *Named_type_listContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseKuneiformParserVisitor) VisitTyped_variable_list(ctx *Typed_variable_listContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseKuneiformParserVisitor) VisitMin_constraint(ctx *Min_constraintContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseKuneiformParserVisitor) VisitMax_constraint(ctx *Max_constraintContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseKuneiformParserVisitor) VisitMin_len_constraint(ctx *Min_len_constraintContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseKuneiformParserVisitor) VisitMax_len_constraint(ctx *Max_len_constraintContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseKuneiformParserVisitor) VisitNot_null_constraint(ctx *Not_null_constraintContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseKuneiformParserVisitor) VisitPrimary_key_constraint(ctx *Primary_key_constraintContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseKuneiformParserVisitor) VisitDefault_constraint(ctx *Default_constraintContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseKuneiformParserVisitor) VisitUnique_constraint(ctx *Unique_constraintContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseKuneiformParserVisitor) VisitAccess_modifier(ctx *Access_modifierContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseKuneiformParserVisitor) VisitAction_declaration(ctx *Action_declarationContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseKuneiformParserVisitor) VisitProcedure_declaration(ctx *Procedure_declarationContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseKuneiformParserVisitor) VisitForeign_procedure_declaration(ctx *Foreign_procedure_declarationContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseKuneiformParserVisitor) VisitProcedure_return(ctx *Procedure_returnContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseKuneiformParserVisitor) VisitSql(ctx *SqlContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseKuneiformParserVisitor) VisitSql_statement(ctx *Sql_statementContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseKuneiformParserVisitor) VisitCommon_table_expression(ctx *Common_table_expressionContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseKuneiformParserVisitor) VisitSelect_statement(ctx *Select_statementContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseKuneiformParserVisitor) VisitCompound_operator(ctx *Compound_operatorContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseKuneiformParserVisitor) VisitOrdering_term(ctx *Ordering_termContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseKuneiformParserVisitor) VisitSelect_core(ctx *Select_coreContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseKuneiformParserVisitor) VisitTable_relation(ctx *Table_relationContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseKuneiformParserVisitor) VisitSubquery_relation(ctx *Subquery_relationContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseKuneiformParserVisitor) VisitFunction_relation(ctx *Function_relationContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseKuneiformParserVisitor) VisitJoin(ctx *JoinContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseKuneiformParserVisitor) VisitExpression_result_column(ctx *Expression_result_columnContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseKuneiformParserVisitor) VisitWildcard_result_column(ctx *Wildcard_result_columnContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseKuneiformParserVisitor) VisitUpdate_statement(ctx *Update_statementContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseKuneiformParserVisitor) VisitUpdate_set_clause(ctx *Update_set_clauseContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseKuneiformParserVisitor) VisitInsert_statement(ctx *Insert_statementContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseKuneiformParserVisitor) VisitUpsert_clause(ctx *Upsert_clauseContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseKuneiformParserVisitor) VisitDelete_statement(ctx *Delete_statementContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseKuneiformParserVisitor) VisitColumn_sql_expr(ctx *Column_sql_exprContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseKuneiformParserVisitor) VisitLogical_sql_expr(ctx *Logical_sql_exprContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseKuneiformParserVisitor) VisitArray_access_sql_expr(ctx *Array_access_sql_exprContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseKuneiformParserVisitor) VisitField_access_sql_expr(ctx *Field_access_sql_exprContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseKuneiformParserVisitor) VisitComparison_sql_expr(ctx *Comparison_sql_exprContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseKuneiformParserVisitor) VisitLiteral_sql_expr(ctx *Literal_sql_exprContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseKuneiformParserVisitor) VisitBetween_sql_expr(ctx *Between_sql_exprContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseKuneiformParserVisitor) VisitFunction_call_sql_expr(ctx *Function_call_sql_exprContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseKuneiformParserVisitor) VisitParen_sql_expr(ctx *Paren_sql_exprContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseKuneiformParserVisitor) VisitCollate_sql_expr(ctx *Collate_sql_exprContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseKuneiformParserVisitor) VisitVariable_sql_expr(ctx *Variable_sql_exprContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseKuneiformParserVisitor) VisitIs_sql_expr(ctx *Is_sql_exprContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseKuneiformParserVisitor) VisitLike_sql_expr(ctx *Like_sql_exprContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseKuneiformParserVisitor) VisitArithmetic_sql_expr(ctx *Arithmetic_sql_exprContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseKuneiformParserVisitor) VisitSubquery_sql_expr(ctx *Subquery_sql_exprContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseKuneiformParserVisitor) VisitUnary_sql_expr(ctx *Unary_sql_exprContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseKuneiformParserVisitor) VisitCase_expr(ctx *Case_exprContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseKuneiformParserVisitor) VisitIn_sql_expr(ctx *In_sql_exprContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseKuneiformParserVisitor) VisitSql_expr_list(ctx *Sql_expr_listContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseKuneiformParserVisitor) VisitNormal_call_sql(ctx *Normal_call_sqlContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseKuneiformParserVisitor) VisitForeign_call_sql(ctx *Foreign_call_sqlContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseKuneiformParserVisitor) VisitAction_block(ctx *Action_blockContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseKuneiformParserVisitor) VisitSql_action(ctx *Sql_actionContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseKuneiformParserVisitor) VisitLocal_action(ctx *Local_actionContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseKuneiformParserVisitor) VisitExtension_action(ctx *Extension_actionContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseKuneiformParserVisitor) VisitProcedure_block(ctx *Procedure_blockContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseKuneiformParserVisitor) VisitField_access_procedure_expr(ctx *Field_access_procedure_exprContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseKuneiformParserVisitor) VisitLiteral_procedure_expr(ctx *Literal_procedure_exprContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseKuneiformParserVisitor) VisitParen_procedure_expr(ctx *Paren_procedure_exprContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseKuneiformParserVisitor) VisitVariable_procedure_expr(ctx *Variable_procedure_exprContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseKuneiformParserVisitor) VisitMake_array_procedure_expr(ctx *Make_array_procedure_exprContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseKuneiformParserVisitor) VisitProcedure_expr_arithmetic(ctx *Procedure_expr_arithmeticContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseKuneiformParserVisitor) VisitUnary_procedure_expr(ctx *Unary_procedure_exprContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseKuneiformParserVisitor) VisitComparison_procedure_expr(ctx *Comparison_procedure_exprContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseKuneiformParserVisitor) VisitFunction_call_procedure_expr(ctx *Function_call_procedure_exprContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseKuneiformParserVisitor) VisitArray_access_procedure_expr(ctx *Array_access_procedure_exprContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseKuneiformParserVisitor) VisitProcedure_expr_list(ctx *Procedure_expr_listContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseKuneiformParserVisitor) VisitStmt_variable_declaration(ctx *Stmt_variable_declarationContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseKuneiformParserVisitor) VisitStmt_procedure_call(ctx *Stmt_procedure_callContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseKuneiformParserVisitor) VisitStmt_variable_assignment(ctx *Stmt_variable_assignmentContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseKuneiformParserVisitor) VisitStmt_for_loop(ctx *Stmt_for_loopContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseKuneiformParserVisitor) VisitStmt_if(ctx *Stmt_ifContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseKuneiformParserVisitor) VisitStmt_sql(ctx *Stmt_sqlContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseKuneiformParserVisitor) VisitStmt_break(ctx *Stmt_breakContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseKuneiformParserVisitor) VisitStmt_return(ctx *Stmt_returnContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseKuneiformParserVisitor) VisitStmt_return_next(ctx *Stmt_return_nextContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseKuneiformParserVisitor) VisitVariable_or_underscore(ctx *Variable_or_underscoreContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseKuneiformParserVisitor) VisitNormal_call_procedure(ctx *Normal_call_procedureContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseKuneiformParserVisitor) VisitForeign_call_procedure(ctx *Foreign_call_procedureContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseKuneiformParserVisitor) VisitIf_then_block(ctx *If_then_blockContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseKuneiformParserVisitor) VisitRange(ctx *RangeContext) interface{} {
	return v.VisitChildren(ctx)
}
