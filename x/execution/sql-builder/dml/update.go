package dml

import (
	"fmt"
	anytype "kwil/x/types/data_types/any_type"
	"kwil/x/types/databases"

	"github.com/doug-martin/goqu/v9"
)

func BuildUpdate(schemaName, table string, params []*databases.Parameter[anytype.KwilAny], whereClauses []*databases.WhereClause[anytype.KwilAny]) (string, error) {
	tblName := makeTableName(schemaName, table)

	// converting parameters to goqu records
	rec := make(goqu.Record)
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
