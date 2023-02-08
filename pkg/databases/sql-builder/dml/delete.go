package dml

import (
	"fmt"
	"github.com/doug-martin/goqu/v9"
	"kwil/pkg/databases"
	"kwil/pkg/types/data_types/any_type"
)

func BuildDelete(schemaName, table string, wheres []*databases.WhereClause[anytype.KwilAny]) (string, error) {
	tblName := makeTableName(schemaName, table)

	// converting the where clauses to goqu expressions
	var whereArray []goqu.Expression
	for _, where := range wheres {
		exp, err := operatorToGoquExpression(where.Operator, where.Column)
		if err != nil {
			return "", fmt.Errorf("error converting comparison operator: %w", err)
		}

		whereArray = append(whereArray, exp)
	}

	stmt, _, err := goqu.Dialect("postgres").Delete(tblName).Prepared(true).Where(whereArray...).ToSQL()
	return stmt, err
}
