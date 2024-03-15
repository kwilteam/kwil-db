// Code generated from SQLParser.g4 by ANTLR 4.13.1. DO NOT EDIT.

package sqlgrammar // SQLParser
import "github.com/antlr4-go/antlr/v4"

// A complete Visitor for a parse tree produced by SQLParser.
type SQLParserVisitor interface {
	antlr.ParseTreeVisitor

	// Visit a parse tree produced by SQLParser#statements.
	VisitStatements(ctx *StatementsContext) interface{}

	// Visit a parse tree produced by SQLParser#sql_stmt_list.
	VisitSql_stmt_list(ctx *Sql_stmt_listContext) interface{}

	// Visit a parse tree produced by SQLParser#sql_stmt.
	VisitSql_stmt(ctx *Sql_stmtContext) interface{}

	// Visit a parse tree produced by SQLParser#indexed_column.
	VisitIndexed_column(ctx *Indexed_columnContext) interface{}

	// Visit a parse tree produced by SQLParser#cte_table_name.
	VisitCte_table_name(ctx *Cte_table_nameContext) interface{}

	// Visit a parse tree produced by SQLParser#common_table_expression.
	VisitCommon_table_expression(ctx *Common_table_expressionContext) interface{}

	// Visit a parse tree produced by SQLParser#common_table_stmt.
	VisitCommon_table_stmt(ctx *Common_table_stmtContext) interface{}

	// Visit a parse tree produced by SQLParser#delete_core.
	VisitDelete_core(ctx *Delete_coreContext) interface{}

	// Visit a parse tree produced by SQLParser#delete_stmt.
	VisitDelete_stmt(ctx *Delete_stmtContext) interface{}

	// Visit a parse tree produced by SQLParser#variable.
	VisitVariable(ctx *VariableContext) interface{}

	// Visit a parse tree produced by SQLParser#function_call.
	VisitFunction_call(ctx *Function_callContext) interface{}

	// Visit a parse tree produced by SQLParser#column_ref.
	VisitColumn_ref(ctx *Column_refContext) interface{}

	// Visit a parse tree produced by SQLParser#when_clause.
	VisitWhen_clause(ctx *When_clauseContext) interface{}

	// Visit a parse tree produced by SQLParser#subquery_expr.
	VisitSubquery_expr(ctx *Subquery_exprContext) interface{}

	// Visit a parse tree produced by SQLParser#logical_not_expr.
	VisitLogical_not_expr(ctx *Logical_not_exprContext) interface{}

	// Visit a parse tree produced by SQLParser#comparison_expr.
	VisitComparison_expr(ctx *Comparison_exprContext) interface{}

	// Visit a parse tree produced by SQLParser#like_expr.
	VisitLike_expr(ctx *Like_exprContext) interface{}

	// Visit a parse tree produced by SQLParser#null_expr.
	VisitNull_expr(ctx *Null_exprContext) interface{}

	// Visit a parse tree produced by SQLParser#column_expr.
	VisitColumn_expr(ctx *Column_exprContext) interface{}

	// Visit a parse tree produced by SQLParser#in_subquery_expr.
	VisitIn_subquery_expr(ctx *In_subquery_exprContext) interface{}

	// Visit a parse tree produced by SQLParser#arithmetic_expr.
	VisitArithmetic_expr(ctx *Arithmetic_exprContext) interface{}

	// Visit a parse tree produced by SQLParser#logical_binary_expr.
	VisitLogical_binary_expr(ctx *Logical_binary_exprContext) interface{}

	// Visit a parse tree produced by SQLParser#variable_expr.
	VisitVariable_expr(ctx *Variable_exprContext) interface{}

	// Visit a parse tree produced by SQLParser#unary_expr.
	VisitUnary_expr(ctx *Unary_exprContext) interface{}

	// Visit a parse tree produced by SQLParser#collate_expr.
	VisitCollate_expr(ctx *Collate_exprContext) interface{}

	// Visit a parse tree produced by SQLParser#parenthesized_expr.
	VisitParenthesized_expr(ctx *Parenthesized_exprContext) interface{}

	// Visit a parse tree produced by SQLParser#between_expr.
	VisitBetween_expr(ctx *Between_exprContext) interface{}

	// Visit a parse tree produced by SQLParser#expr_list_expr.
	VisitExpr_list_expr(ctx *Expr_list_exprContext) interface{}

	// Visit a parse tree produced by SQLParser#in_list_expr.
	VisitIn_list_expr(ctx *In_list_exprContext) interface{}

	// Visit a parse tree produced by SQLParser#literal_expr.
	VisitLiteral_expr(ctx *Literal_exprContext) interface{}

	// Visit a parse tree produced by SQLParser#is_expr.
	VisitIs_expr(ctx *Is_exprContext) interface{}

	// Visit a parse tree produced by SQLParser#case_expr.
	VisitCase_expr(ctx *Case_exprContext) interface{}

	// Visit a parse tree produced by SQLParser#function_expr.
	VisitFunction_expr(ctx *Function_exprContext) interface{}

	// Visit a parse tree produced by SQLParser#subquery.
	VisitSubquery(ctx *SubqueryContext) interface{}

	// Visit a parse tree produced by SQLParser#expr_list.
	VisitExpr_list(ctx *Expr_listContext) interface{}

	// Visit a parse tree produced by SQLParser#comparisonOperator.
	VisitComparisonOperator(ctx *ComparisonOperatorContext) interface{}

	// Visit a parse tree produced by SQLParser#cast_type.
	VisitCast_type(ctx *Cast_typeContext) interface{}

	// Visit a parse tree produced by SQLParser#type_cast.
	VisitType_cast(ctx *Type_castContext) interface{}

	// Visit a parse tree produced by SQLParser#boolean_value.
	VisitBoolean_value(ctx *Boolean_valueContext) interface{}

	// Visit a parse tree produced by SQLParser#string_value.
	VisitString_value(ctx *String_valueContext) interface{}

	// Visit a parse tree produced by SQLParser#numeric_value.
	VisitNumeric_value(ctx *Numeric_valueContext) interface{}

	// Visit a parse tree produced by SQLParser#literal.
	VisitLiteral(ctx *LiteralContext) interface{}

	// Visit a parse tree produced by SQLParser#value_row.
	VisitValue_row(ctx *Value_rowContext) interface{}

	// Visit a parse tree produced by SQLParser#values_clause.
	VisitValues_clause(ctx *Values_clauseContext) interface{}

	// Visit a parse tree produced by SQLParser#insert_core.
	VisitInsert_core(ctx *Insert_coreContext) interface{}

	// Visit a parse tree produced by SQLParser#insert_stmt.
	VisitInsert_stmt(ctx *Insert_stmtContext) interface{}

	// Visit a parse tree produced by SQLParser#returning_clause.
	VisitReturning_clause(ctx *Returning_clauseContext) interface{}

	// Visit a parse tree produced by SQLParser#upsert_update.
	VisitUpsert_update(ctx *Upsert_updateContext) interface{}

	// Visit a parse tree produced by SQLParser#upsert_clause.
	VisitUpsert_clause(ctx *Upsert_clauseContext) interface{}

	// Visit a parse tree produced by SQLParser#select_stmt_no_cte.
	VisitSelect_stmt_no_cte(ctx *Select_stmt_no_cteContext) interface{}

	// Visit a parse tree produced by SQLParser#select_stmt.
	VisitSelect_stmt(ctx *Select_stmtContext) interface{}

	// Visit a parse tree produced by SQLParser#join_relation.
	VisitJoin_relation(ctx *Join_relationContext) interface{}

	// Visit a parse tree produced by SQLParser#relation.
	VisitRelation(ctx *RelationContext) interface{}

	// Visit a parse tree produced by SQLParser#select_core.
	VisitSelect_core(ctx *Select_coreContext) interface{}

	// Visit a parse tree produced by SQLParser#table_or_subquery.
	VisitTable_or_subquery(ctx *Table_or_subqueryContext) interface{}

	// Visit a parse tree produced by SQLParser#result_column.
	VisitResult_column(ctx *Result_columnContext) interface{}

	// Visit a parse tree produced by SQLParser#returning_clause_result_column.
	VisitReturning_clause_result_column(ctx *Returning_clause_result_columnContext) interface{}

	// Visit a parse tree produced by SQLParser#join_operator.
	VisitJoin_operator(ctx *Join_operatorContext) interface{}

	// Visit a parse tree produced by SQLParser#join_constraint.
	VisitJoin_constraint(ctx *Join_constraintContext) interface{}

	// Visit a parse tree produced by SQLParser#compound_operator.
	VisitCompound_operator(ctx *Compound_operatorContext) interface{}

	// Visit a parse tree produced by SQLParser#update_set_subclause.
	VisitUpdate_set_subclause(ctx *Update_set_subclauseContext) interface{}

	// Visit a parse tree produced by SQLParser#update_core.
	VisitUpdate_core(ctx *Update_coreContext) interface{}

	// Visit a parse tree produced by SQLParser#update_stmt.
	VisitUpdate_stmt(ctx *Update_stmtContext) interface{}

	// Visit a parse tree produced by SQLParser#column_name_list.
	VisitColumn_name_list(ctx *Column_name_listContext) interface{}

	// Visit a parse tree produced by SQLParser#qualified_table_name.
	VisitQualified_table_name(ctx *Qualified_table_nameContext) interface{}

	// Visit a parse tree produced by SQLParser#order_by_stmt.
	VisitOrder_by_stmt(ctx *Order_by_stmtContext) interface{}

	// Visit a parse tree produced by SQLParser#limit_stmt.
	VisitLimit_stmt(ctx *Limit_stmtContext) interface{}

	// Visit a parse tree produced by SQLParser#ordering_term.
	VisitOrdering_term(ctx *Ordering_termContext) interface{}

	// Visit a parse tree produced by SQLParser#asc_desc.
	VisitAsc_desc(ctx *Asc_descContext) interface{}

	// Visit a parse tree produced by SQLParser#function_keyword.
	VisitFunction_keyword(ctx *Function_keywordContext) interface{}

	// Visit a parse tree produced by SQLParser#function_name.
	VisitFunction_name(ctx *Function_nameContext) interface{}

	// Visit a parse tree produced by SQLParser#table_name.
	VisitTable_name(ctx *Table_nameContext) interface{}

	// Visit a parse tree produced by SQLParser#table_alias.
	VisitTable_alias(ctx *Table_aliasContext) interface{}

	// Visit a parse tree produced by SQLParser#column_name.
	VisitColumn_name(ctx *Column_nameContext) interface{}

	// Visit a parse tree produced by SQLParser#column_alias.
	VisitColumn_alias(ctx *Column_aliasContext) interface{}

	// Visit a parse tree produced by SQLParser#collation_name.
	VisitCollation_name(ctx *Collation_nameContext) interface{}

	// Visit a parse tree produced by SQLParser#index_name.
	VisitIndex_name(ctx *Index_nameContext) interface{}
}
