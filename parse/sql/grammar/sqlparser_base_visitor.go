// Code generated from SQLParser.g4 by ANTLR 4.13.1. DO NOT EDIT.

package grammar // SQLParser
import "github.com/antlr4-go/antlr/v4"

type BaseSQLParserVisitor struct {
	*antlr.BaseParseTreeVisitor
}

func (v *BaseSQLParserVisitor) VisitStatements(ctx *StatementsContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSQLParserVisitor) VisitSql_stmt_list(ctx *Sql_stmt_listContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSQLParserVisitor) VisitSql_stmt(ctx *Sql_stmtContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSQLParserVisitor) VisitIndexed_column(ctx *Indexed_columnContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSQLParserVisitor) VisitCte_table_name(ctx *Cte_table_nameContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSQLParserVisitor) VisitCommon_table_expression(ctx *Common_table_expressionContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSQLParserVisitor) VisitCommon_table_stmt(ctx *Common_table_stmtContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSQLParserVisitor) VisitDelete_core(ctx *Delete_coreContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSQLParserVisitor) VisitDelete_stmt(ctx *Delete_stmtContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSQLParserVisitor) VisitVariable(ctx *VariableContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSQLParserVisitor) VisitFunction_call(ctx *Function_callContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSQLParserVisitor) VisitColumn_ref(ctx *Column_refContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSQLParserVisitor) VisitWhen_clause(ctx *When_clauseContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSQLParserVisitor) VisitSubquery_expr(ctx *Subquery_exprContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSQLParserVisitor) VisitLogical_not_expr(ctx *Logical_not_exprContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSQLParserVisitor) VisitBoolean_literal_expr(ctx *Boolean_literal_exprContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSQLParserVisitor) VisitComparison_expr(ctx *Comparison_exprContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSQLParserVisitor) VisitLike_expr(ctx *Like_exprContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSQLParserVisitor) VisitNull_expr(ctx *Null_exprContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSQLParserVisitor) VisitColumn_expr(ctx *Column_exprContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSQLParserVisitor) VisitIn_subquery_expr(ctx *In_subquery_exprContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSQLParserVisitor) VisitArithmetic_expr(ctx *Arithmetic_exprContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSQLParserVisitor) VisitLogical_binary_expr(ctx *Logical_binary_exprContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSQLParserVisitor) VisitVariable_expr(ctx *Variable_exprContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSQLParserVisitor) VisitText_literal_expr(ctx *Text_literal_exprContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSQLParserVisitor) VisitUnary_expr(ctx *Unary_exprContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSQLParserVisitor) VisitCollate_expr(ctx *Collate_exprContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSQLParserVisitor) VisitParenthesized_expr(ctx *Parenthesized_exprContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSQLParserVisitor) VisitBetween_expr(ctx *Between_exprContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSQLParserVisitor) VisitExpr_list_expr(ctx *Expr_list_exprContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSQLParserVisitor) VisitNumeric_literal_expr(ctx *Numeric_literal_exprContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSQLParserVisitor) VisitNull_literal_expr(ctx *Null_literal_exprContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSQLParserVisitor) VisitIn_list_expr(ctx *In_list_exprContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSQLParserVisitor) VisitIs_expr(ctx *Is_exprContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSQLParserVisitor) VisitCase_expr(ctx *Case_exprContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSQLParserVisitor) VisitFunction_expr(ctx *Function_exprContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSQLParserVisitor) VisitBlob_literal_expr(ctx *Blob_literal_exprContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSQLParserVisitor) VisitSubquery(ctx *SubqueryContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSQLParserVisitor) VisitExpr_list(ctx *Expr_listContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSQLParserVisitor) VisitComparisonOperator(ctx *ComparisonOperatorContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSQLParserVisitor) VisitCast_type(ctx *Cast_typeContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSQLParserVisitor) VisitType_cast(ctx *Type_castContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSQLParserVisitor) VisitValue_row(ctx *Value_rowContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSQLParserVisitor) VisitValues_clause(ctx *Values_clauseContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSQLParserVisitor) VisitInsert_core(ctx *Insert_coreContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSQLParserVisitor) VisitInsert_stmt(ctx *Insert_stmtContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSQLParserVisitor) VisitReturning_clause(ctx *Returning_clauseContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSQLParserVisitor) VisitUpsert_update(ctx *Upsert_updateContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSQLParserVisitor) VisitUpsert_clause(ctx *Upsert_clauseContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSQLParserVisitor) VisitSelect_core(ctx *Select_coreContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSQLParserVisitor) VisitSelect_stmt(ctx *Select_stmtContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSQLParserVisitor) VisitJoin_relation(ctx *Join_relationContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSQLParserVisitor) VisitRelation(ctx *RelationContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSQLParserVisitor) VisitSimple_select(ctx *Simple_selectContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSQLParserVisitor) VisitTable_or_subquery(ctx *Table_or_subqueryContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSQLParserVisitor) VisitResult_column(ctx *Result_columnContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSQLParserVisitor) VisitReturning_clause_result_column(ctx *Returning_clause_result_columnContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSQLParserVisitor) VisitJoin_operator(ctx *Join_operatorContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSQLParserVisitor) VisitJoin_constraint(ctx *Join_constraintContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSQLParserVisitor) VisitCompound_operator(ctx *Compound_operatorContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSQLParserVisitor) VisitUpdate_set_subclause(ctx *Update_set_subclauseContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSQLParserVisitor) VisitUpdate_core(ctx *Update_coreContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSQLParserVisitor) VisitUpdate_stmt(ctx *Update_stmtContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSQLParserVisitor) VisitColumn_name_list(ctx *Column_name_listContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSQLParserVisitor) VisitQualified_table_name(ctx *Qualified_table_nameContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSQLParserVisitor) VisitOrder_by_stmt(ctx *Order_by_stmtContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSQLParserVisitor) VisitLimit_stmt(ctx *Limit_stmtContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSQLParserVisitor) VisitOrdering_term(ctx *Ordering_termContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSQLParserVisitor) VisitAsc_desc(ctx *Asc_descContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSQLParserVisitor) VisitFunction_keyword(ctx *Function_keywordContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSQLParserVisitor) VisitFunction_name(ctx *Function_nameContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSQLParserVisitor) VisitTable_name(ctx *Table_nameContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSQLParserVisitor) VisitTable_alias(ctx *Table_aliasContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSQLParserVisitor) VisitColumn_name(ctx *Column_nameContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSQLParserVisitor) VisitColumn_alias(ctx *Column_aliasContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSQLParserVisitor) VisitCollation_name(ctx *Collation_nameContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseSQLParserVisitor) VisitIndex_name(ctx *Index_nameContext) interface{} {
	return v.VisitChildren(ctx)
}
