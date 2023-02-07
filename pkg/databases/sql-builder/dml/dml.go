package dml

import (
	"fmt"
	"github.com/doug-martin/goqu/v9"
	_ "github.com/doug-martin/goqu/v9/dialect/postgres"
	"github.com/doug-martin/goqu/v9/exp"
	"kwil/pkg/databases"
)

func makeTableName(schemaName, table string) string {
	if schemaName != "" {
		return schemaName + "." + table
	}
	return table
}

func operatorToGoquExpression(op databases.ComparisonOperatorType, column string) (exp.Expression, error) {
	i := int8(1) // Goqu doesn't always like empty interfaces{} when preparing statements but does fine with bools
	switch op {
	case databases.EQUAL:
		return goqu.C(column).Eq(i), nil
	case databases.NOT_EQUAL:
		return goqu.C(column).Neq(i), nil
	case databases.GREATER_THAN:
		return goqu.C(column).Gt(i), nil
	case databases.GREATER_THAN_OR_EQUAL:
		return goqu.C(column).Gte(i), nil
	case databases.LESS_THAN:
		return goqu.C(column).Lt(i), nil
	case databases.LESS_THAN_OR_EQUAL:
		return goqu.C(column).Lte(i), nil
	}

	return nil, fmt.Errorf("unknown operator: %s", op.String())
}
