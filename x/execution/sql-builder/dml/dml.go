package dml

import (
	"fmt"
	"kwil/x/execution"

	"github.com/doug-martin/goqu/v9"
	_ "github.com/doug-martin/goqu/v9/dialect/postgres"
	"github.com/doug-martin/goqu/v9/exp"
)

func makeTableName(schemaName, table string) string {
	if schemaName != "" {
		return schemaName + "." + table
	}
	return table
}

func operatorToGoquExpression(op execution.ComparisonOperatorType, column string) (exp.Expression, error) {
	i := int8(1) // Goqu doesn't always like empty interfaces{} when preparing statements but does fine with bools
	switch op {
	case execution.EQUAL:
		return goqu.C(column).Eq(i), nil
	case execution.NOT_EQUAL:
		return goqu.C(column).Neq(i), nil
	case execution.GREATER_THAN:
		return goqu.C(column).Gt(i), nil
	case execution.GREATER_THAN_OR_EQUAL:
		return goqu.C(column).Gte(i), nil
	case execution.LESS_THAN:
		return goqu.C(column).Lt(i), nil
	case execution.LESS_THAN_OR_EQUAL:
		return goqu.C(column).Lte(i), nil
	}

	return nil, fmt.Errorf("unknown operator: %s", op.String())
}
