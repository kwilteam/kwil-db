// Code generated from KuneiformParser.g4 by ANTLR 4.13.1. DO NOT EDIT.

package gen // KuneiformParser
import "github.com/antlr4-go/antlr/v4"

// A complete Visitor for a parse tree produced by KuneiformParser.
type KuneiformParserVisitor interface {
	antlr.ParseTreeVisitor

	// Visit a parse tree produced by KuneiformParser#src.
	VisitSrc(ctx *SrcContext) interface{}

	// Visit a parse tree produced by KuneiformParser#schema_entry.
	VisitSchema_entry(ctx *Schema_entryContext) interface{}

	// Visit a parse tree produced by KuneiformParser#sql_entry.
	VisitSql_entry(ctx *Sql_entryContext) interface{}

	// Visit a parse tree produced by KuneiformParser#action_entry.
	VisitAction_entry(ctx *Action_entryContext) interface{}

	// Visit a parse tree produced by KuneiformParser#procedure_entry.
	VisitProcedure_entry(ctx *Procedure_entryContext) interface{}

	// Visit a parse tree produced by KuneiformParser#string_literal.
	VisitString_literal(ctx *String_literalContext) interface{}

	// Visit a parse tree produced by KuneiformParser#integer_literal.
	VisitInteger_literal(ctx *Integer_literalContext) interface{}

	// Visit a parse tree produced by KuneiformParser#decimal_literal.
	VisitDecimal_literal(ctx *Decimal_literalContext) interface{}

	// Visit a parse tree produced by KuneiformParser#boolean_literal.
	VisitBoolean_literal(ctx *Boolean_literalContext) interface{}

	// Visit a parse tree produced by KuneiformParser#null_literal.
	VisitNull_literal(ctx *Null_literalContext) interface{}

	// Visit a parse tree produced by KuneiformParser#binary_literal.
	VisitBinary_literal(ctx *Binary_literalContext) interface{}

	// Visit a parse tree produced by KuneiformParser#identifier.
	VisitIdentifier(ctx *IdentifierContext) interface{}

	// Visit a parse tree produced by KuneiformParser#identifier_list.
	VisitIdentifier_list(ctx *Identifier_listContext) interface{}

	// Visit a parse tree produced by KuneiformParser#type.
	VisitType(ctx *TypeContext) interface{}

	// Visit a parse tree produced by KuneiformParser#type_cast.
	VisitType_cast(ctx *Type_castContext) interface{}

	// Visit a parse tree produced by KuneiformParser#variable.
	VisitVariable(ctx *VariableContext) interface{}

	// Visit a parse tree produced by KuneiformParser#variable_list.
	VisitVariable_list(ctx *Variable_listContext) interface{}

	// Visit a parse tree produced by KuneiformParser#schema.
	VisitSchema(ctx *SchemaContext) interface{}

	// Visit a parse tree produced by KuneiformParser#annotation.
	VisitAnnotation(ctx *AnnotationContext) interface{}

	// Visit a parse tree produced by KuneiformParser#database_declaration.
	VisitDatabase_declaration(ctx *Database_declarationContext) interface{}

	// Visit a parse tree produced by KuneiformParser#use_declaration.
	VisitUse_declaration(ctx *Use_declarationContext) interface{}

	// Visit a parse tree produced by KuneiformParser#table_declaration.
	VisitTable_declaration(ctx *Table_declarationContext) interface{}

	// Visit a parse tree produced by KuneiformParser#column_def.
	VisitColumn_def(ctx *Column_defContext) interface{}

	// Visit a parse tree produced by KuneiformParser#c_column_def.
	VisitC_column_def(ctx *C_column_defContext) interface{}

	// Visit a parse tree produced by KuneiformParser#index_def.
	VisitIndex_def(ctx *Index_defContext) interface{}

	// Visit a parse tree produced by KuneiformParser#c_index_def.
	VisitC_index_def(ctx *C_index_defContext) interface{}

	// Visit a parse tree produced by KuneiformParser#foreign_key_def.
	VisitForeign_key_def(ctx *Foreign_key_defContext) interface{}

	// Visit a parse tree produced by KuneiformParser#foreign_key_action.
	VisitForeign_key_action(ctx *Foreign_key_actionContext) interface{}

	// Visit a parse tree produced by KuneiformParser#type_list.
	VisitType_list(ctx *Type_listContext) interface{}

	// Visit a parse tree produced by KuneiformParser#named_type_list.
	VisitNamed_type_list(ctx *Named_type_listContext) interface{}

	// Visit a parse tree produced by KuneiformParser#typed_variable_list.
	VisitTyped_variable_list(ctx *Typed_variable_listContext) interface{}

	// Visit a parse tree produced by KuneiformParser#constraint.
	VisitConstraint(ctx *ConstraintContext) interface{}

	// Visit a parse tree produced by KuneiformParser#inline_constraint.
	VisitInline_constraint(ctx *Inline_constraintContext) interface{}

	// Visit a parse tree produced by KuneiformParser#fk_action.
	VisitFk_action(ctx *Fk_actionContext) interface{}

	// Visit a parse tree produced by KuneiformParser#fk_constraint.
	VisitFk_constraint(ctx *Fk_constraintContext) interface{}

	// Visit a parse tree produced by KuneiformParser#access_modifier.
	VisitAccess_modifier(ctx *Access_modifierContext) interface{}

	// Visit a parse tree produced by KuneiformParser#action_declaration.
	VisitAction_declaration(ctx *Action_declarationContext) interface{}

	// Visit a parse tree produced by KuneiformParser#procedure_declaration.
	VisitProcedure_declaration(ctx *Procedure_declarationContext) interface{}

	// Visit a parse tree produced by KuneiformParser#procedure_return.
	VisitProcedure_return(ctx *Procedure_returnContext) interface{}

	// Visit a parse tree produced by KuneiformParser#sql.
	VisitSql(ctx *SqlContext) interface{}

	// Visit a parse tree produced by KuneiformParser#sql_statement.
	VisitSql_statement(ctx *Sql_statementContext) interface{}

	// Visit a parse tree produced by KuneiformParser#common_table_expression.
	VisitCommon_table_expression(ctx *Common_table_expressionContext) interface{}

	// Visit a parse tree produced by KuneiformParser#create_table_statement.
	VisitCreate_table_statement(ctx *Create_table_statementContext) interface{}

	// Visit a parse tree produced by KuneiformParser#constraint_def.
	VisitConstraint_def(ctx *Constraint_defContext) interface{}

	// Visit a parse tree produced by KuneiformParser#unnamed_constraint.
	VisitUnnamed_constraint(ctx *Unnamed_constraintContext) interface{}

	// Visit a parse tree produced by KuneiformParser#alter_table_statement.
	VisitAlter_table_statement(ctx *Alter_table_statementContext) interface{}

	// Visit a parse tree produced by KuneiformParser#add_column_constraint.
	VisitAdd_column_constraint(ctx *Add_column_constraintContext) interface{}

	// Visit a parse tree produced by KuneiformParser#drop_column_constraint.
	VisitDrop_column_constraint(ctx *Drop_column_constraintContext) interface{}

	// Visit a parse tree produced by KuneiformParser#add_column.
	VisitAdd_column(ctx *Add_columnContext) interface{}

	// Visit a parse tree produced by KuneiformParser#drop_column.
	VisitDrop_column(ctx *Drop_columnContext) interface{}

	// Visit a parse tree produced by KuneiformParser#rename_column.
	VisitRename_column(ctx *Rename_columnContext) interface{}

	// Visit a parse tree produced by KuneiformParser#rename_table.
	VisitRename_table(ctx *Rename_tableContext) interface{}

	// Visit a parse tree produced by KuneiformParser#add_table_constraint.
	VisitAdd_table_constraint(ctx *Add_table_constraintContext) interface{}

	// Visit a parse tree produced by KuneiformParser#drop_table_constraint.
	VisitDrop_table_constraint(ctx *Drop_table_constraintContext) interface{}

	// Visit a parse tree produced by KuneiformParser#create_index_statement.
	VisitCreate_index_statement(ctx *Create_index_statementContext) interface{}

	// Visit a parse tree produced by KuneiformParser#drop_index_statement.
	VisitDrop_index_statement(ctx *Drop_index_statementContext) interface{}

	// Visit a parse tree produced by KuneiformParser#select_statement.
	VisitSelect_statement(ctx *Select_statementContext) interface{}

	// Visit a parse tree produced by KuneiformParser#compound_operator.
	VisitCompound_operator(ctx *Compound_operatorContext) interface{}

	// Visit a parse tree produced by KuneiformParser#ordering_term.
	VisitOrdering_term(ctx *Ordering_termContext) interface{}

	// Visit a parse tree produced by KuneiformParser#select_core.
	VisitSelect_core(ctx *Select_coreContext) interface{}

	// Visit a parse tree produced by KuneiformParser#table_relation.
	VisitTable_relation(ctx *Table_relationContext) interface{}

	// Visit a parse tree produced by KuneiformParser#subquery_relation.
	VisitSubquery_relation(ctx *Subquery_relationContext) interface{}

	// Visit a parse tree produced by KuneiformParser#join.
	VisitJoin(ctx *JoinContext) interface{}

	// Visit a parse tree produced by KuneiformParser#expression_result_column.
	VisitExpression_result_column(ctx *Expression_result_columnContext) interface{}

	// Visit a parse tree produced by KuneiformParser#wildcard_result_column.
	VisitWildcard_result_column(ctx *Wildcard_result_columnContext) interface{}

	// Visit a parse tree produced by KuneiformParser#update_statement.
	VisitUpdate_statement(ctx *Update_statementContext) interface{}

	// Visit a parse tree produced by KuneiformParser#update_set_clause.
	VisitUpdate_set_clause(ctx *Update_set_clauseContext) interface{}

	// Visit a parse tree produced by KuneiformParser#insert_statement.
	VisitInsert_statement(ctx *Insert_statementContext) interface{}

	// Visit a parse tree produced by KuneiformParser#upsert_clause.
	VisitUpsert_clause(ctx *Upsert_clauseContext) interface{}

	// Visit a parse tree produced by KuneiformParser#delete_statement.
	VisitDelete_statement(ctx *Delete_statementContext) interface{}

	// Visit a parse tree produced by KuneiformParser#column_sql_expr.
	VisitColumn_sql_expr(ctx *Column_sql_exprContext) interface{}

	// Visit a parse tree produced by KuneiformParser#logical_sql_expr.
	VisitLogical_sql_expr(ctx *Logical_sql_exprContext) interface{}

	// Visit a parse tree produced by KuneiformParser#field_access_sql_expr.
	VisitField_access_sql_expr(ctx *Field_access_sql_exprContext) interface{}

	// Visit a parse tree produced by KuneiformParser#array_access_sql_expr.
	VisitArray_access_sql_expr(ctx *Array_access_sql_exprContext) interface{}

	// Visit a parse tree produced by KuneiformParser#comparison_sql_expr.
	VisitComparison_sql_expr(ctx *Comparison_sql_exprContext) interface{}

	// Visit a parse tree produced by KuneiformParser#literal_sql_expr.
	VisitLiteral_sql_expr(ctx *Literal_sql_exprContext) interface{}

	// Visit a parse tree produced by KuneiformParser#between_sql_expr.
	VisitBetween_sql_expr(ctx *Between_sql_exprContext) interface{}

	// Visit a parse tree produced by KuneiformParser#function_call_sql_expr.
	VisitFunction_call_sql_expr(ctx *Function_call_sql_exprContext) interface{}

	// Visit a parse tree produced by KuneiformParser#paren_sql_expr.
	VisitParen_sql_expr(ctx *Paren_sql_exprContext) interface{}

	// Visit a parse tree produced by KuneiformParser#collate_sql_expr.
	VisitCollate_sql_expr(ctx *Collate_sql_exprContext) interface{}

	// Visit a parse tree produced by KuneiformParser#variable_sql_expr.
	VisitVariable_sql_expr(ctx *Variable_sql_exprContext) interface{}

	// Visit a parse tree produced by KuneiformParser#window_function_call_sql_expr.
	VisitWindow_function_call_sql_expr(ctx *Window_function_call_sql_exprContext) interface{}

	// Visit a parse tree produced by KuneiformParser#is_sql_expr.
	VisitIs_sql_expr(ctx *Is_sql_exprContext) interface{}

	// Visit a parse tree produced by KuneiformParser#arithmetic_sql_expr.
	VisitArithmetic_sql_expr(ctx *Arithmetic_sql_exprContext) interface{}

	// Visit a parse tree produced by KuneiformParser#like_sql_expr.
	VisitLike_sql_expr(ctx *Like_sql_exprContext) interface{}

	// Visit a parse tree produced by KuneiformParser#subquery_sql_expr.
	VisitSubquery_sql_expr(ctx *Subquery_sql_exprContext) interface{}

	// Visit a parse tree produced by KuneiformParser#unary_sql_expr.
	VisitUnary_sql_expr(ctx *Unary_sql_exprContext) interface{}

	// Visit a parse tree produced by KuneiformParser#case_expr.
	VisitCase_expr(ctx *Case_exprContext) interface{}

	// Visit a parse tree produced by KuneiformParser#in_sql_expr.
	VisitIn_sql_expr(ctx *In_sql_exprContext) interface{}

	// Visit a parse tree produced by KuneiformParser#window.
	VisitWindow(ctx *WindowContext) interface{}

	// Visit a parse tree produced by KuneiformParser#when_then_clause.
	VisitWhen_then_clause(ctx *When_then_clauseContext) interface{}

	// Visit a parse tree produced by KuneiformParser#sql_expr_list.
	VisitSql_expr_list(ctx *Sql_expr_listContext) interface{}

	// Visit a parse tree produced by KuneiformParser#normal_call_sql.
	VisitNormal_call_sql(ctx *Normal_call_sqlContext) interface{}

	// Visit a parse tree produced by KuneiformParser#action_block.
	VisitAction_block(ctx *Action_blockContext) interface{}

	// Visit a parse tree produced by KuneiformParser#sql_action.
	VisitSql_action(ctx *Sql_actionContext) interface{}

	// Visit a parse tree produced by KuneiformParser#local_action.
	VisitLocal_action(ctx *Local_actionContext) interface{}

	// Visit a parse tree produced by KuneiformParser#extension_action.
	VisitExtension_action(ctx *Extension_actionContext) interface{}

	// Visit a parse tree produced by KuneiformParser#procedure_block.
	VisitProcedure_block(ctx *Procedure_blockContext) interface{}

	// Visit a parse tree produced by KuneiformParser#field_access_procedure_expr.
	VisitField_access_procedure_expr(ctx *Field_access_procedure_exprContext) interface{}

	// Visit a parse tree produced by KuneiformParser#literal_procedure_expr.
	VisitLiteral_procedure_expr(ctx *Literal_procedure_exprContext) interface{}

	// Visit a parse tree produced by KuneiformParser#paren_procedure_expr.
	VisitParen_procedure_expr(ctx *Paren_procedure_exprContext) interface{}

	// Visit a parse tree produced by KuneiformParser#variable_procedure_expr.
	VisitVariable_procedure_expr(ctx *Variable_procedure_exprContext) interface{}

	// Visit a parse tree produced by KuneiformParser#make_array_procedure_expr.
	VisitMake_array_procedure_expr(ctx *Make_array_procedure_exprContext) interface{}

	// Visit a parse tree produced by KuneiformParser#is_procedure_expr.
	VisitIs_procedure_expr(ctx *Is_procedure_exprContext) interface{}

	// Visit a parse tree produced by KuneiformParser#procedure_expr_arithmetic.
	VisitProcedure_expr_arithmetic(ctx *Procedure_expr_arithmeticContext) interface{}

	// Visit a parse tree produced by KuneiformParser#unary_procedure_expr.
	VisitUnary_procedure_expr(ctx *Unary_procedure_exprContext) interface{}

	// Visit a parse tree produced by KuneiformParser#comparison_procedure_expr.
	VisitComparison_procedure_expr(ctx *Comparison_procedure_exprContext) interface{}

	// Visit a parse tree produced by KuneiformParser#function_call_procedure_expr.
	VisitFunction_call_procedure_expr(ctx *Function_call_procedure_exprContext) interface{}

	// Visit a parse tree produced by KuneiformParser#logical_procedure_expr.
	VisitLogical_procedure_expr(ctx *Logical_procedure_exprContext) interface{}

	// Visit a parse tree produced by KuneiformParser#array_access_procedure_expr.
	VisitArray_access_procedure_expr(ctx *Array_access_procedure_exprContext) interface{}

	// Visit a parse tree produced by KuneiformParser#procedure_expr_list.
	VisitProcedure_expr_list(ctx *Procedure_expr_listContext) interface{}

	// Visit a parse tree produced by KuneiformParser#stmt_variable_declaration.
	VisitStmt_variable_declaration(ctx *Stmt_variable_declarationContext) interface{}

	// Visit a parse tree produced by KuneiformParser#stmt_procedure_call.
	VisitStmt_procedure_call(ctx *Stmt_procedure_callContext) interface{}

	// Visit a parse tree produced by KuneiformParser#stmt_variable_assignment.
	VisitStmt_variable_assignment(ctx *Stmt_variable_assignmentContext) interface{}

	// Visit a parse tree produced by KuneiformParser#stmt_for_loop.
	VisitStmt_for_loop(ctx *Stmt_for_loopContext) interface{}

	// Visit a parse tree produced by KuneiformParser#stmt_if.
	VisitStmt_if(ctx *Stmt_ifContext) interface{}

	// Visit a parse tree produced by KuneiformParser#stmt_sql.
	VisitStmt_sql(ctx *Stmt_sqlContext) interface{}

	// Visit a parse tree produced by KuneiformParser#stmt_break.
	VisitStmt_break(ctx *Stmt_breakContext) interface{}

	// Visit a parse tree produced by KuneiformParser#stmt_return.
	VisitStmt_return(ctx *Stmt_returnContext) interface{}

	// Visit a parse tree produced by KuneiformParser#stmt_return_next.
	VisitStmt_return_next(ctx *Stmt_return_nextContext) interface{}

	// Visit a parse tree produced by KuneiformParser#variable_or_underscore.
	VisitVariable_or_underscore(ctx *Variable_or_underscoreContext) interface{}

	// Visit a parse tree produced by KuneiformParser#normal_call_procedure.
	VisitNormal_call_procedure(ctx *Normal_call_procedureContext) interface{}

	// Visit a parse tree produced by KuneiformParser#if_then_block.
	VisitIf_then_block(ctx *If_then_blockContext) interface{}

	// Visit a parse tree produced by KuneiformParser#range.
	VisitRange(ctx *RangeContext) interface{}
}
