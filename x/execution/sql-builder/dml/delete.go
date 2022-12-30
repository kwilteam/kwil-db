package dml

import (
	"fmt"
	"kwil/x/execution/dto"

	"github.com/doug-martin/goqu/v9"
)

func BuildDelete(schemaName, table string, wheres []*dto.WhereClause) (string, error) {
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

	stmt, _, err := goqu.Dialect("postgres").Delete(tblName).Where(whereArray...).ToSQL()
	return stmt, err
}
