package executables

import (
	"fmt"
	"kwil/x/execution"
	execTypes "kwil/x/types/execution"

	"github.com/doug-martin/goqu/v9"
	"github.com/doug-martin/goqu/v9/exp"
)

type row struct {
	column string
	value  any
}

type where struct {
	column   string
	value    any
	operator execution.ComparisonOperatorType
}

func (e *executableInterface) prepareInsert(exec *execTypes.Executable, params []*row) (string, []any, error) {
	tableName := makeTableName(e.getDbId(), exec.Table)

	record := make(goqu.Record)
	for _, row := range params {
		record[row.column] = row.value
	}

	// build the statement
	return goqu.Dialect("postgres").Insert(tableName).Prepared(true).Rows(record).ToSQL()
}

func (e *executableInterface) prepareUpdate(exec *execTypes.Executable, params []*row, wheres []*where) (string, []any, error) {
	tableName := makeTableName(e.getDbId(), exec.Table)

	// build the SET records
	record := make(goqu.Record)
	for _, row := range params {
		record[row.column] = row.value
	}

	// build the where clauses
	var whereArray []goqu.Expression
	for _, where := range wheres {
		exp, err := operatorToGoquExpression(where.operator, where.column, where.value)
		if err != nil {
			return "", nil, fmt.Errorf("error converting comparison operator: %w", err)
		}

		whereArray = append(whereArray, exp)
	}

	// build the statement
	return goqu.Dialect("postgres").Update(tableName).Prepared(true).Set(record).Where(whereArray...).ToSQL()
}

func (e *executableInterface) prepareDelete(exec *execTypes.Executable, wheres []*where) (string, []any, error) {
	tableName := makeTableName(e.getDbId(), exec.Table)

	// build the where clauses
	var whereArray []goqu.Expression
	for _, where := range wheres {
		exp, err := operatorToGoquExpression(where.operator, where.column, where.value)
		if err != nil {
			return "", nil, fmt.Errorf("error converting comparison operator: %w", err)
		}

		whereArray = append(whereArray, exp)
	}

	// build the statement
	return goqu.Dialect("postgres").Delete(tableName).Prepared(true).Where(whereArray...).ToSQL()
}

func makeTableName(schemaName, table string) string {
	if schemaName != "" {
		return schemaName + "." + table
	}
	return table
}

func operatorToGoquExpression(op execution.ComparisonOperatorType, column string, val any) (exp.Expression, error) {
	//val := "hi"
	switch op {
	case execution.EQUAL:
		return goqu.C(column).Eq(val), nil
	case execution.NOT_EQUAL:
		return goqu.C(column).Neq(val), nil
	case execution.GREATER_THAN:
		return goqu.C(column).Gt(val), nil
	case execution.GREATER_THAN_OR_EQUAL:
		return goqu.C(column).Gte(val), nil
	case execution.LESS_THAN:
		return goqu.C(column).Lt(val), nil
	case execution.LESS_THAN_OR_EQUAL:
		return goqu.C(column).Lte(val), nil
	}

	return nil, fmt.Errorf("unknown operator: %s", op.String())
}
