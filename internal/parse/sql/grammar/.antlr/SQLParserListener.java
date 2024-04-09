// Generated from /Users/brennanlamey/kwil-db/internal/parse/sql/grammar/SQLParser.g4 by ANTLR 4.13.1
import org.antlr.v4.runtime.tree.ParseTreeListener;

/**
 * This interface defines a complete listener for a parse tree produced by
 * {@link SQLParser}.
 */
public interface SQLParserListener extends ParseTreeListener {
	/**
	 * Enter a parse tree produced by {@link SQLParser#statements}.
	 * @param ctx the parse tree
	 */
	void enterStatements(SQLParser.StatementsContext ctx);
	/**
	 * Exit a parse tree produced by {@link SQLParser#statements}.
	 * @param ctx the parse tree
	 */
	void exitStatements(SQLParser.StatementsContext ctx);
	/**
	 * Enter a parse tree produced by {@link SQLParser#sql_stmt_list}.
	 * @param ctx the parse tree
	 */
	void enterSql_stmt_list(SQLParser.Sql_stmt_listContext ctx);
	/**
	 * Exit a parse tree produced by {@link SQLParser#sql_stmt_list}.
	 * @param ctx the parse tree
	 */
	void exitSql_stmt_list(SQLParser.Sql_stmt_listContext ctx);
	/**
	 * Enter a parse tree produced by {@link SQLParser#sql_stmt}.
	 * @param ctx the parse tree
	 */
	void enterSql_stmt(SQLParser.Sql_stmtContext ctx);
	/**
	 * Exit a parse tree produced by {@link SQLParser#sql_stmt}.
	 * @param ctx the parse tree
	 */
	void exitSql_stmt(SQLParser.Sql_stmtContext ctx);
	/**
	 * Enter a parse tree produced by {@link SQLParser#indexed_column}.
	 * @param ctx the parse tree
	 */
	void enterIndexed_column(SQLParser.Indexed_columnContext ctx);
	/**
	 * Exit a parse tree produced by {@link SQLParser#indexed_column}.
	 * @param ctx the parse tree
	 */
	void exitIndexed_column(SQLParser.Indexed_columnContext ctx);
	/**
	 * Enter a parse tree produced by {@link SQLParser#cte_table_name}.
	 * @param ctx the parse tree
	 */
	void enterCte_table_name(SQLParser.Cte_table_nameContext ctx);
	/**
	 * Exit a parse tree produced by {@link SQLParser#cte_table_name}.
	 * @param ctx the parse tree
	 */
	void exitCte_table_name(SQLParser.Cte_table_nameContext ctx);
	/**
	 * Enter a parse tree produced by {@link SQLParser#common_table_expression}.
	 * @param ctx the parse tree
	 */
	void enterCommon_table_expression(SQLParser.Common_table_expressionContext ctx);
	/**
	 * Exit a parse tree produced by {@link SQLParser#common_table_expression}.
	 * @param ctx the parse tree
	 */
	void exitCommon_table_expression(SQLParser.Common_table_expressionContext ctx);
	/**
	 * Enter a parse tree produced by {@link SQLParser#common_table_stmt}.
	 * @param ctx the parse tree
	 */
	void enterCommon_table_stmt(SQLParser.Common_table_stmtContext ctx);
	/**
	 * Exit a parse tree produced by {@link SQLParser#common_table_stmt}.
	 * @param ctx the parse tree
	 */
	void exitCommon_table_stmt(SQLParser.Common_table_stmtContext ctx);
	/**
	 * Enter a parse tree produced by {@link SQLParser#delete_core}.
	 * @param ctx the parse tree
	 */
	void enterDelete_core(SQLParser.Delete_coreContext ctx);
	/**
	 * Exit a parse tree produced by {@link SQLParser#delete_core}.
	 * @param ctx the parse tree
	 */
	void exitDelete_core(SQLParser.Delete_coreContext ctx);
	/**
	 * Enter a parse tree produced by {@link SQLParser#delete_stmt}.
	 * @param ctx the parse tree
	 */
	void enterDelete_stmt(SQLParser.Delete_stmtContext ctx);
	/**
	 * Exit a parse tree produced by {@link SQLParser#delete_stmt}.
	 * @param ctx the parse tree
	 */
	void exitDelete_stmt(SQLParser.Delete_stmtContext ctx);
	/**
	 * Enter a parse tree produced by {@link SQLParser#variable}.
	 * @param ctx the parse tree
	 */
	void enterVariable(SQLParser.VariableContext ctx);
	/**
	 * Exit a parse tree produced by {@link SQLParser#variable}.
	 * @param ctx the parse tree
	 */
	void exitVariable(SQLParser.VariableContext ctx);
	/**
	 * Enter a parse tree produced by {@link SQLParser#function_call}.
	 * @param ctx the parse tree
	 */
	void enterFunction_call(SQLParser.Function_callContext ctx);
	/**
	 * Exit a parse tree produced by {@link SQLParser#function_call}.
	 * @param ctx the parse tree
	 */
	void exitFunction_call(SQLParser.Function_callContext ctx);
	/**
	 * Enter a parse tree produced by {@link SQLParser#column_ref}.
	 * @param ctx the parse tree
	 */
	void enterColumn_ref(SQLParser.Column_refContext ctx);
	/**
	 * Exit a parse tree produced by {@link SQLParser#column_ref}.
	 * @param ctx the parse tree
	 */
	void exitColumn_ref(SQLParser.Column_refContext ctx);
	/**
	 * Enter a parse tree produced by {@link SQLParser#when_clause}.
	 * @param ctx the parse tree
	 */
	void enterWhen_clause(SQLParser.When_clauseContext ctx);
	/**
	 * Exit a parse tree produced by {@link SQLParser#when_clause}.
	 * @param ctx the parse tree
	 */
	void exitWhen_clause(SQLParser.When_clauseContext ctx);
	/**
	 * Enter a parse tree produced by the {@code subquery_expr}
	 * labeled alternative in {@link SQLParser#expr}.
	 * @param ctx the parse tree
	 */
	void enterSubquery_expr(SQLParser.Subquery_exprContext ctx);
	/**
	 * Exit a parse tree produced by the {@code subquery_expr}
	 * labeled alternative in {@link SQLParser#expr}.
	 * @param ctx the parse tree
	 */
	void exitSubquery_expr(SQLParser.Subquery_exprContext ctx);
	/**
	 * Enter a parse tree produced by the {@code logical_not_expr}
	 * labeled alternative in {@link SQLParser#expr}.
	 * @param ctx the parse tree
	 */
	void enterLogical_not_expr(SQLParser.Logical_not_exprContext ctx);
	/**
	 * Exit a parse tree produced by the {@code logical_not_expr}
	 * labeled alternative in {@link SQLParser#expr}.
	 * @param ctx the parse tree
	 */
	void exitLogical_not_expr(SQLParser.Logical_not_exprContext ctx);
	/**
	 * Enter a parse tree produced by the {@code boolean_literal_expr}
	 * labeled alternative in {@link SQLParser#expr}.
	 * @param ctx the parse tree
	 */
	void enterBoolean_literal_expr(SQLParser.Boolean_literal_exprContext ctx);
	/**
	 * Exit a parse tree produced by the {@code boolean_literal_expr}
	 * labeled alternative in {@link SQLParser#expr}.
	 * @param ctx the parse tree
	 */
	void exitBoolean_literal_expr(SQLParser.Boolean_literal_exprContext ctx);
	/**
	 * Enter a parse tree produced by the {@code comparison_expr}
	 * labeled alternative in {@link SQLParser#expr}.
	 * @param ctx the parse tree
	 */
	void enterComparison_expr(SQLParser.Comparison_exprContext ctx);
	/**
	 * Exit a parse tree produced by the {@code comparison_expr}
	 * labeled alternative in {@link SQLParser#expr}.
	 * @param ctx the parse tree
	 */
	void exitComparison_expr(SQLParser.Comparison_exprContext ctx);
	/**
	 * Enter a parse tree produced by the {@code like_expr}
	 * labeled alternative in {@link SQLParser#expr}.
	 * @param ctx the parse tree
	 */
	void enterLike_expr(SQLParser.Like_exprContext ctx);
	/**
	 * Exit a parse tree produced by the {@code like_expr}
	 * labeled alternative in {@link SQLParser#expr}.
	 * @param ctx the parse tree
	 */
	void exitLike_expr(SQLParser.Like_exprContext ctx);
	/**
	 * Enter a parse tree produced by the {@code null_expr}
	 * labeled alternative in {@link SQLParser#expr}.
	 * @param ctx the parse tree
	 */
	void enterNull_expr(SQLParser.Null_exprContext ctx);
	/**
	 * Exit a parse tree produced by the {@code null_expr}
	 * labeled alternative in {@link SQLParser#expr}.
	 * @param ctx the parse tree
	 */
	void exitNull_expr(SQLParser.Null_exprContext ctx);
	/**
	 * Enter a parse tree produced by the {@code column_expr}
	 * labeled alternative in {@link SQLParser#expr}.
	 * @param ctx the parse tree
	 */
	void enterColumn_expr(SQLParser.Column_exprContext ctx);
	/**
	 * Exit a parse tree produced by the {@code column_expr}
	 * labeled alternative in {@link SQLParser#expr}.
	 * @param ctx the parse tree
	 */
	void exitColumn_expr(SQLParser.Column_exprContext ctx);
	/**
	 * Enter a parse tree produced by the {@code in_subquery_expr}
	 * labeled alternative in {@link SQLParser#expr}.
	 * @param ctx the parse tree
	 */
	void enterIn_subquery_expr(SQLParser.In_subquery_exprContext ctx);
	/**
	 * Exit a parse tree produced by the {@code in_subquery_expr}
	 * labeled alternative in {@link SQLParser#expr}.
	 * @param ctx the parse tree
	 */
	void exitIn_subquery_expr(SQLParser.In_subquery_exprContext ctx);
	/**
	 * Enter a parse tree produced by the {@code arithmetic_expr}
	 * labeled alternative in {@link SQLParser#expr}.
	 * @param ctx the parse tree
	 */
	void enterArithmetic_expr(SQLParser.Arithmetic_exprContext ctx);
	/**
	 * Exit a parse tree produced by the {@code arithmetic_expr}
	 * labeled alternative in {@link SQLParser#expr}.
	 * @param ctx the parse tree
	 */
	void exitArithmetic_expr(SQLParser.Arithmetic_exprContext ctx);
	/**
	 * Enter a parse tree produced by the {@code logical_binary_expr}
	 * labeled alternative in {@link SQLParser#expr}.
	 * @param ctx the parse tree
	 */
	void enterLogical_binary_expr(SQLParser.Logical_binary_exprContext ctx);
	/**
	 * Exit a parse tree produced by the {@code logical_binary_expr}
	 * labeled alternative in {@link SQLParser#expr}.
	 * @param ctx the parse tree
	 */
	void exitLogical_binary_expr(SQLParser.Logical_binary_exprContext ctx);
	/**
	 * Enter a parse tree produced by the {@code variable_expr}
	 * labeled alternative in {@link SQLParser#expr}.
	 * @param ctx the parse tree
	 */
	void enterVariable_expr(SQLParser.Variable_exprContext ctx);
	/**
	 * Exit a parse tree produced by the {@code variable_expr}
	 * labeled alternative in {@link SQLParser#expr}.
	 * @param ctx the parse tree
	 */
	void exitVariable_expr(SQLParser.Variable_exprContext ctx);
	/**
	 * Enter a parse tree produced by the {@code text_literal_expr}
	 * labeled alternative in {@link SQLParser#expr}.
	 * @param ctx the parse tree
	 */
	void enterText_literal_expr(SQLParser.Text_literal_exprContext ctx);
	/**
	 * Exit a parse tree produced by the {@code text_literal_expr}
	 * labeled alternative in {@link SQLParser#expr}.
	 * @param ctx the parse tree
	 */
	void exitText_literal_expr(SQLParser.Text_literal_exprContext ctx);
	/**
	 * Enter a parse tree produced by the {@code unary_expr}
	 * labeled alternative in {@link SQLParser#expr}.
	 * @param ctx the parse tree
	 */
	void enterUnary_expr(SQLParser.Unary_exprContext ctx);
	/**
	 * Exit a parse tree produced by the {@code unary_expr}
	 * labeled alternative in {@link SQLParser#expr}.
	 * @param ctx the parse tree
	 */
	void exitUnary_expr(SQLParser.Unary_exprContext ctx);
	/**
	 * Enter a parse tree produced by the {@code collate_expr}
	 * labeled alternative in {@link SQLParser#expr}.
	 * @param ctx the parse tree
	 */
	void enterCollate_expr(SQLParser.Collate_exprContext ctx);
	/**
	 * Exit a parse tree produced by the {@code collate_expr}
	 * labeled alternative in {@link SQLParser#expr}.
	 * @param ctx the parse tree
	 */
	void exitCollate_expr(SQLParser.Collate_exprContext ctx);
	/**
	 * Enter a parse tree produced by the {@code parenthesized_expr}
	 * labeled alternative in {@link SQLParser#expr}.
	 * @param ctx the parse tree
	 */
	void enterParenthesized_expr(SQLParser.Parenthesized_exprContext ctx);
	/**
	 * Exit a parse tree produced by the {@code parenthesized_expr}
	 * labeled alternative in {@link SQLParser#expr}.
	 * @param ctx the parse tree
	 */
	void exitParenthesized_expr(SQLParser.Parenthesized_exprContext ctx);
	/**
	 * Enter a parse tree produced by the {@code between_expr}
	 * labeled alternative in {@link SQLParser#expr}.
	 * @param ctx the parse tree
	 */
	void enterBetween_expr(SQLParser.Between_exprContext ctx);
	/**
	 * Exit a parse tree produced by the {@code between_expr}
	 * labeled alternative in {@link SQLParser#expr}.
	 * @param ctx the parse tree
	 */
	void exitBetween_expr(SQLParser.Between_exprContext ctx);
	/**
	 * Enter a parse tree produced by the {@code expr_list_expr}
	 * labeled alternative in {@link SQLParser#expr}.
	 * @param ctx the parse tree
	 */
	void enterExpr_list_expr(SQLParser.Expr_list_exprContext ctx);
	/**
	 * Exit a parse tree produced by the {@code expr_list_expr}
	 * labeled alternative in {@link SQLParser#expr}.
	 * @param ctx the parse tree
	 */
	void exitExpr_list_expr(SQLParser.Expr_list_exprContext ctx);
	/**
	 * Enter a parse tree produced by the {@code numeric_literal_expr}
	 * labeled alternative in {@link SQLParser#expr}.
	 * @param ctx the parse tree
	 */
	void enterNumeric_literal_expr(SQLParser.Numeric_literal_exprContext ctx);
	/**
	 * Exit a parse tree produced by the {@code numeric_literal_expr}
	 * labeled alternative in {@link SQLParser#expr}.
	 * @param ctx the parse tree
	 */
	void exitNumeric_literal_expr(SQLParser.Numeric_literal_exprContext ctx);
	/**
	 * Enter a parse tree produced by the {@code null_literal_expr}
	 * labeled alternative in {@link SQLParser#expr}.
	 * @param ctx the parse tree
	 */
	void enterNull_literal_expr(SQLParser.Null_literal_exprContext ctx);
	/**
	 * Exit a parse tree produced by the {@code null_literal_expr}
	 * labeled alternative in {@link SQLParser#expr}.
	 * @param ctx the parse tree
	 */
	void exitNull_literal_expr(SQLParser.Null_literal_exprContext ctx);
	/**
	 * Enter a parse tree produced by the {@code in_list_expr}
	 * labeled alternative in {@link SQLParser#expr}.
	 * @param ctx the parse tree
	 */
	void enterIn_list_expr(SQLParser.In_list_exprContext ctx);
	/**
	 * Exit a parse tree produced by the {@code in_list_expr}
	 * labeled alternative in {@link SQLParser#expr}.
	 * @param ctx the parse tree
	 */
	void exitIn_list_expr(SQLParser.In_list_exprContext ctx);
	/**
	 * Enter a parse tree produced by the {@code is_expr}
	 * labeled alternative in {@link SQLParser#expr}.
	 * @param ctx the parse tree
	 */
	void enterIs_expr(SQLParser.Is_exprContext ctx);
	/**
	 * Exit a parse tree produced by the {@code is_expr}
	 * labeled alternative in {@link SQLParser#expr}.
	 * @param ctx the parse tree
	 */
	void exitIs_expr(SQLParser.Is_exprContext ctx);
	/**
	 * Enter a parse tree produced by the {@code case_expr}
	 * labeled alternative in {@link SQLParser#expr}.
	 * @param ctx the parse tree
	 */
	void enterCase_expr(SQLParser.Case_exprContext ctx);
	/**
	 * Exit a parse tree produced by the {@code case_expr}
	 * labeled alternative in {@link SQLParser#expr}.
	 * @param ctx the parse tree
	 */
	void exitCase_expr(SQLParser.Case_exprContext ctx);
	/**
	 * Enter a parse tree produced by the {@code function_expr}
	 * labeled alternative in {@link SQLParser#expr}.
	 * @param ctx the parse tree
	 */
	void enterFunction_expr(SQLParser.Function_exprContext ctx);
	/**
	 * Exit a parse tree produced by the {@code function_expr}
	 * labeled alternative in {@link SQLParser#expr}.
	 * @param ctx the parse tree
	 */
	void exitFunction_expr(SQLParser.Function_exprContext ctx);
	/**
	 * Enter a parse tree produced by the {@code blob_literal_expr}
	 * labeled alternative in {@link SQLParser#expr}.
	 * @param ctx the parse tree
	 */
	void enterBlob_literal_expr(SQLParser.Blob_literal_exprContext ctx);
	/**
	 * Exit a parse tree produced by the {@code blob_literal_expr}
	 * labeled alternative in {@link SQLParser#expr}.
	 * @param ctx the parse tree
	 */
	void exitBlob_literal_expr(SQLParser.Blob_literal_exprContext ctx);
	/**
	 * Enter a parse tree produced by {@link SQLParser#subquery}.
	 * @param ctx the parse tree
	 */
	void enterSubquery(SQLParser.SubqueryContext ctx);
	/**
	 * Exit a parse tree produced by {@link SQLParser#subquery}.
	 * @param ctx the parse tree
	 */
	void exitSubquery(SQLParser.SubqueryContext ctx);
	/**
	 * Enter a parse tree produced by {@link SQLParser#expr_list}.
	 * @param ctx the parse tree
	 */
	void enterExpr_list(SQLParser.Expr_listContext ctx);
	/**
	 * Exit a parse tree produced by {@link SQLParser#expr_list}.
	 * @param ctx the parse tree
	 */
	void exitExpr_list(SQLParser.Expr_listContext ctx);
	/**
	 * Enter a parse tree produced by {@link SQLParser#comparisonOperator}.
	 * @param ctx the parse tree
	 */
	void enterComparisonOperator(SQLParser.ComparisonOperatorContext ctx);
	/**
	 * Exit a parse tree produced by {@link SQLParser#comparisonOperator}.
	 * @param ctx the parse tree
	 */
	void exitComparisonOperator(SQLParser.ComparisonOperatorContext ctx);
	/**
	 * Enter a parse tree produced by {@link SQLParser#cast_type}.
	 * @param ctx the parse tree
	 */
	void enterCast_type(SQLParser.Cast_typeContext ctx);
	/**
	 * Exit a parse tree produced by {@link SQLParser#cast_type}.
	 * @param ctx the parse tree
	 */
	void exitCast_type(SQLParser.Cast_typeContext ctx);
	/**
	 * Enter a parse tree produced by {@link SQLParser#type_cast}.
	 * @param ctx the parse tree
	 */
	void enterType_cast(SQLParser.Type_castContext ctx);
	/**
	 * Exit a parse tree produced by {@link SQLParser#type_cast}.
	 * @param ctx the parse tree
	 */
	void exitType_cast(SQLParser.Type_castContext ctx);
	/**
	 * Enter a parse tree produced by {@link SQLParser#value_row}.
	 * @param ctx the parse tree
	 */
	void enterValue_row(SQLParser.Value_rowContext ctx);
	/**
	 * Exit a parse tree produced by {@link SQLParser#value_row}.
	 * @param ctx the parse tree
	 */
	void exitValue_row(SQLParser.Value_rowContext ctx);
	/**
	 * Enter a parse tree produced by {@link SQLParser#values_clause}.
	 * @param ctx the parse tree
	 */
	void enterValues_clause(SQLParser.Values_clauseContext ctx);
	/**
	 * Exit a parse tree produced by {@link SQLParser#values_clause}.
	 * @param ctx the parse tree
	 */
	void exitValues_clause(SQLParser.Values_clauseContext ctx);
	/**
	 * Enter a parse tree produced by {@link SQLParser#insert_core}.
	 * @param ctx the parse tree
	 */
	void enterInsert_core(SQLParser.Insert_coreContext ctx);
	/**
	 * Exit a parse tree produced by {@link SQLParser#insert_core}.
	 * @param ctx the parse tree
	 */
	void exitInsert_core(SQLParser.Insert_coreContext ctx);
	/**
	 * Enter a parse tree produced by {@link SQLParser#insert_stmt}.
	 * @param ctx the parse tree
	 */
	void enterInsert_stmt(SQLParser.Insert_stmtContext ctx);
	/**
	 * Exit a parse tree produced by {@link SQLParser#insert_stmt}.
	 * @param ctx the parse tree
	 */
	void exitInsert_stmt(SQLParser.Insert_stmtContext ctx);
	/**
	 * Enter a parse tree produced by {@link SQLParser#returning_clause}.
	 * @param ctx the parse tree
	 */
	void enterReturning_clause(SQLParser.Returning_clauseContext ctx);
	/**
	 * Exit a parse tree produced by {@link SQLParser#returning_clause}.
	 * @param ctx the parse tree
	 */
	void exitReturning_clause(SQLParser.Returning_clauseContext ctx);
	/**
	 * Enter a parse tree produced by {@link SQLParser#upsert_update}.
	 * @param ctx the parse tree
	 */
	void enterUpsert_update(SQLParser.Upsert_updateContext ctx);
	/**
	 * Exit a parse tree produced by {@link SQLParser#upsert_update}.
	 * @param ctx the parse tree
	 */
	void exitUpsert_update(SQLParser.Upsert_updateContext ctx);
	/**
	 * Enter a parse tree produced by {@link SQLParser#upsert_clause}.
	 * @param ctx the parse tree
	 */
	void enterUpsert_clause(SQLParser.Upsert_clauseContext ctx);
	/**
	 * Exit a parse tree produced by {@link SQLParser#upsert_clause}.
	 * @param ctx the parse tree
	 */
	void exitUpsert_clause(SQLParser.Upsert_clauseContext ctx);
	/**
	 * Enter a parse tree produced by {@link SQLParser#select_core}.
	 * @param ctx the parse tree
	 */
	void enterSelect_core(SQLParser.Select_coreContext ctx);
	/**
	 * Exit a parse tree produced by {@link SQLParser#select_core}.
	 * @param ctx the parse tree
	 */
	void exitSelect_core(SQLParser.Select_coreContext ctx);
	/**
	 * Enter a parse tree produced by {@link SQLParser#select_stmt}.
	 * @param ctx the parse tree
	 */
	void enterSelect_stmt(SQLParser.Select_stmtContext ctx);
	/**
	 * Exit a parse tree produced by {@link SQLParser#select_stmt}.
	 * @param ctx the parse tree
	 */
	void exitSelect_stmt(SQLParser.Select_stmtContext ctx);
	/**
	 * Enter a parse tree produced by {@link SQLParser#join_relation}.
	 * @param ctx the parse tree
	 */
	void enterJoin_relation(SQLParser.Join_relationContext ctx);
	/**
	 * Exit a parse tree produced by {@link SQLParser#join_relation}.
	 * @param ctx the parse tree
	 */
	void exitJoin_relation(SQLParser.Join_relationContext ctx);
	/**
	 * Enter a parse tree produced by {@link SQLParser#relation}.
	 * @param ctx the parse tree
	 */
	void enterRelation(SQLParser.RelationContext ctx);
	/**
	 * Exit a parse tree produced by {@link SQLParser#relation}.
	 * @param ctx the parse tree
	 */
	void exitRelation(SQLParser.RelationContext ctx);
	/**
	 * Enter a parse tree produced by {@link SQLParser#simple_select}.
	 * @param ctx the parse tree
	 */
	void enterSimple_select(SQLParser.Simple_selectContext ctx);
	/**
	 * Exit a parse tree produced by {@link SQLParser#simple_select}.
	 * @param ctx the parse tree
	 */
	void exitSimple_select(SQLParser.Simple_selectContext ctx);
	/**
	 * Enter a parse tree produced by {@link SQLParser#table_or_subquery}.
	 * @param ctx the parse tree
	 */
	void enterTable_or_subquery(SQLParser.Table_or_subqueryContext ctx);
	/**
	 * Exit a parse tree produced by {@link SQLParser#table_or_subquery}.
	 * @param ctx the parse tree
	 */
	void exitTable_or_subquery(SQLParser.Table_or_subqueryContext ctx);
	/**
	 * Enter a parse tree produced by {@link SQLParser#result_column}.
	 * @param ctx the parse tree
	 */
	void enterResult_column(SQLParser.Result_columnContext ctx);
	/**
	 * Exit a parse tree produced by {@link SQLParser#result_column}.
	 * @param ctx the parse tree
	 */
	void exitResult_column(SQLParser.Result_columnContext ctx);
	/**
	 * Enter a parse tree produced by {@link SQLParser#returning_clause_result_column}.
	 * @param ctx the parse tree
	 */
	void enterReturning_clause_result_column(SQLParser.Returning_clause_result_columnContext ctx);
	/**
	 * Exit a parse tree produced by {@link SQLParser#returning_clause_result_column}.
	 * @param ctx the parse tree
	 */
	void exitReturning_clause_result_column(SQLParser.Returning_clause_result_columnContext ctx);
	/**
	 * Enter a parse tree produced by {@link SQLParser#join_operator}.
	 * @param ctx the parse tree
	 */
	void enterJoin_operator(SQLParser.Join_operatorContext ctx);
	/**
	 * Exit a parse tree produced by {@link SQLParser#join_operator}.
	 * @param ctx the parse tree
	 */
	void exitJoin_operator(SQLParser.Join_operatorContext ctx);
	/**
	 * Enter a parse tree produced by {@link SQLParser#join_constraint}.
	 * @param ctx the parse tree
	 */
	void enterJoin_constraint(SQLParser.Join_constraintContext ctx);
	/**
	 * Exit a parse tree produced by {@link SQLParser#join_constraint}.
	 * @param ctx the parse tree
	 */
	void exitJoin_constraint(SQLParser.Join_constraintContext ctx);
	/**
	 * Enter a parse tree produced by {@link SQLParser#compound_operator}.
	 * @param ctx the parse tree
	 */
	void enterCompound_operator(SQLParser.Compound_operatorContext ctx);
	/**
	 * Exit a parse tree produced by {@link SQLParser#compound_operator}.
	 * @param ctx the parse tree
	 */
	void exitCompound_operator(SQLParser.Compound_operatorContext ctx);
	/**
	 * Enter a parse tree produced by {@link SQLParser#update_set_subclause}.
	 * @param ctx the parse tree
	 */
	void enterUpdate_set_subclause(SQLParser.Update_set_subclauseContext ctx);
	/**
	 * Exit a parse tree produced by {@link SQLParser#update_set_subclause}.
	 * @param ctx the parse tree
	 */
	void exitUpdate_set_subclause(SQLParser.Update_set_subclauseContext ctx);
	/**
	 * Enter a parse tree produced by {@link SQLParser#update_core}.
	 * @param ctx the parse tree
	 */
	void enterUpdate_core(SQLParser.Update_coreContext ctx);
	/**
	 * Exit a parse tree produced by {@link SQLParser#update_core}.
	 * @param ctx the parse tree
	 */
	void exitUpdate_core(SQLParser.Update_coreContext ctx);
	/**
	 * Enter a parse tree produced by {@link SQLParser#update_stmt}.
	 * @param ctx the parse tree
	 */
	void enterUpdate_stmt(SQLParser.Update_stmtContext ctx);
	/**
	 * Exit a parse tree produced by {@link SQLParser#update_stmt}.
	 * @param ctx the parse tree
	 */
	void exitUpdate_stmt(SQLParser.Update_stmtContext ctx);
	/**
	 * Enter a parse tree produced by {@link SQLParser#column_name_list}.
	 * @param ctx the parse tree
	 */
	void enterColumn_name_list(SQLParser.Column_name_listContext ctx);
	/**
	 * Exit a parse tree produced by {@link SQLParser#column_name_list}.
	 * @param ctx the parse tree
	 */
	void exitColumn_name_list(SQLParser.Column_name_listContext ctx);
	/**
	 * Enter a parse tree produced by {@link SQLParser#qualified_table_name}.
	 * @param ctx the parse tree
	 */
	void enterQualified_table_name(SQLParser.Qualified_table_nameContext ctx);
	/**
	 * Exit a parse tree produced by {@link SQLParser#qualified_table_name}.
	 * @param ctx the parse tree
	 */
	void exitQualified_table_name(SQLParser.Qualified_table_nameContext ctx);
	/**
	 * Enter a parse tree produced by {@link SQLParser#order_by_stmt}.
	 * @param ctx the parse tree
	 */
	void enterOrder_by_stmt(SQLParser.Order_by_stmtContext ctx);
	/**
	 * Exit a parse tree produced by {@link SQLParser#order_by_stmt}.
	 * @param ctx the parse tree
	 */
	void exitOrder_by_stmt(SQLParser.Order_by_stmtContext ctx);
	/**
	 * Enter a parse tree produced by {@link SQLParser#limit_stmt}.
	 * @param ctx the parse tree
	 */
	void enterLimit_stmt(SQLParser.Limit_stmtContext ctx);
	/**
	 * Exit a parse tree produced by {@link SQLParser#limit_stmt}.
	 * @param ctx the parse tree
	 */
	void exitLimit_stmt(SQLParser.Limit_stmtContext ctx);
	/**
	 * Enter a parse tree produced by {@link SQLParser#ordering_term}.
	 * @param ctx the parse tree
	 */
	void enterOrdering_term(SQLParser.Ordering_termContext ctx);
	/**
	 * Exit a parse tree produced by {@link SQLParser#ordering_term}.
	 * @param ctx the parse tree
	 */
	void exitOrdering_term(SQLParser.Ordering_termContext ctx);
	/**
	 * Enter a parse tree produced by {@link SQLParser#asc_desc}.
	 * @param ctx the parse tree
	 */
	void enterAsc_desc(SQLParser.Asc_descContext ctx);
	/**
	 * Exit a parse tree produced by {@link SQLParser#asc_desc}.
	 * @param ctx the parse tree
	 */
	void exitAsc_desc(SQLParser.Asc_descContext ctx);
	/**
	 * Enter a parse tree produced by {@link SQLParser#function_keyword}.
	 * @param ctx the parse tree
	 */
	void enterFunction_keyword(SQLParser.Function_keywordContext ctx);
	/**
	 * Exit a parse tree produced by {@link SQLParser#function_keyword}.
	 * @param ctx the parse tree
	 */
	void exitFunction_keyword(SQLParser.Function_keywordContext ctx);
	/**
	 * Enter a parse tree produced by {@link SQLParser#function_name}.
	 * @param ctx the parse tree
	 */
	void enterFunction_name(SQLParser.Function_nameContext ctx);
	/**
	 * Exit a parse tree produced by {@link SQLParser#function_name}.
	 * @param ctx the parse tree
	 */
	void exitFunction_name(SQLParser.Function_nameContext ctx);
	/**
	 * Enter a parse tree produced by {@link SQLParser#table_name}.
	 * @param ctx the parse tree
	 */
	void enterTable_name(SQLParser.Table_nameContext ctx);
	/**
	 * Exit a parse tree produced by {@link SQLParser#table_name}.
	 * @param ctx the parse tree
	 */
	void exitTable_name(SQLParser.Table_nameContext ctx);
	/**
	 * Enter a parse tree produced by {@link SQLParser#table_alias}.
	 * @param ctx the parse tree
	 */
	void enterTable_alias(SQLParser.Table_aliasContext ctx);
	/**
	 * Exit a parse tree produced by {@link SQLParser#table_alias}.
	 * @param ctx the parse tree
	 */
	void exitTable_alias(SQLParser.Table_aliasContext ctx);
	/**
	 * Enter a parse tree produced by {@link SQLParser#column_name}.
	 * @param ctx the parse tree
	 */
	void enterColumn_name(SQLParser.Column_nameContext ctx);
	/**
	 * Exit a parse tree produced by {@link SQLParser#column_name}.
	 * @param ctx the parse tree
	 */
	void exitColumn_name(SQLParser.Column_nameContext ctx);
	/**
	 * Enter a parse tree produced by {@link SQLParser#column_alias}.
	 * @param ctx the parse tree
	 */
	void enterColumn_alias(SQLParser.Column_aliasContext ctx);
	/**
	 * Exit a parse tree produced by {@link SQLParser#column_alias}.
	 * @param ctx the parse tree
	 */
	void exitColumn_alias(SQLParser.Column_aliasContext ctx);
	/**
	 * Enter a parse tree produced by {@link SQLParser#collation_name}.
	 * @param ctx the parse tree
	 */
	void enterCollation_name(SQLParser.Collation_nameContext ctx);
	/**
	 * Exit a parse tree produced by {@link SQLParser#collation_name}.
	 * @param ctx the parse tree
	 */
	void exitCollation_name(SQLParser.Collation_nameContext ctx);
	/**
	 * Enter a parse tree produced by {@link SQLParser#index_name}.
	 * @param ctx the parse tree
	 */
	void enterIndex_name(SQLParser.Index_nameContext ctx);
	/**
	 * Exit a parse tree produced by {@link SQLParser#index_name}.
	 * @param ctx the parse tree
	 */
	void exitIndex_name(SQLParser.Index_nameContext ctx);
}