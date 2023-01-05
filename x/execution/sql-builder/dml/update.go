package dml

import (
	"fmt"
	"kwil/x/execution/dto"

	"github.com/doug-martin/goqu/v9"
	"github.com/doug-martin/goqu/v9/exp"
)

func BuildUpdate(schemaName, table string, params []*dto.Parameter, whereClauses []*dto.WhereClause) (string, error) {
	tblName := makeTableName(schemaName, table)

	// converting parameters to goqu records
	rec := make(exp.Record)
	for _, param := range params {
		rec[param.Column] = false // any value will do, but it can't be a struct{}{}
	}

	// converting where clauses to goqu expressions
	var whereArray []goqu.Expression
	for _, where := range whereClauses {
		exp, err := operatorToGoquExpression(where.Operator, where.Column)
		if err != nil {
			return "", fmt.Errorf("error converting comparison operator: %w", err)
		}

		whereArray = append(whereArray, exp)
	}

	stmt, _, err := goqu.Dialect("postgres").Update(tblName).Prepared(true).Set(rec).Where(whereArray...).ToSQL()
	return stmt, err
}
